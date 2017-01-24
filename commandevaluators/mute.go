package commandevaluators

import (
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

//Mute implements CommandEvaluation
type Mute struct {
}

/*
 	Evalute takes a public room struct, scans the struct and builds any needed
	actions based on the contents of the struct.
*/
func (p *Mute) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {

	log.Printf("Evaluating mute command.")

	actions := []base.ActionStructure{}

	//scan the room struct
	if room.AudioDevices != nil {
		log.Printf("Audio device request recieved. Scanning all devices.")

		//generate commands
		for _, audioDevice := range room.AudioDevices {
			if audioDevice.Muted != nil {

				//get the device
				device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
				if err != nil {
					return []base.ActionStructure{}, err
				}

				//generate a command based on desired state
				action := ""

				if *audioDevice.Muted == true {
					//append command to mute device
					log.Printf("Muting %v", audioDevice.Name)
					action = "Mute"

				} else {
					//append command to un-mute device
					log.Printf("Un-muting %v", audioDevice.Name)
					action = "UnMute"
				}

				actions = append(actions, base.ActionStructure{
					Action:              action,
					GeneratingEvaluator: "Mute",
					Device:              device,
					DeviceSpecific:      true,
				})

			}

		}

	}

	return actions, nil

}

/*
	  Validate takes an action structure (for the command) and validates
		that the device and parameter are valid for the command.
*/
func (p *Mute) Validate(room base.ActionStructure) error {
	return nil
}

/*
   GetIncompatableActions returns a list of commands that are incompatable
   with this one (i.e. 'standby' and 'power on', or 'mute' and 'volume up')
*/
func (p *Mute) GetIncompatableCommands() []string {
	return nil
}
