package helpers

import (
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
	room, err := dbo.GetRoomByInfo(roomInQuestion.Room, roomInQuestion.Building)
	if err != nil {
		return
	}

	actionList := []base.ActionStructure{}

	//for each command in the configuration, evaluate and validate.
	for _, c := range room.Configuration.Commands {
		curEvaluator := evaluators[c.CommandKey]
		subList := []base.ActionStructure{}

		//Evaluate
		subList, err = curEvaluator.Evaluate(roomInQuestion)
		if err != nil {
			return
		}

		//Validate
		for _, action := range subList {
			err = curEvaluator.Validate(action)
			if err != nil {
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
	report, err = commandEvaluators.ExecuteActions(actionList)

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
