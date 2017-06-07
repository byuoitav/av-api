package commandevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

type UnBlankDisplayDefault struct {
}

//Evaluate creates UnBlank actions for the entire room and for individual devices
func (p *UnBlankDisplayDefault) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	var actions []base.ActionStructure
	eventInfo := eventinfrastructure.EventInfo{
		Type:           eventinfrastructure.CORESTATE,
		EventCause:     eventinfrastructure.USERINPUT,
		EventInfoKey:   "blanked",
		EventInfoValue: "false",
	}

	if room.Blanked != nil && !*room.Blanked {

		log.Printf("Room-wide UnBlank request received. Retrieving all devices.")

		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "VideoOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		log.Printf("Un-Blanking all displays in room.")

		for _, device := range devices {

			if device.Output {

				log.Printf("Adding Device %+v", device.Name)

				eventInfo.Device = device.Name
				actions = append(actions, base.ActionStructure{
					Action:              "UnblankDisplay",
					GeneratingEvaluator: "UnBlankDisplayDefault",
					Device:              device,
					DeviceSpecific:      false,
					EventLog:            []eventinfrastructure.EventInfo{eventInfo},
				})
			}

		}

	}

	log.Printf("Evaluating individial displays for unblanking.")

	for _, display := range room.Displays {

		log.Printf("Adding device %+v", display.Name)

		if display.Blanked != nil && !*display.Blanked {

			device, err := dbo.GetDeviceByName(room.Building, room.Room, display.Name)
			if err != nil {
				return []base.ActionStructure{}, err
			}

			eventInfo.Device = device.Name
			actions = append(actions, base.ActionStructure{
				Action:              "UnblankDisplay",
				GeneratingEvaluator: "UnBlankDisplayDefault",
				Device:              device,
				DeviceSpecific:      true,
				EventLog:            []eventinfrastructure.EventInfo{eventInfo},
			})

		}
	}

	log.Printf("Evaluation complete; %v actions generated.", len(actions))

	return actions, nil
}

//Validate returns an error if a command is invalid for a device
func (p *UnBlankDisplayDefault) Validate(action base.ActionStructure) error {
	log.Printf("Validating action for command \"UnBlank\"")

	ok, _ := CheckCommands(action.Device.Commands, "UnblankDisplay")

	if !ok || !strings.EqualFold(action.Action, "UnblankDisplay") {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	log.Printf("Done.")
	return nil
}

//GetIncompatibleCommands returns a string array containing commands incompatible with UnBlank Display
func (p *UnBlankDisplayDefault) GetIncompatibleCommands() (incompatibleActions []string) {
	incompatibleActions = []string{
		"BlankScreen",
	}

	return
}
