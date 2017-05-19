package commandevaluators

import (
	"errors"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

//ChangeVideoInputDefault is struct that implements the CommandEvaluation struct
type ChangeVideoInputDefault struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (p *ChangeVideoInputDefault) Evaluate(room base.PublicRoom) (actions []base.ActionStructure, err error) {
	//RoomWideSetVideoInput
	if len(room.CurrentVideoInput) > 0 { // Check if the user sent a PUT body changing the current video input
		var tempActions []base.ActionStructure

		tempActions, err = generateChangeInputByRole(
			"VideoOut",
			room.CurrentVideoInput,
			room.Room,
			room.Building,
			"ChangeVideoInputDefault",
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

		action, err = generateChangeInputByDevice(d.Device, room.Room, room.Building, "ChangeVideoInputDefault")
		if err != nil {
			return
		}
		actions = append(actions, action)
	}

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

func generateChangeInputByDevice(dev base.Device, room string, building string, generatingEvaluator string) (action base.ActionStructure, err error) {

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

	eventInfo := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.USERACTION,
		EventCause:     eventinfrastructure.USERINPUT,
		Device:         dev.Name,
		EventInfoKey:   "input",
		EventInfoValue: paramMap["port"],
	}

	action = base.ActionStructure{
		Action:              "ChangeInput",
		GeneratingEvaluator: generatingEvaluator,
		Device:              curDevice,
		Parameters:          paramMap,
		DeviceSpecific:      true,
		Overridden:          false,
		EventLog:            []eventinfrastructure.EventInfo{eventInfo},
	}

	return
}

func generateChangeInputByRole(role string, input string, room string, building string, generatingEvaluator string) (actions []base.ActionStructure, err error) {
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

		eventInfo := eventinfrastructure.EventInfo{
			Type:           eventinfrastructure.USERACTION,
			EventCause:     eventinfrastructure.USERINPUT,
			Device:         d.Name,
			EventInfoKey:   "input",
			EventInfoValue: paramMap["port"],
		}

		action := base.ActionStructure{
			Action:              "ChangeInput",
			GeneratingEvaluator: generatingEvaluator,
			Device:              d,
			Parameters:          paramMap,
			DeviceSpecific:      false,
			Overridden:          false,
			EventLog:            []eventinfrastructure.EventInfo{eventInfo},
		}

		actions = append(actions, action)

		return
	}

	return
}
