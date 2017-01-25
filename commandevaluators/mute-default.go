package commandevaluators

import (
	"log"

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

	log.Printf("Evaluating mute command.")

	//create array of type ActionStructure
	actions := []base.ActionStructure{}

	if room.Muted != nil {

		//general mute command
		log.Printf("Room-wide mute request detected. Retrieving all devices.")

		//create action string
		action := ""
		if *room.Muted == true {
			action = "Mute"
		} else {
			action = "UnMute"
		}

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
					Action:              action,
					GeneratingEvaluator: "MuteDefault",
					Device:              devices[i],
					DeviceSpecific:      false,
				})
			}
		}
	}

	//scan the room struct
	if len(room.AudioDevices) != 0 {
		log.Printf("Device-specific request recieved. Scanning devices.")

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
					GeneratingEvaluator: "MuteDefault",
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
func (p *MuteDefault) Validate(room base.ActionStructure) error {
	return nil
}

/*
   GetIncompatableActions returns a list of commands that are incompatable
   with this one (i.e. 'standby' and 'power on', or 'mute' and 'volume up')
*/
func (p *MuteDefault) GetIncompatableCommands() []string {
	return nil
}
