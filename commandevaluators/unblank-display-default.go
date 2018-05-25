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

type UnBlankDisplayDefault struct {
}

//Evaluate creates UnBlank actions for the entire room and for individual devices
func (p *UnBlankDisplayDefault) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	var actions []base.ActionStructure

	eventInfo := events.EventInfo{
		Type:           events.CORESTATE,
		EventCause:     events.USERINPUT,
		EventInfoKey:   "blanked",
		EventInfoValue: "false",
		Requestor:      requestor,
	}

	destination := base.DestinationDevice{Display: true}

	if room.Blanked != nil && !*room.Blanked {

		base.Log("Room-wide UnBlank request received. Retrieving all devices.")

		roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
		devices, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "VideoOut")
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		base.Log("Un-Blanking all displays in room.")

		for _, device := range devices {

			if device.Type.Output {

				base.Log("Adding Device %+v", device.Name)

				eventInfo.Device = device.Name
				destination.Device = device

				if structs.HasRole(device, "AudioOut") {
					destination.AudioDevice = true
				}

				actions = append(actions, base.ActionStructure{
					Action:              "UnblankDisplay",
					GeneratingEvaluator: "UnBlankDisplayDefault",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []events.EventInfo{eventInfo},
				})
			}

		}

	}

	base.Log("Evaluating individial displays for unblanking.")

	for _, display := range room.Displays {

		base.Log("Adding device %+v", display.Name)

		if display.Blanked != nil && !*display.Blanked {

			deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, display.Name)
			device, err := db.GetDB().GetDevice(deviceID)
			if err != nil {
				return []base.ActionStructure{}, 0, err
			}

			eventInfo.Device = device.Name
			destination.Device = device

			if structs.HasRole(device, "AudioOut") {
				destination.AudioDevice = true
			}

			actions = append(actions, base.ActionStructure{
				Action:              "UnblankDisplay",
				GeneratingEvaluator: "UnBlankDisplayDefault",
				Device:              device,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []events.EventInfo{eventInfo},
			})

		}
	}

	base.Log("Evaluation complete; %v actions generated.", len(actions))

	return actions, len(actions), nil
}

//Validate returns an error if a command is invalid for a device
func (p *UnBlankDisplayDefault) Validate(action base.ActionStructure) error {
	base.Log("Validating action for command \"UnBlank\"")

	ok, _ := CheckCommands(action.Device.Type.Commands, "UnblankDisplay")

	if !ok || !strings.EqualFold(action.Action, "UnblankDisplay") {
		base.Log("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	base.Log("Done.")
	return nil
}

//GetIncompatibleCommands returns a string array containing commands incompatible with UnBlank Display
func (p *UnBlankDisplayDefault) GetIncompatibleCommands() (incompatibleActions []string) {
	incompatibleActions = []string{
		"BlankScreen",
	}

	return
}
