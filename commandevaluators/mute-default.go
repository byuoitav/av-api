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
)

//MuteDefault implements the CommandEvaluation struct.
type MuteDefault struct {
}

/*Evaluate takes a public room struct, scans the struct and builds any needed
actions based on the contents of the struct.*/
func (p *MuteDefault) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	log.L.Info("[command_evaluators] Evaluating for Mute command.")

	var actions []base.ActionStructure

	destination := base.DestinationDevice{
		AudioDevice: true,
	}

	eventInfo := events.EventInfo{
		Type:           events.CORESTATE,
		EventCause:     events.USERINPUT,
		EventInfoKey:   "muted",
		EventInfoValue: "true",
		Requestor:      requestor,
	}

	if room.Muted != nil && *room.Muted {

		log.L.Info("[command_evaluators] Room-wide Mute request recieved. Retrieving all devices.")

		roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
		devices, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "AudioOut")
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		log.L.Info("[command_evaluators] Muting all devices in room.")

		for _, device := range devices {
			if device.Type.Output {
				log.L.Infof("[command_evaluators] Adding device %+v", device.Name)

				eventInfo.Device = device.Name
				eventInfo.DeviceID = device.ID
				destination.Device = device

				if structs.HasRole(device, "VideoOut") {
					destination.Display = true
				}

				actions = append(actions, base.ActionStructure{
					Action:              "Mute",
					GeneratingEvaluator: "MuteDefault",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []events.EventInfo{eventInfo},
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

							cmd := DX.GetCommandByName("MuteDefault")
							if len(cmd.ID) < 1 {
								return actions, len(actions), nil
							}

							log.L.Info("[command_evaluators] Adding mirror device %+v", DX.Name)

							eventInfo.Device = DX.Name
							eventInfo.DeviceID = DX.ID

							actions = append(actions, base.ActionStructure{
								Action:              "Mute",
								GeneratingEvaluator: "MuteDefault",
								Device:              DX,
								DestinationDevice:   destination,
								DeviceSpecific:      false,
								EventLog:            []events.EventInfo{eventInfo},
							})
						}
					}
				}
				///// MIRROR STUFF /////
				////////////////////////
			}
		}
	}

	//scan the room struct
	log.L.Info("[command_evaluators] Evaluating audio devices for Mute command.")

	//generate commands
	for _, audioDevice := range room.AudioDevices {
		if audioDevice.Muted != nil && *audioDevice.Muted {

			//get the device
			deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, audioDevice.Name)
			device, err := db.GetDB().GetDevice(deviceID)
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
				EventLog:            []events.EventInfo{eventInfo},
			})

			////////////////////////
			///// MIRROR STUFF /////
			if structs.HasRole(device, "MirrorMaster") {
				for _, port := range device.Ports {
					if port.ID == "mirror" {
						DX, err := db.GetDB().GetDevice(port.DestinationDevice)
						if err != nil {
							return []base.ActionStructure{}, 0, err
						}

						cmd := DX.GetCommandByName("MuteDefault")
						if len(cmd.ID) < 1 {
							return actions, len(actions), nil
						}

						log.L.Info("[command_evaluators] Adding mirror device %+v", DX.Name)

						eventInfo.Device = DX.Name
						eventInfo.DeviceID = DX.ID

						actions = append(actions, base.ActionStructure{
							Action:              "Mute",
							GeneratingEvaluator: "MuteDefault",
							Device:              DX,
							DestinationDevice:   destination,
							DeviceSpecific:      true,
							EventLog:            []events.EventInfo{eventInfo},
						})
					}
				}
			}
			///// MIRROR STUFF /////
			////////////////////////
		}

	}

	return actions, len(actions), nil

}

// Validate takes an ActionStructure and determines if the command and parameter are valid for the device specified
func (p *MuteDefault) Validate(action base.ActionStructure) error {

	log.L.Info("[command_evaluators] Validating for command \"Mute\".")

	ok, _ := CheckCommands(action.Device.Type.Commands, "Mute")

	// fmt.Printf("action.Device.Commands contains: %+v\n", action.Device.Commands)
	log.L.Infof("[command_evaluators] Device ID: %v\n", action.Device.ID)
	log.L.Infof("[command_evaluators] CheckCommands returns: %v\n", ok)

	if !ok || !strings.EqualFold(action.Action, "Mute") {
		msg := fmt.Sprintf("[command_evaluators] ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		log.L.Error(msg)
		return errors.New(msg)
	}

	log.L.Info("[command_evaluators] Done.")

	return nil
}

// GetIncompatibleCommands returns a list of commands that are incompatabl with this one (i.e. 'standby' and 'power on', or 'mute' and 'volume up')
func (p *MuteDefault) GetIncompatibleCommands() (incompatibleActions []string) {

	incompatibleActions = []string{
		"UnMute",
	}

	return
}
