package commandevaluators

import (
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

//This file contains common 'helper' functions.

//Checks an action list to see if it has a device (by name, room, and building) already in it,
//if so, it returns the index of the device, if not -1.
func checkActionListForDevice(a []base.ActionStructure, d string, room string, building string) (index int) {
	for i, curDevice := range a {
		if checkDevicesEqual(curDevice.Device, d, room, building) {
			return i
		}
	}
	return -1
}

func checkDevicesEqual(dev structs.Device, name string, room string, building string) bool {
	splits := strings.Split(dev.GetDeviceRoomID(), "-")
	return strings.EqualFold(dev.ID, name) &&
		strings.EqualFold(splits[1], room) &&
		strings.EqualFold(splits[0], building)
}

// CheckCommands searches a list of Commands to see if it contains any command by the name given.
// returns T/F, as well as the command if true.
func CheckCommands(commands []structs.Command, commandName string) (bool, structs.Command) {
	for _, c := range commands {
		if strings.EqualFold(c.ID, commandName) {
			return true, c
		}
	}
	return false, structs.Command{}
}

func markAsOverridden(action base.ActionStructure, structs ...[]*base.ActionStructure) {
	for i := 0; i < len(structs); i++ {
		for j := 0; j < len(structs[i]); j++ {
			if structs[i][j].Equals(action) {
				structs[i][j].Overridden = true
			}
		}
	}
}

// FindDevice searches a list of devices for the device specified by the given ID and returns it
func FindDevice(deviceList []structs.Device, searchID string) structs.Device {
	for i := range deviceList {
		if deviceList[i].ID == searchID {
			return deviceList[i]
		}
	}

	return structs.Device{}
}

// FilterDevicesByRole searches a list of devices for the devices that have the given roles, and returns a new list of those devices
func FilterDevicesByRole(deviceList []structs.Device, roleID string) []structs.Device {
	var toReturn []structs.Device

	for _, device := range deviceList {
		if device.HasRole(roleID) {
			toReturn = append(toReturn, device)
		}
	}

	return toReturn
}
