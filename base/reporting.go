package base

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
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

// Reports an event to ELK
// (depreciated) now reported by the event-translator-microservice
func ReportToELK(e eventinfrastructure.Event) error {
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

func Event(e eventinfrastructure.Event) error {
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

	log.Printf("Event to send: %+v", e)

	toSend, err := json.Marshal(&e)
	if err != nil {
		return err
	}

	log.Print("Sending event to event router.")

	/* send error */
	_, err = http.Post("hi",
		"application/json",
		bytes.NewBuffer(toSend))

	return err
}
