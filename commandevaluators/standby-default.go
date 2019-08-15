package commandevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
)

// StandbyDefault implements the CommandEvaluator struct.
type StandbyDefault struct {
}

// Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (s *StandbyDefault) Evaluate(dbRoom structs.Room, room base.PublicRoom, requestor string) (actions []base.ActionStructure, count int, err error) {

	log.L.Info("[command_evaluators] Evaluating for Standby Command.")

	var devices []structs.Device

	roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)

	eventInfo := events.Event{
		Key:   "power",
		Value: "standby",
		User:  requestor,
	}

	eventInfo.AddToTags(events.CoreState, events.UserGenerated)

	if strings.EqualFold(room.Power, "standby") {

		log.L.Info("[command_evaluators] Room-wide power set. Retrieving all devices.")

		log.L.Debugf("[command_evaluators] Setting power to 'standby' state for all devices with a 'standby' power state, that are also output devices.")
		for _, device := range dbRoom.Devices {

			if device.Type.Output {
				//check to see if it has the standby command

				cmd := device.GetCommandByID("Standby")
				if len(cmd.ID) < 1 {
					log.L.Debugf("Device %v doesn't have standby command. Skipping.")
					continue
				}

				log.L.Debugf("[command_evaluators] Adding device %+v from room-wide standby", device.Name)

				dest := base.DestinationDevice{
					Device: device,
				}

				if structs.HasRole(device, "AudioOut") {
					dest.AudioDevice = true
				}

				if structs.HasRole(device, "VideoOut") {
					dest.AudioDevice = true
				}

				eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

				eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(device.ID)

				actions = append(actions, base.ActionStructure{
					Action:              "Standby",
					Device:              device,
					DestinationDevice:   dest,
					GeneratingEvaluator: "StandbyDefault",
					DeviceSpecific:      false,
					EventLog:            []events.Event{eventInfo},
				})
			}
		}
	}

	// now we go through and check if power 'standby' was set for any other device.
	for _, device := range room.Displays {
		log.L.Info("[command_evaluators] Evaluating displays for command power standby. ")

		destination := base.DestinationDevice{AudioDevice: false, Display: true}

		actions, err = s.evaluateDevice(device.Device, destination, actions, devices, room.Room, room.Building, eventInfo)
		if err != nil {
			return
		}
	}

	for _, device := range room.AudioDevices {
		log.L.Info("[command_evaluators] Evaluating audio devices for command power on. ")
		destination := base.DestinationDevice{AudioDevice: true, Display: false}
		actions, err = s.evaluateDevice(device.Device, destination, actions, devices, room.Room, room.Building, eventInfo)
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
func (s *StandbyDefault) Validate(action base.ActionStructure) (err error) {
	log.L.Info("[command_evaluators] Validating action for command Standby.")

	ok, _ := CheckCommands(action.Device.Type.Commands, "Standby")
	if !ok || !strings.EqualFold(action.Action, "Standby") {
		msg := fmt.Sprintf("[command_evaluators] ERROR. %s is an invalid command for %s", action.Action, action.Device.ID)
		log.L.Error(msg)
		return errors.New(msg)
	}

	log.L.Info("[command_evaluators] Done.")
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
func (s *StandbyDefault) evaluateDevice(device base.Device, destination base.DestinationDevice, actions []base.ActionStructure, devices []structs.Device, room string, building string, eventInfo events.Event) ([]base.ActionStructure, error) {
	// Check if we even need to start anything
	if strings.EqualFold(device.Power, "standby") {
		roomID := fmt.Sprintf("%s-%s", building, room)

		// check if we already added it
		index := checkActionListForDevice(actions, device.Name, room, building)
		if index == -1 {

			// Get the device, check the list of already retreived devices first, if not there,
			// hit the DB up for it.
			dev, err := getDevice(devices, device.Name, room, building)
			if err != nil {
				return actions, err
			}

			eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

			eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(dev.ID)

			destination.Device = dev

			actions = append(actions, base.ActionStructure{
				Action:              "Standby",
				Device:              dev,
				DestinationDevice:   destination,
				GeneratingEvaluator: "StandbyDefault",
				DeviceSpecific:      true,
				EventLog:            []events.Event{eventInfo},
			})

			////////////////////////
			///// MIRROR STUFF /////
			if structs.HasRole(dev, "MirrorMaster") {
				for _, port := range dev.Ports {
					if port.ID == "mirror" {
						DX, err := db.GetDB().GetDevice(port.DestinationDevice)
						if err != nil {
							return actions, err
						}

						cmd := DX.GetCommandByID("Standby")
						if len(cmd.ID) < 1 {
							continue
						}

						log.L.Info("[command_evaluators] Adding mirror device %+v", DX.Name)

						eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

						eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(DX.ID)

						actions = append(actions, base.ActionStructure{
							Action:              "Standby",
							Device:              DX,
							DestinationDevice:   destination,
							GeneratingEvaluator: "StandbyDefault",
							DeviceSpecific:      true,
							EventLog:            []events.Event{eventInfo},
						})
					}
				}
			}
			///// MIRROR STUFF /////
			////////////////////////
		}
	}
	return actions, nil
}
