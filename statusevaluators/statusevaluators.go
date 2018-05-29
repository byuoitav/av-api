package statusevaluators

import (
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// StatusEvaluator defines the common functions for all StatusEvaluators.
type StatusEvaluator interface {

	//Identifies relevant devices
	GetDevices(room structs.Room) ([]structs.Device, error)

	//Generates action list
	GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error)

	//Evaluate Response
	EvaluateResponse(label string, value interface{}, Source structs.Device, Destination base.DestinationDevice) (string, interface{}, error)
}

// StatusEvaluatorMap is a map of the different StatusEvaluators used.
//TODO: we shoud grab the keys from constants in the evaluators themselves
var StatusEvaluatorMap = map[string]StatusEvaluator{
	"STATUS_PowerDefault":       &PowerDefault{},
	"STATUS_BlankedDefault":     &BlankedDefault{},
	"STATUS_MutedDefault":       &MutedDefault{},
	"STATUS_InputDefault":       &InputDefault{},
	"STATUS_VolumeDefault":      &VolumeDefault{},
	"STATUS_InputVideoSwitcher": &InputVideoSwitcher{},
	"STATUS_InputDSP":           &InputDSP{},
	"STATUS_MutedDSP":           &MutedDSP{},
	"STATUS_VolumeDSP":          &VolumeDSP{},
	"STATUS_Tiered_Switching":   &InputTieredSwitcher{},
}

func generateStandardStatusCommand(devices []structs.Device, evaluatorName string, commandName string) ([]StatusCommand, int, error) {

	var count int

	log.L.Infof("[statusevals] Generating status commands from %v", evaluatorName)
	var output []StatusCommand

	//iterate over each device
	for _, device := range devices {

		log.L.Infof("[statusevals] Considering device: %s", device.Name)

		for _, command := range device.Type.Commands {
			if strings.HasPrefix(command.ID, FLAG) && strings.Contains(command.ID, commandName) {
				log.L.Info("[statusevals] Command found")

				//every power command needs an address parameter
				parameters := make(map[string]string)
				parameters["address"] = device.Address

				//build destination device
				var destinationDevice base.DestinationDevice
				for _, role := range device.Roles {

					if role.ID == "AudioOut" {
						destinationDevice.AudioDevice = true
					}

					if role.ID == "VideoOut" {
						destinationDevice.Display = true
					}

				}

				destinationDevice.Device = device

				log.L.Infof("[statusevals] Adding command: %s to action list with device %s", command.ID, device.ID)
				output = append(output, StatusCommand{
					Action:            command,
					Device:            device,
					Parameters:        parameters,
					DestinationDevice: destinationDevice,
					Generator:         evaluatorName,
				})
				count++

			}

		}

	}
	return output, count, nil

}
