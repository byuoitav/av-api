package state

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/byuoitav/av-api/actionreconcilers"
	"github.com/byuoitav/av-api/base"
	ce "github.com/byuoitav/av-api/commandevaluators"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

//for each command in the configuration, evaluate and validate.
func GenerateActions(dbRoom structs.Room, bodyRoom base.PublicRoom) (batches []base.ActionStructure, err error) {

	color.Set(color.FgYellow)
	log.Printf("[state] generating actions...")
	color.Unset()

	var actions []base.ActionStructure
	for _, evaluator := range dbRoom.Configuration.Evaluators {

		if strings.Contains(evaluator.EvaluatorKey, "STATUS") {
			continue
		}

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

	color.Set(color.FgYellow)
	log.Printf("[state] done generating actions.")
	color.Unset()

	return ReconcileActions(dbRoom, actions)
}

//produces a DAG
func ReconcileActions(room structs.Room, actions []base.ActionStructure) (batches []base.ActionStructure, err error) {

	color.Set(color.FgYellow)
	log.Printf("[state] Reconciling actions...")
	color.Unset()

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

	color.Set(color.FgYellow)
	log.Printf("[state] Done reconcililing actions.")
	color.Unset()

	return
}

//@pre TODO DestinationDevice field is populated for every action!!
//ExecuteActions carries out the actions defined in the struct
func ExecuteActions(DAG []base.ActionStructure) ([]se.StatusResponse, error) {

	color.Set(color.FgYellow)
	log.Printf("[state] Executing actions...")
	color.Unset()

	var output []se.StatusResponse

	responses := make(chan se.StatusResponse, len(DAG))
	var done sync.WaitGroup

	for _, child := range DAG[0].Children {

		done.Add(1)
		go ExecuteAction(*child, responses, &done)
	}

	log.Printf("[state] waiting for responses...")
	done.Wait()

	log.Printf("[state] done executing actions, closing channel...")
	close(responses)

	for response := range responses {
		output = append(output, response)
	}

	color.Set(color.FgYellow)
	color.Unset()

	return output, nil
}

//builds a status response
func ExecuteAction(action base.ActionStructure, responses chan<- se.StatusResponse, control *sync.WaitGroup) {

	log.Printf("[state] Executing action %s against device %s...", action.Action, action.Device.Name)

	if action.Overridden {
		log.Printf("Action %s on device %s have been overridden. Continuing.",
			action.Action, action.Device.Name)
		control.Done()
		return
	}

	has, cmd := ce.CheckCommands(action.Device.Commands, action.Action)
	if !has {
		errorStr := fmt.Sprintf("Error retrieving the command %s for device %s.", action.Action, action.Device.GetFullName())
		log.Printf(errorStr)
		PublishError(errorStr, action)
		control.Done()
		return
	}

	//replace the address
	endpoint := ReplaceIPAddressEndpoint(cmd.Endpoint.Path, action.Device.Address)

	endpoint, err := ReplaceParameters(endpoint, action.Parameters)
	if err != nil {
		errorString := fmt.Sprintf("Error building endpoint for command %s against device %s: %s", action.Action, action.Device.GetFullName(), err.Error())
		log.Printf(errorString)
		PublishError(errorString, action)
		control.Done()
		return
	}

	//Execute the command.
	status := ExecuteCommand(action, cmd, endpoint)

	log.Printf("[state] Writing response to channel...")
	responses <- status
	log.Printf("[state] microservice reported status: %s", status.Status)

	for _, child := range action.Children {

		log.Printf("[state] found child: %s. Executing...", child.Action)

		control.Add(1)
		go ExecuteAction(*child, responses, control)
	}

	log.Printf("[state] started child actions.")

	control.Done()
}

//this is where we decide which status evaluator is used to evalutate the resultant status of a command that sets state
var SET_STATE_STATUS_EVALUATORS = map[string]string{

	"PowerOnDefault":                "STATUS_PowerDefault",
	"StandbyDefault":                "STATUS_PowerDefault",
	"ChangeVideoInputDefault":       "STATUS_InputDefault",
	"ChangeAudioInputDefault":       "STATUS_InputDefault",
	"ChangeVideoInputVideoSwitcher": "STATUS_InputVideoSwitcher",
	"BlankDisplayDefault":           "STATUS_BlankedDefault",
	"UnBlankDisplayDefault":         "STATUS_BlankedDefault",
	"MuteDefault":                   "STATUS_MutedDefault",
	"UnMuteDefault":                 "STATUS_MutedDefault",
	"SetVolumeDefault":              "STATUS_VolumeDefault",
	"SetVolumeTecLite":              "STATUS_VolumeDefault",
	"MuteDSP":                       "STATUS_MutedDSP",
	"UnmuteDSP":                     "STATUS_MutedDSP",
	"SetVolumeDSP":                  "STATUS_VolumeDSP",
}
