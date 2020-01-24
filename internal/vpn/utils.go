package vpn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const risBase = 0x1F1E6

func country2flag(cc string) (string, error) {
	if len(cc) != 2 {
		return "", fmt.Errorf("Country code must be exactly 2 charaters long: %s", cc)
	}
	cc = strings.ToLower(cc)
	flag := ""
	for _, c := range cc {
		if !(('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')) {
			return "", fmt.Errorf("Illegal character in country code: %c", c)
		}
		ris := int(c - 97 + risBase)
		flag += fmt.Sprintf("&#x%X;", ris)
	}

	return flag, nil
}

func loadVPNProfiles(path string) (*vpnProfiles, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Could not open %s: %s", path, err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Could not read %s: %s", path, err)
	}

	list := new(vpnProfiles)
	err = json.Unmarshal([]byte(data), list)
	if err != nil {
		return nil, fmt.Errorf("Could not parse %s: %s", path, err)
	}

	for _, v := range list.Profiles {
		v.Running = false
		flag, err := country2flag(v.Country)
		if err != nil {
			return nil, err
		}
		v.Flag = flag
	}

	return list, err
}
