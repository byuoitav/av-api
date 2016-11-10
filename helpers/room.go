package helpers

import (
	"fmt"
	"strings"

	"github.com/byuoitav/av-api/actionReconcilers"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/commandEvaluators"
	"github.com/byuoitav/av-api/dbo"
)

//EditRoomState actually carries out the room state changes
func EditRoomState(roomInQuestion base.PublicRoom) (report []commandEvaluators.CommandExecutionReporting, err error) {

	//Initialize
	evaluators := commandEvaluators.Init()
	reconcilers := actionReconcilers.Init()

	//get our room
	room := dbo.GetRoomByInfo(roomInQuestion.Room, roomInQuestion.Building)

	actionList := []base.ActionStructure{}

	//for each command in the configuration, evaluate and validate.
	for c := range room.Configuration.Commands {
		curEvaluator := evaluators[c.CommandKey]
		subList := []string{}

		//Evaluate
		subList, err = curEvaluator.Evaluate(roomInQuestion)
		if err != nil {
			return
		}

		//Validate
		for action := range subList {
			if !curEvaluator.Validate(action) {
				err = errors.new(fmt.Sprintf("Could not validate action %+v", action))
				return
			}
			actionList = append(actionList, action)
		}

		return
	}

	//Reconcile actions
	curReconciler := reconcilers[room.Configuration.RoomKey]

	actionList, err = curReconciler.Reconcile(actionList)
	if err != nil {
		return
	}

	//execute actions.
	report, err = commands.ExecuteActions(actionList)

	return
}

/*
ReplaceIPAddressEndpoint is a simple helper
*/
func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}
