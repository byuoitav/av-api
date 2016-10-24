package helpers

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/commands"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//EditRoomStateNew is just a placeholder
func EditRoomStateNew(roomInQuestion base.PublicRoom) error {

	log.Printf("Room: %v\n", roomInQuestion)

	//Evaluate commands
	evaluateCommands(roomInQuestion)
	return nil
}

/*
	Note that is is important to add a command to this list and set the rules surounding that command (functionally mapping) property -> command
	here.
*/
func evaluateCommands(roomInQuestion base.PublicRoom) (actions []commands.ActionStructure, err error) {

	//getAllCommands
	log.Printf("Getting command orders.")
	commands, err := dbo.GetAllRawCommands()

	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	//order commands by priority
	commands = orderCommands(commands)
	fmt.Printf("%+v", commands)
	//Switch on each command.

	for _, c := range commands {
		switch c.Name {

		}
	}

	return
}

func orderCommands(commands []accessors.RawCommand) []accessors.RawCommand {
	sorter := accessors.CommandSorterByPriority{Commands: commands}
	sort.Sort(&sorter)
	return sorter.Commands
}

//EditRoomState actually carries out the room
func EditRoomState(roomInQuestion base.PublicRoom, building string, room string) error {

	return nil
}

/*
ReplaceIPAddressEndpoint is a simple helper
*/
func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}
