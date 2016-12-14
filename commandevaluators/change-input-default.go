package commandevaluators

import (
	"errors"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//ChangeInput is struct that implements the CommandEvaluation struct
type ChangeInputDefault struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (p *ChangeInputDefault) Evaluate(room base.PublicRoom) (actions []base.ActionStructure, err error) {
	//RoomWideSetVideoInput
	if len(room.CurrentVideoInput) > 0 { // Check if the user sent a PUT body changing the current video input
		var tempActions []base.ActionStructure

		tempActions, err = generateChangeInputByRole(
			"VideoOut",
			room.CurrentVideoInput,
			room.Room,
			room.Building,
		)

		if err != nil {
			return
		}

		actions = append(actions, tempActions...)
	}

	//RoomWideSetAudioInput
	if len(room.CurrentAudioInput) > 0 { // Check if the user sent a PUT body changing the current audio input
		var tempActions []base.ActionStructure

		//generate action
		tempActions, err = generateChangeInputByRole(
			"AudioOut",
			room.CurrentVideoInput,
			room.Room,
			room.Building,
		)
		if err != nil {
			return
		}
		actions = append(actions, tempActions...)
	}

	//Displays
	for _, d := range room.Displays { // Loop through the devices array (potentially) passed in the user's PUT body
		var action base.ActionStructure
		action, err = generateChangeInputByDevice(d.Device, room.Room, room.Building)
		if err != nil {
			return
		}
		actions = append(actions, action)
	}

	//AudioDevice
	for _, d := range room.AudioDevices { // Loop through the audio devices array (potentially) passed in the user's PUT body
		var action base.ActionStructure
		action, err = generateChangeInputByDevice(d.Device, room.Room, room.Building)
		if err != nil {
			return
		}
		actions = append(actions, action)
	}

	return
}

//Validate fulfills the Fulfill requirement on the command interface
func (p *ChangeInputDefault) Validate(action base.ActionStructure) (err error) {
	return nil
}

//GetIncompatableCommands keeps track of actions that are incompatable (on the same device)
func (p *ChangeInputDefault) GetIncompatableCommands() (incompatableActions []string) {
	return
}

func generateChangeInputByDevice(dev base.Device, room string, building string) (action base.ActionStructure, err error) {
	var curDevice accessors.Device

	curDevice, err = dbo.GetDeviceByName(building, room, dev.Name)
	if err != nil {
		return
	}

	paramMap := make(map[string]string)

	for _, port := range curDevice.Ports {
		if strings.EqualFold(port.Source, dev.Input) {
			paramMap["port"] = port.Name
			break
		}
	}

	if len(paramMap) == 0 {
		err = errors.New("No port found for input.")
		return
	}

	action = base.ActionStructure{
		Action:              "change-input",
		GeneratingEvaluator: "changeInput",
		Device:              curDevice,
		Parameters:          paramMap,
		DeviceSpecific:      true,
		Overridden:          false,
	}

	return
}

func generateChangeInputByRole(role string, input string, room string, building string) (actions []base.ActionStructure, err error) {
	devicesToChange, err := dbo.GetDevicesByBuildingAndRoomAndRole(building, room, role)
	if err != nil {
		return
	}

	for _, d := range devicesToChange { // Loop through the devices in the room
		paramMap := make(map[string]string) // Start building parameter map

		//Get the port mapping for the device
		for _, curPort := range d.Ports { // Loop through the found ports
			if strings.EqualFold(curPort.Source, input) {
				paramMap["port"] = curPort.Name
				break
			}
		}

		if len(paramMap) == 0 {
			err = errors.New("No port found for input.")
			return
		}

		action := base.ActionStructure{
			Action:              "ChangeInput",
			GeneratingEvaluator: "ChangeInputDefault",
			Device:              d,
			Parameters:          paramMap,
			DeviceSpecific:      false,
			Overridden:          false,
		}

		actions = append(actions, action)

		return
	}

	return
}
