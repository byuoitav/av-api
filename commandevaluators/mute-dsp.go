package commandevaluators

import (
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

type MuteDSP struct{}

func (p *MuteDSP) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Evaluating PUT body for MUTE command in DSP context...")

	var actions []base.ActionStructure
	eventInfo := ei.EventInfo{
		Type:           ei.CORESTATE,
		EventCause:     ei.USERINPUT,
		EventInfoKey:   "muted",
		EventInfoValue: "true",
	}

	if room.Muted != nil {

		generalActions, err := evaluateGeneralMute(room, eventInfo)
		if err != nil {
			log.Printf("Error evaluating general mute command %s", err.Error())
			return []base.ActionStructure{}, err
		}

		actions = generalActions

	}

	if len(room.AudioDevices) > 0 {

		specializedActions, err := evaluateSpecificMute(room, eventInfo)
		if err != nil {
			log.Printf("Error evalutating mute commands %s", err.Error())
			return []base.ActionStructure{}, err
		}

		actions = append(actions, specializedActions...)

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

func evaluateGeneralMute(room base.PublicRoom, eventInfo ei.EventInfo) ([]base.ActionStructure, error) {

	log.Printf("Detected general mute request")

	var actions []base.ActionStructure

	devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
	if err != nil {
		log.Printf("Error getting devices %s", err.Error())
		return []base.ActionStructure{}, err
	}

	for _, device := range devices {

		if device.Output {

			log.Printf("Adding mute command for device %s", device.Name)

			eventInfo.Device = device.Name
			actions = append(actions, base.ActionStructure{
				Action:              "Mute",
				GeneratingEvaluator: "MuteDSP",
				Device:              device,
				DeviceSpecific:      false,
				EventLog:            []ei.EventInfo{eventInfo},
			})

		}
	}

	return actions, nil

}

func evaluateSpecificMute(room base.PublicRoom, eventInfo ei.EventInfo) ([]base.ActionStructure, error) {

	log.Printf("Detected specific mute requests. Evaluating devices...")

	var actions []base.ActionStructure

	for _, audioDevice := range room.AudioDevices {

		if *audioDevice.Muted {

			log.Printf("Adding device %s", audioDevice.Name)

			device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
			if err != nil {
				log.Printf("Error getting device from database %s", err.Error())
				return []base.ActionStructure{}, err
			}

			parameters := make(map[string]string)
			if device.HasRole("AudioIn") {

				ports, err := dbo.GetPortConfigurationsBySourceDevice(device)
				if err != nil {
					log.Printf("Error getting port configurations of device %s: %s", device.Name, err.Error())
					return []base.ActionStructure{}, err
				}

				for _, port := range ports {

					destinationDevice, err := dbo.GetDeviceByName(room.Building, room.Room, port.Destination)
					if err != nil {
						log.Printf("Error getting destination device %s: %s", port.Destination, err.Error())
						return []base.ActionStructure{}, err
					}

					if destinationDevice.HasRole("DSP") { //identified DSP

						log.Printf("Identified DSP: %s corresponding to audio input device %s", destinationDevice.Name, device.Name)

						parameters["input"] = port.Name
						parameters["address"] = destinationDevice.Address
						actions = append(actions, base.ActionStructure{
							Action:              "Mute",
							GeneratingEvaluator: "MuteDSP",
							Device:              destinationDevice,
							Parameters:          parameters,
							EventLog:            []ei.EventInfo{eventInfo},
						})
					}
				}

			}
			if device.HasRole("AudioOut") { //business as usual

				parameters["address"] = device.Address
				actions = append(actions, base.ActionStructure{
					Action:              "Mute",
					GeneratingEvaluator: "MuteDSP",
					Device:              device,
					Parameters:          parameters,
					EventLog:            []ei.EventInfo{eventInfo},
				})

			}

		}
	}

	return actions, nil
}
