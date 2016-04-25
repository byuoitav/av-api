package helpers

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/byuoitav/av-api/packages/fusion"
)

// GetHealth checks a number of status items for each box and returns a JSON object representing the test restults
func GetHealth(address string) (fusion.Health, error) {
	pingIn, err := checkPingIn(address)
	if err != nil {
		return fusion.Health{}, err
	}

	pingOut, err := checkPingOut(address)
	if err != nil {
		return fusion.Health{}, err
	}

	return fusion.Health{PingIn: pingIn, PingOut: pingOut}, nil
}

func checkPingIn(address string) (bool, error) {
	timeout := 2 * time.Second

	connection, err := net.DialTimeout("tcp", address+":23", timeout)
	if err == nil {
		connection.Close()
		return true, nil
	}

	return false, nil
}

func checkPingOut(address string) (bool, error) {
	url := "http://avmetrics1.byu.edu:8001/sendCommand/"

	var jsonStr = []byte(`{"Command":"ping avmetrics1.byu.edu", "IPAddress": "` + address + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if strings.Contains(string(body), "alive") {
		return true, nil
	}

	return false, nil
}
