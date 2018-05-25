package commandevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/structs"
)

type UnMuteDefault struct {
}

func (p *UnMuteDefault) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {
	base.Log("Evaluating UnMute command.")

	var actions []base.ActionStructure
	eventInfo := events.EventInfo{
		Type:           events.CORESTATE,
		EventCause:     events.USERINPUT,
		EventInfoKey:   "muted",
		EventInfoValue: "false",
		Requestor:      requestor,
	}

	destination := base.DestinationDevice{AudioDevice: true}

	//check if request is a roomwide unmute
	if room.Muted != nil && !*room.Muted {

		base.Log("Room-wide UnMute request recieved. Retrieving all devices")

		roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
		devices, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "AudioOut")
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		base.Log("UnMuting all devices in room.")

		for _, device := range devices {

			if device.Type.Output {

				base.Log("Adding device %+v", device.Name)

				eventInfo.Device = device.Name
				destination.Device = device

				if structs.HasRole(device, "VideoOut") {
					destination.Display = true
				}

				actions = append(actions, base.ActionStructure{
					Action:              "UnMute",
					GeneratingEvaluator: "UnMuteDefault",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []events.EventInfo{eventInfo},
				})

			}

		}

	}

	//check specific devices
	base.Log("Evaluating individual audio devices for unmuting.")

	for _, audioDevice := range room.AudioDevices {

		base.Log("Adding device %+v", audioDevice.Name)

		if audioDevice.Muted != nil && !*audioDevice.Muted {

			deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, audioDevice.Name)
			device, err := db.GetDB().GetDevice(deviceID)
			if err != nil {
				return []base.ActionStructure{}, 0, err
			}

			eventInfo.Device = device.Name
			destination.Device = device

			if structs.HasRole(device, "VideoOut") {
				destination.Display = true
			}

			actions = append(actions, base.ActionStructure{
				Action:              "UnMute",
				GeneratingEvaluator: "UnMuteDefault",
				Device:              device,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []events.EventInfo{eventInfo},
			})

		}

	}

	base.Log("%v actions generated.", len(actions))
	base.Log("Evalutation complete.")

	return actions, len(actions), nil

}

func (p *UnMuteDefault) Validate(action base.ActionStructure) error {

	base.Log("Validating action for command \"UnMute\"")

	ok, _ := CheckCommands(action.Device.Type.Commands, "UnMute")

	if !ok || !strings.EqualFold(action.Action, "UnMute") {
		base.Log("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	base.Log("Done.")
	return nil
}

func (p *UnMuteDefault) GetIncompatibleCommands() (incompatibleActions []string) {

	incompatibleActions = []string{
		"Mute",
	}

	return
}
