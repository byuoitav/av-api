package actionreconcilers

import (
	"sort"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

//DefaultReconciler is the Default Reconciler
//Sorts by device, then by priority
type DefaultReconciler struct{}

//Reconcile sorts through the list of actions to determine the execution order.
func (d *DefaultReconciler) Reconcile(actions []base.ActionStructure, inCount int) ([]base.ActionStructure, int, error) {

	log.L.Debug("[reconciler] Removing incompatible actions...")

	// First we will map device IDs to the action related to them.
	actionMap := make(map[string][]base.ActionStructure)

	for _, action := range actions {
		actionMap[action.Device.ID] = append(actionMap[action.Device.ID], action)
	}
	// Next we will make a list of actions to output.
	output := []base.ActionStructure{base.ActionStructure{
		Action:              "Start",
		Device:              structs.Device{ID: "DefaultReconciler"},
		GeneratingEvaluator: "DefaultReconciler",
		Overridden:          true,
	},
	}

	var count int

	// As we iterate through the actionMap, we will sort the actions by device and priority.
	for device, actionList := range actionMap {
		actionList, c, err := StandardReconcile(device, inCount, actionList)
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		sort.Sort(base.ActionByPriority(actionList))
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		// Some actions are dependent on others, so we will map that relationship as well.
		for i := range actionList {
			if i != len(actionList)-1 {
				log.L.Debugf("[reconciler] creating relationship %s, %s -> %s, %s", actionList[i].Action, actionList[i].Device.Name, actionList[i+1].Action, actionList[i+1].Device.Name)
				actionList[i].Children = append(actionList[i].Children, &actionList[i+1])
			}
		}

		// After sorting, we add the sorted actions and their children to the output.
		output[0].Children = append(output[0].Children, &actionList[0])
		output = append(output, actionList...)
		count = c
	}

	// Finally, we return the sorted list of actions.
	return output, count, nil
}

// CreateChildRelationships establishes the relationship hierarchy between any actions that are dependent on others.
func CreateChildRelationships(actions []base.ActionStructure) ([]base.ActionStructure, error) {

	color.Set(color.FgHiMagenta)
	log.L.Debug("[reconciler] creating child relationships...")

	for i, action := range actions {

		log.L.Debugf("[reconciler] considering action %s against device %s...", action.Action, action.Device.Name)

		if i != len(actions)-1 {

			log.L.Debugf("[reconciler] creating relationship %s, %s -> %s, %s", action.Action, action.Device.Name, actions[i+1].Action, actions[i+1].Device.Name)

			action.Children = append(action.Children, &actions[i+1])
		}
	}

	color.Unset()
	return actions, nil
}
