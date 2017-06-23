package commandevaluators

import (
	"fmt"
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"

	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

type SetVolumeDSP struct{}

func (p *SetVolumeDSP) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Evaluating commands in a DSP context...")

	var actions []base.ActionStructure

	eventInfo := ei.EventInfo{
		Type:         ei.CORESTATE,
		EventCause:   ei.USERINPUT,
		EventInfoKey: "volume",
	}

	if room.Volume != nil {

		log.Printf("General volume request detected.")
		actions, err := EvaluateGeneral(eventInfo, room)
		if err != nil {
			return actions, err
		}
	}

	if len(room.AudioDevices) != 0 {

		log.Printf("Device-specific request detected.")
		specializedActions, err := EvaluateSpecific(eventInfo, room)
		if err != nil {
			return specializedActions, err
		}

		actions = append(actions, specializedActions...)

	}

	return actions, nil
}

func (p *SetVolumeDSP) Validate(action base.ActionStructure) (err error) {
	return
}

func (p *SetVolumeDSP) GetIncompatibleCommands() (incompatibleActions []string) {
	return []string{}
}

func EvaluateGeneral(eventInfo ei.EventInfo, room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Builing action structrures for entire room...")

	var actions []base.ActionStructure
	devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
	if err != nil {
		return []base.ActionStructure{}, err
	}

	for _, device := range devices {

		if device.Output {

			parameters := make(map[string]string)
			parameters["level"] = fmt.Sprintf("%v", *room.Volume)

			eventInfo.EventInfoValue = fmt.Sprintf("%v", *room.Volume)
			eventInfo.Device = device.Name
			actions = append(actions, base.ActionStructure{
				Action:              "SetVolume",
				Parameters:          parameters,
				GeneratingEvaluator: "SetVolumeDSP",
				Device:              device,
				DeviceSpecific:      false,
				EventLog:            []ei.EventInfo{eventInfo},
			})

		}

	}

	return actions, nil
}

func EvaluateSpecific(eventInfo ei.EventInfo, room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Scanning devices")

	var actions []base.ActionStructure

	for _, audioDevice := range room.AudioDevices {

		if audioDevice.Volume != nil {

			log.Printf("Adding device %s", audioDevice.Name)

			device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
			if err != nil {
				return []base.ActionStructure{}, err
			}

			//trying to set the volume on an audio input device implies using the DSP
			for _, role := range device.Roles {

				if role == "AudioIn" {

					//send command to DSP

					//figure out which port the mic corresponds to
					//the name of the port the mic has is a parameter for the SetVolumeDSP endpoint

					//identify the correct ports

					for _, port := range device.Ports {

						destinationDevice, err := dbo.GetDeviceByName(room.Building, room.Room, port.Destination)
						if err != nil {
							return []base.ActionStructure{}, err
						}

						for _, destRole := range destinationDevice.Roles {

							if destRole == "DSP" { //we found a DSP to send the command to
								parameters := make(map[string]string)
								parameters["level"] = fmt.Sprintf("%v", *audioDevice.Volume)
								parameters["input"] = port.Name

								eventInfo.EventInfoValue = fmt.Sprintf("%v", *room.Volume)
								eventInfo.Device = device.Name
								actions = append(actions, base.ActionStructure{
									Action:              "SetVolume",
									GeneratingEvaluator: "SetVolumeDSP",
									Device:              destinationDevice,
									DeviceSpecific:      true,
									Parameters:          parameters,
									EventLog:            []ei.EventInfo{eventInfo},
								})
							}
						}
					}
				}
			}
		}
	}

	return actions, nil
}
