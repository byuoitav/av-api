package commandevaluators

import (
	"errors"
	"fmt"
	"log"
	"strconv"

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
		generalActions, err := EvaluateGeneral(eventInfo, room)
		if err != nil {
			return []base.ActionStructure{}, err
		}

		actions = generalActions
	}

	if len(room.AudioDevices) != 0 {

		log.Printf("Device-specific request detected.")
		specializedActions, err := EvaluateSpecificVolume(eventInfo, room)
		if err != nil {
			return []base.ActionStructure{}, err
		}

		actions = append(actions, specializedActions...)

	}

	log.Printf("%v actions generated.", len(actions))
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

func EvaluateSpecificVolume(eventInfo ei.EventInfo, room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Scanning devices")

	var actions []base.ActionStructure

	for _, audioDevice := range room.AudioDevices {

		if audioDevice.Volume != nil {

			log.Printf("Adding device %s", audioDevice.Name)

			device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
			if err != nil {
				return []base.ActionStructure{}, err
			}

			parameters := make(map[string]string)
			parameters["level"] = fmt.Sprintf("%v", *audioDevice.Volume)
			eventInfo.EventInfoValue = fmt.Sprintf("%v", *room.Volume)
			eventInfo.Device = device.Name
			//trying to set the volume on an audio input device implies using the DSP
			if device.HasRole("AudioIn") {

				ports, err := dbo.GetPortConfigurationsBySourceDevice(device)
				if err != nil {
					log.Printf("Error getting port configurations of device %s: %s", device.Name, err.Error())
					continue
				}

				for _, port := range ports {

					destinationDevice, err := dbo.GetDeviceByName(room.Building, room.Room, port.Destination)
					if err != nil {
						log.Printf("Error getting destination device %s: %s", port.Destination, err.Error())
						continue
					}

					if destinationDevice.HasRole("DSP") {

						log.Printf("Identified DSP: %s corresponding to audio input device %s:", destinationDevice.Name, device.Name)

						parameters["input"] = port.Name

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

			if device.HasRole("AudioOut") {

				actions = append(actions, base.ActionStructure{
					Action:              "SetVolume",
					GeneratingEvaluator: "SetVolumeDSP",
					Device:              device,
					DeviceSpecific:      true,
					Parameters:          parameters,
					EventLog:            []ei.EventInfo{eventInfo},
				})
			}
		}
	}
	return actions, nil
}
