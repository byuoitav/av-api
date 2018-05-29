package actionreconcilers

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

//DefaultReconciler is the Default Reconciler
//Sorts by device, then by priority
type DefaultReconciler struct{}

//Reconcile sorts through the list of actions to determine the execution order.
func (d *DefaultReconciler) Reconcile(actions []base.ActionStructure, inCount int) ([]base.ActionStructure, int, error) {

	log.L.Info("[reconciler] Removing incompatible actions...")
	var buffer bytes.Buffer

	// First we will map device IDs to the action related to them.
	actionMap := make(map[string][]base.ActionStructure)

	for _, action := range actions {
		buffer.WriteString(action.Device.ID + " ")
		actionMap[action.Device.ID] = append(actionMap[action.Device.ID], action)
	}

	// Next we will make a list of actions to output.
	output := []base.ActionStructure{
		base.ActionStructure{
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

		actionList, err = SortActionsByPriority(actionList)
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		// Some actions are dependent on others, so we will map that relationship as well.
		for i := range actionList {

			if i != len(actionList)-1 {

				log.L.Infof("[reconciler] creating relationship %s, %s -> %s, %s", actionList[i].Action, actionList[i].Device.Name, actionList[i+1].Action, actionList[i+1].Device.Name)

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

// SortActionsByPriority sorts the list of actions by their priority integer value.
func SortActionsByPriority(actions []base.ActionStructure) (output []base.ActionStructure, err error) {

	color.Set(color.FgHiMagenta)
	log.L.Info("[reconciler] sorting actions by priority...")
	color.Unset()

	// Map priority values to actions.
	actionMap := make(map[int][]base.ActionStructure)

	for _, action := range actions {

		deviceType, err := db.GetDB().GetDeviceType(action.Device.Type.ID)
		if err != nil {
			errorMessage := fmt.Sprintf("Problem getting the room for %s", action.Device.ID)
			log.L.Error(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		// Obtain the list of commands for this type of device.
		commands := deviceType.Commands

		for _, command := range commands {

			if command.ID == action.Action {

				actionMap[command.Priority] = append(actionMap[command.Priority], action)
			}
		}
	}

	// Create a list of the priority keys.
	var keys []int
	for key := range actionMap {
		keys = append(keys, key)
	}

	// Sort the list of priority keys.
	sort.Ints(keys)

	// Append the actions to the output in order of highest priority first. (1 being the highest possible)
	output = append(output, actionMap[keys[0]]...) //parents of everything
	marker := len(output) - 1
	delete(actionMap, keys[0])

	for len(actionMap) != 0 {
		for index, key := range keys {

			if index == 0 {
				continue
			}

			output = append(output, actionMap[key]...)
			marker = len(output) - 1
			for _, action := range actionMap[key] {

				output[marker].Children = append(output[marker].Children, &action)
			}

			delete(actionMap, key)
		}

	}
	return output, nil
}

// CreateChildRelationships establishes the relationship hierarchy between any actions that are dependent on others.
func CreateChildRelationships(actions []base.ActionStructure) ([]base.ActionStructure, error) {

	color.Set(color.FgHiMagenta)
	log.L.Info("[reconciler] creating child relationships...")

	for i, action := range actions {

		log.L.Infof("[reconciler] considering action %s against device %s...", action.Action, action.Device.Name)

		if i != len(actions)-1 {

			log.L.Infof("[reconciler] creating relationship %s, %s -> %s, %s", action.Action, action.Device.Name, actions[i+1].Action, actions[i+1].Device.Name)

			action.Children = append(action.Children, &actions[i+1])
		}
	}

	color.Unset()
	return actions, nil
}
