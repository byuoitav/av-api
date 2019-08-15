package state

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/byuoitav/av-api/actionreconcilers"
	"github.com/byuoitav/av-api/base"
	ce "github.com/byuoitav/av-api/commandevaluators"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

//GenerateActions evaluates and validates each command in the configuration.
func GenerateActions(dbRoom structs.Room, bodyRoom base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {

	log.L.Infof("%s", color.HiBlueString("[state] generating actions..."))

	var count int

	var output []base.ActionStructure
	for _, evaluator := range dbRoom.Configuration.Evaluators {

		if strings.Contains(evaluator.CodeKey, "STATUS") {
			continue
		}

		curEvaluator := ce.EVALUATORS[evaluator.CodeKey]
		if curEvaluator == nil {
			msg := fmt.Sprintf("no evaluator corresponding to key: %s", evaluator.CodeKey)
			log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
			return []base.ActionStructure{}, 0, errors.New(msg)
		}

		actions, c, err := curEvaluator.Evaluate(dbRoom, bodyRoom, requestor)
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		for _, action := range actions {
			err := curEvaluator.Validate(action)
			if err != nil {
				msg := fmt.Sprintf("action %s not valid with evaluator %s: %s", action.Action, evaluator.CodeKey, err.Error())
				log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
				return []base.ActionStructure{}, 0, errors.New(msg)
			}

			// Provide a map from the generating evaluator to the generated action in
			// case they want to use the Incompatable actions in the reconcilers.
			action.GeneratingEvaluator = evaluator.CodeKey
		}

		output = append(output, actions...)
		count += c
	}

	log.L.Infof("%s", color.HiBlueString("[state] generated %v total actions.", len(output)))

	batches, count, err := ReconcileActions(dbRoom, output, count)

	return batches, count, err
}

//ReconcileActions produces a DAG
func ReconcileActions(room structs.Room, actions []base.ActionStructure, inCount int) (batches []base.ActionStructure, count int, err error) {

	log.L.Infof("%s", color.HiBlueString("[state] reconciling actions..."))

	//Initialize map of strings to commandevaluators
	reconcilers := actionreconcilers.Init()

	curReconciler := reconcilers[room.Configuration.Description]
	if curReconciler == nil {
		err = fmt.Errorf("no reconciler corresponding to key: %s ", room.Configuration.Description)
		return
	}

	batches, count, err = curReconciler.Reconcile(actions, inCount)
	if err != nil {
		return
	}

	log.L.Infof("%s", color.HiBlueString("[state] done reconciling actions."))

	return
}

//ExecuteActions carries out the actions defined in the struct
//@pre TODO DestinationDevice field is populated for every action!!
func ExecuteActions(DAG []base.ActionStructure, requestor string) ([]se.StatusResponse, error) {
	// get total number of actions in dag
	log.L.Infof("%s", color.HiBlueString("[state] executing actions..."))

	if len(DAG) == 0 {
		return []se.StatusResponse{}, errors.New("no actions generated")
	}

	var output []se.StatusResponse

	responses := make(chan se.StatusResponse, len(DAG))
	var done sync.WaitGroup

	for _, child := range DAG[0].Children {
		done.Add(1)
		go ExecuteAction(*child, responses, &done, requestor)
	}

	log.L.Info("[state] waiting for responses...")
	done.Wait()

	log.L.Info("[state] done executing actions, closing channel...")
	close(responses)

	if len(responses) < len(DAG)-1 {
		log.L.Errorf("%s", color.HiRedString("[error] expecting %v responses, found %v", len(DAG), len(responses)))
	}

	for response := range responses {
		output = append(output, response)
	}

	log.L.Infof("%s", color.HiBlueString("[state] done executing actions"))

	return output, nil
}

// ExecuteAction builds a status response
func ExecuteAction(action base.ActionStructure, responses chan<- se.StatusResponse, control *sync.WaitGroup, requestor string) {
	log.L.Infof("[state] Executing action %s against device %s...", action.Action, action.Device.Name)

	if action.Overridden {
		log.L.Infof("[state] Action %s on device %s have been overridden. Continuing.",
			action.Action, action.Device.Name)
		control.Done()
		return
	}

	/*
		has, cmd := ce.CheckCommands(action.Device.Type.Commands, action.Action)
		if !has {
			errorStr := fmt.Sprintf("[state] Error retrieving the command %s for device %s.", action.Action, action.Device.ID)
			log.L.Error(errorStr)
			PublishError(errorStr, action, requestor)
			control.Done()
			return
		}

		endpoint := ReplaceIPAddressEndpoint(cmd.Endpoint.Path, action.Device.Address)
		endpoint, err := ReplaceParameters(endpoint, action.Parameters)
		if err != nil {
			msg := fmt.Sprintf("Error building endpoint for command %s against device %s: %s", action.Action, action.Device.ID, err.Error())
			log.L.Errorf("%s", color.HiRedString("[state] %s", msg))
			PublishError(msg, action, requestor)
			control.Done()
			return
		}
	*/

	url, err := action.Device.BuildCommandURL(action.Action)
	if err != nil {
		msg := fmt.Sprintf("unable to execute action '%s' on %s: %s", action.Action, action.Device.ID, err.Error())
		log.L.Errorf("%s", color.HiRedString("[state] %s", msg))
		PublishError(msg, action, requestor)
		control.Done()
		return
	}

	url = strings.Replace(url, ":address", action.Device.Address, -1)

	var gerr error
	url, gerr = ReplaceParameters(url, action.Parameters)
	if gerr != nil {
		msg := fmt.Sprintf("unable to execute action '%s' on %s: %s", action.Action, action.Device.ID, gerr.Error())
		log.L.Errorf("%s", color.HiRedString("[state] %s", msg))
		PublishError(msg, action, requestor)
		control.Done()
		return
	}

	//Execute the command.
	status := ExecuteCommand(action, url, requestor)

	log.L.Info("[state] Writing response to channel...")
	responses <- status
	log.L.Infof("[state] microservice reported status: %v", status.Status)

	for _, child := range action.Children {
		log.L.Infof("[state] found child: %s. Executing...", child.Action)
		control.Add(1)
		go ExecuteAction(*child, responses, control, requestor)
	}

	control.Done()
}

//SET_STATE_STATUS_EVALUATORS is the map containing the definitions of our evaluator strings.
//this is where we decide which status evaluator is used to evalutate the resultant status of a command that sets state
var SET_STATE_STATUS_EVALUATORS = map[string]string{
	"PowerOnDefault":                 "STATUS_PowerDefault",
	"StandbyDefault":                 "STATUS_PowerDefault",
	"ChangeVideoInputDefault":        "STATUS_InputDefault",
	"ChangeAudioInputDefault":        "STATUS_InputDefault",
	"ChangeVideoInputVideoSwitcher":  "STATUS_InputVideoSwitcher",
	"ChangeVideoInputTieredSwitcher": "STATUS_InputVideoSwitcher",
	"BlankDisplayDefault":            "STATUS_BlankedDefault",
	"UnBlankDisplayDefault":          "STATUS_BlankedDefault",
	"MuteDefault":                    "STATUS_MutedDefault",
	"UnMuteDefault":                  "STATUS_MutedDefault",
	"SetVolumeDefault":               "STATUS_VolumeDefault",
	"SetVolumeTecLite":               "STATUS_VolumeDefault",
	"MuteDSP":                        "STATUS_MutedDSP",
	"UnmuteDSP":                      "STATUS_MutedDSP",
	"SetVolumeDSP":                   "STATUS_VolumeDSP",
}
