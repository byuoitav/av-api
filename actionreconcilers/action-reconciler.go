package actionreconcilers

import (
	"bytes"
	"errors"
	"strings"

	"github.com/byuoitav/av-api/base"
	ce "github.com/byuoitav/av-api/commandevaluators"
	"github.com/fatih/color"
)

/*
ActionReconciler is an interface that builds a reconciler for a room configuration.
The purpose of a reconciler is to
*/
type ActionReconciler interface {
	/*
	   Reconcile takes a slice of ActionStructure objects, and returns a slice of
	   slices representing a DAG representing execution order, e.g. each outer slice
	   represents a level in the DAG

	   It is the purpose of the reconcile function to allow control of the interplay of
	   commands within a room (order of execution, mutually exclusive commands, etc.)

	   The ActionStructure elements will be evaluated (executed) in the order returned
	   from Reconcile.
	*/
	Reconcile([]base.ActionStructure, int) ([]base.ActionStructure, int, error)
}

//reconcilerMap is a singleton that maps known keys to their reconciler struct.
var reconcilerMap = make(map[string]ActionReconciler)
var reconcilerMapInitialized = false

//Init adds the reconcilers to the reconcilerMap.
func Init() map[string]ActionReconciler {
	if !reconcilerMapInitialized {
		//-------------------------------
		//Add reconcilers to the map here
		//-------------------------------
		reconcilerMap["Default"] = &DefaultReconciler{}

		reconcilerMapInitialized = true
	}

	return reconcilerMap
}

func StandardReconcile(device string, inCount int, actions []base.ActionStructure) ([]base.ActionStructure, int, error) {

	color.Set(color.FgHiMagenta)
	base.Log("[reconciler] performing standard reconcile...")
	color.Unset()

	//for each device, construct set of actions
	actionsForEvaluation := make(map[string]base.ActionStructure)
	incompatible := make(map[string]base.ActionStructure)

	for _, action := range actions {
		actionsForEvaluation[action.Action] = action
		//for each device, construct set of incompatible actions
		//Value is the action that generated the incompatible action.
		evaluator := ce.EVALUATORS[action.GeneratingEvaluator]

		if evaluator == nil {
			color.Set(color.FgHiRed)
			base.Log("Alert! Nil pointer for evaluator: %s", action.GeneratingEvaluator)
			color.Unset()
			continue
		}

		incompatibleActions := evaluator.GetIncompatibleCommands()

		for _, incompatibleAction := range incompatibleActions {
			incompatible[incompatibleAction] = action
		}
	}

	//find intersection of sets.

	//baseAction is the actionStructure generating the action (for cur action)
	//incompatibleBaseAction is the actionStructure that generated the incompatible action.
	for curAction, baseAction := range actionsForEvaluation {
		if baseAction.Overridden {
			continue
		}

		for incompatibleAction, incompatibleBaseAction := range incompatible {
			if incompatibleBaseAction.Overridden {
				continue
			}

			if strings.EqualFold(curAction, incompatibleAction) { //we've found an incompatible action
				base.Log("%s is incompatible with %s.", incompatibleAction, incompatibleBaseAction.Action)
				// if one of them is room wide and the other is not override the room-wide
				// action.

				if !baseAction.DeviceSpecific && incompatibleBaseAction.DeviceSpecific {
					base.Log("%s is a device specific command. Overriding %s in favor of device-specific command %s.",
						incompatibleBaseAction.Action, baseAction.Action, incompatibleBaseAction.Action)
					inCount--
					baseAction.Overridden = true

				} else if baseAction.DeviceSpecific && !incompatibleBaseAction.DeviceSpecific {
					base.Log("%s is a device specific command. Overriding %s in favor of device-specific command %s.",
						baseAction.Action, incompatibleBaseAction.Action, baseAction.Action)
					inCount--
					incompatibleBaseAction.Overridden = true
				} else {
					errorString := incompatibleAction + " is an incompatible action with " + incompatibleBaseAction.Action + " for device with ID: " +
						string(device)
					base.Log("%s", errorString)
					return []base.ActionStructure{}, 0, errors.New(errorString)
				}
			}
		}
	}
	//DEBUG ==============================================================================================================================================

	var buffer bytes.Buffer
	for i, a := range actions {

		buffer.WriteString(a.Action)
		if i != len(actions)-1 {
			buffer.WriteString(", ")
		}
	}
	base.Log("[reconciler] actions after standard reconcile: %s", buffer.String())
	//=====================================================================================================================================================

	return actions, inCount, nil
}
