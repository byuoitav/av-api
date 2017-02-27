package commandevaluators

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

type SetVolumeDMPS struct {
}

//Validate checks for a volume for the entire room or the volume of a specific device
func (*SetVolumeDMPS) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	actions := []base.ActionStructure{}

	// general room volume
	if room.Volume != nil {

		newVol := remapVolume(*room.Volume)

		log.Printf("General volume request detected.")

		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		for i := range devices {

			if devices[i].Output {

				parameters := make(map[string]string)
				parameters["level"] = fmt.Sprintf("%v", newVol)

				actions = append(actions, base.ActionStructure{
					Action:              "SetVolume",
					Parameters:          parameters,
					GeneratingEvaluator: "SetVolumeDMPS",
					Device:              devices[i],
					DeviceSpecific:      false,
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
				newVol := remapVolume(*audioDevice.Volume)

				device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
				if err != nil {
					return []base.ActionStructure{}, err
				}

				parameters := make(map[string]string)
				parameters["level"] = string(*audioDevice.Volume)

				actions = append(actions, base.ActionStructure{
					Action:              "SetVolume",
					GeneratingEvaluator: "SetVolumeDMPS",
					Device:              device,
					DeviceSpecific:      true,
				})

			}

		}

	}

	log.Printf("%v actions generated.", len(actions))
	log.Printf("Evaluation complete.")

	fmt.Printf("Generated: %+v\n", actions)

	return actions, nil
}

func remapVolume(int oldLevel) int {
	MinLevel := 0
	MaxLevel := 65534

	return (oldLevel * (MaxLevel - MinLevel) / 100) + MinLevel
}

//Evaluate returns an error if the volume is greater than 100 or less than 0
func (p *SetVolumeDMPS) Validate(action base.ActionStructure) error {
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

	return nil

}

//GetIncompatibleCommands returns a string array of commands incompatible with setting the volume
func (p *SetVolumeDMPS) GetIncompatibleCommands() []string {
	return nil
}
