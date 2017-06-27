package commandevaluators

import (
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
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

	if room.Muted != nil && *room.Muted == false {

		generalActions, err := evaluateGeneralUnMute(room, eventInfo)
		if err != nil {
			log.Printf("Error evaluating general unmute command %s", err.Error())
			return []base.ActionStructure{}, err
		}

		actions = generalActions

	}

	if len(room.AudioDevices) > 0 {

		specializedActions, err := evaluateSpecificUnMute(room, eventInfo)
		if err != nil {
			log.Printf("Error evalutating unmute commands %s", err.Error())
			return []base.ActionStructure{}, err
		}

		actions = append(actions, specializedActions...)

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

func evaluateGeneralUnMute(room base.PublicRoom, eventInfo ei.EventInfo) ([]base.ActionStructure, error) {

	log.Printf("Detected general unmute request")

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
				Action:              "UnMute",
				GeneratingEvaluator: "UnMuteDSP",
				Device:              device,
				DeviceSpecific:      false,
				EventLog:            []ei.EventInfo{eventInfo},
			})

		}
	}

	return actions, nil

}

func evaluateSpecificUnMute(room base.PublicRoom, eventInfo ei.EventInfo) ([]base.ActionStructure, error) {

	log.Printf("Detected specific unmute requests. Evaluating devices...")

	var actions []base.ActionStructure

	for _, audioDevice := range room.AudioDevices {

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
						Action:              "UnMute",
						GeneratingEvaluator: "UnMuteDSP",
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
				Action:              "UnMute",
				GeneratingEvaluator: "UnMuteDSP",
				Device:              device,
				Parameters:          parameters,
				EventLog:            []ei.EventInfo{eventInfo},
			})

		}

	}

	return actions, nil
}
