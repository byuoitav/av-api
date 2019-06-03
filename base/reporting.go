package base

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/v2/events"
)

// Messenger is the variable used through the AV-API package to send events.
var Messenger *messenger.Messenger

// PublishHealth is a wrapper function to publish an Event that is not an error.
func PublishHealth(e events.Event) {
	SendEvent(e)
}

// SendEvent sends a pre-made Event to the hub.
func SendEvent(e events.Event) error {
	if Messenger == nil {
		return nil
	}

	var err error

	if len(e.Key) == 0 || len(e.Value) == 0 {
		return nil
	}
	if len(e.EventTags) == 0 {
		e.EventTags = []string{events.CoreState}

	}

	// Add some more information to the Event, such as hostname and a timestamp.
	e.Timestamp = time.Now()
	if len(os.Getenv("ROOM_SYSTEM")) > 0 {
		e.GeneratingSystem = os.Getenv("SYSTEM_ID")
		// +deploy not_required
		if len(os.Getenv("DEVELOPMENT_HOSTNAME")) > 0 {
			// +deploy not_required
			e.GeneratingSystem = os.Getenv("DEVELOPMENT_HOSTNAME")
		}
	} else {
		// isn't it running in a docker container in aws? this won't work?
		e.GeneratingSystem, err = os.Hostname()
	}
	if err != nil {
		return err
	}

	if len(os.Getenv("ROOM_SYSTEM")) > 0 {
		e.AddToTags(events.RoomSystem)
	}

	Messenger.SendEvent(e)

	return err
}

// PublishError takes an error message and cause for the error, and then builds an Event to send to the event router.
func PublishError(errorStr string, cause string, device string) {
	deviceInfo := strings.Split(device, "-")
	room := fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1])

	e := events.Event{
		TargetDevice: events.GenerateBasicDeviceInfo(device),
		AffectedRoom: events.GenerateBasicRoomInfo(room),
		Key:          "Error String",
		Value:        errorStr,
	}

	e.AddToTags(events.Error, cause)

	SendEvent(e)
}
