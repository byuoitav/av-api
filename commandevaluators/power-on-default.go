package commandevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/fatih/color"
)

// PowerOn is struct that implements the CommandEvaluation struct
type PowerOnDefault struct {
}

// Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (p *PowerOnDefault) Evaluate(room base.PublicRoom, requestor string) (actions []base.ActionStructure, count int, err error) {
	count = 0

	log.Printf("Evaluating for PowerOn command.")
	color.Set(color.FgYellow, color.Bold)
	log.Printf("requestor: %s", requestor)
	color.Unset()

	eventInfo := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.CORESTATE,
		EventCause:     eventinfrastructure.USERINPUT,
		EventInfoKey:   "power",
		EventInfoValue: "on",
		Requestor:      requestor,
	}

	var devices []structs.Device
	if strings.EqualFold(room.Power, "on") {

		log.Printf("Room-wide PowerOn request received. Retrieving all devices.")

		devices, err = dbo.GetDevicesByRoom(room.Building, room.Room)
		if err != nil {
			return
		}

		log.Printf("Setting power 'on' state for all output devices.")

		for _, device := range devices {

			if device.Output {

				destination := se.DestinationDevice{
					Device: device,
				}

				if device.HasRole("AudioOut") {
					destination.AudioDevice = true
				}

				if device.HasRole("VideoOut") {
					destination.Display = true
				}

				log.Printf("Adding device %+v", device.Name)

				eventInfo.Device = device.Name
				actions = append(actions, base.ActionStructure{
					Action:              "PowerOn",
					Device:              device,
					DestinationDevice:   destination,
					GeneratingEvaluator: "PowerOnDefault",
					DeviceSpecific:      false,
					EventLog:            []eventinfrastructure.EventInfo{eventInfo},
				})
			}
		}
	}

	// Now we go through and check if power 'on' was set for any other device.
	log.Printf("Evaluating displays for power on command.")
	for _, device := range room.Displays {

		actions, err = p.evaluateDevice(device.Device, actions, devices, room.Room, room.Building, eventInfo)
		if err != nil {
			return
		}
	}

	for _, device := range room.AudioDevices {

		log.Printf("Evaluating audio devices for command power on. ")

		actions, err = p.evaluateDevice(device.Device, actions, devices, room.Room, room.Building, eventInfo)
		if err != nil {
			return
		}
	}

	log.Printf("%v actions generated.", len(actions))
	log.Printf("Evaluation complete.")

	count = len(actions)
	return
}

// Validate fulfills the Fulfill requirement on the command interface
func (p *PowerOnDefault) Validate(action base.ActionStructure) (err error) {

	log.Printf("Validating action for comand PowerOn")

	ok, _ := CheckCommands(action.Device.Commands, "PowerOn")
	if !ok || !strings.EqualFold(action.Action, "PowerOn") {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	log.Printf("Done.")
	return
}

// GetIncompatibleCommands keeps track of actions that are incompatable (on the same device)
func (p *PowerOnDefault) GetIncompatibleCommands() (incompatableActions []string) {
	incompatableActions = []string{
		"standby",
	}

	return
}

// Evaluate devices just pulls out the process we do with the audio-devices and displays into one function.
func (p *PowerOnDefault) evaluateDevice(device base.Device,
	actions []base.ActionStructure,
	devices []structs.Device,
	room string,
	building string,
	eventInfo eventinfrastructure.EventInfo) ([]base.ActionStructure, error) {

	// Check if we even need to start anything
	if strings.EqualFold(device.Power, "on") {
		// check if we already added it
		index := checkActionListForDevice(actions, device.Name, room, building)
		if index == -1 {
			// Get the device, check the list of already retreived devices first, if not there,
			// hit the DB up for it.
			dev, err := getDevice(devices, device.Name, room, building)
			if err != nil {
				return actions, err
			}

			destination := se.DestinationDevice{
				Device: dev,
			}

			if dev.HasRole("AudioOut") {
				destination.AudioDevice = true
			}

			if dev.HasRole("VideoOut") {
				destination.Display = true
			}

			eventInfo.Device = dev.Name
			destination.Device = dev

			actions = append(actions, base.ActionStructure{
				Action:              "PowerOn",
				Device:              dev,
				DestinationDevice:   destination,
				GeneratingEvaluator: "PowerOnDefault",
				DeviceSpecific:      true,
				EventLog:            []eventinfrastructure.EventInfo{eventInfo},
			})
		}
	}
	return actions, nil
}
