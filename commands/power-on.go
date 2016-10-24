package commands

import (
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//PowerOn is struct that implements the CommandEvaluation struct
type PowerOn struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (p *PowerOn) Evaluate(room base.PublicRoom) (actions []ActionStructure, err error) {
	var devices []accessors.Device
	if strings.EqualFold(room.Power, "on") {
		//Get all devices.
		devices, err = dbo.GetDevicesByRoom(room.Room, room.Building)
		if err != nil {
			return
		}

		//Currently we only check for output devices.
		for _, device := range devices {
			if device.Output {
				actions = append(actions, ActionStructure{Action: "PowerOn", Device: &device})
			}
		}
	}

	var dev *accessors.Device

	//now we go through and check if power 'on' was set for any other device.
	for _, device := range room.Displays {
		if strings.EqualFold(device.Power, "on") {
			//check if we already added it
			index := checkActionListForDevice(actions, device.Name, room.Room, room.Building)
			if index == -1 {

				//Get the device, check the list of already retreived devices first, if not there,
				//hit the DB up for it.
				dev, err = getDevice(devices, device.Name, room.Room, room.Building)
				if err != nil {
					return
				}
				actions = append(actions, ActionStructure{Action: "PowerOn", Device: dev})
			}
		}
	}

	for _, device := range room.AudioDevices {
		if strings.EqualFold(device.Power, "on") {
			//check if we already added it
			index := checkActionListForDevice(actions, device.Name, room.Room, room.Building)
			if index == -1 {

				//Get the device, check the list of already retreived devices first, if not there,
				//hit the DB up for it.
				dev, err = getDevice(devices, device.Name, room.Room, room.Building)
				if err != nil {
					return
				}
				actions = append(actions, ActionStructure{Action: "PowerOn", Device: dev})
			}
		}
	}
	return
}

//Validate fulfills the Fulfill requirement on the command interface
func (p *PowerOn) Validate(actions []ActionStructure) (b bool, err error) {

	return
}
