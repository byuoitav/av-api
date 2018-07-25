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
	"github.com/fatih/color"
)

// PowerOnDefault implements the CommandEvaluation struct.
type PowerOnDefault struct {
}

// Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (p *PowerOnDefault) Evaluate(room base.PublicRoom, requestor string) (actions []base.ActionStructure, count int, err error) {
	count = 0

	log.L.Info("[command_evaluators] Evaluating for PowerOn command.")
	color.Set(color.FgYellow, color.Bold)
	log.L.Infof("[command_evaluators] Requestor: %s", requestor)
	color.Unset()

	eventInfo := events.EventInfo{
		Type:           events.CORESTATE,
		EventCause:     events.USERINPUT,
		EventInfoKey:   "power",
		EventInfoValue: "on",
		Requestor:      requestor,
	}

	var devices []structs.Device
	if strings.EqualFold(room.Power, "on") {

		log.L.Info("[command_evaluators] Room-wide PowerOn request received. Retrieving all devices.")

		roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
		devices, err = db.GetDB().GetDevicesByRoom(roomID)
		if err != nil {
			return
		}

		log.L.Info("[command_evaluators] Setting power 'on' state for all output devices.")

		for _, device := range devices {

			if device.Type.Output {

				destination := base.DestinationDevice{
					Device: device,
				}

				if structs.HasRole(device, "AudioOut") {
					destination.AudioDevice = true
				}

				if structs.HasRole(device, "VideoOut") {
					destination.Display = true
				}

				log.L.Info("[command_evaluators] Adding device %+v", device.Name)

				eventInfo.Device = device.Name
				actions = append(actions, base.ActionStructure{
					Action:              "PowerOn",
					Device:              device,
					DestinationDevice:   destination,
					GeneratingEvaluator: "PowerOnDefault",
					DeviceSpecific:      false,
					EventLog:            []events.EventInfo{eventInfo},
				})
			}
		}
	}

	// Now we go through and check if power 'on' was set for any other device.
	log.L.Info("[command_evaluators] Evaluating displays for power on command.")
	for _, device := range room.Displays {

		actions, err = p.evaluateDevice(device.Device, actions, devices, room.Room, room.Building, eventInfo)
		if err != nil {
			return
		}
	}

	for _, device := range room.AudioDevices {

		log.L.Info("[command_evaluators] Evaluating audio devices for command power on. ")

		actions, err = p.evaluateDevice(device.Device, actions, devices, room.Room, room.Building, eventInfo)
		if err != nil {
			return
		}
	}

	log.L.Infof("[command_evaluators] %v actions generated.", len(actions))
	log.L.Info("[command_evaluators] Evaluation complete.")

	count = len(actions)
	return
}

// Validate fulfills the Fulfill requirement on the command interface
func (p *PowerOnDefault) Validate(action base.ActionStructure) (err error) {

	log.L.Info("[command_evaluators] Validating action for comand PowerOn")

	ok, _ := CheckCommands(action.Device.Type.Commands, "PowerOn")
	if !ok || !strings.EqualFold(action.Action, "PowerOn") {
		msg := fmt.Sprintf("[command_evaluators] ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		log.L.Error(msg)
		return errors.New(msg)
	}

	log.L.Info("[command_evaluators] Done.")
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
	eventInfo events.EventInfo) ([]base.ActionStructure, error) {

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

			destination := base.DestinationDevice{
				Device: dev,
			}

			if structs.HasRole(dev, "AudioOut") {
				destination.AudioDevice = true
			}

			if structs.HasRole(dev, "VideoOut") {
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
				EventLog:            []events.EventInfo{eventInfo},
			})
		}
	}
	return actions, nil
}
