package commandevaluators

import (
	"errors"
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

/*
	With tiered switchers we basically have to build a connection 'graph' and then traverse that graph to get all of the command necessary to fulfil a path from source to destination.
*/

//ChangeVideoInputVideoswitcher the struct that implements the CommandEvaluation struct
type ChagneVideoInputTieredSwitchers struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement
func (c *ChangeVideoInputTieredSwitchers) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, error) {
	//so first we need to go through and see if anyone even wants a piece of us, is there an 'input' field that isn't empty.

	has := (len(room.CurrentVideoInput) > 0) 
	for d := range room.Displays {
		if len(room.Displays[d].Input) != 0 {
			has = true
			break
		}
	}
	if !has {
		//there's nothing to do in the room
		return []base.ActionStructure, nil
	}


	//build the graph

	//get all the devices from the room
	devices, err := dbo.GetDevicesByRoom(room.Building, room.Room)
	if err != nil {

		log.Printf(color.FgHiRedString("There was an issue getting the devices from the room: %v", err.Error()))
		return []base.ActionStructure{}, err
	}

	graph, err := inputGraph.buildInputGraph(devicies)


	return []base.ActionStructure{}, nil

}

//GetSwitcherAndCreateAction gets the videoswitcher in a room, matches the destination port to the new port
// and creates an action
func GetSwitcherAndCreateAction(room base.PublicRoom, device structs.Device, selectedInput, generatingEvaluator, requestor string) (base.ActionStructure, error) {

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
				Type:           eventinfrastructure.CORESTATE,
				EventCause:     eventinfrastructure.USERINPUT,
				Device:         device.Name,
				EventInfoKey:   "input",
				EventInfoValue: selectedInput,
				Requestor:      requestor,
			}

			destination := statusevaluators.DestinationDevice{
				Device: device,
			}

			if device.HasRole("AudioOut") {
				destination.AudioDevice = true
			}

			if device.HasRole("VideoOut") {
				destination.Display = true
			}

			tempAction := base.ActionStructure{
				Action:              "ChangeInput",
				GeneratingEvaluator: generatingEvaluator,
				Device:              switcher[0],
				DestinationDevice:   destination,
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
func (c *ChangeVideoInputTieredSwitchers) Validate(action base.ActionStructure) error {
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
func (c *ChangeVideoInputTieredSwitchers) GetIncompatibleCommands() []string {
	return nil
}
