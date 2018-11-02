package commandevaluators

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/byuoitav/common/log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
)

// SetVolumeDefault implements the CommandEvaluation struct.
type SetVolumeDefault struct {
}

//Evaluate checks for a volume for the entire room or the volume of a specific device
func (*SetVolumeDefault) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	var actions []base.ActionStructure

	roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)

	eventInfo := events.Event{
		Key:  "volume",
		User: requestor,
	}

	eventInfo.AddToTags(events.CoreState, events.UserGenerated)

	destination := base.DestinationDevice{
		AudioDevice: true,
	}

	// general room volume
	if room.Volume != nil {

		log.L.Info("[command_evaluators] General volume request detected.")

		devices, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "AudioOut")
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		for _, device := range devices {

			if device.Type.Output {

				parameters := make(map[string]string)
				parameters["level"] = fmt.Sprintf("%v", *room.Volume)

				eventInfo.Value = fmt.Sprintf("%v", *room.Volume)

				eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

				eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(device.ID)

				destination.Device = device

				if structs.HasRole(device, "VideoOut") {
					destination.Display = true
				}

				actions = append(actions, base.ActionStructure{
					Action:              "SetVolume",
					Parameters:          parameters,
					GeneratingEvaluator: "SetVolumeDefault",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []events.Event{eventInfo},
				})

				////////////////////////
				///// MIRROR STUFF /////
				if structs.HasRole(device, "MirrorMaster") {
					for _, port := range device.Ports {
						if port.ID == "mirror" {
							DX, err := db.GetDB().GetDevice(port.DestinationDevice)
							if err != nil {
								return []base.ActionStructure{}, 0, err
							}

							cmd := DX.GetCommandByName("SetVolume")
							if len(cmd.ID) < 1 {
								continue
							}

							eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

							eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(DX.ID)

							log.L.Info("[command_evaluators] Adding mirror device %+v", DX.Name)

							actions = append(actions, base.ActionStructure{
								Action:              "SetVolume",
								Parameters:          parameters,
								GeneratingEvaluator: "SetVolumeDefault",
								Device:              DX,
								DestinationDevice:   destination,
								DeviceSpecific:      false,
								EventLog:            []events.Event{eventInfo},
							})
						}
					}
				}
				///// MIRROR STUFF /////
				////////////////////////
			}

		}

	}

	//identify devices in request body
	if len(room.AudioDevices) != 0 {

		log.L.Info("[command_evaluators] Device specific request detected. Scanning devices")

		for _, audioDevice := range room.AudioDevices {
			// create actions based on request

			if audioDevice.Volume != nil {
				log.L.Info("[command_evaluators] Adding device %+v", audioDevice.Name)

				deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, audioDevice.Name)
				device, err := db.GetDB().GetDevice(deviceID)
				if err != nil {
					return []base.ActionStructure{}, 0, err
				}

				parameters := make(map[string]string)
				parameters["level"] = fmt.Sprintf("%v", *audioDevice.Volume)
				log.L.Info("[command_evaluators] %+v", parameters)

				eventInfo.Value = fmt.Sprintf("%v", *audioDevice.Volume)

				eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

				eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(device.ID)

				destination.Device = device

				if structs.HasRole(device, "VideoOut") {
					destination.Display = true
				}

				actions = append(actions, base.ActionStructure{
					Action:              "SetVolume",
					GeneratingEvaluator: "SetVolumeDefault",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      true,
					Parameters:          parameters,
					EventLog:            []events.Event{eventInfo},
				})

				////////////////////////
				///// MIRROR STUFF /////
				if structs.HasRole(device, "MirrorMaster") {
					for _, port := range device.Ports {
						if port.ID == "mirror" {
							DX, err := db.GetDB().GetDevice(port.DestinationDevice)
							if err != nil {
								return actions, len(actions), err
							}

							cmd := DX.GetCommandByName("SetVolume")
							if len(cmd.ID) < 1 {
								continue
							}

							eventInfo.AffectedRoom = events.GenerateBasicRoomInfo(roomID)

							eventInfo.TargetDevice = events.GenerateBasicDeviceInfo(DX.ID)

							log.L.Info("[command_evaluators] Adding mirror device %+v", DX.Name)

							actions = append(actions, base.ActionStructure{
								Action:              "SetVolume",
								GeneratingEvaluator: "SetVolumeDefault",
								Device:              DX,
								DestinationDevice:   destination,
								DeviceSpecific:      true,
								Parameters:          parameters,
								EventLog:            []events.Event{eventInfo},
							})
						}
					}
				}
				///// MIRROR STUFF /////
				////////////////////////
			}

		}

	}

	log.L.Infof("[command_evaluators] %v actions generated.", len(actions))
	log.L.Info("[command_evaluators] Evaluation complete.")

	return actions, len(actions), nil
}

func validateSetVolumeMaxMin(action base.ActionStructure, maximum int, minimum int) error {
	level, err := strconv.Atoi(action.Parameters["level"])
	if err != nil {
		return err
	}

	if level > maximum || level < minimum {
		msg := fmt.Sprintf("[command_evaluators] ERROR. %v is an invalid volume level for %s", action.Parameters["level"], action.Device.Name)
		log.L.Error(msg)
		return errors.New(msg)
	}
	return nil
}

//Validate returns an error if the volume is greater than 100 or less than 0
func (p *SetVolumeDefault) Validate(action base.ActionStructure) error {
	maximum := 100
	minimum := 0

	return validateSetVolumeMaxMin(action, maximum, minimum)

}

//GetIncompatibleCommands returns a string array of commands incompatible with setting the volume
func (p *SetVolumeDefault) GetIncompatibleCommands() []string {
	return nil
}
