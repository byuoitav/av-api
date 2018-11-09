package commandevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
)

// BlankDisplayDefault is struct that implements the CommandEvaluation struct
type BlankDisplayDefault struct {
}

// Evaluate verifies the information for a BlankDisplayDefault object and generates a list of actions based on the command.
func (p *BlankDisplayDefault) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	log.L.Info("[command_evaluators] Evaluating BlankDisplay commands...")

	var actions []base.ActionStructure

	//build event info
	event := events.Event{
		Key:   "blanked",
		Value: "true",
		User:  requestor,
	}

	event.EventTags = append(event.EventTags, events.CoreState, events.UserGenerated)

	// Check for room-wide blanking
	if room.Blanked != nil && *room.Blanked {
		log.L.Info("[command_evaluators] Room-wide blank request received. Retrieving all devices...")

		// Get all devices
		roomID := fmt.Sprintf("%v-%v", room.Building, room.Room)
		devices, err := db.GetDB().GetDevicesByRoomAndRole(roomID, "VideoOut")
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		log.L.Infof("[command_evaluators] VideoOut devices: %v\n", devices)

		log.L.Info("[command_evaluators] Assigning BlankDisplay commands...")
		// Currently we only check for output devices
		for _, device := range devices {

			if device.Type.Output {

				log.L.Infof("[command_evaluators] Adding device %v", device.Name)

				destination := base.DestinationDevice{
					Device:  device,
					Display: true,
				}

				if structs.HasRole(device, "AudioOut") {
					destination.AudioDevice = true
				}

				event.AffectedRoom = events.GenerateBasicRoomInfo(fmt.Sprintf("%s-%s", room.Building, room.Room))
				event.TargetDevice = events.GenerateBasicDeviceInfo(destination.ID)

				actions = append(actions, base.ActionStructure{
					Action:              "BlankDisplay",
					GeneratingEvaluator: "BlankDisplayDefault",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []events.Event{event},
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

							cmd := DX.GetCommandByName("BlankDisplay")

							if len(cmd.ID) == 0 || cmd.ID != "BlankDisplay" {
								continue
							}

							log.L.Info("[command_evaluators] Adding device %v", DX.Name)

							event.TargetDevice = events.GenerateBasicDeviceInfo(DX.ID)

							actions = append(actions, base.ActionStructure{
								Action:              "BlankDisplay",
								Device:              DX,
								DestinationDevice:   destination,
								GeneratingEvaluator: "BlankDisplayDefault",
								DeviceSpecific:      false,
								EventLog:            []events.Event{event},
							})
						}
					}
				}
				///// MIRROR STUFF /////
				////////////////////////
			}
		}
	}

	log.L.Info("[command_evaluators] Evaluating individual displays for blanking.")

	for _, display := range room.Displays {
		log.L.Infof("[command_evaluators] Adding device %v", display.Name)

		if display.Blanked != nil && *display.Blanked {

			// Retrieve device information from the database.
			deviceID := fmt.Sprintf("%v-%v-%v", room.Building, room.Room, display.Name)
			device, err := db.GetDB().GetDevice(deviceID)
			if err != nil {
				return []base.ActionStructure{}, 0, err
			}

			destination := base.DestinationDevice{
				Device:  device,
				Display: true,
			}

			event.AffectedRoom = events.GenerateBasicRoomInfo(fmt.Sprintf("%s-%s", room.Building, room.Room))
			event.TargetDevice = events.GenerateBasicDeviceInfo(destination.ID)

			if structs.HasRole(device, "AudioOut") {
				destination.AudioDevice = true
			}

			event.TargetDevice = events.GenerateBasicDeviceInfo(device.ID)

			actions = append(actions, base.ActionStructure{
				Action:              "BlankDisplay",
				GeneratingEvaluator: "BlankDisplayDefault",
				Device:              device,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []events.Event{event},
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

						cmd := DX.GetCommandByName("BlankDisplay")

						if cmd.ID != "BlankDisplay" {
							continue
						}

						log.L.Info("[command_evaluators] Adding mirror device %v", DX.Name)

						event.TargetDevice = events.GenerateBasicDeviceInfo(DX.ID)

						actions = append(actions, base.ActionStructure{
							Action:              "BlankDisplay",
							Device:              DX,
							DestinationDevice:   destination,
							GeneratingEvaluator: "BlankDisplayDefault",
							DeviceSpecific:      false,
							EventLog:            []events.Event{event},
						})
					}
				}
			}
			///// MIRROR STUFF /////
			////////////////////////
		}
	}

	log.L.Infof("[command_evaluators] %v actions generated.", len(actions))
	log.L.Info("[command_evaluators] Evaluation complete.")

	return actions, len(actions), nil
}

// Validate fulfills the Fulfill requirement on the command interface
func (p *BlankDisplayDefault) Validate(action base.ActionStructure) (err error) {
	log.L.Infof("[command_evaluators] Validating action for command %v", action.Action)

	// Check if the BlankDisplay command is a valid name of a command
	ok, _ := CheckCommands(action.Device.Type.Commands, "BlankDisplay")

	// Return an error if the BlankDisplay command doesn't exist or the command in question isn't a BlankDisplay command
	if !ok || !strings.EqualFold(action.Action, "BlankDisplay") {
		log.L.Errorf("[command_evaluators] ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	log.L.Info("[command_evaluators] Done.")
	return
}

// GetIncompatibleCommands keeps track of actions that are incompatable (on the same device)
func (p *BlankDisplayDefault) GetIncompatibleCommands() (incompatableActions []string) {
	incompatableActions = []string{
		"UnblankDisplay",
	}

	return
}
