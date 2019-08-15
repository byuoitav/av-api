package commandevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	ei "github.com/byuoitav/common/v2/events"
)

/**
ASSUMPTIONS:

a) there is only 1 DSP in a given room

b) there is only 1 video switcher in a given room

c) the switcher has access to all the media audio

d) a room-wide audio input request implies sending a command to the DSP and muting all devices designatied as 'AudioOut'

e) microphones are not affected by actions generated in this command evaluator

**/

// ChangeAudioInputDSP implements the CommandEvaluation struct.
type ChangeAudioInputDSP struct{}

// Evaluate verifies the information for a ChangeAudioInputDSP object and generates action based on the command.
func (p *ChangeAudioInputDSP) Evaluate(dbRoom structs.Room, room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	log.L.Info("[command_evaluators] Evaluating PUT body for \"ChangeInput\" command in an audio DSP context...")

	var actions []base.ActionStructure

	e := ei.Event{
		Key:  "input",
		User: requestor,
	}

	e.EventTags = append(e.EventTags, ei.CoreState, ei.UserGenerated)

	destination := base.DestinationDevice{
		AudioDevice: true,
	}

	if len(room.CurrentAudioInput) > 0 { //

		generalAction, err := GetDSPMediaInputAction(dbRoom, room, e, room.CurrentAudioInput, false, destination)
		if err != nil {
			errorMessage := "[command_evaluators] Could not generate actions for room-wide \"ChangeInput\" request: " + err.Error()
			log.L.Error(errorMessage)
			return []base.ActionStructure{}, 0, errors.New(errorMessage)
		}

		actions = append(actions, generalAction)

		devices := FilterDevicesByRole(dbRoom.Devices, "AudioOut")

		for _, device := range devices {

			if device.Type.Output && !structs.HasRole(device, "Microphone") {

				log.L.Infof("[command_evaluators] Adding device %+v", device.Name)

				deviceInfo := strings.Split(device.ID, "-")
				e.TargetDevice = ei.BasicDeviceInfo{
					BasicRoomInfo: ei.BasicRoomInfo{
						BuildingID: deviceInfo[0],
						RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
					},
					DeviceID: device.ID,
				}

				actions = append(actions, base.ActionStructure{
					Action:              "Mute",
					GeneratingEvaluator: "ChangeAudioInputDSP",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []ei.Event{e},
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

							cmd := DX.GetCommandByID("ChangeAudioInputDSP")
							if len(cmd.ID) < 1 {
								continue
							}

							log.L.Infof("[command_evaluators] Adding device %+v", DX.Name)

							deviceInfo := strings.Split(DX.ID, "-")
							e.TargetDevice = ei.BasicDeviceInfo{
								BasicRoomInfo: ei.BasicRoomInfo{
									BuildingID: deviceInfo[0],
									RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
								},
								DeviceID: device.ID,
							}

							actions = append(actions, base.ActionStructure{
								Action:              "Mute",
								GeneratingEvaluator: "ChangeAudioInputDSP",
								Device:              DX,
								DestinationDevice:   destination,
								DeviceSpecific:      false,
								EventLog:            []ei.Event{e},
							})
						}
					}
				}
				///// MIRROR STUFF /////
				////////////////////////
			}

		}

	}

	//TODO will this be a problem if the slice is nil?
	if len(room.AudioDevices) > 0 {

		for _, audioDevice := range room.AudioDevices {

			if len(audioDevice.Input) > 0 {

				deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, audioDevice.Name)
				device := FindDevice(dbRoom.Devices, deviceID)

				if structs.HasRole(device, "DSP") {

					dspAction, err := GetDSPMediaInputAction(dbRoom, room, e, room.AudioDevices[0].Input, true, destination)
					if err != nil {
						errorMessage := "[command_evaluators] Could not generate actions for specific \"ChangeInput\" requests: " + err.Error()
						log.L.Error(errorMessage)
						return []base.ActionStructure{}, 0, errors.New(errorMessage)
					}

					actions = append(actions, dspAction)

				} else if structs.HasRole(device, "AudioOut") && !structs.HasRole(device, "Microphone") {

					mediaAction, err := generateChangeInputByDevice(dbRoom, audioDevice.Device, room.Room, room.Building, "ChangeAudioInputDefault", requestor)
					if err != nil {
						msg := fmt.Sprintf("[command_evaluators] Unable to generate actions corresponding to \"ChangeInput\" request against device: %s: %s", device.Name, err.Error())
						log.L.Error(msg)
						return []base.ActionStructure{}, 0, errors.New(msg)
					}
					actions = append(actions, mediaAction...)
				}
			}
		}
	}

	log.L.Infof("[commandevaluators] Evaluation complete: %s actions generated.", len(actions))

	return actions, len(actions), nil
}

// GetDSPMediaInputAction determines the devices affected and actions needed for this command.
func GetDSPMediaInputAction(dbRoom structs.Room, room base.PublicRoom, event ei.Event, input string, deviceSpecific bool, destination base.DestinationDevice) (base.ActionStructure, error) {

	//get DSP
	dsps := FilterDevicesByRole(dbRoom.Devices, "DSP")

	//validate number of DSPs
	if len(dsps) != 1 {
		errorMessage := "[command_evaluators] Invalid DSP configuration detected in room"
		log.L.Info(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//get switcher
	switchers := FilterDevicesByRole(dbRoom.Devices, "VideoSwitcher")

	//validate number of switchers
	if len(switchers) != 1 {
		errorMessage := "[command_evaluators] Invalid video switch configuration detected in room"
		log.L.Info(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//get requested device
	deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, input)
	device := FindDevice(dbRoom.Devices, deviceID)

	//find the port where the host is the switcher and the destination is the DSP
	for _, port := range device.Ports {

		if port.DestinationDevice == dsps[0].ID {
			//once we find the port, send the command to the switcher

			switcherPorts := strings.Split(port.ID, ":")
			if len(switcherPorts) != 2 {
				return base.ActionStructure{}, errors.New("[command_evaluators] Invalid video switcher port")
			}

			parameters := make(map[string]string)
			parameters["input"] = switcherPorts[0]
			parameters["output"] = switcherPorts[1]

			deviceInfo := strings.Split(device.ID, "-")
			event.TargetDevice = ei.BasicDeviceInfo{
				BasicRoomInfo: ei.BasicRoomInfo{
					BuildingID: deviceInfo[0],
					RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
				},
				DeviceID: device.ID,
			}

			event.Value = input

			destination.Device = device

			return base.ActionStructure{
				Action:              "ChangeInput",
				GeneratingEvaluator: "ChangeAudioInputDSP",
				Device:              switchers[0],
				DestinationDevice:   destination,
				DeviceSpecific:      deviceSpecific,
				Parameters:          parameters,
				EventLog:            []ei.Event{event},
			}, nil

		}

	}

	return base.ActionStructure{}, errors.New("[command_evaluators] No port found for given input")

}
