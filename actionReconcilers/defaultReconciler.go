package actionReconcilers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

//DefaultReconciler is the Default Reconciler
type DefaultReconciler struct{}

//Reconcile fulfills the requirement to be a Reconciler.
func (d *DefaultReconciler) Reconcile(actions []ActionStructure) (actionsNew []ActionStructure, err error) {
	log.Printf("Reconciling actions.")
	deviceActionMap := make(map[int][]ActionStructure)

	log.Printf("Generating device action set.")
	//generate a set of actions for each device.
	for _, a := range *actions {
		if _, has := deviceActionMap[a.Device.ID]; has {
			deviceActionMap[a.Device.ID] = append(deviceActionMap[a.Device.ID], a)
		} else {
			deviceActionMap[a.Device.ID] = []ActionStructure{a}
		}
	}

	log.Printf("Checking for incompatable actions.")
	for devID, v := range deviceActionMap {
		//for each device, construct set of actions
		actionsForEvaluation := make(map[string]ActionStructure)
		incompat := make(map[string]ActionStructure)

		for i := 0; i < len(v); i++ {
			actionsForEvaluation[v[i].Action] = v[i]
			//for each device, construct set of incompatable actions
			//Value is the action that generated the incompatable action.
			incompatableActions := CommandMap[v[i].Action].GetIncompatableCommands()
			for _, incompatAct := range incompatableActions {
				incompat[incompatAct] = v[i]
			}
		}

		//find intersection of sets.

		//baseAction is the actionStructure generating the action (for cur action)
		//incompatableBaseAction is the actionStructure that generated the incompatable action.
		for curAction, baseAction := range actionsForEvaluation {
			fmt.Printf("%v: %+v\n", curAction, baseAction)
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
						/*
							We have to mark it as incompatable in three places, incompat (so it doesn't cause problems for other commands),
							actionsForEvaluation for the same reason, and in actions so the action won't get sent. We were using pointers, but
							for simplicity and readability in code, we pulled them out.

							Don't judge. :D
						*/
						//markAsOverridden(incompatableBaseAction, &incompat, &actionsForEvaluation, actions)
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

	//Debug
	b, _ := json.Marshal(&actions)
	fmt.Printf("%s", b)
	//END Debug

	log.Printf("Done.")
	return
}
