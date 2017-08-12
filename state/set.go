package state

import (
	"errors"
	"fmt"
	"log"

	"github.com/byuoitav/av-api/base"
	ce "github.com/byuoitav/av-api/commandevaluators"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
)

//for each command in the configuration, evaluate and validate.
func GenerateActions(dbRoom structs.Room, bodyRoom base.PublicRoom) (actions []base.ActionStructure, err error) {

	log.Printf("Generating actions...")
	evaluators := ce.EVALUATORS

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

//ExecuteActions carries out the actions defined in the struct
func ExecuteActions(actions []base.ActionStructure) ([]se.StatusResponse, error) {

	var output []se.StatusResponse
	for _, a := range actions {

		if a.Overridden {
			log.Printf("Action %s on device %s have been overridden. Continuing.",
				a.Action, a.Device.Name)
			continue
		}

		has, cmd := ce.CheckCommands(a.Device.Commands, a.Action)
		if !has {
			errorStr := fmt.Sprintf("Error retrieving the command %s for device %s.", a.Action, a.Device.GetFullName())
			log.Printf(errorStr)
			//return base.PublicRoom{}, errors.New(errorStr)
		}

		//replace the address
		endpoint := ReplaceIPAddressEndpoint(cmd.Endpoint.Path, a.Device.Address)

		endpoint, err := ReplaceParameters(endpoint, a.Parameters)
		if err != nil {
			errorString := fmt.Sprintf("Error building endpoint for command %s against device %s: %s", a.Action, a.Device.GetFullName(), err.Error())
			log.Printf(errorString)
			//return base.PublicRoom{}, errors.New(errorString)
		}

		//Execute the command.
		status := ExecuteCommand(a, cmd, endpoint)
		log.Printf("Status: %v", status)
	}

	return output, nil
}
