package commandevaluators

/**
ASSUMPTIONS

a) there is only 1 DSP in a given room

b) microphones only have one port configuration and the DSP is the destination device

c) room-wide requests do not affect microphones

**/

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"

	ei "github.com/byuoitav/common/v2/events"
)

// SetVolumeDSP implements the CommandEvaluation struct.
type SetVolumeDSP struct{}

// Evaluate generates a list of actions based on the room information.
func (p *SetVolumeDSP) Evaluate(dbRoom structs.Room, room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	log.L.Info("[command_evaluators] Evaluating SetVolume command in DSP context...")

	eventInfo := ei.Event{
		Key:  "volume",
		User: requestor,
	}

	eventInfo.AddToTags(ei.CoreState, ei.UserGenerated)

	var actions []base.ActionStructure

	if room.Volume != nil {

		log.L.Info("[command_evaluators] Room-wide request detected")

		eventInfo.Value = strconv.Itoa(*room.Volume)

		actions, err := GetGeneralVolumeRequestActionsDSP(dbRoom, room, eventInfo)
		if err != nil {
			errorMessage := "[command_evaluators] Could not generate actions for room-wide \"SetVolume\" request: " + err.Error()
			log.L.Error(errorMessage)
			return []base.ActionStructure{}, 0, errors.New(errorMessage)
		}

		actions = append(actions, actions...)
	}

	if len(room.AudioDevices) > 0 {

		for _, audioDevice := range room.AudioDevices {

			if audioDevice.Volume != nil {

				eventInfo.Value = strconv.Itoa(*audioDevice.Volume)

				deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, audioDevice.Name)
				device := FindDevice(dbRoom.Devices, deviceID)

				if structs.HasRole(device, "Microphone") {

					action, err := GetMicVolumeAction(dbRoom, device, room, eventInfo, *audioDevice.Volume)
					if err != nil {
						return []base.ActionStructure{}, 0, err
					}

					actions = append(actions, action)

				} else if structs.HasRole(device, "DSP") {

					dspActions, err := GetDSPMediaVolumeAction(dbRoom, device, room, eventInfo, *audioDevice.Volume)
					if err != nil {
						return []base.ActionStructure{}, 0, err
					}

					actions = append(actions, dspActions...)

				} else if structs.HasRole(device, "AudioOut") {

					action, err := GetDisplayVolumeAction(device, room, eventInfo, *audioDevice.Volume)
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

								cmd := DX.GetCommandByID("SetVolumeDSP")
								if len(cmd.ID) < 1 {
									continue
								}

								log.L.Info("[command_evaluators] Adding mirror device %+v", DX.Name)

								action, err := GetDisplayVolumeAction(DX, room, eventInfo, *audioDevice.Volume)
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
					errorMessage := "[command_evaluators] Cannot set volume of device: " + device.Name + " in given context"
					log.L.Error(errorMessage)
					return []base.ActionStructure{}, 0, errors.New(errorMessage)
				}
			}
		}
	}

	log.L.Infof("[command_evaluators] %v actions generated.", len(actions))

	for _, a := range actions {
		log.L.Infof("[command_evaluators] %v, %v", a.Action, a.Parameters)

	}

	log.L.Info("[command_evaluators] Evaluation complete.")
	return actions, len(actions), nil
}

// Validate verifies that the action information is correct.
func (p *SetVolumeDSP) Validate(action base.ActionStructure) (err error) {
	maximum := 100
	minimum := 0

	level, err := strconv.Atoi(action.Parameters["level"])
	if err != nil {
		return err
	}

	if level > maximum || level < minimum {
		msg := fmt.Sprintf("[command_evaluators] ERROR. %v is an invalid volume level for %s", action.Parameters["level"], action.Device.Name)
		log.L.Error(msg)
		return errors.New(msg)
	}

	return
}

// GetIncompatibleCommands determines the commands from the room that are incompatible with this evaluator.
func (p *SetVolumeDSP) GetIncompatibleCommands() (incompatibleActions []string) {
	return nil
}

// GetGeneralVolumeRequestActionsDSP generates a list of actions based on the room and DSP info.
func GetGeneralVolumeRequestActionsDSP(dbRoom structs.Room, room base.PublicRoom, eventInfo ei.Event) ([]base.ActionStructure, error) {

	log.L.Info("[command_evaluators] Generating actions for room-wide \"SetVolume\" request")

	var actions []base.ActionStructure

	dsp := FilterDevicesByRole(dbRoom.Devices, "DSP")

	//verify that there is only one DSP
	if len(dsp) != 1 {
		errorMessage := "[command_evaluators] Invalid DSP configuration detected in room."
		log.L.Error(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	dspActions, err := GetDSPMediaVolumeAction(dbRoom, dsp[0], room, eventInfo, *room.Volume)
	if err != nil {
		errorMessage := "[command_evaluators] Could not generate action corresponding to general mute request in room " + room.Room + ", building " + room.Building + ": " + err.Error()
		log.L.Error(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	actions = append(actions, dspActions...)

	audioDevices := FilterDevicesByRole(dbRoom.Devices, "AudioOut")

	for _, device := range audioDevices {
		if structs.HasRole(device, "DSP") {
			continue
		}

		action, err := GetDisplayVolumeAction(device, room, eventInfo, *room.Volume)
		if err != nil {
			errorMessage := "[command_evaluators] Could not generate mute action for display " + device.Name + " in room " + room.Room + ", building " + room.Building + ": " + err.Error()
			log.L.Error(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, action)
	}

	return actions, nil
}

// GetMicVolumeAction generates an action based on the room, microphone and event information.
//we assume microphones are only connected to a DSP
//commands regarding microphones are only issued to DSP
func GetMicVolumeAction(dbRoom structs.Room, mic structs.Device, room base.PublicRoom, eventInfo ei.Event, volume int) (base.ActionStructure, error) {

	log.L.Info("[command_evaluators] Identified microphone volume request")

	destination := base.DestinationDevice{
		Device:      mic,
		AudioDevice: true,
	}

	//get DSP
	dsps := FilterDevicesByRole(dbRoom.Devices, "DSP")

	//verify that there is only one DSP
	if len(dsps) != 1 {
		errorMessage := "[command_evaluators] Invalid DSP configuration detected in room."
		log.L.Error(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	dsp := dsps[0]
	parameters := make(map[string]string)

	if volume < 0 || volume > 100 {
		errorMessage := "[command_evaluators] Invalid volume parameter: " + strconv.Itoa(volume)
		log.L.Error(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	for _, port := range dsp.Ports {

		if port.SourceDevice == mic.ID {

			eventInfo.AffectedRoom = ei.BasicRoomInfo{
				BuildingID: room.Building,
				RoomID:     fmt.Sprintf("%s-%s", room.Building, room.Room),
			}

			eventInfo.Value = strconv.Itoa(volume)
			deviceInfo := strings.Split(mic.ID, "-")

			eventInfo.TargetDevice = ei.BasicDeviceInfo{
				BasicRoomInfo: ei.BasicRoomInfo{
					BuildingID: deviceInfo[0],
					RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
				},
				DeviceID: mic.ID,
			}

			parameters["level"] = strconv.Itoa(volume)
			parameters["input"] = port.ID

			return base.ActionStructure{
				Action:              "SetVolume",
				GeneratingEvaluator: "SetVolumeDSP",
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

// GetDSPMediaVolumeAction generates a list of actions based on the room, DSP, and event information.
func GetDSPMediaVolumeAction(dbRoom structs.Room, dsp structs.Device, room base.PublicRoom, eventInfo ei.Event, volume int) ([]base.ActionStructure, error) { //commands are issued to whatever port doesn't have a mic connected
	log.L.Infof("[command_evaluators] %v", volume)

	log.L.Info("[command_evaluators] Generating action for command SetVolume on media routed through DSP")

	var output []base.ActionStructure

	for _, port := range dsp.Ports {
		parameters := make(map[string]string)
		parameters["level"] = fmt.Sprintf("%v", volume)

		eventInfo.Value = fmt.Sprintf("%v", volume)

		eventInfo.AffectedRoom = ei.BasicRoomInfo{
			BuildingID: room.Building,
			RoomID:     fmt.Sprintf("%s-%s", room.Building, room.Room),
		}

		eventInfo.Value = strconv.Itoa(volume)
		deviceInfo := strings.Split(dsp.ID, "-")

		eventInfo.TargetDevice = ei.BasicDeviceInfo{
			BasicRoomInfo: ei.BasicRoomInfo{
				BuildingID: deviceInfo[0],
				RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
			},
			DeviceID: dsp.ID,
		}

		deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, port.SourceDevice)
		sourceDevice := FindDevice(dbRoom.Devices, deviceID)

		if !(structs.HasRole(sourceDevice, "Microphone")) {

			destination := base.DestinationDevice{
				Device:      dsp,
				AudioDevice: true,
			}

			parameters["input"] = port.ID
			action := base.ActionStructure{
				Action:              "SetVolume",
				GeneratingEvaluator: "SetVolumeDSP",
				Device:              dsp,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []ei.Event{eventInfo},
				Parameters:          parameters,
			}

			output = append(output, action)
		}
	}

	return output, nil

}

// GetDisplayVolumeAction generates an action based on the room, display and event information.
func GetDisplayVolumeAction(device structs.Device, room base.PublicRoom, eventInfo ei.Event, volume int) (base.ActionStructure, error) { //commands are issued to devices, e.g. they aren't connected to the DSP

	log.L.Infof("[command_evaluators] Generating action for SetVolume on device %s external to DSP", device.Name)

	parameters := make(map[string]string)

	destination := base.DestinationDevice{
		Device:      device,
		AudioDevice: true,
	}

	if structs.HasRole(device, "VideoOut") {
		destination.Display = true
	}

	eventInfo.Value = strconv.Itoa(volume)

	eventInfo.AffectedRoom = ei.BasicRoomInfo{
		BuildingID: room.Building,
		RoomID:     fmt.Sprintf("%s-%s", room.Building, room.Room),
	}

	eventInfo.Value = strconv.Itoa(volume)
	deviceInfo := strings.Split(device.ID, "-")

	eventInfo.TargetDevice = ei.BasicDeviceInfo{
		BasicRoomInfo: ei.BasicRoomInfo{
			BuildingID: deviceInfo[0],
			RoomID:     fmt.Sprintf("%s-%s", deviceInfo[0], deviceInfo[1]),
		},
		DeviceID: device.ID,
	}

	parameters["level"] = strconv.Itoa(volume)

	action := base.ActionStructure{
		Action:              "SetVolume",
		GeneratingEvaluator: "SetVolumeDSP",
		Device:              device,
		DestinationDevice:   destination,
		DeviceSpecific:      true,
		EventLog:            []ei.Event{eventInfo},
		Parameters:          parameters,
	}

	return action, nil
}
