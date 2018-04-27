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

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"

	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

type SetVolumeDSP struct{}

func (p *SetVolumeDSP) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	base.Log("Evaluating SetVolume command in DSP context...")

	eventInfo := ei.EventInfo{
		Type:         ei.CORESTATE,
		EventCause:   ei.USERINPUT,
		EventInfoKey: "volume",
		Requestor:    requestor,
	}

	var actions []base.ActionStructure

	if room.Volume != nil {

		base.Log("Room-wide request detected")

		eventInfo.EventInfoValue = strconv.Itoa(*room.Volume)

		actions, err := GetGeneralVolumeRequestActionsDSP(room, eventInfo)
		if err != nil {
			errorMessage := "Could not generate actions for room-wide \"SetVolume\" request: " + err.Error()
			base.Log(errorMessage)
			return []base.ActionStructure{}, 0, errors.New(errorMessage)
		}

		actions = append(actions, actions...)
	}

	if len(room.AudioDevices) > 0 {

		for _, audioDevice := range room.AudioDevices {

			if audioDevice.Volume != nil {

				eventInfo.EventInfoValue = strconv.Itoa(*audioDevice.Volume)

				device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
				if err != nil {
					base.Log("Error getting device %s from database: %s", audioDevice.Name, err.Error())
				}

				if device.HasRole("Microphone") {

					action, err := GetMicVolumeAction(device, room, eventInfo, *audioDevice.Volume)
					if err != nil {
						return []base.ActionStructure{}, 0, err
					}

					actions = append(actions, action)

				} else if device.HasRole("DSP") {

					dspActions, err := GetDSPMediaVolumeAction(device, room, eventInfo, *audioDevice.Volume)
					if err != nil {
						return []base.ActionStructure{}, 0, err
					}

					actions = append(actions, dspActions...)

				} else if device.HasRole("AudioOut") {

					action, err := GetDisplayVolumeAction(device, room, eventInfo, *audioDevice.Volume)
					if err != nil {
						return []base.ActionStructure{}, 0, err
					}

					actions = append(actions, action)

				} else { //bad device
					errorMessage := "Cannot set volume of device: " + device.Name + " in given context"
					base.Log(errorMessage)
					return []base.ActionStructure{}, 0, errors.New(errorMessage)
				}
			}
		}
	}

	base.Log("%v actions generated.", len(actions))

	for _, a := range actions {
		base.Log("%v, %v", a.Action, a.Parameters)

	}

	base.Log("Evaluation complete.")
	return actions, len(actions), nil
}

func (p *SetVolumeDSP) Validate(action base.ActionStructure) (err error) {
	maximum := 100
	minimum := 0

	level, err := strconv.Atoi(action.Parameters["level"])
	if err != nil {
		return err
	}

	if level > maximum || level < minimum {
		base.Log("ERROR. %v is an invalid volume level for %s", action.Parameters["level"], action.Device.Name)
		return errors.New(action.Action + " is an invalid command for " + action.Device.Name)
	}

	return
}

func (p *SetVolumeDSP) GetIncompatibleCommands() (incompatibleActions []string) {
	return nil
}

func GetGeneralVolumeRequestActionsDSP(room base.PublicRoom, eventInfo ei.EventInfo) ([]base.ActionStructure, error) {

	base.Log("Generating actions for room-wide \"SetVolume\" request")

	var actions []base.ActionStructure

	dsp, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "DSP")
	if err != nil {
		base.Log("Error getting devices %s", err.Error)
		return []base.ActionStructure{}, err
	}

	//verify that there is only one DSP
	if len(dsp) != 1 {
		errorMessage := "Invalid DSP configuration detected in room."
		base.Log(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	dspActions, err := GetDSPMediaVolumeAction(dsp[0], room, eventInfo, *room.Volume)
	if err != nil {
		errorMessage := "Could not generate action corresponding to general mute request in room " + room.Room + ", building " + room.Building + ": " + err.Error()
		base.Log(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	actions = append(actions, dspActions...)

	audioDevices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
	if err != nil {
		base.Log("Error getting devices %s", err.Error())
		return []base.ActionStructure{}, err
	}

	for _, device := range audioDevices {
		if device.HasRole("DSP") {
			continue
		}

		action, err := GetDisplayVolumeAction(device, room, eventInfo, *room.Volume)
		if err != nil {
			errorMessage := "Could not generate mute action for display " + device.Name + " in room " + room.Room + ", building " + room.Building + ": " + err.Error()
			base.Log(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, action)
	}

	return actions, nil
}

//we assume microphones are only connected to a DSP
//commands regarding microphones are only issued to DSP
func GetMicVolumeAction(mic structs.Device, room base.PublicRoom, eventInfo ei.EventInfo, volume int) (base.ActionStructure, error) {

	base.Log("Identified microphone volume request")

	destination := base.DestinationDevice{
		Device:      mic,
		AudioDevice: true,
	}

	//get DSP
	dsps, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "DSP")
	if err != nil {
		errorMessage := "Error getting DSP configuration for building " + room.Building + ", room " + room.Room + ": " + err.Error()
		base.Log(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//verify that there is only one DSP
	if len(dsps) != 1 {
		errorMessage := "Invalid DSP configuration detected in room."
		base.Log(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	dsp := dsps[0]
	parameters := make(map[string]string)

	if volume < 0 || volume > 100 {
		errorMessage := "Invalid volume parameter: " + strconv.Itoa(volume)
		base.Log(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	for _, port := range dsp.Ports {

		if port.Source == mic.Name {

			eventInfo.EventInfoValue = strconv.Itoa(volume)
			eventInfo.Device = mic.Name
			parameters["level"] = strconv.Itoa(volume)
			parameters["input"] = port.Name

			return base.ActionStructure{
				Action:              "SetVolume",
				GeneratingEvaluator: "SetVolumeDSP",
				Device:              dsp,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []ei.EventInfo{eventInfo},
				Parameters:          parameters,
			}, nil
		}
	}

	return base.ActionStructure{}, errors.New("Could not find port for mic " + mic.Name)
}

func GetDSPMediaVolumeAction(dsp structs.Device, room base.PublicRoom, eventInfo ei.EventInfo, volume int) ([]base.ActionStructure, error) { //commands are issued to whatever port doesn't have a mic connected
	base.Log("%v", volume)

	base.Log("Generating action for command SetVolume on media routed through DSP")

	var output []base.ActionStructure

	for _, port := range dsp.Ports {
		parameters := make(map[string]string)
		parameters["level"] = fmt.Sprintf("%v", volume)
		eventInfo.EventInfoValue = fmt.Sprintf("%v", volume)
		eventInfo.Device = dsp.Name

		sourceDevice, err := dbo.GetDeviceByName(room.Building, room.Room, port.Source)
		if err != nil {
			errorMessage := "Could not get device " + port.Source + " from database: " + err.Error()
			base.Log(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		if !(sourceDevice.HasRole("Microphone")) {

			destination := base.DestinationDevice{
				Device:      dsp,
				AudioDevice: true,
			}

			parameters["input"] = port.Name
			action := base.ActionStructure{
				Action:              "SetVolume",
				GeneratingEvaluator: "SetVolumeDSP",
				Device:              dsp,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []ei.EventInfo{eventInfo},
				Parameters:          parameters,
			}

			output = append(output, action)
		}
	}

	return output, nil

}

func GetDisplayVolumeAction(device structs.Device, room base.PublicRoom, eventInfo ei.EventInfo, volume int) (base.ActionStructure, error) { //commands are issued to devices, e.g. they aren't connected to the DSP

	base.Log("Generating action for SetVolume on device %s external to DSP", device.Name)

	parameters := make(map[string]string)

	destination := base.DestinationDevice{
		Device:      device,
		AudioDevice: true,
	}

	if device.HasRole("VideoOut") {
		destination.Display = true
	}

	eventInfo.EventInfoValue = strconv.Itoa(volume)
	eventInfo.Device = device.Name
	parameters["level"] = strconv.Itoa(volume)

	action := base.ActionStructure{
		Action:              "SetVolume",
		GeneratingEvaluator: "SetVolumeDSP",
		Device:              device,
		DestinationDevice:   destination,
		DeviceSpecific:      true,
		EventLog:            []ei.EventInfo{eventInfo},
		Parameters:          parameters,
	}

	return action, nil
}
