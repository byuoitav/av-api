package commandevaluators

import (
	"fmt"
	"strings"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
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
	var output structs.Device
	var input structs.Device

	roomID := fmt.Sprintf("%v-%v", building, room)
	devices, err := db.GetDB().GetDevicesByRoom(roomID)
	if err != nil {
		return
	}

	// get the input/output devices
	for _, device := range devices {
		if strings.EqualFold(device.Name, dev.Name) {
			output = device
		} else if strings.EqualFold(device.Name, dev.Input) {
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

	deviceInfo := strings.Split(output.ID, "-")

	roomInfo := events.BasicRoomInfo{
		BuildingID: building,
		RoomID:     roomID,
	}

	eventInfo := events.Event{
		TargetDevice: events.BasicDeviceInfo{
			BasicRoomInfo: events.BasicRoomInfo{
				BuildingID: deviceInfo[0],
				RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
			},
			DeviceID: output.ID,
		},
		AffectedRoom: roomInfo,
		Key:          "input",
		Value:        input.ID,
		User:         requestor,
	}

	eventInfo.EventTags = append(eventInfo.EventTags, events.CoreState, events.UserGenerated)

	action = base.ActionStructure{
		Action:              "ChangeInput",
		GeneratingEvaluator: generatingEvaluator,
		Device:              output,
		DestinationDevice:   destination,
		Parameters:          paramMap,
		DeviceSpecific:      true,
		Overridden:          false,
		EventLog:            []events.Event{eventInfo},
	}

	return
}

func generateChangeInputByRole(role, input, room, building, generatingEvaluator, requestor string) (actions []base.ActionStructure, err error) {
	roomID := fmt.Sprintf("%v-%v", building, room)
	devicesToChange, err := db.GetDB().GetDevicesByRoomAndRole(roomID, role)
	if err != nil {
		return
	}

	// get the input device
	inputDevice, err := db.GetDB().GetDevice(input)
	if err != nil {
		return
	}

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

		deviceInfo := strings.Split(d.ID, "-")

		roomInfo := events.BasicRoomInfo{
			BuildingID: building,
			RoomID:     roomID,
		}

		eventInfo := events.Event{
			TargetDevice: events.BasicDeviceInfo{
				BasicRoomInfo: events.BasicRoomInfo{
					BuildingID: deviceInfo[0],
					RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
				},
				DeviceID: d.ID,
			},
			AffectedRoom: roomInfo,
			Key:          "input",
			Value:        inputDevice.ID,
			User:         requestor,
		}

		eventInfo.EventTags = append(eventInfo.EventTags, events.CoreState, events.UserGenerated)

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
