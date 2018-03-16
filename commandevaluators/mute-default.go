package commandevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

//MuteDefault implements CommandEvaluation
type MuteDefault struct {
}

/*
 	Evalute takes a public room struct, scans the struct and builds any needed
	actions based on the contents of the struct.
*/
func (p *MuteDefault) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	base.Log("Evaluating for Mute command.")

	var actions []base.ActionStructure

	destination := base.DestinationDevice{
		AudioDevice: true,
	}

	eventInfo := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.CORESTATE,
		EventCause:     eventinfrastructure.USERINPUT,
		EventInfoKey:   "muted",
		EventInfoValue: "true",
		Requestor:      requestor,
	}

	if room.Muted != nil && *room.Muted {

		base.Log("Room-wide Mute request recieved. Retrieving all devices.")

		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		base.Log("Muting all devices in room.")

		for _, device := range devices {
			if device.Output {
				base.Log("Adding device %+v", device.Name)

				eventInfo.Device = device.Name
				destination.Device = device

				if device.HasRole("VideoOut") {
					destination.Display = true
				}

				actions = append(actions, base.ActionStructure{
					Action:              "Mute",
					GeneratingEvaluator: "MuteDefault",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []eventinfrastructure.EventInfo{eventInfo},
				})
			}
		}
	}

	//scan the room struct
	base.Log("Evaluating audio devices for Mute command.")

	//generate commands
	for _, audioDevice := range room.AudioDevices {
		if audioDevice.Muted != nil && *audioDevice.Muted {

			//get the device
			device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
			if err != nil {
				return []base.ActionStructure{}, 0, err
			}

			eventInfo.Device = device.Name
			destination.Device = device

			actions = append(actions, base.ActionStructure{
				Action:              "Mute",
				GeneratingEvaluator: "MuteDefault",
				Device:              device,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []eventinfrastructure.EventInfo{eventInfo},
			})

		}

	}

	return actions, len(actions), nil

}

// Validate takes an ActionStructure and determines if the command and parameter are valid for the device specified
func (p *MuteDefault) Validate(action base.ActionStructure) error {

	base.Log("Validating for command \"Mute\".")

	ok, _ := CheckCommands(action.Device.Commands, "Mute")

	// fmt.Printf("action.Device.Commands contains: %+v\n", action.Device.Commands)
	fmt.Printf("Device ID: %v\n", action.Device.ID)
	fmt.Printf("checkCommands returns: %v\n", ok)

	if !ok || !strings.EqualFold(action.Action, "Mute") {
		base.Log("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for " + action.Device.Name)
	}

	base.Log("Done.")

	return nil
}

//  GetIncompatableActions returns a list of commands that are incompatabl with this one (i.e. 'standby' and 'power on', or 'mute' and 'volume up')
func (p *MuteDefault) GetIncompatibleCommands() (incompatibleActions []string) {

	incompatibleActions = []string{
		"UnMute",
	}

	return
}
