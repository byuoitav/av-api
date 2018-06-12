package commandevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/structs"
)

//ChangeVideoInputDefault is struct that implements the CommandEvaluation struct
type ChangeVideoInputDefault struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (p *ChangeVideoInputDefault) Evaluate(room base.PublicRoom, requestor string) (actions []base.ActionStructure, count int, err error) {
	count = 0
	//RoomWideSetVideoInput
	if len(room.CurrentVideoInput) > 0 { // Check if the user sent a PUT body changing the current video input
		var tempActions []base.ActionStructure

		tempActions, err = generateChangeInputByRole(
			"VideoOut",
			room.CurrentVideoInput,
			room.Room,
			room.Building,
			"ChangeVideoInputDefault",
			requestor,
		)

		if err != nil {
			return
		}

		actions = append(actions, tempActions...)
	}

	//Displays
	for _, d := range room.Displays { // Loop through the devices array (potentially) passed in the user's PUT body
		if len(d.Input) < 1 {
			continue
		}

		var action base.ActionStructure

		action, err = generateChangeInputByDevice(d.Device, room.Room, room.Building, "ChangeVideoInputDefault", requestor)
		if err != nil {
			return
		}
		actions = append(actions, action)
	}

	count = len(actions)
	return
}

//Validate fulfills the Fulfill requirement on the command interface
func (p *ChangeVideoInputDefault) Validate(action base.ActionStructure) (err error) {
	return nil
}

//GetIncompatibleCommands keeps track of actions that are incompatable (on the same device)
func (p *ChangeVideoInputDefault) GetIncompatibleCommands() (incompatableActions []string) {
	return
}

func generateChangeInputByDevice(dev base.Device, room, building, generatingEvaluator, requestor string) (action base.ActionStructure, err error) {
	var curDevice structs.Device

	roomID := fmt.Sprintf("%v-%v", building, room)
	devices, err := db.GetDB().GetDevicesByRoom(roomID)
	if err != nil {
		return
	}

	inputID := getDeviceIDFromShortname(dev.Input, devices)

	for _, device := range devices {
		if strings.EqualFold(device.Name, dev.Name) {
			curDevice = device
			break
		}
	}

	if len(curDevice.ID) == 0 {
		err = errors.New(fmt.Sprintf("unable to find a device in the room matching the name: %s", dev.Name))
		return
	}

	paramMap := make(map[string]string)
	var portSource string

	for _, port := range curDevice.Ports {
		if strings.EqualFold(port.SourceDevice, inputID) {
			paramMap["port"] = port.ID
			portSource = port.SourceDevice
			break
		}
	}

	if len(paramMap) == 0 {
		log.L.Error("[command_evaluators] No port found for input.")
		return
	}

	destination := base.DestinationDevice{
		Device: curDevice,
	}

	if structs.HasRole(curDevice, "AudioOut") {
		destination.AudioDevice = true
	}

	if structs.HasRole(curDevice, "VideoOut") {
		destination.Display = true
	}

	eventInfo := events.EventInfo{
		Type:           events.CORESTATE,
		EventCause:     events.USERINPUT,
		Device:         dev.Name,
		EventInfoKey:   "input",
		EventInfoValue: portSource,
		Requestor:      requestor,
	}

	action = base.ActionStructure{
		Action:              "ChangeInput",
		GeneratingEvaluator: generatingEvaluator,
		Device:              curDevice,
		Parameters:          paramMap,
		DeviceSpecific:      true,
		Overridden:          false,
		EventLog:            []events.EventInfo{eventInfo},
	}

	return
}

func generateChangeInputByRole(role, input, room, building, generatingEvaluator, requestor string) (actions []base.ActionStructure, err error) {
	roomID := fmt.Sprintf("%v-%v", building, room)
	devicesToChange, err := db.GetDB().GetDevicesByRoomAndRole(roomID, role)
	if err != nil {
		return
	}

	var source string

	for _, d := range devicesToChange { // Loop through the devices in the room
		paramMap := make(map[string]string) // Start building parameter map

		//Get the port mapping for the device
		for _, curPort := range d.Ports { // Loop through the found ports

			if strings.EqualFold(curPort.SourceDevice, input) {

				paramMap["port"] = curPort.ID
				source = curPort.SourceDevice
				break

			}

		}

		if len(paramMap) == 0 {
			log.L.Error("[command_evaluators] No port found for input.")
			return
		}

		dest := base.DestinationDevice{
			Device: d,
		}

		if structs.HasRole(d, "AudioOut") {
			dest.AudioDevice = true
		}

		if structs.HasRole(d, "VideoOut") {
			dest.Display = true
		}

		eventInfo := events.EventInfo{
			Type:           events.USERACTION,
			EventCause:     events.USERINPUT,
			Device:         d.Name,
			EventInfoKey:   "input",
			EventInfoValue: source,
			Requestor:      requestor,
		}

		action := base.ActionStructure{
			Action:              "ChangeInput",
			GeneratingEvaluator: generatingEvaluator,
			Device:              d,
			DestinationDevice:   dest,
			Parameters:          paramMap,
			DeviceSpecific:      false,
			Overridden:          false,
			EventLog:            []events.EventInfo{eventInfo},
		}

		actions = append(actions, action)

		return
	}

	return
}
