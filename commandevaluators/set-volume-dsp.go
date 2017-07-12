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
	"log"
	"strconv"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"

	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

type SetVolumeDSP struct{}

func (p *SetVolumeDSP) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Evaluating SetVolume command in DSP context...")

	eventInfo := ei.EventInfo{
		Type:         ei.CORESTATE,
		EventCause:   ei.USERINPUT,
		EventInfoKey: "volume",
	}

	var actions []base.ActionStructure

	if room.Volume != nil {

		log.Printf("Room-wide request detected")

		eventInfo.EventInfoValue = strconv.Itoa(*room.Volume)

		actions, err := GetGeneralVolumeRequestActionsDSP(room, eventInfo)
		if err != nil {
			errorMessage := "Could not generate actions for room-wide \"SetVolume\" request: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, actions...)
	}

	if len(room.AudioDevices) > 0 {

		for _, audioDevice := range room.AudioDevices {

			if audioDevice.Volume != nil {

				eventInfo.EventInfoValue = strconv.Itoa(*audioDevice.Volume)

				device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
				if err != nil {
					log.Printf("Error getting device %s from database: %s", audioDevice.Name, err.Error())
				}

				if device.HasRole("Microphone") {

					action, err := GetMicVolumeAction(device, room, eventInfo, *audioDevice.Volume)
					if err != nil {
						return []base.ActionStructure{}, err
					}

					actions = append(actions, action)

				} else if device.HasRole("DSP") {

					dspActions, err := GetDSPMediaVolumeAction(device, room, eventInfo, *audioDevice.Volume)
					if err != nil {
						return []base.ActionStructure{}, err
					}

					actions = append(actions, dspActions...)

				} else if device.HasRole("AudioOut") {

					action, err := GetDisplayVolumeAction(device, room, eventInfo, *audioDevice.Volume)
					if err != nil {
						return []base.ActionStructure{}, err
					}

					actions = append(actions, action)

				} else { //bad device
					errorMessage := "Cannot set volume of device: " + device.Name + " in given context"
					log.Printf(errorMessage)
					return []base.ActionStructure{}, errors.New(errorMessage)
				}
			}
		}
	}

	log.Printf("%v actions generated.", len(actions))

	for _, a := range actions {
		log.Printf("%v, %v", a.Action, a.Parameters)

	}

	log.Printf("Evaluation complete.")
	return actions, nil
}

func (p *SetVolumeDSP) Validate(action base.ActionStructure) (err error) {
	maximum := 100
	minimum := 0

	level, err := strconv.Atoi(action.Parameters["level"])
	if err != nil {
		return err
	}

	if level > maximum || level < minimum {
		log.Printf("ERROR. %v is an invalid volume level for %s", action.Parameters["level"], action.Device.Name)
		return errors.New(action.Action + " is an invalid command for " + action.Device.Name)
	}

	return
}

func (p *SetVolumeDSP) GetIncompatibleCommands() (incompatibleActions []string) {
	return nil
}

func GetGeneralVolumeRequestActionsDSP(room base.PublicRoom, eventInfo ei.EventInfo) ([]base.ActionStructure, error) {

	log.Printf("Generating actions for room-wide \"SetVolume\" request")

	var actions []base.ActionStructure

	dsp, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "DSP")
	if err != nil {
		log.Printf("Error getting devices %s", err.Error)
		return []base.ActionStructure{}, err
	}

	dspActions, err := GetDSPMediaVolumeAction(dsp[0], room, eventInfo, *room.Volume)
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
		if device.HasRole("DSP") {
			continue
		}

		action, err := GetDisplayVolumeAction(device, room, eventInfo, *room.Volume)
		if err != nil {
			errorMessage := "Could not generate mute action for display " + device.Name + " in room " + room.Room + ", building " + room.Building + ": " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, action)
	}

	return actions, nil
}

//we assume microphones are only connected to a DSP
//commands regarding microphones are only issued to DSP
func GetMicVolumeAction(mic accessors.Device, room base.PublicRoom, eventInfo ei.EventInfo, volume int) (base.ActionStructure, error) {

	log.Printf("Identified microphone volume request")

	parameters := make(map[string]string)

	eventInfo.EventInfoValue = string(volume)
	parameters["level"] = string(volume)
	parameters["input"] = mic.Ports[0].Name

	dsp, err := dbo.GetDeviceByName(room.Building, room.Room, mic.Ports[0].Destination)
	if err != nil {
		errorMessage := "Could not get DSP corresponding to mic " + mic.Name + ": " + err.Error()
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	return base.ActionStructure{
		Action:              "SetVolume",
		GeneratingEvaluator: "SetVolumeDSP",
		Device:              dsp,
		DeviceSpecific:      true,
		EventLog:            []ei.EventInfo{eventInfo},
		Parameters:          parameters,
	}, nil

}

func GetDSPMediaVolumeAction(device accessors.Device, room base.PublicRoom, eventInfo ei.EventInfo, volume int) ([]base.ActionStructure, error) { //commands are issued to whatever port doesn't have a mic connected
	log.Printf("%v", volume)

	log.Printf("Generating action for command SetVolume on media routed through DSP")

	var output []base.ActionStructure

	for _, port := range device.Ports {
		parameters := make(map[string]string)
		parameters["level"] = fmt.Sprintf("%v", volume)
		eventInfo.EventInfoValue = fmt.Sprintf("%v", volume)

		sourceDevice, err := dbo.GetDeviceByName(room.Building, room.Room, port.Source)
		if err != nil {
			errorMessage := "Could not get device " + port.Source + " from database: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		if !(sourceDevice.HasRole("Microphone")) {

			parameters["input"] = port.Name
			action := base.ActionStructure{
				Action:              "SetVolume",
				GeneratingEvaluator: "SetVolumeDSP",
				Device:              device,
				DeviceSpecific:      true,
				EventLog:            []ei.EventInfo{eventInfo},
				Parameters:          parameters,
			}

			output = append(output, action)
		}
	}

	return output, nil

}

func GetDisplayVolumeAction(device accessors.Device, room base.PublicRoom, eventInfo ei.EventInfo, volume int) (base.ActionStructure, error) { //commands are issued to devices, e.g. they aren't connected to the DSP

	log.Printf("Generating action for SetVolume on device %s external to DSP", device.Name)

	parameters := make(map[string]string)

	eventInfo.EventInfoValue = strconv.Itoa(volume)
	parameters["level"] = strconv.Itoa(volume)

	action := base.ActionStructure{
		Action:              "SetVolume",
		GeneratingEvaluator: "SetVolumeDSP",
		Device:              device,
		DeviceSpecific:      true,
		EventLog:            []ei.EventInfo{eventInfo},
		Parameters:          parameters,
	}

	return action, nil
}
