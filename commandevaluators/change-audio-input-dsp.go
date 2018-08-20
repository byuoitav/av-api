package commandevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	ei "github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
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
func (p *ChangeAudioInputDSP) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	log.L.Info("[command_evaluators] Evaluating PUT body for \"ChangeInput\" command in an audio DSP context...")

	var actions []base.ActionStructure

	eventInfo := ei.EventInfo{
		Type:         ei.CORESTATE,
		EventCause:   ei.USERINPUT,
		EventInfoKey: "input",
		Requestor:    requestor,
	}

	destination := base.DestinationDevice{
		AudioDevice: true,
	}

	if len(room.CurrentAudioInput) > 0 { //

		generalAction, err := GetDSPMediaInputAction(room, eventInfo, room.CurrentAudioInput, false, destination)
		if err != nil {
			errorMessage := "[command_evaluators] Could not generate actions for room-wide \"ChangeInput\" request: " + err.Error()
			log.L.Error(errorMessage)
			return []base.ActionStructure{}, 0, errors.New(errorMessage)
		}

		actions = append(actions, generalAction)

		roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
		devices, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "AudioOut")
		if err != nil {
			errorMessage := "[command_evaluators] Could not generate actions for room-wide \"ChangeInput\" request: " + err.Error()
			log.L.Error(errorMessage)
			return []base.ActionStructure{}, 0, errors.New(errorMessage)
		}

		for _, device := range devices {

			if device.Type.Output && !structs.HasRole(device, "Microphone") {

				log.L.Infof("[command_evaluators] Adding device %+v", device.Name)

				eventInfo.Device = device.Name
				eventInfo.DeviceID = device.ID

				actions = append(actions, base.ActionStructure{
					Action:              "Mute",
					GeneratingEvaluator: "ChangeAudioInputDSP",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []ei.EventInfo{eventInfo},
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

							log.L.Infof("[command_evaluators] Adding device %+v", DX.Name)

							eventInfo.Device = DX.Name
							eventInfo.DeviceID = DX.ID

							actions = append(actions, base.ActionStructure{
								Action:              "Mute",
								GeneratingEvaluator: "ChangeAudioInputDSP",
								Device:              DX,
								DestinationDevice:   destination,
								DeviceSpecific:      false,
								EventLog:            []ei.EventInfo{eventInfo},
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
				device, err := db.GetDB().GetDevice(deviceID)
				if err != nil {
					errorMessage := "[command_evaluators] Could not get device: " + audioDevice.Name + " from database: " + err.Error()
					log.L.Error(errorMessage)
					return []base.ActionStructure{}, 0, errors.New(errorMessage)
				}

				if structs.HasRole(device, "DSP") {

					dspAction, err := GetDSPMediaInputAction(room, eventInfo, room.AudioDevices[0].Input, true, destination)
					if err != nil {
						errorMessage := "[command_evaluators] Could not generate actions for specific \"ChangeInput\" requests: " + err.Error()
						log.L.Error(errorMessage)
						return []base.ActionStructure{}, 0, errors.New(errorMessage)
					}

					actions = append(actions, dspAction)

				} else if structs.HasRole(device, "AudioOut") && !structs.HasRole(device, "Microphone") {

					mediaAction, err := generateChangeInputByDevice(audioDevice.Device, room.Room, room.Building, "ChangeAudioInputDefault", requestor)
					if err != nil {
						msg := fmt.Sprintf("[command_evaluators] Unable to generate actions corresponding to \"ChangeInput\" request against device: %s: %s", device.Name, err.Error())
						log.L.Error(msg)
						return []base.ActionStructure{}, 0, errors.New(msg)
					}
					actions = append(actions, mediaAction)
				}
			}
		}
	}

	log.L.Infof("[commandevaluators] Evaluation complete: %s actions generated.", len(actions))

	return actions, len(actions), nil
}

// GetDSPMediaInputAction determines the devices affected and actions needed for this command.
func GetDSPMediaInputAction(room base.PublicRoom, eventInfo ei.EventInfo, input string, deviceSpecific bool, destination base.DestinationDevice) (base.ActionStructure, error) {

	//get DSP
	roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
	dsp, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "DSP")
	if err != nil {
		errorMessage := "[command_evaluators] Problem getting device " + input + " from database " + err.Error()
		log.L.Info(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//validate number of DSPs
	if len(dsp) != 1 {
		errorMessage := "[command_evaluators] Invalid DSP configuration detected in room"
		log.L.Info(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//get switcher
	switchers, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "VideoSwitcher")
	if err != nil {
		errorMessage := "[command_evaluators] Could not get room switch in room " + room.Room + ", building " + room.Building + ": " + err.Error()
		log.L.Info(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//validate number of switchers
	if len(switchers) != 1 {
		errorMessage := "[command_evaluators] Invalid video switch configuration detected in room"
		log.L.Info(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//get requested device
	deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, input)
	device, err := db.GetDB().GetDevice(deviceID)
	if err != nil {
		errorMessage := "[command_evaluators] Problem getting device " + input + " from database " + err.Error()
		log.L.Info(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//find the port where the host is the switcher and the destination is the DSP
	for _, port := range device.Ports {

		if port.DestinationDevice == dsp[0].ID {
			//once we find the port, send the command to the switcher

			switcherPorts := strings.Split(port.ID, ":")
			if len(switcherPorts) != 2 {
				return base.ActionStructure{}, errors.New("[command_evaluators] Invalid video switcher port")
			}

			parameters := make(map[string]string)
			parameters["input"] = switcherPorts[0]
			parameters["output"] = switcherPorts[1]

			eventInfo.Device = switchers[0].Name
			eventInfo.DeviceID = switchers[0].ID
			eventInfo.EventInfoValue = input

			destination.Device = device

			return base.ActionStructure{
				Action:              "ChangeInput",
				GeneratingEvaluator: "ChangeAudioInputDSP",
				Device:              switchers[0],
				DestinationDevice:   destination,
				DeviceSpecific:      deviceSpecific,
				Parameters:          parameters,
				EventLog:            []ei.EventInfo{eventInfo},
			}, nil

		}

	}

	return base.ActionStructure{}, errors.New("[command_evaluators] No port found for given input")

}
