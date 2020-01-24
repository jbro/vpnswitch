package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/handlers"
	"github.com/jbro/vpnswitch/internal/vpn"
)

func main() {
	vman, err := vpn.NewManager(os.Args[1])
	if err != nil {
		log.Fatalf("Could not create VPN Manager: %s", err)
	}

	muxer := http.NewServeMux()
	muxer.Handle("/", http.FileServer(http.Dir("web/")))
	muxer.HandleFunc("/vpn/connect", vman.VPNConnectHandler)
	muxer.HandleFunc("/vpn/disconnect", vman.VPNDisconnectHandler)
	muxer.HandleFunc("/vpn/list", vman.VPNListHandler)
	muxer.HandleFunc("/vpn/stream", vman.SSEHandler)

	server := http.Server{}
	server.Addr = os.Args[2]
	server.Handler = handlers.LoggingHandler(os.Stdout, muxer)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Closing")
		vman.Shutdown()
		log.Println("Closed")
		server.Shutdown(context.Background())
	}()

	vman.Start()
	server.ListenAndServe()
}
