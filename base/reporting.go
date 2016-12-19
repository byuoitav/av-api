package base

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

/*
Event is the struct we push up to ELK.
{
  hostname: "",
  timestamp: RFC 3339 Format,
  localEnvironment: bool,
  callingIP: "",
  event: " ",
  responseCode: int,
  building: "",
  room: ""
}
*/
type Event struct {
	Hostname         string `json:"hostname,omitempty"`
	Timestamp        string `json:"timestamp,omitempty"`
	LocalEnvironment bool   `json:"localEnvironment,omitempty"`
	Event            string `json:"event,omitempty"`
	ResponseCode     int    `json:"responseCode,omitempty"`
	Success          bool   `json:"success,omitempty"`
	Building         string `json:"building,omitempty"`
	Room             string `json:"room,omitempty"`
	Device           string `json:"device,omitempty"`
}

func ReportToELK(e Event) error {
	var err error

	e.Timestamp = time.Now().Format(time.RFC3339)
	e.Hostname, err = os.Hostname()
	if err != nil {
		return err
	}
	e.LocalEnvironment = len(os.Getenv("LOCAL_ENVIRONMENT")) > 0

	toSend, err := json.Marshal(&e)
	if err != nil {
		return err
	}

	_, err = http.Post(os.Getenv("ELASTIC_API_EVENTS"),
		"application/json",
		bytes.NewBuffer(toSend))

	return err
}
