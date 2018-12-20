package commandevaluators

/**
ASSUMPTIONS:

a) there is only 1 DSP in a given room

b) microphones only have one port configuration and the DSP is the destination device

c) room-wide requests do not affect microphones

**/

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"
	ei "github.com/byuoitav/common/v2/events"
)

// MuteDSP implements the CommandEvaluation struct.
type MuteDSP struct{}

// Evaluate takes the information given and generates a list of actions.
func (p *MuteDSP) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	log.L.Info("[command_evaluators] Evaluating PUT body for \"Mute\" command in DSP context...")

	var actions []base.ActionStructure

	// eventInfo := ei.EventInfo{
	// 	Type:           ei.CORESTATE,
	// 	EventCause:     ei.USERINPUT,
	// 	EventInfoKey:   "muted",
	// 	EventInfoValue: "true",
	// 	Requestor:      requestor,
	// }

	eventInfo := ei.Event{
		Key:   "muted",
		Value: "true",
		User:  requestor,
	}

	eventInfo.AddToTags(ei.CoreState, ei.UserGenerated)

	destination := base.DestinationDevice{
		AudioDevice: true,
	}

	if room.Muted != nil && *room.Muted {

		generalActions, err := GetGeneralMuteRequestActionsDSP(room, eventInfo, destination)
		if err != nil {
			errorMessage := "[command_evaluators] Could not generate actions for room-wide \"Mute\" request: " + err.Error()
			log.L.Error(errorMessage)
			return []base.ActionStructure{}, 0, errors.New(errorMessage)
		}

		actions = append(actions, generalActions...)
	}

	if len(room.AudioDevices) > 0 {

		for _, audioDevice := range room.AudioDevices {

			if audioDevice.Muted == nil || !(*audioDevice.Muted) {
				continue
			}

			deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, audioDevice.Name)
			device, err := db.GetDB().GetDevice(deviceID)
			if err != nil {
				log.L.Errorf("[command_evaluators] Error getting device %s from database: %s", audioDevice.Name, err.Error())
			}

			destination.Device = device //if we've made it this far, the destination device is this audio device

			if structs.HasRole(device, "Microphone") {

				action, err := GetMicMuteAction(device, room, eventInfo)
				if err != nil {
					return []base.ActionStructure{}, 0, err
				}

				actions = append(actions, action)

			} else if structs.HasRole(device, "DSP") {

				dspActions, err := GetDSPMediaMuteAction(device, room, eventInfo, true)
				if err != nil {
					return []base.ActionStructure{}, 0, err
				}

				actions = append(actions, dspActions...)

			} else if structs.HasRole(device, "AudioOut") {

				action, err := GetDisplayMuteAction(device, room, eventInfo, true)
				if err != nil {
					return []base.ActionStructure{}, 0, err
				}

				actions = append(actions, action)

				////////////////////////
				///// MIRROR STUFF /////
				if structs.HasRole(device, "MirrorMaster") {
					for _, port := range device.Ports {
						if port.ID == "mirror" {
							DX, err := db.GetDB().GetDevice(port.DestinationDevice)
							if err != nil {
								return actions, len(actions), err
							}

							cmd := DX.GetCommandByID("MuteDSP")
							if len(cmd.ID) < 1 {
								continue
							}

							log.L.Info("[command_evaluators] Adding mirror device %+v", DX.Name)

							action, err := GetDisplayMuteAction(DX, room, eventInfo, true)
							if err != nil {
								return actions, len(actions), err
							}

							actions = append(actions, action)
						}
					}
				}
				///// MIRROR STUFF /////
				////////////////////////

			} else { //bad device
				errorMessage := "[command_evaluators] Cannot set volume of device " + device.Name
				log.L.Error(errorMessage)
				return []base.ActionStructure{}, 0, errors.New(errorMessage)
			}
		}
	}

	log.L.Info("[command_evaluators] %s actions generated.", len(actions))
	log.L.Info("[command_evaluators] Evaluation complete.")

	return actions, len(actions), nil

}

// Validate verifies that the action given has correct information.
func (p *MuteDSP) Validate(base.ActionStructure) error {
	//TODO make sure the device actually can be muted
	return nil
}

// GetIncompatibleCommands returns the list of commands that are incompatible with this device.
func (p *MuteDSP) GetIncompatibleCommands() []string {
	return nil
}

// GetGeneralMuteRequestActionsDSP assumes only one DSP, but allows for the possiblity of multiple devices not routed through the DSP
//room-wide mute requests DO NOT include mics
func GetGeneralMuteRequestActionsDSP(room base.PublicRoom, eventInfo ei.Event, destination base.DestinationDevice) ([]base.ActionStructure, error) {

	log.L.Info("[command_evaluators] Generating actions for room-wide \"Mute\" request")

	var actions []base.ActionStructure

	roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
	dsp, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "DSP")
	if err != nil {
		msg := fmt.Sprintf("[command_evaluators] Error getting devices %s", err.Error())
		log.L.Error(msg)
		return []base.ActionStructure{}, err
	}

	if len(dsp) != 1 {
		errorMessage := "[command_evaluators] Invalid number of DSP devices found in room: " + room.Room + " in building " + room.Building
		log.L.Error(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	dspActions, err := GetDSPMediaMuteAction(dsp[0], room, eventInfo, false)
	if err != nil {
		errorMessage := "[command_evaluators] Could not generate action corresponding to general mute request in room " + room.Room + ", building " + room.Building + ": " + err.Error()
		log.L.Error(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	actions = append(actions, dspActions...)

	audioDevices, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "AudioOut")
	if err != nil {
		log.L.Errorf("[command_evaluators] Error getting devices %s", err.Error())
		return []base.ActionStructure{}, err
	}

	for _, device := range audioDevices {

		action, err := GetDisplayMuteAction(device, room, eventInfo, false)
		if err != nil {
			errorMessage := "[command_evaluators] Could not generate mute action for display " + device.Name + " in room " + room.Room + ", building " + room.Building + ": " + err.Error()
			log.L.Error(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, action)
	}

	return actions, nil
}

// GetMicMuteAction takes the room information and a microphone and generates an action.
//assumes the mic is only connected to a single DSP
func GetMicMuteAction(mic structs.Device, room base.PublicRoom, eventInfo ei.Event) (base.ActionStructure, error) {

	log.L.Infof("[command_evaluators] Generating action for command \"Mute\" on microphone %s", mic.Name)

	destination := base.DestinationDevice{
		Device:      mic,
		AudioDevice: true,
	}

	//get DSP
	roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
	dsps, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "DSP")
	if err != nil {
		errorMessage := "[command_evaluators] Error getting DSP configuration in building " + room.Building + ", room " + room.Room + ": " + err.Error()
		log.L.Error(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//verify DSP configuration
	if len(dsps) != 1 {
		errorMessage := "[command_evaluators] Invalid DSP configuration detected."
		log.L.Error(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	dsp := dsps[0]

	parameters := make(map[string]string)
	for _, port := range dsp.Ports {

		if port.SourceDevice == mic.ID {

			parameters["input"] = port.ID

			deviceInfo := strings.Split(mic.ID, "-")

			eventInfo.TargetDevice = ei.BasicDeviceInfo{
				BasicRoomInfo: ei.BasicRoomInfo{
					BuildingID: deviceInfo[0],
					RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
				},
				DeviceID: mic.ID,
			}

			return base.ActionStructure{
				Action:              "Mute",
				GeneratingEvaluator: "MuteDSP",
				Device:              dsp,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []ei.Event{eventInfo},
				Parameters:          parameters,
			}, nil
		}
	}

	return base.ActionStructure{}, errors.New("[command_evaluators] Could not find port for mic " + mic.Name)
}

// GetDSPMediaMuteAction generates a list of actions based on information about the room and the DSP.
func GetDSPMediaMuteAction(dsp structs.Device, room base.PublicRoom, eventInfo ei.Event, deviceSpecific bool) ([]base.ActionStructure, error) {

	log.L.Info("[command_evaluators] Generating action for command Mute on media routed through DSP")

	var output []base.ActionStructure
	// eventInfo.Device = dsp.Name
	// eventInfo.DeviceID = dsp.ID

	deviceInfo := strings.Split(dsp.ID, "-")

	eventInfo.TargetDevice = ei.BasicDeviceInfo{
		BasicRoomInfo: ei.BasicRoomInfo{
			BuildingID: deviceInfo[0],
			RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
		},
		DeviceID: dsp.ID,
	}

	for _, port := range dsp.Ports {
		parameters := make(map[string]string)

		deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, port.SourceDevice)
		sourceDevice, err := db.GetDB().GetDevice(deviceID)
		if err != nil {
			errorMessage := "Could not get device " + port.SourceDevice + " from database " + err.Error()
			log.L.Error(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		if !structs.HasRole(sourceDevice, "Microphone") {

			destination := base.DestinationDevice{
				Device:      dsp,
				AudioDevice: true,
			}

			parameters["input"] = port.ID
			action := base.ActionStructure{
				Action:              "Mute",
				GeneratingEvaluator: "MuteDSP",
				Device:              dsp,
				DestinationDevice:   destination,
				DeviceSpecific:      deviceSpecific,
				EventLog:            []ei.Event{eventInfo},
				Parameters:          parameters,
			}

			output = append(output, action)
		}
	}

	return output, nil
}

// GetDisplayMuteAction generates an action based on the information about the room and display.
func GetDisplayMuteAction(device structs.Device, room base.PublicRoom, eventInfo ei.Event, deviceSpecific bool) (base.ActionStructure, error) {

	log.L.Infof("Generating action for command \"Mute\" for device %s external to DSP", device.Name)

	deviceInfo := strings.Split(device.ID, "-")

	eventInfo.TargetDevice = ei.BasicDeviceInfo{
		BasicRoomInfo: ei.BasicRoomInfo{
			BuildingID: deviceInfo[0],
			RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
		},
		DeviceID: device.ID,
	}

	destination := base.DestinationDevice{
		Device:      device,
		AudioDevice: true,
	}

	if structs.HasRole(device, "VideoOut") {
		destination.Display = true
	}

	return base.ActionStructure{
		Action:              "Mute",
		GeneratingEvaluator: "MuteDSP",
		Device:              device,
		DestinationDevice:   destination,
		DeviceSpecific:      deviceSpecific,
		EventLog:            []ei.Event{eventInfo},
	}, nil
}
