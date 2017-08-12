package state

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/byuoitav/av-api/actionreconcilers"
	"github.com/byuoitav/av-api/base"
	ce "github.com/byuoitav/av-api/commandevaluators"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
)

//for each command in the configuration, evaluate and validate.
func GenerateActions(dbRoom structs.Room, bodyRoom base.PublicRoom) (batches [][]base.ActionStructure, err error) {

	log.Printf("Generating actions...")

	var actions []base.ActionStructure
	for _, evaluator := range dbRoom.Configuration.Evaluators {

		log.Printf("Considering evaluator %s", evaluator.EvaluatorKey)

		curEvaluator := ce.EVALUATORS[evaluator.EvaluatorKey]
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

	return ReconcileActions(dbRoom, actions)
}

//produces a DAG
func ReconcileActions(room structs.Room, actions []base.ActionStructure) (batches [][]base.ActionStructure, err error) {

	log.Printf("Reconciling actions...")

	//Initialize map of strings to commandevaluators
	reconcilers := actionreconcilers.Init()

	curReconciler := reconcilers[room.Configuration.RoomKey]
	if curReconciler == nil {
		err = errors.New("No reconciler corresponding to key " + room.Configuration.RoomKey)
		return
	}

	batches, err = curReconciler.Reconcile(actions)
	if err != nil {
		return
	}

	return
}

//@pre TODO DestinationDevice field is populated for every action!!
//ExecuteActions carries out the actions defined in the struct
func ExecuteActions(DAG [][]base.ActionStructure) ([]se.StatusResponse, error) {

	var output []se.StatusResponse

	responses := make(chan se.StatusResponse)
	var done sync.WaitGroup

	for _, level := range DAG {

		go func() {

			done.Add(1)

			for _, a := range level { // these commands can be executed in parallel

				if a.Overridden {
					log.Printf("Action %s on device %s have been overridden. Continuing.",
						a.Action, a.Device.Name)
					continue
				}

				has, cmd := ce.CheckCommands(a.Device.Commands, a.Action)
				if !has {
					errorStr := fmt.Sprintf("Error retrieving the command %s for device %s.", a.Action, a.Device.GetFullName())
					log.Printf(errorStr)
					PublishError(errorStr, a)
					continue
				}

				//replace the address
				endpoint := ReplaceIPAddressEndpoint(cmd.Endpoint.Path, a.Device.Address)

				endpoint, err := ReplaceParameters(endpoint, a.Parameters)
				if err != nil {
					errorString := fmt.Sprintf("Error building endpoint for command %s against device %s: %s", a.Action, a.Device.GetFullName(), err.Error())
					log.Printf(errorString)
					PublishError(errorString, a)
					continue
				}

				//Execute the command.
				status := ExecuteCommand(a, cmd, endpoint)
				responses <- status
				log.Printf("Status: %v", status)
			}

			done.Done()
		}()
	}

	done.Wait()

	for response := range responses {
		output = append(output, response)
	}

	return output, nil
}
