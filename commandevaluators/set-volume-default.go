package commandevaluators

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

type SetVolumeDefault struct {
}

//Validate checks for a volume for the entire room or the volume of a specific device
func (*SetVolumeDefault) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	var actions []base.ActionStructure

	eventInfo := eventinfrastructure.EventInfo{
		Type:         eventinfrastructure.USERACTION,
		EventCause:   eventinfrastructure.USERINPUT,
		EventInfoKey: "volume",
	}

	// general room volume
	if room.Volume != nil {

		log.Printf("General volume request detected.")

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
					GeneratingEvaluator: "SetVolumeDefault",
					Device:              device,
					DeviceSpecific:      false,
					EventLog:            []eventinfrastructure.EventInfo{eventInfo},
				})

			}

		}

	}

	//identify devices in request body
	if len(room.AudioDevices) != 0 {

		log.Printf("Device specific request detected. Scanning devices")

		for _, audioDevice := range room.AudioDevices {
			// create actions based on request

			if audioDevice.Volume != nil {
				log.Printf("Adding device %+v", audioDevice.Name)

				device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
				if err != nil {
					return []base.ActionStructure{}, err
				}

				parameters := make(map[string]string)
				parameters["level"] = fmt.Sprintf("%v", *audioDevice.Volume)
				log.Printf("%+v", parameters)

				eventInfo.EventInfoValue = fmt.Sprintf("%v", *audioDevice.Volume)
				eventInfo.Device = device.Name
				actions = append(actions, base.ActionStructure{
					Action:              "SetVolume",
					GeneratingEvaluator: "SetVolumeDefault",
					Device:              device,
					DeviceSpecific:      true,
					Parameters:          parameters,
					EventLog:            []eventinfrastructure.EventInfo{eventInfo},
				})

			}

		}

	}

	log.Printf("%v actions generated.", len(actions))
	log.Printf("Evaluation complete.")

	return actions, nil
}

func validateSetVolumeMaxMin(action base.ActionStructure, maximum int, minimum int) error {
	level, err := strconv.Atoi(action.Parameters["level"])
	if err != nil {
		return err
	}

	if level > maximum || level < minimum {
		log.Printf("ERROR. %v is an invalid volume level for %s", action.Parameters["level"], action.Device.Name)
		return errors.New(action.Action + " is an invalid command for " + action.Device.Name)
	}
	return nil
}

//Evaluate returns an error if the volume is greater than 100 or less than 0
func (p *SetVolumeDefault) Validate(action base.ActionStructure) error {
	maximum := 100
	minimum := 0

	return validateSetVolumeMaxMin(action, maximum, minimum)

}

//GetIncompatibleCommands returns a string array of commands incompatible with setting the volume
func (p *SetVolumeDefault) GetIncompatibleCommands() []string {
	return nil
}
