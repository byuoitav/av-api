package actionreconcilers

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/commandevaluators"
)

//DefaultReconciler is the Default Reconciler
type DefaultReconciler struct{}

//Reconcile fulfills the requirement to be a Reconciler.
func (d *DefaultReconciler) Reconcile(actions []base.ActionStructure) (actionsNew []base.ActionStructure, err error) {

	CommandMap := commandevaluators.Init()

	log.Printf("Reconciling actions.")
	deviceActionMap := make(map[int][]base.ActionStructure)

	log.Printf("Generating device action set.")
	//generate a set of actions for each device.
	for _, a := range actions {
		if _, has := deviceActionMap[a.Device.ID]; has {
			deviceActionMap[a.Device.ID] = append(deviceActionMap[a.Device.ID], a)
		} else {
			deviceActionMap[a.Device.ID] = []base.ActionStructure{a}
		}
	}

	log.Printf("Checking for incompatable actions.")
	for devID, v := range deviceActionMap {
		//for each device, construct set of actions
		actionsForEvaluation := make(map[string]base.ActionStructure)
		incompat := make(map[string]base.ActionStructure)

		for i := 0; i < len(v); i++ {
			actionsForEvaluation[v[i].Action] = v[i]
			//for each device, construct set of incompatable actions
			//Value is the action that generated the incompatable action.
			incompatableActions := CommandMap[v[i].GeneratingEvaluator].GetIncompatibleCommands()

			for _, incompatAct := range incompatableActions {
				incompat[incompatAct] = v[i]
			}
		}

		//find intersection of sets.

		//baseAction is the actionStructure generating the action (for cur action)
		//incompatableBaseAction is the actionStructure that generated the incompatable action.
		for curAction, baseAction := range actionsForEvaluation {
			if baseAction.Overridden {
				continue
			}

			for incompatableAction, incompatableBaseAction := range incompat {
				if incompatableBaseAction.Overridden {
					continue
				}

				if strings.EqualFold(curAction, incompatableAction) {
					log.Printf("%s is incompatable with %s.", incompatableAction, incompatableBaseAction.Action)
					// if one of them is room wide and the other is not override the room-wide
					// action.

					if !baseAction.DeviceSpecific && incompatableBaseAction.DeviceSpecific {
						log.Printf("%s is a device specific command. Overriding %s in favor of device-specific command %s.",
							incompatableBaseAction.Action, baseAction.Action, incompatableBaseAction.Action)
						baseAction.Overridden = true

					} else if baseAction.DeviceSpecific && !incompatableBaseAction.DeviceSpecific {
						log.Printf("%s is a device specific command. Overriding %s in favor of device-specific command %s.",
							baseAction.Action, incompatableBaseAction.Action, baseAction.Action)

						incompatableBaseAction.Overridden = true
					} else {
						errorString := incompatableAction + " is an incompatable action with " + incompatableBaseAction.Action + " for device with ID: " +
							string(devID)
						log.Printf("%s", errorString)
						err = errors.New(errorString)
						return
					}
				}
			}
		}
	}

	log.Printf("Done.")
	actionsNew = actions
	return
}
