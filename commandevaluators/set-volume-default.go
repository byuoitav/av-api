package commandevaluators

import (
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

type SetVolumeDefault struct {
}

//Validate checks for a volume for the entire room or the volume of a specific device
func (*SetVolumeDefault) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	actions := []base.ActionStructure{}

	// general room volume
	if room.Volume != nil {

		log.Printf("General volume request detected.")

		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Room, room.Building, "AudioOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		for i := range devices {

			if devices[i].Output {

				parameters := make(map[string]string)
				parameters["level"] = string(*room.Volume)

				actions = append(actions, base.ActionStructure{
					Action:              "SetVolume",
					Parameters:          parameters,
					GeneratingEvaluator: "SetVolumeDefault",
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

				device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
				if err != nil {
					return []base.ActionStructure{}, err
				}

				parameters := make(map[string]string)
				parameters["port"] = string(*audioDevice.Volume)

				actions = append(actions, base.ActionStructure{
					Action:              "SetVolume",
					GeneratingEvaluator: "SetVolumeDefault",
					Device:              device,
					DeviceSpecific:      true,
				})

			}

		}

	}

	return actions, nil
}