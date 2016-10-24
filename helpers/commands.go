package helpers

import (
	"strings"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//ActionStructure is the internal struct we use to pass commands around once
//they've been evaluated.
type ActionStructure struct {
	Action    string            `json:"action"`
	Device    *accessors.Device `json:"device"`
	Parameter string            `json:"parameter"`
}

/*
CommandEvaluation is an interface that must be implemented for each command to be
evaluated.
*/
type CommandEvaluation interface {
	/*
		 	Evalute takes a public room struct, scans the struct and builds any needed
			actions based on the contents of the struct.
	*/
	Evaluate(PublicRoom) ([]ActionStructure, error)
	/*
		  Validate takes an action structure (for the command) and validates that the
			device and parameter are valid for the comamnd.
	*/
	Validate(ActionStructure) (bool, error)
	/*
			   GetIncompatableActions returns A list of commands that are incompatable
		     with this one (i.e. 'standby' and 'power on', or 'mute' and 'volume up')
	*/
	GetIncompatableActions() []string
}

//PowerOn is struct that implements the CommandEvaluation struct
type PowerOn struct {
}

//Evaluate fulfills the CommmandEvaluation evaluate requirement.
func (p *PowerOn) Evaluate(room PublicRoom) (actions []ActionStructure, err error) {
	var devices []accessors.Device
	if strings.EqualFold(room.Power, "on") {
		//Get all devices.
		devices, err = getDevicesByRoom(room.Room, room.Building)
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

func getDevice(devs []accessors.Device, d string, room string, building string) (dev *accessors.Device, err error) {
	for i, curDevice := range devs {
		if checkDevicesEqual(&curDevice, d, room, building) {
			dev = &devs[i]
			return
		}
	}
	var device accessors.Device

	device, err = GetDeviceByName(room, building, d)
	if err != nil {
		return
	}
	dev = &device
	return
}

//Checks an action list to see if it has a device (by name, room, and building) already in it,
//if so, it returns the index of the device, if not -1.
func checkActionListForDevice(a []ActionStructure, d string, room string, building string) (index int) {
	for i, curDevice := range a {
		if checkDevicesEqual(curDevice.Device, d, room, building) {
			return i
		}
	}
	return -1
}

func checkDevicesEqual(dev *accessors.Device, name string, room string, building string) bool {
	return strings.EqualFold(dev.Name, name) &&
		strings.EqualFold(dev.Room.Name, room) &&
		strings.EqualFold(dev.Building.Shortname, building)
}
