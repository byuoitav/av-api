package commandevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

// BlankDisplay is struct that implements the CommandEvaluation struct
type BlankDisplayDefault struct {
}

// Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (p *BlankDisplayDefault) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {
	log.Printf("Evaluating for BlankDisplay command.")

	actions := []base.ActionStructure{}

	// Check for room-wide blanking
	if room.Blanked != nil && *room.Blanked {
		log.Printf("Room-wide blank request received. Retrieving all devices.")

		// Get all devices
		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Room, room.Building, "VideoOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		log.Printf("Blanking all displays in room.")
		// Currently we only check for output devices
		for i := range devices {
			if devices[i].Output {
				log.Printf("Adding device %+v", devices[i].Name)

				actions = append(actions, base.ActionStructure{
					Action:              "BlankScreen",
					GeneratingEvaluator: "BlankDisplayDefault",
					Device:              devices[i],
					DeviceSpecific:      false,
				})
			}
		}
	}

	// Now we go through and check if blank was set for a specific device
	log.Printf("Evaluating individual displays for blanking.")

	for _, display := range room.Displays { // Loop through the provided displays array from the PUT body
		log.Printf("Adding device %+v", display.Name)

		if display.Blanked != nil && *display.Blanked { // If the user passed in a blanked state with the device

			device, err := dbo.GetDeviceByName(room.Building, room.Room, display.Name)
			if err != nil {
				return []base.ActionStructure{}, err
			}

			actions = append(actions, base.ActionStructure{
				Action:              "BlankScreen",
				GeneratingEvaluator: "BlankDisplayDefault",
				Device:              device,
				DeviceSpecific:      true,
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
	ok1, _ := checkCommands(action.Device.Commands, "BlankScreen")
	// Return an error if the BlankDisplay command doesn't exist or the command in question isn't a BlankDisplay command
	if !ok1 || !strings.EqualFold(action.Action, "BlankScreen") {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	log.Printf("Done.")
	return
}

// GetIncompatableCommands keeps track of actions that are incompatable (on the same device)
func (p *BlankDisplayDefault) GetIncompatableCommands() (incompatableActions []string) {
	incompatableActions = []string{
		"UnblankScreen"
	}

	return
}
