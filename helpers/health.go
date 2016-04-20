package helpers

import (
	"net"
	"strings"
	"time"

	"github.com/byuoitav/av-api/packages/cretelnet"
)

// Health represents the results of various health checks run on each box
type Health struct {
	PingIn  bool
	PingOut bool
}

// GetHealth checks a number of status items for each box and returns a JSON object representing the test restults
func GetHealth(address string) (Health, error) {
	pingIn, err := checkPingIn(address)
	if err != nil {
		return Health{}, err
	}

	pingOut, err := checkPingOut(address)
	if err != nil {
		return Health{}, err
	}

	return Health{PingIn: pingIn, PingOut: pingOut}, nil
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
	response, err := cretelnet.GetOutput(address, "DMPS-300-C>", "ping avmetrics1.byu.edu")
	if err != nil {
		return false, err
	}

	if strings.Contains(response, "alive") {
		return true, nil
	}

	return false, nil
}
