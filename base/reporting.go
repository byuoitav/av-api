package base

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/log"
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
	var err error

	if len(e.Key) == 0 || len(e.Value) == 0 {
		return nil
	}

	// Add some more information to the Event, such as hostname and a timestamp.
	e.Timestamp = time.Now()
	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {
		e.GeneratingSystem = os.Getenv("SYSTEM_ID")
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
		e.AddToTags(os.Getenv("LOCAL_ENVIRONMENT"))
	}

	eventBytes, err := json.Marshal(e)
	if err != nil {
		log.L.Errorf("failed to marshal event : %s", err.Error())
		return err
	}

	Messenger.SendEvent(base.EventWrapper{
		Room:  fmt.Sprintf("%s-%s", e.AffectedRoom.BuildingID, e.AffectedRoom.RoomID),
		Event: eventBytes,
	})

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
