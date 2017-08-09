package helpers

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/actionreconcilers"
	"github.com/byuoitav/av-api/base"
	ce "github.com/byuoitav/av-api/commandevaluators"
	"github.com/byuoitav/av-api/dbo"
)

//EditRoomState actually carries out the room state changes
func EditRoomState(roomInQuestion base.PublicRoom) (report base.PublicRoom, err error) {

	//Initialize map of strings to commandevaluators
	evaluators := ce.Init()
	reconcilers := actionreconcilers.Init()

	//get accessors.Room (as it exists in the database)
	room, err := dbo.GetRoomByInfo(roomInQuestion.Building, roomInQuestion.Room)
	if err != nil {
		return
	}

	var actions []base.ActionStructure

	//for each command in the configuration, evaluate and validate.
	for _, evaluator := range room.Configuration.Evaluators {

		log.Printf("Considering evaluator %s", evaluator.EvaluatorKey)

		curEvaluator := evaluators[evaluator.EvaluatorKey]
		if curEvaluator == nil {
			err = errors.New("No evaluator corresponding to key " + evaluator.EvaluatorKey)
			return
		}

		actions, err = curEvaluator.Evaluate(roomInQuestion)
		if err != nil {
			return
		}

		for _, action := range actions {
			err = curEvaluator.Validate(action)
			if err != nil {
				log.Printf("Error on validation of %s on evaluator %s", action.Action, evaluator.EvaluatorKey)
				return
			}

			// Provide a map from the generating evaluator to the generated action in
			// case they want to use the Incompatable actions in the reconcilers.
			action.GeneratingEvaluator = evaluator.EvaluatorKey
			actions = append(actions, action)
		}
	}

	log.Printf("Evaluation complete, starting reconciliation...")
	curReconciler := reconcilers[room.Configuration.RoomKey]
	if curReconciler == nil {
		err = errors.New("No reconciler corresponding to key " + room.Configuration.RoomKey)
		return
	}

	actions, err = curReconciler.Reconcile(actions)
	if err != nil {
		return
	}

	//execute actions.
	report, err = ce.ExecuteActions(actions)

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
