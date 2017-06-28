package commandevaluators

import (
	"errors"
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

/**
ASSUMPTIONS:

a) there is only 1 DSP in a given room

b) there is only 1 video switcher in a given room

c) all the media audio routed through the DSP is controlled by the video switcher

d) the video switcher has an "AudioOut" role

e) all audio inputs are one edge away from the video switcher

d) microphones are not affected by actions generated in this command evaluator

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

	if len(room.CurrentAudioInput) > 0 {

		log.Printf("Room-wide audio input request detected")

		generalActions, err := GetGeneralAudioInputRequestActionsDSP(room, eventInfo, room.CurrentAudioInput)
		if err != nil {
			errorMessage := "Could not generate actions for room-wide \"ChangeInput\" request: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, generalActions...)
	}

	if len(room.AudioDevices) > 0 {

		specificActions, err := GetSpecifcAudioInputRequestActionsDSP(room, eventInfo)
		if err != nil {
			errorMessage := "Could not generate actions for specific \"ChangeInput\" requests: " + err.Error()
			log.Printf(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		actions = append(actions, specificActions...)
	}

	log.Printf("%s actions generated.", len(actions))
	log.Printf("Evalutation complete")

	return actions, nil
}

func GetGeneralAudioInputRequestActionsDSP(room base.PublicRoom, eventInfo ei.EventInfo, input string) ([]base.ActionStructure, error) {

	var actions []base.ActionStructure

	//anything that's not a microphone input is coming from a video switcher
	//send the command to the video switcher

	action, err := GetDSPMediaInputAction(room, eventInfo, input)
	if err != nil {
		errorMessage := "Could not generate action for \"ChangeInput\" request directed at DSP: " + err.Error()
		log.Printf(errorMessage)
		return []base.ActionStructure{}, errors.New(errorMessage)
	}

	actions = append(actions, action)

	return actions, nil

}

func GetSpecifcAudioInputRequestActionsDSP(room base.PublicRoom, eventInfo ei.EventInfo) ([]base.ActionStructure, error) {

	var actions []base.ActionStructure

	return actions, nil

}

func GetDSPMediaInputAction(room base.PublicRoom, eventInfo ei.EventInfo, input string) (base.ActionStructure, error) {

	var action base.ActionStructure

	//get DSP
	dsp, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "DSP")
	if err != nil {
		errorMessage := "Problem getting device " + input + " from database " + err.Error()
		log.Printf(errorMessage)
		return base.ActionStructure{}, errors.New(errorMessage)
	}

	//validate number of DSPs
	if len(dsp) != 1 {
	}

	//get switcher
	switchers, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoSwitcher")
	if err != nil {
	}

	//validate number of switchers
	if len(switchers) != 1 {
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

		//once we find the port, send the command to the switcher

	}
	return action, nil

}
