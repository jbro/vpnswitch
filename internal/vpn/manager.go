package vpn

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-systemd/dbus"
)

type vpnProfile struct {
	Country string `json:"country"`
	Flag    string `json:"flag"`
	City    string `json:"city"`
	Running bool   `json:"running"`
}

type vpnProfiles struct {
	ProfilePath string                 `json:"profile_path"`
	Profiles    map[string]*vpnProfile `json:"profiles"`
}

type msg struct {
	event string
	data  string
}

type subscriber chan msg

type Manager struct {
	list        *vpnProfiles
	dbusConn    *dbus.Conn
	subscribe   chan *subscriber
	unsubscribe chan *subscriber
	subscribers map[*subscriber]bool
	shutdown    chan bool
	outstanding *sync.WaitGroup
}

func NewManager(vpnConfig string) (*Manager, error) {
	l, err := loadVPNProfiles(vpnConfig)
	if err != nil {
		return nil, err
	}
	conn, err := dbus.New()
	if err != nil {
		return nil, fmt.Errorf("Could not connect to systemd via Dbus: %s", err)
	}

	v := Manager{
		list:        l,
		dbusConn:    conn,
		subscribe:   make(chan *subscriber),
		unsubscribe: make(chan *subscriber),
		subscribers: make(map[*subscriber]bool),
		shutdown:    make(chan bool),
		outstanding: new(sync.WaitGroup),
	}

	return &v, nil
}

func (vp *vpnProfiles) updateState(changes map[string]*dbus.UnitStatus) []msg {
	var m []msg
	for k, u := range changes {
		if strings.HasPrefix(k, "openvpn-client@") {
			n := k
			n = strings.Split(n, "@")[1]
			n = strings.Split(n, ".")[0]

			oldState := vp.Profiles[n].Running
			newState := (u != nil)
			if oldState == true && newState == false {
				m = append(m, msg{event: "disconnect", data: n})
			} else if oldState == false && newState == true {
				m = append(m, msg{event: "connect", data: n})
			}

			vp.Profiles[n].Running = newState
		}
	}
	return m
}

func (vm *Manager) Start() {
	go func() {
		defer vm.dbusConn.Close()
		vm.dbusConn.Subscribe()
		defer vm.dbusConn.Unsubscribe()
		status, _ := vm.dbusConn.SubscribeUnits(1 * time.Second)

		vm.outstanding.Add(1)
		defer vm.outstanding.Done()

		for {
			select {
			case sub := <-vm.subscribe:
				vm.subscribers[sub] = true
			case unsub := <-vm.unsubscribe:
				delete(vm.subscribers, unsub)
				close(*unsub)
			case changes := <-status:
				log.Println("Recieved status changes from systemd")
				chgs := vm.list.updateState(changes)
				for _, c := range chgs {
					for recv := range vm.subscribers {
						log.Println("Dispatching")
						*recv <- c
					}
				}
			case <-vm.shutdown:
				for recv := range vm.subscribers {
					close(*recv)
				}
				return
			}
		}
	}()
}

func (vm *Manager) Shutdown() {
	close(vm.shutdown)
	vm.outstanding.Wait()
}

func (vm *Manager) VPNListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json")
	if r.Method == http.MethodGet {
		data, _ := json.MarshalIndent(vm.list, "", "  ")
		w.Write(data)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not supported")
	}
}

func (vm *Manager) VPNConnectHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	profile := q["profile"][0]

	if len(profile) > 1 {
		if _, ok := vm.list.Profiles[profile]; ok {
			wait := make(chan string)
			vm.dbusConn.StartUnit("openvpn-client@"+profile+".service", "replace", wait)
			<-wait
		}
	}
}

func (vm *Manager) VPNDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	profile := q["profile"][0]

	if len(profile) > 1 {
		if _, ok := vm.list.Profiles[profile]; ok {
			wait := make(chan string)
			vm.dbusConn.StopUnit("openvpn-client@"+profile+".service", "replace", wait)
			<-wait
		}
	}

}

func (vm *Manager) SSEHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming is not supported", http.StatusInternalServerError)
		return
	}

	sub := make(subscriber)
	vm.subscribe <- &sub
	defer func() { vm.unsubscribe <- &sub }()

	vm.outstanding.Add(1)
	defer vm.outstanding.Done()
	defer log.Println("Stream closed")
	log.Println("Stream opened")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-sub:
			if !ok {
				return
			}
			log.Printf("Event: %s", event.event)
			fmt.Fprintf(w, "event: %s\n", event.event)
			for _, line := range strings.Split(event.data, "\n") {
				log.Printf("Data: %s", line)
				fmt.Fprintf(w, "data: %s\n", line)
			}
			fmt.Fprint(w, "\n")
			flusher.Flush()
		}
	}
}
