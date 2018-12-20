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

// UnMuteDefault implements the CommandEvaluator struct.
type UnMuteDefault struct {
}

// Evaluate generates a list of actions based on the room information.
func (p *UnMuteDefault) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {
	log.L.Info("[command_evaluators] Evaluating UnMute command.")

	var actions []base.ActionStructure

	eventInfo := events.Event{
		Key:   "muted",
		Value: "false",
		User:  requestor,
	}

	roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)

	eventInfo.AddToTags(events.CoreState, events.UserGenerated)

	destination := base.DestinationDevice{AudioDevice: true}

	//check if request is a roomwide unmute
	if room.Muted != nil && !*room.Muted {

		log.L.Info("[command_evaluators] Room-wide UnMute request recieved. Retrieving all devices")

		devices, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "AudioOut")
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		log.L.Info("[command_evaluators] UnMuting all devices in room.")

		for _, device := range devices {

			if device.Type.Output {

				log.L.Infof("[command_evaluators] Adding device %+v", device.Name)

				eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

				eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(device.ID)

				destination.Device = device

				if structs.HasRole(device, "VideoOut") {
					destination.Display = true
				}

				actions = append(actions, base.ActionStructure{
					Action:              "UnMute",
					GeneratingEvaluator: "UnMuteDefault",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []events.Event{eventInfo},
				})

				////////////////////////
				///// MIRROR STUFF /////
				if structs.HasRole(device, "MirrorMaster") {
					for _, port := range device.Ports {
						if port.ID == "mirror" {
							DX, err := db.GetDB().GetDevice(port.DestinationDevice)
							if err != nil {
								return actions, len(actions), err
							}

							cmd := DX.GetCommandByID("UnMute")
							if len(cmd.ID) < 1 {
								continue
							}

							eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

							eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(DX.ID)

							log.L.Info("[command_evaluators] Adding mirror device %+v", DX.Name)

							actions = append(actions, base.ActionStructure{
								Action:              "UnMute",
								GeneratingEvaluator: "UnMuteDefault",
								Device:              DX,
								DestinationDevice:   destination,
								DeviceSpecific:      false,
								EventLog:            []events.Event{eventInfo},
							})
						}
					}
				}
				///// MIRROR STUFF /////
				////////////////////////
			}
		}
	}

	//check specific devices
	log.L.Info("[command_evaluators] Evaluating individual audio devices for unmuting.")

	for _, audioDevice := range room.AudioDevices {

		log.L.Infof("[command_evaluators] Adding device %+v", audioDevice.Name)

		if audioDevice.Muted != nil && !*audioDevice.Muted {

			deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, audioDevice.Name)
			device, err := db.GetDB().GetDevice(deviceID)
			if err != nil {
				return []base.ActionStructure{}, 0, err
			}

			eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

			eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(device.ID)

			destination.Device = device

			if structs.HasRole(device, "VideoOut") {
				destination.Display = true
			}

			actions = append(actions, base.ActionStructure{
				Action:              "UnMute",
				GeneratingEvaluator: "UnMuteDefault",
				Device:              device,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []events.Event{eventInfo},
			})

		}

	}

	log.L.Infof("[command_evaluators] %v actions generated.", len(actions))
	log.L.Info("[command_evaluators] Evalutation complete.")

	return actions, len(actions), nil

}

// Validate verified that the action information is correct.
func (p *UnMuteDefault) Validate(action base.ActionStructure) error {

	log.L.Info("[command_evaluators] Validating action for command \"UnMute\"")

	ok, _ := CheckCommands(action.Device.Type.Commands, "UnMute")

	if !ok || !strings.EqualFold(action.Action, "UnMute") {
		msg := fmt.Sprintf("[command_evaluators] ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		log.L.Error(msg)
		return errors.New(msg)
	}

	log.L.Info("[command_evaluators] Done.")
	return nil
}

// GetIncompatibleCommands determines the list of incompatible commands for this evaluator.
func (p *UnMuteDefault) GetIncompatibleCommands() (incompatibleActions []string) {

	incompatibleActions = []string{
		"Mute",
	}

	return
}
