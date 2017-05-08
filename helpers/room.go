package helpers

import (
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/byuoitav/av-api/actionreconcilers"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/commandevaluators"
	"github.com/byuoitav/av-api/dbo"
)

//EditRoomState actually carries out the room state changes
func EditRoomState(roomInQuestion base.PublicRoom) (report []commandevaluators.CommandExecutionReporting, err error) {

	//Initialize map of strings to commandevaluators
	evaluators := commandevaluators.Init()
	reconcilers := actionreconcilers.Init()

	//get accessors.Room (as it exists in the database)
	room, err := dbo.GetRoomByInfo(roomInQuestion.Building, roomInQuestion.Room)
	if err != nil {
		return
	}

	actionList := []base.ActionStructure{}

	re := regexp.MustCompile(".*-RPC$")

	//for each command in the configuration, evaluate and validate.
	for _, c := range room.Configuration.Evaluators {

		if re.MatchString(c.EvaluatorKey) {
			continue
		}

		log.Printf("Starting evaluation with evaluator %s", c.EvaluatorKey)

		curEvaluator := evaluators[c.EvaluatorKey]
		if curEvaluator == nil {
			err = errors.New("No evaluator corresponding to key " + c.EvaluatorKey)
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
				log.Printf("Error on validation of %s on evaluator %s", action.Action, c.EvaluatorKey)
				return
			}

			// Provide a map from the generating evaluator to the generated action in
			// case they want to use the Incompatable actions in the reconcilers.
			action.GeneratingEvaluator = c.EvaluatorKey
			actionList = append(actionList, action)
		}
	}
	log.Printf("Done with Evaluation.")
	log.Printf("Starting reconcilliation")

	//Reconcile actions
	curReconciler := reconcilers[room.Configuration.RoomKey]
	if curReconciler == nil {
		err = errors.New("No reconciler corresponding to key " + room.Configuration.RoomKey)
		return
	}
	log.Printf("Startin reconciliation")

	actionList, err = curReconciler.Reconcile(actionList)
	if err != nil {
		return
	}

	//execute actions.
	report, err = commandevaluators.ExecuteActions(actionList)

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
