package commandevaluators

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
)

//ChangeVideoInputDefault is struct that implements the CommandEvaluation struct
type ChangeVideoInputDefault struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (p *ChangeVideoInputDefault) Evaluate(dbRoom structs.Room, room base.PublicRoom, requestor string) (actions []base.ActionStructure, count int, err error) {
	count = 0
	//RoomWideSetVideoInput
	if len(room.CurrentVideoInput) > 0 { // Check if the user sent a PUT body changing the current video input
		var tempActions []base.ActionStructure

		tempActions, err = generateChangeInputByRole(
			dbRoom,
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

		var action []base.ActionStructure

		action, err = generateChangeInputByDevice(dbRoom, d.Device, room.Room, room.Building, "ChangeVideoInputDefault", requestor)
		if err != nil {
			return
		}
		actions = append(actions, action...)
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

func generateChangeInputByDevice(dbRoom structs.Room, dev base.Device, room, building, generatingEvaluator, requestor string) (actions []base.ActionStructure, err error) {
	var output structs.Device
	var input structs.Device
	var streamURL string

	inputDeviceString := dev.Input

	//Check for stream delimiter
	streamParams := make(map[string]string)
	streamDelimiterIndex := strings.Index(inputDeviceString, "|")
	if streamDelimiterIndex != -1 {
		streamChars := []rune(inputDeviceString)
		streamURL = url.QueryEscape(string(streamChars[(streamDelimiterIndex + 1):len(inputDeviceString)]))
		inputDeviceString = string(streamChars[0:streamDelimiterIndex])
		streamParams["streamURL"] = streamURL
	}

	// get the input/output devices
	for _, device := range dbRoom.Devices {
		if strings.EqualFold(device.Name, dev.Name) {
			output = device
		} else if strings.EqualFold(device.Name, inputDeviceString) {
			input = device
		}
	}

	if len(output.ID) == 0 {
		err = fmt.Errorf("unable to find a device in the room matching the name %s", dev.Name)
		return
	}

	if len(input.ID) == 0 {
		err = fmt.Errorf("unable to find a device in the room matching the name %s", dev.Input)
		return
	}

	paramMap := make(map[string]string)

	for _, port := range output.Ports {
		if strings.EqualFold(port.SourceDevice, input.ID) {
			paramMap["port"] = port.ID
			break
		}
	}

	if len(paramMap) == 0 {
		log.L.Error("[command_evaluators] No port found for input.")
		return
	}

	destination := base.DestinationDevice{
		Device: output,
	}

	if structs.HasRole(output, "AudioOut") {
		destination.AudioDevice = true
	}

	if structs.HasRole(output, "VideoOut") {
		destination.Display = true
	}

	eventInfo := events.Event{
		TargetDevice: events.GenerateBasicDeviceInfo(output.ID),
		AffectedRoom: events.GenerateBasicRoomInfo(dbRoom.ID),
		Key:          "input",
		Value:        input.Name,
		User:         requestor,
	}

	eventInfo.AddToTags(events.CoreState, events.UserGenerated)

	action := base.ActionStructure{
		Action:              "ChangeInput",
		GeneratingEvaluator: generatingEvaluator,
		Device:              output,
		DestinationDevice:   destination,
		Parameters:          paramMap,
		DeviceSpecific:      true,
		Overridden:          false,
		EventLog:            []events.Event{eventInfo},
	}

	actions = append(actions, action)

	if streamDelimiterIndex != -1 {
		for i := range actions {
			if actions[i].Action == "ChangeInput" {
				for j := range actions[i].EventLog {
					if actions[i].EventLog[j].Key == "input" {
						actions[i].EventLog[j].Value = fmt.Sprintf("%s|%s", actions[i].EventLog[j].Value, streamURL)
					}
				}
			}
		}

		actions = append(actions, base.ActionStructure{
			Action:              "ChangeStream",
			GeneratingEvaluator: generatingEvaluator,
			Device:              output,
			DestinationDevice:   destination,
			Parameters:          streamParams,
			DeviceSpecific:      true,
			Overridden:          false,
			EventLog:            []events.Event{},
		})
	}

	return
}

func generateChangeInputByRole(dbRoom structs.Room, role, input, room, building, generatingEvaluator, requestor string) (actions []base.ActionStructure, err error) {
	devicesToChange := FilterDevicesByRole(dbRoom.Devices, role)

	// get the input device
	inputDevice := FindDevice(dbRoom.Devices, input)

	for _, d := range devicesToChange { // Loop through the devices in the room
		paramMap := make(map[string]string) // Start building parameter map

		//Get the port mapping for the device
		for _, curPort := range d.Ports { // Loop through the found ports
			if strings.EqualFold(curPort.SourceDevice, inputDevice.ID) {
				paramMap["port"] = curPort.ID
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

		eventInfo := events.Event{
			TargetDevice: events.GenerateBasicDeviceInfo(d.ID),
			AffectedRoom: events.GenerateBasicRoomInfo(dbRoom.ID),
			Key:          "input",
			Value:        inputDevice.Name,
			User:         requestor,
		}

		eventInfo.AddToTags(events.CoreState, events.UserGenerated)

		action := base.ActionStructure{
			Action:              "ChangeInput",
			GeneratingEvaluator: generatingEvaluator,
			Device:              d,
			DestinationDevice:   dest,
			Parameters:          paramMap,
			DeviceSpecific:      false,
			Overridden:          false,
			EventLog:            []events.Event{eventInfo},
		}

		actions = append(actions, action)

		return
	}

	return
}
