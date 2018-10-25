package base

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/common/v2/events"
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

	if len(e.Key) == 0 || len(e.Value) == 0 {
		return nil
	}

	// Add some more information to the Event, such as hostname and a timestamp.
	e.Timestamp = time.Now()
	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		e.GeneratingSystem = os.Getenv("PI_HOSTNAME")
		if len(os.Getenv("DEVELOPMENT_HOSTNAME")) > 0 {
			e.GeneratingSystem = os.Getenv("DEVELOPMENT_HOSTNAME")
		}
	} else {
		// isn't it running in a docker container in aws? this won't work?
		e.GeneratingSystem, err = os.Hostname()
	}
	if err != nil {
		return err
	}

	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		e.EventTags = append(e.EventTags, os.Getenv("LOCAL_ENVIRONMENT"))
	}

	if !Error {
		EventNode.PublishEvent(events.APISuccess, e)
	} else {
		EventNode.PublishEvent(events.APIError, e)
	}

	return err
}

// SendEvent builds and then sends the Event to the event router.
func SendEvent(e events.Event, Error bool) error {

	err := Publish(e, Error)

	return err
}

// PublishError takes an error message and cause for the error, and then builds an Event to send to the event router.
func PublishError(errorStr string, cause string, device string) {
	deviceInfo := strings.Split(device, "-")
	building := deviceInfo[0]
	room := fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1])

	roomInfo := events.BasicRoomInfo{
		BuildingID: building,
		RoomID:     room,
	}

	e := events.Event{
		TargetDevice: events.BasicDeviceInfo{
			BasicRoomInfo: roomInfo,
			DeviceID:      device,
		},
		AffectedRoom: roomInfo,
		Key:          "Error String",
		Value:        errorStr,
	}

	e.EventTags = append(e.EventTags, events.Error, cause)

	Publish(e, true)
}
