package helpers

import (
	"errors"
	"log"
	"strings"

	ar "github.com/byuoitav/av-api/actionreconcilers"
	"github.com/byuoitav/av-api/base"
	ce "github.com/byuoitav/av-api/commandevaluators"
	"github.com/byuoitav/av-api/dbo"
	ex "github.com/byuoitav/av-api/execute"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//EditRoomState actually carries out the room state changes
func EditRoomState(roomInQuestion base.PublicRoom) (report base.PublicRoom, err error) {

	//get accessors.Room (as it exists in the database)
	room, err := dbo.GetRoomByInfo(roomInQuestion.Building, roomInQuestion.Room)
	if err != nil {
		return
	}

	actions, err := GenerateActions(room, roomInQuestion)
	if err != nil {
		return
	}

	batches, err := ReconcileActions(room, actions)
	if err != nil {
		return
	}

	//execute actions.
	report, err = ex.ExecuteActions(batches)

	return
}

//for each command in the configuration, evaluate and validate.
func GenerateActions(dbRoom accessors.Room, bodyRoom base.PublicRoom) (actions []base.ActionStructure, err error) {

	log.Printf("Generating actions...")
	evaluators := ce.Init()

	for _, evaluator := range dbRoom.Configuration.Evaluators {

		log.Printf("Considering evaluator %s", evaluator.EvaluatorKey)

		curEvaluator := evaluators[evaluator.EvaluatorKey]
		if curEvaluator == nil {
			err = errors.New("No evaluator corresponding to key " + evaluator.EvaluatorKey)
			return
		}

		actions, err = curEvaluator.Evaluate(bodyRoom)
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

	return
}

func ReconcileActions(room accessors.Room, actions []base.ActionStructure) (batches [][]base.ActionStructure, err error) {

	log.Printf("Reconciling actions...")

	//Initialize map of strings to commandevaluators
	reconcilers := ar.Init()

	curReconciler := reconcilers[room.Configuration.RoomKey]
	if curReconciler == nil {
		err = errors.New("No reconciler corresponding to key " + room.Configuration.RoomKey)
		return
	}

	actions, err = curReconciler.Reconcile(actions)
	if err != nil {
		return
	}

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
