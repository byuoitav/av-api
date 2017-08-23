package commandevaluators

/**
ASSUMPTIONS:

a) there is only 1 DSP in a given room

b) microphones only have one port configuration and the DSP is the destination device

c) room-wide requests do not affect microphones

d) media devices connected to the DSP have the role "AudioOut"

**/

import (
	"errors"
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

type UnMuteDSP struct{}

func (p *UnMuteDSP) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Evaluating PUT body for UNMUTE command in DSP context...")

	var actions []base.ActionStructure
	eventInfo := ei.EventInfo{
		Type:           ei.CORESTATE,
		EventCause:     ei.USERINPUT,
		EventInfoKey:   "muted",
		EventInfoValue: "false",
	}

	if room.Muted != nil && !(*room.Muted) {

		generalActions, err := GetGeneralUnMuteRequestActionsDSP(room, eventInfo)
		if err != nil {
			errorMessage := "Could not generate actions for room-wide \"UnMute\" request: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, generalActions...)
	}

	if len(room.AudioDevices) > 0 {

		for _, audioDevice := range room.AudioDevices {

			if audioDevice.Muted != nil && !(*audioDevice.Muted) {

				device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
				if err != nil {
					log.Printf("Error getting device %s from database: %s", audioDevice.Name, err.Error())
				}

				if device.HasRole("Microphone") {

					action, err := GetMicUnMuteAction(device, room, eventInfo)
					if err != nil {
						return []base.ActionStructure{}, err
					}

					actions = append(actions, action)

				} else if device.HasRole("DSP") {

					action, err := GetDSPMediaUnMuteAction(device, room, eventInfo, true)
					if err != nil {
						return []base.ActionStructure{}, err
					}

					actions = append(actions, action...)

				} else if device.HasRole("AudioOut") {

					action, err := GetDisplayUnMuteAction(device, room, eventInfo, true)
					if err != nil {
						return []base.ActionStructure{}, err
					}

					actions = append(actions, action)

				} else { //bad device
					errorMessage := "Cannot set volume of device " + device.Name
					log.Printf(errorMessage)
					return []base.ActionStructure{}, errors.New(errorMessage)
				}
			}
		}
	}

	log.Printf("%s actions generated.", len(actions))
	log.Printf("Evaluation complete.")

	return actions, nil
}

func (p *UnMuteDSP) Validate(base.ActionStructure) error {
	//TODO make sure the device actually can be muted
	return nil
}

func (p *UnMuteDSP) GetIncompatibleCommands() []string {
	return nil
}

//assumes only one DSP, but allows for the possiblity of multiple devices not routed through the DSP
//room-wide mute requests DO NOT include mics
func GetGeneralUnMuteRequestActionsDSP(room base.PublicRoom, eventInfo ei.EventInfo) ([]base.ActionStructure, error) {

	log.Printf("Generating actions for room-wide \"UnMute\" request")

	var actions []base.ActionStructure

	dsp, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "DSP")
	if err != nil {
		log.Printf("Error getting devices %s", err.Error())
		return []base.ActionStructure{}, err
	}

	if len(dsp) != 1 {
		errorMessage := "Invalid number of DSP devices found in room: " + room.Room + " in building " + room.Building
		log.Printf(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	action, err := GetDSPMediaUnMuteAction(dsp[0], room, eventInfo, false)
	if err != nil {
		errorMessage := "Could not generate action corresponding to general mute request in room " + room.Room + ", building " + room.Building + ": " + err.Error()
		log.Printf(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	actions = append(actions, action...)

	audioDevices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
	if err != nil {
		log.Printf("Error getting devices %s", err.Error())
		return []base.ActionStructure{}, err
	}

	for _, device := range audioDevices {
		if device.HasRole("DSP") {
			continue
		}

		action, err := GetDisplayUnMuteAction(device, room, eventInfo, false)
		if err != nil {
			errorMessage := "Could not generate mute action for display " + device.Name + " in room " + room.Room + ", building " + room.Building + ": " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, action)
	}

	return actions, nil
}

//assumes the mic is only connected to a single DSP
func GetMicUnMuteAction(mic structs.Device, room base.PublicRoom, eventInfo ei.EventInfo) (base.ActionStructure, error) {

	log.Printf("Generating action for command \"UnMute\" on microphone %s", mic.Name)

	destination := se.DestinationDevice{
		Device:      mic,
		AudioDevice: true,
	}

	//TODO move me and parameterize the DSP for efficiency!
	//get DSP
	dsps, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "DSP")
	if err != nil {
		errorMessage := "Error getting DSP in building " + room.Building + ", room " + room.Room + ": " + err.Error()
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	if len(dsps) != 1 {
		errorMessage := "Invalid DSP configuration detected."
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	dsp := dsps[0]

	parameters := make(map[string]string)
	for _, port := range dsp.Ports {

		if port.Source == mic.Name {
			parameters["input"] = port.Name
			eventInfo.Device = mic.Name

			return base.ActionStructure{
				Action:              "UnMute",
				GeneratingEvaluator: "UnmuteDSP",
				Device:              dsp,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []ei.EventInfo{eventInfo},
				Parameters:          parameters,
			}, nil
		}

	}

	return base.ActionStructure{}, errors.New("Couldn't find port configuration for mic " + mic.Name)
}

func GetDSPMediaUnMuteAction(dsp structs.Device, room base.PublicRoom, eventInfo ei.EventInfo, deviceSpecific bool) ([]base.ActionStructure, error) {
	toReturn := []base.ActionStructure{}

	destination := se.DestinationDevice{
		Device:      dsp,
		AudioDevice: true,
	}

	log.Printf("Generating action for command UnMute on media routed through DSP")

	for _, port := range dsp.Ports {
		parameters := make(map[string]string)

		sourceDevice, err := dbo.GetDeviceByName(room.Building, room.Room, port.Source)
		if err != nil {
			errorMessage := "Could not get device " + port.Source + " from database " + err.Error()
			log.Printf(errorMessage)
			return toReturn, errors.New(errorMessage)
		}

		if sourceDevice.HasRole("Microphone") {
			continue
		}

		if sourceDevice.HasRole("AudioOut") || sourceDevice.HasRole("VideoSwitcher") {

			parameters["input"] = port.Name
			eventInfo.Device = dsp.Name

			toReturn = append(toReturn, base.ActionStructure{
				Action:              "UnMute",
				GeneratingEvaluator: "UnmuteDSP",
				Device:              dsp,
				DestinationDevice:   destination,
				DeviceSpecific:      deviceSpecific,
				EventLog:            []ei.EventInfo{eventInfo},
				Parameters:          parameters,
			})
		}
	}

	return toReturn, nil
}

func GetDisplayUnMuteAction(device structs.Device, room base.PublicRoom, eventInfo ei.EventInfo, deviceSpecific bool) (base.ActionStructure, error) {

	log.Printf("Generating action for command \"UnMute\" for device %s external to DSP", device.Name)

	eventInfo.Device = device.Name

	destination := se.DestinationDevice{
		Device:      device,
		AudioDevice: true,
	}

	if device.HasRole("VideoOut") {
		destination.Display = true
	}

	return base.ActionStructure{
		Action:              "UnMute",
		GeneratingEvaluator: "UnmuteDSP",
		Device:              device,
		DestinationDevice:   destination,
		DeviceSpecific:      deviceSpecific,
		EventLog:            []ei.EventInfo{eventInfo},
	}, nil
}
