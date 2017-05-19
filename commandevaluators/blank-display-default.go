package commandevaluators

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

// BlankDisplay is struct that implements the CommandEvaluation struct
type BlankDisplayDefault struct {
}

// Takes a PublicRoom and builds a slice of ActionStructures
func (p *BlankDisplayDefault) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Evaluating for BlankDisplay command.")

	var actions []base.ActionStructure

	// Get all devices
	devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoOut")
	if err != nil {
		return []base.ActionStructure{}, err
	}

	//build event info
	eventInfo := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.USERACTION,
		EventCause:     eventinfrastructure.USERINPUT,
		EventInfoKey:   "blanked",
		EventInfoValue: "true",
	}

	// Check for room-wide blanking
	if room.Blanked != nil && *room.Blanked {
		log.Printf("Room-wide blank request received. Retrieving all devices.")

		fmt.Printf("VideoOut devices: %+v\n", devices)

		log.Printf("Assigning BlankDisplayCommands")
		// Currently we only check for output devices
		for _, device := range devices {
			if device.Output {
				log.Printf("Adding device %+v", device.Name)

				eventInfo.Device = device.Name
				actions = append(actions, base.ActionStructure{
					Action:              "BlankDisplay",
					GeneratingEvaluator: "BlankDisplayDefault",
					Device:              device,
					DeviceSpecific:      false,
					EventLog:            []eventinfrastructure.EventInfo{eventInfo},
				})
			}
		}
	}

	log.Printf("Evaluating individual displays for blanking.")

	for _, display := range room.Displays {
		log.Printf("Adding device %+v", display.Name)

		if display.Blanked != nil && *display.Blanked {

			device, err := dbo.GetDeviceByName(room.Building, room.Room, display.Name)
			if err != nil {
				return []base.ActionStructure{}, err
			}

			eventInfo.Device = device.Name
			actions = append(actions, base.ActionStructure{
				Action:              "BlankDisplay",
				GeneratingEvaluator: "BlankDisplayDefault",
				Device:              device,
				DeviceSpecific:      true,
				EventLog:            []eventinfrastructure.EventInfo{eventInfo},
			})
		}
	}

	log.Printf("%v actions generated.", len(actions))
	log.Printf("Evaluation complete.")

	return actions, nil
}

// Validate fulfills the Fulfill requirement on the command interface
func (p *BlankDisplayDefault) Validate(action base.ActionStructure) (err error) {
	log.Printf("Validating action for command %v", action.Action)

	// Check if the BlankDisplay command is a valid name of a command
	ok, _ := CheckCommands(action.Device.Commands, "BlankDisplay")
	// Return an error if the BlankDisplay command doesn't exist or the command in question isn't a BlankDisplay command
	if !ok || !strings.EqualFold(action.Action, "BlankDisplay") {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	log.Printf("Done.")
	return
}

// GetIncompatibleCommands keeps track of actions that are incompatable (on the same device)
func (p *BlankDisplayDefault) GetIncompatibleCommands() (incompatableActions []string) {
	incompatableActions = []string{
		"UnblankDisplay",
	}

	return
}
