package commandevaluators

import (
	"errors"
	"fmt"
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

//ChangeVideoInputVideoswitcher the struct that implements the CommandEvaluation struct
type ChangeVideoInputDMPS struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement
func (c *ChangeVideoInputDMPS) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {
	actionList := []base.ActionStructure{}

	if len(room.CurrentVideoInput) != 0 {
		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		for _, device := range devices {
			action, err := GetSwitcherAndCreateAction(room, device, room.CurrentVideoInput, "ChangeVideoInputDMPS")
			if err != nil {
				return []base.ActionStructure{}, err
			}
			actionList = append(actionList, action)
		}
	}

	// if there is at least one display
	if len(room.Displays) != 0 {

		// interate through all displays in the room, create an ActionStructure if it has an input
		for _, display := range room.Displays {

			// if the display has an input, create the action
			if len(display.Input) != 0 {
				device, err := dbo.GetDeviceByName(room.Building, room.Room, display.Name)
				if err != nil {
					return []base.ActionStructure{}, err
				}

				action, err := GetSwitcherAndCreateAction(room, device, display.Input, "ChangeVideoInputDMPS")
				if err != nil {
					return []base.ActionStructure{}, err
				}

				fmt.Printf("Adding device signal name for DMPS specific Video Switching")
				deviceSignalName := ""
				switch display.Name {
				case "D1":
					deviceSignalName = "Display1_Select"
					break
				case "D2":
					deviceSignalName = "Display2_Select"
					break
				case "D3":
					deviceSignalName = "Display3_Select"
					break

				}

				fmt.Printf("Signal name to add: %s", deviceSignalName)

				action.Parameters["input"] = deviceSignalName
				actionList = append(actionList, action)
			}
		}

	}
	return actionList, nil
}

//Validate f
func (c *ChangeVideoInputDMPS) Validate(action base.ActionStructure) error {
	log.Printf("Validating action for command %v", action.Action)

	// check if ChangeInput is a valid name of a command (ok is a bool)
	ok, _ := CheckCommands(action.Device.Commands, "ChangeInput")

	// returns and error if the ChangeInput command doesn't exist or if the command isn't ChangeInput
	if !ok || action.Action != "ChangeInput" {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + "is not an invalid command for " + action.Device.Name)
	}

	log.Print("done.")
	return nil
}

//GetIncompatibleCommands f
func (c *ChangeVideoInputDMPS) GetIncompatibleCommands() []string {
	return nil
}
