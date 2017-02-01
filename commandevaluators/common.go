package commandevaluators

import (
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/configuration-database-microservice/accessors"
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

func checkDevicesEqual(dev accessors.Device, name string, room string, building string) bool {
	return strings.EqualFold(dev.Name, name) &&
		strings.EqualFold(dev.Room.Name, room) &&
		strings.EqualFold(dev.Building.Shortname, building)
}

func CheckCommands(commands []accessors.Command, commandName string) (bool, accessors.Command) {
	for _, c := range commands {
		if strings.EqualFold(c.Name, commandName) {
			return true, c
		}
	}
	return false, accessors.Command{}
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
