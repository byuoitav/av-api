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
	"github.com/byuoitav/configuration-database-microservice/accessors"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

type UnMuteDSP struct{}

func (p *UnMuteDSP) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Evaluating PUT body for MUTE command in DSP context...")

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

					actions = append(actions, action)

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

	actions = append(actions, action)

	audioDevices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
	if err != nil {
		log.Printf("Error getting devices %s", err.Error())
		return []base.ActionStructure{}, err
	}

	for _, device := range audioDevices {

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
func GetMicUnMuteAction(mic accessors.Device, room base.PublicRoom, eventInfo ei.EventInfo) (base.ActionStructure, error) {

	log.Printf("Generating action for command \"UnMute\" on microphone %s", mic.Name)

	parameters := make(map[string]string)
	parameters["input"] = mic.Ports[0].Name

	dsp, err := dbo.GetDeviceByName(room.Building, room.Room, mic.Ports[0].Destination)
	if err != nil {
		errorMessage := "Could not get DSP corresponding to mic " + mic.Name + ": " + err.Error()
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}
	return base.ActionStructure{
		Action:              "UnMute",
		GeneratingEvaluator: "UnMuteDSP",
		Device:              dsp,
		DeviceSpecific:      true,
		EventLog:            []ei.EventInfo{eventInfo},
		Parameters:          parameters,
	}, nil
}

func GetDSPMediaUnMuteAction(dsp accessors.Device, room base.PublicRoom, eventInfo ei.EventInfo, deviceSpecific bool) (base.ActionStructure, error) {

	log.Printf("Generating action for command UnMute on media routed through DSP")

	parameters := make(map[string]string)

	for _, port := range dsp.Ports {

		sourceDevice, err := dbo.GetDeviceByName(room.Building, room.Room, port.Source)
		if err != nil {
			errorMessage := "Could not get device " + port.Source + " from database " + err.Error()
			log.Printf(errorMessage)
			return base.ActionStructure{}, errors.New(errorMessage)
		}

		if sourceDevice.HasRole("AudioOut") {

			parameters["input"] = port.Name
			return base.ActionStructure{
				Action:              "UnMute",
				GeneratingEvaluator: "UnMuteDSP",
				Device:              dsp,
				DeviceSpecific:      deviceSpecific,
				EventLog:            []ei.EventInfo{eventInfo},
				Parameters:          parameters,
			}, nil
		}
	}

	return base.ActionStructure{}, nil
}

func GetDisplayUnMuteAction(device accessors.Device, room base.PublicRoom, eventInfo ei.EventInfo, deviceSpecific bool) (base.ActionStructure, error) {

	log.Printf("Generating action for command \"UnMute\" for device %s external to DSP", device.Name)

	return base.ActionStructure{
		Action:              "UnMute",
		GeneratingEvaluator: "UnMuteDSP",
		Device:              device,
		DeviceSpecific:      deviceSpecific,
		EventLog:            []ei.EventInfo{eventInfo},
	}, nil
}
