package commandevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/statusevaluators"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

/**
ASSUMPTIONS:

a) there is only 1 DSP in a given room

b) there is only 1 video switcher in a given room

c) the switcher has access to all the media audio

d) a room-wide audio input request implies sending a command to the DSP and muting all devices designatied as 'AudioOut'

e) microphones are not affected by actions generated in this command evaluator

**/

type ChangeAudioInputDSP struct{}

func (p *ChangeAudioInputDSP) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	log.Printf("Evaluating PUT body for \"ChangeInput\" command in an audio DSP context...")

	var actions []base.ActionStructure

	eventInfo := ei.EventInfo{
		Type:         ei.CORESTATE,
		EventCause:   ei.USERINPUT,
		EventInfoKey: "input",
		Requestor:    requestor,
	}

	destination := statusevaluators.DestinationDevice{
		AudioDevice: true,
	}

	if len(room.CurrentAudioInput) > 0 { //

		generalAction, err := GetDSPMediaInputAction(room, eventInfo, room.CurrentAudioInput, false, destination)
		if err != nil {
			errorMessage := "Could not generate actions for room-wide \"ChangeInput\" request: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, 0, errors.New(errorMessage)
		}

		actions = append(actions, generalAction)

		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
		if err != nil {
			errorMessage := "Could not generate actions for room-wide \"ChangeInput\" request: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, 0, errors.New(errorMessage)
		}

		for _, device := range devices {

			if device.Output && !device.HasRole("Microphone") {

				log.Printf("Adding device %+v", device.Name)

				eventInfo.Device = device.Name
				actions = append(actions, base.ActionStructure{
					Action:              "Mute",
					GeneratingEvaluator: "ChangeAudioInputDSP",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []ei.EventInfo{eventInfo},
				})
			}

		}

	}

	//TODO will this be a problem if the slice is nil?
	if len(room.AudioDevices) > 0 {

		for _, audioDevice := range room.AudioDevices {

			if len(audioDevice.Input) > 0 {

				device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
				if err != nil {
					errorMessage := "Could not get device: " + audioDevice.Name + " from database: " + err.Error()
					log.Printf(errorMessage)
					return []base.ActionStructure{}, 0, errors.New(errorMessage)
				}

				if device.HasRole("DSP") {

					dspAction, err := GetDSPMediaInputAction(room, eventInfo, room.AudioDevices[0].Input, true, destination)
					if err != nil {
						errorMessage := "Could not generate actions for specific \"ChangeInput\" requests: " + err.Error()
						log.Printf(errorMessage)
						return []base.ActionStructure{}, 0, errors.New(errorMessage)
					}

					actions = append(actions, dspAction)

				} else if device.HasRole("AudioOut") && !device.HasRole("Microphone") {

					mediaAction, err := generateChangeInputByDevice(audioDevice.Device, room.Room, room.Building, "ChangeAudioInputDefault", requestor)
					if err != nil {
						errorMessage := "Could not generate actions for specific \"ChangeInput\" request for deivce: " + device.Name + ": " + err.Error()
						log.Printf(errorMessage)
						return []base.ActionStructure{}, 0, errors.New(errorMessage)
					}
					actions = append(actions, mediaAction)
				}
			}
		}
	}

	log.Printf("%s actions generated.", len(actions))
	log.Printf("Evalutation complete")

	return actions, len(actions), nil
}

func GetDSPMediaInputAction(room base.PublicRoom, eventInfo ei.EventInfo, input string, deviceSpecific bool, destination statusevaluators.DestinationDevice) (base.ActionStructure, error) {

	//get DSP
	dsp, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "DSP")
	if err != nil {
		errorMessage := "Problem getting device " + input + " from database " + err.Error()
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//validate number of DSPs
	if len(dsp) != 1 {
		errorMessage := "Invalid DSP configuration detected in room"
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//get switcher
	switchers, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoSwitcher")
	if err != nil {
		errorMessage := "Could not get room switch in room " + room.Room + ", building " + room.Building + ": " + err.Error()
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//validate number of switchers
	if len(switchers) != 1 {
		errorMessage := "Invalid video switch configuration detected in room"
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//get requested device
	device, err := dbo.GetDeviceByName(room.Building, room.Room, input)
	if err != nil {
		errorMessage := "Problem getting device " + input + " from database " + err.Error()
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//find the port where the host is the switcher and the destination is the DSP
	for _, port := range device.Ports {

		if port.Host == switchers[0].Name && port.Destination == dsp[0].Name {
			//once we find the port, send the command to the switcher

			switcherPorts := strings.Split(port.Name, ":")
			if len(switcherPorts) != 2 {
				return base.ActionStructure{}, errors.New("Invalid video switcher port")
			}

			parameters := make(map[string]string)
			parameters["input"] = switcherPorts[0]
			parameters["output"] = switcherPorts[1]

			eventInfo.Device = switchers[0].Name
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

	return base.ActionStructure{}, errors.New("No port found for given input")

}
