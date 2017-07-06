package commandevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

/**
ASSUMPTIONS:

a) there is only 1 DSP in a given room and it is the only device with an 'AudioOut' role

b) there is only 1 video switcher in a given room

c) all the media audio is routed through the switcher

d) all audio inputs are routed throught the switcher and are found exactly one edge away from the video switcher

e) microphones are not affected by actions generated in this command evaluator

**/

type ChangeAudioInputDSP struct{}

func (p *ChangeAudioInputDSP) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Evaluating PUT body for \"ChangeInput\" command in an audio DSP context...")

	var actions []base.ActionStructure

	eventInfo := ei.EventInfo{
		Type:         ei.CORESTATE,
		EventCause:   ei.USERINPUT,
		EventInfoKey: "input",
	}

	if len(room.CurrentAudioInput) > 0 { //

		generalAction, err := GetDSPMediaInputAction(room, eventInfo, room.CurrentAudioInput, false)
		if err != nil {
			errorMessage := "Could not generate actions for room-wide \"ChangeInput\" request: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, generalAction)
	}

	if len(room.AudioDevices) == 1 { //the only device coming here is going to be the DSP

		specificAction, err := GetDSPMediaInputAction(room, eventInfo, room.AudioDevices[0].Input, true)
		if err != nil {
			errorMessage := "Could not generate actions for specific \"ChangeInput\" requests: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, specificAction)
	}

	log.Printf("%s actions generated.", len(actions))
	log.Printf("Evalutation complete")

	return actions, nil
}

func GetDSPMediaInputAction(room base.PublicRoom, eventInfo ei.EventInfo, input string, deviceSpecific bool) (base.ActionStructure, error) {

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

	//get the port configurations where the requested device is source
	ports, err := dbo.GetPortConfigurationsBySourceDevice(device)
	if err != nil {
		errorMessage := "Problem getting port configurations where device " + device.Name + " is source: " + err.Error()
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//find the port where the host is the switcher and the destination is the DSP
	for _, port := range ports {

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

			return base.ActionStructure{
				Action:              "ChangeInput",
				GeneratingEvaluator: "ChangeAudioInputDSP",
				Device:              switchers[0],
				DeviceSpecific:      deviceSpecific,
				Parameters:          parameters,
				EventLog:            []ei.EventInfo{eventInfo},
			}, nil

		}

	}

	return base.ActionStructure{}, errors.New("No port found for given input")

}
