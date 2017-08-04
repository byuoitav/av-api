package commandevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

// Standby is struct that implements the CommandEvaluation struct
type StandbyDefault struct {
}

// Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (s *StandbyDefault) Evaluate(room base.PublicRoom) (actions []base.ActionStructure, err error) {

	log.Printf("Evaluating for Standby Command.")

	var devices []accessors.Device
	eventInfo := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.CORESTATE,
		EventCause:     eventinfrastructure.USERINPUT,
		EventInfoKey:   "power",
		EventInfoValue: "standby",
	}

	if strings.EqualFold(room.Power, "standby") {

		log.Printf("Room-wide power set. Retrieving all devices.")
		devices, err = dbo.GetDevicesByRoom(room.Building, room.Room)
		if err != nil {
			return
		}

		log.Printf("Setting power to 'standby' state for all devices with a 'standby' power state.")
		for _, device := range devices {
			containsStandby := false
			for _, ps := range device.PowerStates {
				if strings.EqualFold(ps, "Standby") {
					containsStandby = true
					break
				}
			}

			if containsStandby {
				log.Printf("Adding device %+v", device.Name)
				eventInfo.Device = device.Name
				actions = append(actions, base.ActionStructure{
					Action:              "Standby",
					Device:              device,
					GeneratingEvaluator: "StandbyDefault",
					DeviceSpecific:      false,
					EventLog:            []eventinfrastructure.EventInfo{eventInfo},
				})
			}
		}
	}

	// now we go through and check if power 'standby' was set for any other device.
	for _, device := range room.Displays {
		log.Printf("Evaluating displays for command power standby. ")
		actions, err = s.evaluateDevice(device.Device, actions, devices, room.Room, room.Building, eventInfo)
		if err != nil {
			return
		}
	}

	for _, device := range room.AudioDevices {
		log.Printf("Evaluating audio devices for command power on. ")
		actions, err = s.evaluateDevice(device.Device, actions, devices, room.Room, room.Building, eventInfo)
		if err != nil {
			return
		}
	}
	log.Printf("%v actions generated.", len(actions))
	log.Printf("Evaluation complete.")

	return
}

// Validate fulfills the Fulfill requirement on the command interface
func (s *StandbyDefault) Validate(action base.ActionStructure) (err error) {
	log.Printf("Validating action for command Standby.")

	ok, _ := CheckCommands(action.Device.Commands, "Standby")
	if !ok || !strings.EqualFold(action.Action, "Standby") {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.GetFullName())
		return errors.New(action.Action + " is an invalid command for" + action.Device.GetFullName())
	}

	log.Printf("Done.")
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
func (s *StandbyDefault) evaluateDevice(device base.Device, actions []base.ActionStructure, devices []accessors.Device, room string, building string, eventInfo eventinfrastructure.EventInfo) ([]base.ActionStructure, error) {
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
