package commandevaluators

import (
	"errors"
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//ChangeVideoInputVideoswitcher the struct that implements the CommandEvaluation struct
type ChangeVideoInputVideoswitcher struct {
}

//Evaluate f
func (c *ChangeVideoInputVideoswitcher) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {
	actionList := []base.ActionStructure{}

	if len(room.CurrentVideoInput) != 0 {
		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		switcher, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoSwitcher")
		if err != nil {
			return []base.ActionStructure{}, err
		}
		if len(switcher) != 1 {
			return []base.ActionStructure{}, errors.New("too many switchers/none available")
		}

		for _, device := range devices {
			for _, port := range switcher[0].Ports {
				if port.Destination == device.Name && port.Source == room.CurrentVideoInput {
					m := make(map[string]string)
					m["port"] = port.Name

					tempAction := base.ActionStructure{
						Action:              "ChangeInput",
						GeneratingEvaluator: "ChangeVideoInputVideoswitcher",
						Device:              switcher[0],
						Parameters:          m,
						DeviceSpecific:      false,
						Overridden:          false,
					}

					actionList = append(actionList, tempAction)
					break
				}
			}
		}
	}

	// if there is at least one display
	if len(room.Displays) != 0 {

		// interate through all displays in the room, create an ActionStructure if it has an input
		for _, display := range room.Displays {

			// if the display has an input, create the action
			if len(display.Input) != 0 {

				tempAction := base.ActionStructure{
					Action:              "ChangeInput",
					GeneratingEvaluator: "ChangeVideoInputVideoswitcher",
					Device:              display, // or is it the switcher[0]?
					Parameters:          m,
					DeviceSpecific:      true,
					Overridden:          false,
				}
			}
		}

	}
	return actionList, nil
}

//GetSwitcherAndCreateAction f
func (c *ChangeVideoInputVideoswitcher) GetSwitcherAndCreateAction(room, device accessors.Device) {

}

//Validate f
func (c *ChangeVideoInputVideoswitcher) Validate(action base.ActionStructure) error {
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
func (c *ChangeVideoInputVideoswitcher) GetIncompatibleCommands() []string {
	return nil
}
