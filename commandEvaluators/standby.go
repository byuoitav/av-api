package commandEvaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//Standby is struct that implements the CommandEvaluation struct
type Standby struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (s *Standby) Evaluate(room base.PublicRoom) (actions []base.ActionStructure, err error) {
	log.Printf("Evaluating for Standby Command.")
	var devices []accessors.Device
	if strings.EqualFold(room.Power, "standby") {
		log.Printf("Room-wide power set. Retrieving all devices.")
		//Get all devices.
		devices, err = dbo.GetDevicesByRoom(room.Building, room.Room)
		if err != nil {
			return
		}
		log.Printf("Setting power to 'standby' state for all output devices.")
		//Currently we only check for output devices.
		for i := range devices {
			if devices[i].Output {
				log.Printf("Adding device %+v", devices[i].Name)
				actions = append(actions, base.ActionStructure{Action: "Standby", Device: devices[i], DeviceSpecific: false})
			}
		}
	}

	//now we go through and check if power 'standby' was set for any other device.
	for _, device := range room.Displays {
		log.Printf("Evaluating displays for command power standby. ")
		actions, err = s.evaluateDevice(device.Device, actions, devices, room.Room, room.Building)
		if err != nil {
			return
		}
	}

	for _, device := range room.AudioDevices {
		log.Printf("Evaluating audio devices for command power on. ")
		actions, err = s.evaluateDevice(device.Device, actions, devices, room.Room, room.Building)
		if err != nil {
			return
		}
	}
	log.Printf("%v actions generated.", len(actions))
	log.Printf("Evaluation complete.")

	return
}

//Validate fulfills the Fulfill requirement on the command interface
func (s *Standby) Validate(action base.ActionStructure) (err error) {
	log.Printf("Validating action for command Standby.")

	if ok, _ := checkCommands(action.Device.Commands, "Standby"); !ok || !strings.EqualFold(action.Action, "Standby") {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	log.Printf("Done.")
	return
}

//GetIncompatableCommands keeps track of actions that are incompatable (on the same device)
func (s *Standby) GetIncompatableCommands() (incompatableActions []string) {
	incompatableActions = []string{
		"PowerOn",
	}
	return
}

//Evaluate devices just pulls out the process we do with the audio-devices and
//displays into one function.
func (s *Standby) evaluateDevice(device base.Device,
	actions []base.ActionStructure,
	devices []accessors.Device,
	room string,
	building string) ([]base.ActionStructure, error) {

	//Check if we even need to start anything
	if strings.EqualFold(device.Power, "standby") {
		//check if we already added it
		index := checkActionListForDevice(actions, device.Name, room, building)
		if index == -1 {

			//Get the device, check the list of already retreived devices first, if not there,
			//hit the DB up for it.
			dev, err := getDevice(devices, device.Name, room, building)
			if err != nil {
				return actions, err
			}
			actions = append(actions, base.ActionStructure{Action: "Standby", Device: dev, DeviceSpecific: true})
		}
	}
	return actions, nil
}
