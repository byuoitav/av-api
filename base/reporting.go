package base

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/xuther/go-message-router/common"
	"github.com/xuther/go-message-router/publisher"
)

var Publisher publisher.Publisher

func Publish(e eventinfrastructure.Event, Error bool) error {
	var err error

	// create the event
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

	toSend, err := json.Marshal(&e)
	if err != nil {
		return err
	}

	header := [24]byte{}
	if !Error {
		copy(header[:], eventinfrastructure.APISuccess)
	} else {
		copy(header[:], eventinfrastructure.APIError)
	}

	log.Printf("Publishing event: %s", toSend)
	Publisher.Write(common.Message{MessageHeader: header, MessageBody: toSend})

	return err
}

func SendEvent(Type eventinfrastructure.EventType,
	Cause eventinfrastructure.EventCause,
	Device string,
	Room string,
	Building string,
	InfoKey string,
	InfoValue string,
	Error bool) error {

	e := eventinfrastructure.EventInfo{
		Type:           Type,
		EventCause:     Cause,
		Device:         Device,
		EventInfoKey:   InfoKey,
		EventInfoValue: InfoValue,
	}

	err := Publish(eventinfrastructure.Event{
		Event:    e,
		Building: Building,
		Room:     Room,
	}, Error)

	return err
}

func PublishError(errorStr string, cause eventinfrastructure.EventCause) {
	e := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.ERROR,
		EventCause:     cause,
		EventInfoKey:   "Error String",
		EventInfoValue: errorStr,
	}

	building := ""
	room := ""

	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		if len(os.Getenv("PI_HOSTNAME")) > 0 {
			name := os.Getenv("PI_HOSTNAME")
			roomInfo := strings.Split(name, "-")
			building = roomInfo[0]
			room = roomInfo[1]
			e.Device = roomInfo[2]
		}
	}

	Publish(eventinfrastructure.Event{
		Event:    e,
		Building: building,
		Room:     room,
	}, true)
}
