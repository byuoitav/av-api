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
	log.Printf("Evaluating for BlankDisplay Command.")

	actions := []base.ActionStructure{}

	// Check for room-wide blanking
	if room.Blanked != nil {
		log.Printf("Room-wide blank request received. Retrieving all devices.")

		// Set the action (blank or unblank) based on the current status of the room
		blankActionFromCurrentStatus := ""
		if *room.Blanked == true {
			blankActionFromCurrentStatus = "BlankDisplay"
		} else {
			blankActionFromCurrentStatus = "UnBlankDisplay"
		}

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
					Action:              blankActionFromCurrentStatus,
					GeneratingEvaluator: "BlankDisplayDefault",
					Device:              devices[i],
					DeviceSpecific:      false,
				})
			}
		}
	}

	// Now we go through and check if blank was set for a specific device
	for _, device := range room.Displays { // Loop through the provided displays array from the PUT body
		log.Printf("Evaluating individual displays for blanking.")
		log.Printf("Adding device %+v", device.Name)

		blankActionFromCurrentStatus := ""

		if device.Blanked != nil { // If the user passed in a blanked state with the device
			if *device.Blanked == true {
				blankActionFromCurrentStatus = "BlankDisplay"
			} else {
				blankActionFromCurrentStatus = "UnBlankDisplay"
			}

			dev, err := dbo.GetDeviceByName(room.Building, room.Room, device.Name)
			if err != nil {
				return []base.ActionStructure{}, err
			}

			actions = append(actions, base.ActionStructure{
				Action:              blankActionFromCurrentStatus,
				GeneratingEvaluator: "BlankDisplayDefault",
				Device:              dev,
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
	log.Printf("Validating action for comand BlankDisplay")

	// Check if the BlankDisplay command is a valid name of a command
	ok, _ := checkCommands(action.Device.Commands, "BlankDisplay")
	// Return an error if the BlankDisplay command doesn't exist or the command in question isn't a BlankDisplay command
	if !ok || !strings.EqualFold(action.Action, "BlankDisplay") {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	log.Printf("Done.")
	return
}

// GetIncompatableCommands keeps track of actions that are incompatable (on the same device)
func (p *BlankDisplayDefault) GetIncompatableCommands() (incompatableActions []string) {
	return // Just return because there are no actions incompatible with BlankDisplay
}
