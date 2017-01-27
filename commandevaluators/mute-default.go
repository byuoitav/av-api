package commandevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

//MuteDefault implements CommandEvaluation
type MuteDefault struct {
}

/*
 	Evalute takes a public room struct, scans the struct and builds any needed
	actions based on the contents of the struct.
*/
func (p *MuteDefault) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Evaluating for Mute command.")

	//create array of type ActionStructure
	actions := []base.ActionStructure{}

	if room.Muted != nil && *room.Muted {

		//general mute command
		log.Printf("Room-wide Mute request recieved. Retrieving all devices.")

		//get all devices
		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Room, room.Building, "AudioOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		log.Printf("Muting all devices in room.")

		for i := range devices {
			if devices[i].Output {
				log.Printf("Adding device %+v", devices[i].Name)

				actions = append(actions, base.ActionStructure{
					Action:              "Mute",
					GeneratingEvaluator: "MuteDefault",
					Device:              devices[i],
					DeviceSpecific:      false,
				})
			}
		}
	}

	//scan the room struct
	log.Printf("Evaluating audio devices for Mute command.")

	//generate commands
	for _, audioDevice := range room.AudioDevices {
		if audioDevice.Muted != nil && *audioDevice.Muted {

			//get the device
			device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
			if err != nil {
				return []base.ActionStructure{}, err
			}

			actions = append(actions, base.ActionStructure{
				Action:              "Mute",
				GeneratingEvaluator: "MuteDefault",
				Device:              device,
				DeviceSpecific:      true,
			})

		}

	}

	return actions, nil

}

// Validate takes an ActionStructure and determines if the command and parameter are valid for the device specified
func (p *MuteDefault) Validate(action base.ActionStructure) error {

	log.Printf("Validating mute action for command \"UnMute\".")

	ok, _ := checkCommands(action.Device.Commands, "UnMute")

	if !ok || !strings.EqualFold(action.Action, "UnMute") {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	log.Printf("Done.")

	return nil
}

//  GetIncompatableActions returns a list of commands that are incompatabl with this one (i.e. 'standby' and 'power on', or 'mute' and 'volume up')
func (p *MuteDefault) GetIncompatableCommands() (incompatibleActions []string) {

	incompatibleActions = []string{
		"UnMute",
	}

	return
}
