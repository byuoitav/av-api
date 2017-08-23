package commandevaluators

/**
ASSUMPTIONS:

a) there is only 1 DSP in a given room

b) microphones only have one port configuration and the DSP is the destination device

c) room-wide requests do not affect microphones

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

type MuteDSP struct{}

func (p *MuteDSP) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, error) {

	log.Printf("Evaluating PUT body for \"Mute\" command in DSP context...")

	var actions []base.ActionStructure

	eventInfo := ei.EventInfo{
		Type:           ei.CORESTATE,
		EventCause:     ei.USERINPUT,
		EventInfoKey:   "muted",
		EventInfoValue: "true",
		Requestor:      requestor,
	}

	destination := se.DestinationDevice{
		AudioDevice: true,
	}

	if room.Muted != nil && *room.Muted {

		generalActions, err := GetGeneralMuteRequestActionsDSP(room, eventInfo, destination)
		if err != nil {
			errorMessage := "Could not generate actions for room-wide \"Mute\" request: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, generalActions...)
	}

	if len(room.AudioDevices) > 0 {

		for _, audioDevice := range room.AudioDevices {

			if audioDevice.Muted == nil || !(*audioDevice.Muted) {
				continue
			}

			device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
			if err != nil {
				log.Printf("Error getting device %s from database: %s", audioDevice.Name, err.Error())
			}

			destination.Device = device //if we've made it this far, the destination device is this audio device

			if device.HasRole("Microphone") {

				action, err := GetMicMuteAction(device, room, eventInfo)
				if err != nil {
					return []base.ActionStructure{}, err
				}

				actions = append(actions, action)

			} else if device.HasRole("DSP") {

				dspActions, err := GetDSPMediaMuteAction(device, room, eventInfo, true)
				if err != nil {
					return []base.ActionStructure{}, err
				}

				actions = append(actions, dspActions...)

			} else if device.HasRole("AudioOut") {

				action, err := GetDisplayMuteAction(device, room, eventInfo, true)
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

	log.Printf("%s actions generated.", len(actions))
	log.Printf("Evaluation complete.")

	return actions, nil

}

func (p *MuteDSP) Validate(base.ActionStructure) error {
	//TODO make sure the device actually can be muted
	return nil
}

func (p *MuteDSP) GetIncompatibleCommands() []string {
	return nil
}

//assumes only one DSP, but allows for the possiblity of multiple devices not routed through the DSP
//room-wide mute requests DO NOT include mics
func GetGeneralMuteRequestActionsDSP(room base.PublicRoom, eventInfo ei.EventInfo, destination se.DestinationDevice) ([]base.ActionStructure, error) {

	log.Printf("Generating actions for room-wide \"Mute\" request")

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

	dspActions, err := GetDSPMediaMuteAction(dsp[0], room, eventInfo, false)
	if err != nil {
		errorMessage := "Could not generate action corresponding to general mute request in room " + room.Room + ", building " + room.Building + ": " + err.Error()
		log.Printf(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	actions = append(actions, dspActions...)

	audioDevices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
	if err != nil {
		log.Printf("Error getting devices %s", err.Error())
		return []base.ActionStructure{}, err
	}

	for _, device := range audioDevices {

		action, err := GetDisplayMuteAction(device, room, eventInfo, false)
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
func GetMicMuteAction(mic structs.Device, room base.PublicRoom, eventInfo ei.EventInfo) (base.ActionStructure, error) {

	log.Printf("Generating action for command \"Mute\" on microphone %s", mic.Name)

	destination := se.DestinationDevice{
		Device:      mic,
		AudioDevice: true,
	}

	//get DSP
	dsps, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "DSP")
	if err != nil {
		errorMessage := "Error getting DSP configuration in building " + room.Building + ", room " + room.Room + ": " + err.Error()
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//verify DSP configuration
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
				Action:              "Mute",
				GeneratingEvaluator: "MuteDSP",
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

func GetDSPMediaMuteAction(dsp structs.Device, room base.PublicRoom, eventInfo ei.EventInfo, deviceSpecific bool) ([]base.ActionStructure, error) {

	log.Printf("Generating action for command Mute on media routed through DSP")

	var output []base.ActionStructure
	eventInfo.Device = dsp.Name

	for _, port := range dsp.Ports {
		parameters := make(map[string]string)

		sourceDevice, err := dbo.GetDeviceByName(room.Building, room.Room, port.Source)
		if err != nil {
			errorMessage := "Could not get device " + port.Source + " from database " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		if !sourceDevice.HasRole("Microphone") {

			destination := se.DestinationDevice{
				Device:      dsp,
				AudioDevice: true,
			}

			parameters["input"] = port.Name
			action := base.ActionStructure{
				Action:              "Mute",
				GeneratingEvaluator: "MuteDSP",
				Device:              dsp,
				DestinationDevice:   destination,
				DeviceSpecific:      deviceSpecific,
				EventLog:            []ei.EventInfo{eventInfo},
				Parameters:          parameters,
			}

			output = append(output, action)
		}
	}

	return output, nil
}

func GetDisplayMuteAction(device structs.Device, room base.PublicRoom, eventInfo ei.EventInfo, deviceSpecific bool) (base.ActionStructure, error) {

	log.Printf("Generating action for command \"Mute\" for device %s external to DSP", device.Name)

	eventInfo.Device = device.Name

	destination := se.DestinationDevice{
		Device:      device,
		AudioDevice: true,
	}

	if device.HasRole("VideoOut") {
		destination.Display = true
	}

	return base.ActionStructure{
		Action:              "Mute",
		GeneratingEvaluator: "MuteDSP",
		Device:              device,
		DestinationDevice:   destination,
		DeviceSpecific:      deviceSpecific,
		EventLog:            []ei.EventInfo{eventInfo},
	}, nil
}
