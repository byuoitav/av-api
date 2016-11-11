package helpers

import (
	"errors"
	"log"
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
		log.Printf("Evaluating command %s.", c.CommandName)

		curEvaluator := evaluators[c.CommandKey]
		if curEvaluator == nil {
			err = errors.New("No evaluator corresponding to key " + c.CommandKey)
			return
		}

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

			// Provide a map from the generating evaluator to the generated action in
			// case they want to use the Incompatable actions in the reconcilers.
			action.GeneratingEvaluator = c.CommandKey
			actionList = append(actionList, action)
		}
	}

	//Reconcile actions
	curReconciler := reconcilers[room.Configuration.RoomKey]
	if curReconciler == nil {
		err = errors.New("No reconciler corresponding to key " + room.Configuration.RoomKey)
		return
	}

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
