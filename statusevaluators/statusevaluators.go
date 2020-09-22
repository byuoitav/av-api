package statusevaluators

import (
	"strings"

	"github.com/byuoitav/common/db"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// StatusEvaluator defines the common functions for all StatusEvaluators.
type StatusEvaluator interface {
	//Generates action list
	GenerateCommands(room structs.Room) ([]StatusCommand, int, error)

	//Evaluate Response
	EvaluateResponse(room structs.Room, label string, value interface{}, Source structs.Device, Destination base.DestinationDevice) (string, interface{}, error)
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
			if strings.HasPrefix(command.ID, FLAG) && strings.EqualFold(command.ID, commandName) {
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

				////////////////////////
				///// MIRROR STUFF /////
				if structs.HasRole(device, "MirrorMaster") {
					for _, port := range device.Ports {
						if port.ID == "mirror" {
							DX, err := db.GetDB().GetDevice(port.DestinationDevice)
							if err != nil {
								return output, count, err
							}

							cmd := DX.GetCommandByID(commandName)
							if len(cmd.ID) < 1 {
								continue
							}

							params := make(map[string]string)
							params["address"] = DX.Address

							destDevice := base.DestinationDevice{
								Device: DX,
							}

							log.L.Infof("[statusevals] Adding command: %s to action list with device %s", cmd.ID, DX.ID)
							output = append(output, StatusCommand{
								Action:            cmd,
								Device:            DX,
								Parameters:        params,
								DestinationDevice: destDevice,
								Generator:         evaluatorName,
							})
							count++
						}
					}
				}
				///// MIRROR STUFF /////
				////////////////////////
			}

		}

	}

	return output, count, nil

}

// FindDevice searches a list of devices for the device specified by the given ID and returns it
func FindDevice(deviceList []structs.Device, searchID string) structs.Device {
	for i := range deviceList {
		if deviceList[i].ID == searchID {
			return deviceList[i]
		}
	}

	return structs.Device{}
}

// FilterDevicesByRole searches a list of devices for the devices that have the given roles, and returns a new list of those devices
func FilterDevicesByRole(deviceList []structs.Device, roleID string) []structs.Device {
	var toReturn []structs.Device

	for _, device := range deviceList {
		if device.HasRole(roleID) {
			toReturn = append(toReturn, device)
		}
	}

	return toReturn
}
