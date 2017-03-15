package base

import (
	"bytes"
	"encoding/json"
	"log"
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
	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		e.Hostname = os.Getenv("PI_HOSTNAME")
	} else {
		e.Hostname, err = os.Hostname()
	}

	if err != nil {
		return err
	}

	e.LocalEnvironment = len(os.Getenv("LOCAL_ENVIRONMENT")) > 0

	log.Printf("Elastic event to send: %+v", e)

	toSend, err := json.Marshal(&e)
	if err != nil {
		return err
	}

	log.Print("Sending event to: " + os.Getenv("ELASTIC_API_EVENTS"))

	_, err = http.Post(os.Getenv("ELASTIC_API_EVENTS"),
		"application/json",
		bytes.NewBuffer(toSend))

	return err
}
