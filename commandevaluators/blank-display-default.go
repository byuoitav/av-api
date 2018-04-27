package commandevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

// BlankDisplay is struct that implements the CommandEvaluation struct
type BlankDisplayDefault struct {
}

// Takes a PublicRoom and builds a slice of ActionStructures
func (p *BlankDisplayDefault) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	base.Log("[command_evaluators] evaluating BlankDisplay commands...")

	var actions []base.ActionStructure

	//build event info
	eventInfo := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.CORESTATE,
		EventCause:     eventinfrastructure.USERINPUT,
		EventInfoKey:   "blanked",
		EventInfoValue: "true",
		Requestor:      requestor,
	}

	// Check for room-wide blanking
	if room.Blanked != nil && *room.Blanked {
		base.Log("[command_evaluators] room-wide blank request received. Retrieving all devices...")

		// Get all devices
		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoOut")
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		fmt.Printf("VideoOut devices: %+v\n", devices)

		base.Log("[command_evaluators] assigning BlankDisplay commands...")
		// Currently we only check for output devices
		for _, device := range devices {

			if device.Output {

				base.Log("[command_evaluators]Adding device %+v", device.Name)

				destination := base.DestinationDevice{
					Device:  device,
					Display: true,
				}

				if device.HasRole("AudioOut") {
					destination.AudioDevice = true
				}

				eventInfo.Device = device.Name
				actions = append(actions, base.ActionStructure{
					Action:              "BlankDisplay",
					GeneratingEvaluator: "BlankDisplayDefault",
					Device:              device,
					DestinationDevice:   destination,
					DeviceSpecific:      false,
					EventLog:            []eventinfrastructure.EventInfo{eventInfo},
				})
			}
		}
	}

	base.Log("[command_evaluators]Evaluating individual displays for blanking.")

	for _, display := range room.Displays {
		base.Log("[command_evaluators]Adding device %+v", display.Name)

		if display.Blanked != nil && *display.Blanked {

			device, err := dbo.GetDeviceByName(room.Building, room.Room, display.Name)
			if err != nil {
				return []base.ActionStructure{}, 0, err
			}

			destination := base.DestinationDevice{
				Device:  device,
				Display: true,
			}

			if device.HasRole("AudioOut") {
				destination.AudioDevice = true
			}

			eventInfo.Device = device.Name
			actions = append(actions, base.ActionStructure{
				Action:              "BlankDisplay",
				GeneratingEvaluator: "BlankDisplayDefault",
				Device:              device,
				DestinationDevice:   destination,
				DeviceSpecific:      true,
				EventLog:            []eventinfrastructure.EventInfo{eventInfo},
			})
		}
	}

	base.Log("[command_evaluators]%v actions generated.", len(actions))
	base.Log("[command_evaluators]Evaluation complete.")

	return actions, len(actions), nil
}

// Validate fulfills the Fulfill requirement on the command interface
func (p *BlankDisplayDefault) Validate(action base.ActionStructure) (err error) {
	base.Log("[command_evaluators] validating action for command %v", action.Action)

	// Check if the BlankDisplay command is a valid name of a command
	ok, _ := CheckCommands(action.Device.Commands, "BlankDisplay")
	// Return an error if the BlankDisplay command doesn't exist or the command in question isn't a BlankDisplay command
	if !ok || !strings.EqualFold(action.Action, "BlankDisplay") {
		base.Log("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	base.Log("[command_evaluators] Done.")
	return
}

// GetIncompatibleCommands keeps track of actions that are incompatable (on the same device)
func (p *BlankDisplayDefault) GetIncompatibleCommands() (incompatableActions []string) {
	incompatableActions = []string{
		"UnblankDisplay",
	}

	return
}
