package commandevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

/*

	Video Switchers are a little different by way of port identification.
	Basically our ports are combinations of input + output

	so 0:0 is the input zero set to output zero. So here we run a split on the ':' to assign the input and output separately.

*/

//ChangeVideoInputVideoswitcher the struct that implements the CommandEvaluation struct
type ChangeVideoInputVideoSwitcher struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement
func (c *ChangeVideoInputVideoSwitcher) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {
	actionList := []base.ActionStructure{}

	if len(room.CurrentVideoInput) != 0 {
		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		for _, device := range devices {
			action, err := GetSwitcherAndCreateAction(room, device, room.CurrentVideoInput, "ChangeVideoInputVideoSwitcher")
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

				action, err := GetSwitcherAndCreateAction(room, device, display.Input, "ChangeVideoInputVideoSwitcher")
				if err != nil {
					return []base.ActionStructure{}, err
				}
				//Undecode the format into the
				actionList = append(actionList, action)
			}
		}

	}

	for _, action := range actionList {
		p := action.Parameters["output"]
		splitP := strings.Split(p, ":")

		if len(splitP) != 2 {
			return actionList, errors.New("Invalid port for a video switcher")
		}

		action.Parameters["input"] = splitP[0]
		action.Parameters["output"] = splitP[1]
	}
	return actionList, nil
}

//GetSwitcherAndCreateAction gets the videoswitcher in a room, matches the destination port to the new port
// and creates an action
func GetSwitcherAndCreateAction(room base.PublicRoom, device accessors.Device, selectedInput string, generatingEvaluator string) (base.ActionStructure, error) {

	switcher, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoSwitcher")
	if err != nil {
		return base.ActionStructure{}, err
	}

	if len(switcher) != 1 {
		return base.ActionStructure{}, errors.New("too many switchers/none available")
	}

	log.Printf("Evaluating device %s for a port connecting %s to %s", switcher[0].GetFullName(), selectedInput, device.GetFullName())
	for _, port := range switcher[0].Ports {

		if port.Destination == device.Name && port.Source == selectedInput {

			m := make(map[string]string)
			m["output"] = port.Name

			eventInfo := eventinfrastructure.EventInfo{
				Type:           eventinfrastructure.USERACTION,
				EventCause:     eventinfrastructure.USERINPUT,
				Device:         switcher[0].Name,
				EventInfoKey:   "input",
				EventInfoValue: m["output"],
			}

			tempAction := base.ActionStructure{
				Action:              "ChangeInput",
				GeneratingEvaluator: generatingEvaluator,
				Device:              switcher[0],
				Parameters:          m,
				DeviceSpecific:      false,
				Overridden:          false,
				EventLog:            []eventinfrastructure.EventInfo{eventInfo},
			}

			return tempAction, nil
		}
	}

	return base.ActionStructure{}, errors.New("no switcher found with the matching port")
}

//Validate f
func (c *ChangeVideoInputVideoSwitcher) Validate(action base.ActionStructure) error {
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
func (c *ChangeVideoInputVideoSwitcher) GetIncompatibleCommands() []string {
	return nil
}
