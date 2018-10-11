package base

import (
	"os"
	"strings"
	"time"

	"github.com/byuoitav/common/events"
)

// EventNode is the event node used through the AV-API package to send events.
var EventNode *events.EventNode

// PublishHealth is a wrapper function to publish an Event that is not an error.
func PublishHealth(e events.Event) {
	Publish(e, false)
}

// Publish sends a pre-made Event to the event router and tags it as a Success or an Error.
func Publish(e events.Event, Error bool) error {
	var err error

	if len(e.Event.EventInfoValue) == 0 || len(e.Event.EventInfoValue) == 0 {
		return nil
	}

	// Add some more information to the Event, such as hostname and a timestamp.
	e.Timestamp = time.Now().Format(time.RFC3339)
	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		e.Hostname = os.Getenv("PI_HOSTNAME")
		if len(os.Getenv("DEVELOPMENT_HOSTNAME")) > 0 {
			e.Hostname = os.Getenv("DEVELOPMENT_HOSTNAME")
		}
	} else {
		// isn't it running in a docker container in aws? this won't work?
		e.Hostname, err = os.Hostname()
	}
	if err != nil {
		return err
	}

	e.LocalEnvironment = len(os.Getenv("LOCAL_ENVIRONMENT")) > 0

	if !Error {
		EventNode.PublishEvent(events.APISuccess, e)
	} else {
		EventNode.PublishEvent(events.APIError, e)
	}

	return err
}

// SendEvent builds and then sends the Event to the event router.
func SendEvent(Type events.EventType,
	Cause events.EventCause,
	Device string,
	Room string,
	Building string,
	InfoKey string,
	InfoValue string,
	Requestor string,
	Error bool) error {

	e := events.EventInfo{
		Type:           Type,
		EventCause:     Cause,
		Device:         Device,
		EventInfoKey:   InfoKey,
		EventInfoValue: InfoValue,
		Requestor:      Requestor,
	}

	err := Publish(events.Event{
		Event:    e,
		Building: Building,
		Room:     Room,
	}, Error)

	return err
}

// PublishError takes an error message and cause for the error, and then builds an Event to send to the event router.
func PublishError(errorStr string, cause events.EventCause) {
	e := events.EventInfo{
		Type:           events.ERROR,
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

	Publish(events.Event{
		Event:    e,
		Building: building,
		Room:     room,
	}, true)
}
