package commandevaluators

import (
	"errors"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

// Standby is struct that implements the CommandEvaluation struct
type StandbyDefault struct {
}

// Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (s *StandbyDefault) Evaluate(room base.PublicRoom, requestor string) (actions []base.ActionStructure, count int, err error) {

	base.Log("Evaluating for Standby Command.")

	var devices []structs.Device
	eventInfo := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.CORESTATE,
		EventCause:     eventinfrastructure.USERINPUT,
		EventInfoKey:   "power",
		EventInfoValue: "standby",
		Requestor:      requestor,
	}

	if strings.EqualFold(room.Power, "standby") {

		base.Log("Room-wide power set. Retrieving all devices.")
		devices, err = dbo.GetDevicesByRoom(room.Building, room.Room)
		if err != nil {
			return
		}

		base.Log("Setting power to 'standby' state for all devices with a 'standby' power state, that are also output devices.")
		for _, device := range devices {

			containsStandby := false
			for _, ps := range device.PowerStates {
				if strings.EqualFold(ps, "Standby") {
					containsStandby = true
					break
				}
			}

			if containsStandby && device.Output {

				base.Log("Adding device %+v", device.Name)

				dest := base.DestinationDevice{
					Device: device,
				}

				if device.HasRole("AudioOut") {
					dest.AudioDevice = true
				}

				if device.HasRole("VideoOut") {
					dest.AudioDevice = true
				}

				eventInfo.Device = device.Name
				actions = append(actions, base.ActionStructure{
					Action:              "Standby",
					Device:              device,
					DestinationDevice:   dest,
					GeneratingEvaluator: "StandbyDefault",
					DeviceSpecific:      false,
					EventLog:            []eventinfrastructure.EventInfo{eventInfo},
				})
			}
		}
	}

	// now we go through and check if power 'standby' was set for any other device.
	for _, device := range room.Displays {
		base.Log("Evaluating displays for command power standby. ")
		destination := base.DestinationDevice{AudioDevice: true}
		actions, err = s.evaluateDevice(device.Device, destination, actions, devices, room.Room, room.Building, eventInfo)
		if err != nil {
			return
		}
	}

	for _, device := range room.AudioDevices {
		base.Log("Evaluating audio devices for command power on. ")
		destination := base.DestinationDevice{AudioDevice: true}
		actions, err = s.evaluateDevice(device.Device, destination, actions, devices, room.Room, room.Building, eventInfo)
		if err != nil {
			return
		}
	}
	base.Log("%v actions generated.", len(actions))
	base.Log("Evaluation complete.")

	count = len(actions)
	return
}

// Validate fulfills the Fulfill requirement on the command interface
func (s *StandbyDefault) Validate(action base.ActionStructure) (err error) {
	base.Log("Validating action for command Standby.")

	ok, _ := CheckCommands(action.Device.Commands, "Standby")
	if !ok || !strings.EqualFold(action.Action, "Standby") {
		base.Log("ERROR. %s is an invalid command for %s", action.Action, action.Device.GetFullName())
		return errors.New(action.Action + " is an invalid command for" + action.Device.GetFullName())
	}

	base.Log("Done.")
	return
}

// GetIncompatibleCommands keeps track of actions that are incompatable (on the same device)
func (s *StandbyDefault) GetIncompatibleCommands() (incompatableActions []string) {
	incompatableActions = []string{
		"PowerOn",
	}

	return
}

// Evaluate devices just pulls out the process we do with the audio-devices and displays into one function.
func (s *StandbyDefault) evaluateDevice(device base.Device, destination base.DestinationDevice, actions []base.ActionStructure, devices []structs.Device, room string, building string, eventInfo eventinfrastructure.EventInfo) ([]base.ActionStructure, error) {
	// Check if we even need to start anything
	if strings.EqualFold(device.Power, "standby") {
		// check if we already added it
		index := checkActionListForDevice(actions, device.Name, room, building)
		if index == -1 {

			// Get the device, check the list of already retreived devices first, if not there,
			// hit the DB up for it.
			dev, err := getDevice(devices, device.Name, room, building)
			if err != nil {
				return actions, err
			}

			eventInfo.Device = device.Name
			destination.Device = dev

			actions = append(actions, base.ActionStructure{
				Action:              "Standby",
				Device:              dev,
				GeneratingEvaluator: "StandbyDefault",
				DeviceSpecific:      true,
				EventLog:            []eventinfrastructure.EventInfo{eventInfo},
			})
		}
	}
	return actions, nil
}
