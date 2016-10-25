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

/*
	Note that is is important to add a command to this list and set the rules surounding that command (functionally mapping) property -> command
	here.
*/
func evaluateCommands(roomInQuestion base.PublicRoom) (actions []commands.ActionStructure, err error) {

	//getAllCommands
	log.Printf("Getting command orders.")
	rawCommands, err := dbo.GetAllRawCommands()

	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	//order commands by priority
	rawCommands = orderCommands(rawCommands)
	fmt.Printf("%+v", rawCommands)

	//Switch on each command.
	var tempActions []commands.ActionStructure
	for _, c := range rawCommands {

		//go through map and call evaluate, and then validate. Add them to action list.
		curCommand := commands.CommandMap[c.Name]

		tempActions, err = curCommand.Evaluate(roomInQuestion)
		if err != nil {
			return
		}

		err = curCommand.Validate(tempActions)
		if err != nil {
			return
		}

		actions = append(actions, tempActions...)
	}

	return
}

func orderCommands(commands []accessors.RawCommand) []accessors.RawCommand {
	sorter := accessors.CommandSorterByPriority{Commands: commands}
	sort.Sort(&sorter)
	return sorter.Commands
}

//EditRoomState actually carries out the room
func EditRoomState(roomInQuestion base.PublicRoom) error {

	//Initialize
	commands.Init()

	//Evaluate commands
	actions, err := evaluateCommands(roomInQuestion)
	if err != nil {
		return err
	}

	//Reconcile actions
	err = commands.ReconcileActions(&actions)
	if err != nil {
		return err
	}

	//execute actions.
	err = commands.ExecuteActions(actions)

	return err
}

/*
ReplaceIPAddressEndpoint is a simple helper
*/
func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}
