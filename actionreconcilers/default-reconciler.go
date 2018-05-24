package actionreconcilers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/byuoitav/common/db"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

//DefaultReconciler is the Default Reconciler
//Sorts by device, then by priority
type DefaultReconciler struct{}

//Reconcile fulfills the requirement to be a Reconciler.
func (d *DefaultReconciler) Reconcile(actions []base.ActionStructure, inCount int) ([]base.ActionStructure, int, error) {

	base.Log("[reconciler] Removing incompatible actions...")
	var buffer bytes.Buffer

	actionMap := make(map[string][]base.ActionStructure)
	for _, action := range actions {

		buffer.WriteString(action.Device.ID + " ")
		actionMap[action.Device.ID] = append(actionMap[action.Device.ID], action) //this should work every time, right?
	}

	output := []base.ActionStructure{
		base.ActionStructure{
			Action:              "Start",
			Device:              structs.Device{ID: "DefaultReconciler"},
			GeneratingEvaluator: "DefaultReconciler",
			Overridden:          true,
		},
	}
	var count int

	for device, actionList := range actionMap {

		actionList, c, err := StandardReconcile(device, inCount, actionList)
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		actionList, err = SortActionsByPriority(actionList)
		if err != nil {
			return []base.ActionStructure{}, 0, err
		}

		//		actionList, err = CreateChildRelationships(actionList)
		//		if err != nil {
		//			return []base.ActionStructure{}, err
		//		}

		for i := range actionList {

			if i != len(actionList)-1 {

				base.Log("[reconciler] creating relationship %s, %s -> %s, %s", actionList[i].Action, actionList[i].Device.Name, actionList[i+1].Action, actionList[i+1].Device.Name)

				actionList[i].Children = append(actionList[i].Children, &actionList[i+1])
			}
		}

		output[0].Children = append(output[0].Children, &actionList[0])
		output = append(output, actionList...)
		count = c
	}

	return output, count, nil
}

func SortActionsByPriority(actions []base.ActionStructure) (output []base.ActionStructure, err error) {

	color.Set(color.FgHiMagenta)
	base.Log("[reconciler] sorting actions by priority...")
	color.Unset()

	actionMap := make(map[int][]base.ActionStructure)

	for _, action := range actions {

		room, err := db.GetDB().GetRoom(action.Device.GetDeviceRoomID())
		if err != nil {
			errorMessage := fmt.Sprintf("Problem getting the room for %s", action.Device.ID)
			base.Log(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		roomConfig, err := db.GetDB().GetRoomConfiguration(room.Configuration.ID)
		if err != nil {
			errorMessage := fmt.Sprintf("Problem getting command evaluators for %s", action.Device.GetDeviceRoomID())
			base.Log(errorMessage)
			return []base.ActionStructure{}, errors.New(errorMessage)
		}

		evaluators := roomConfig.Evaluators

		for _, commandEval := range evaluators {

			log.Printf("Eval ID: %s  Priority: %v -- Action: %s", commandEval.ID, commandEval.Priority, action.Action)
			if commandEval.ID == action.Action {

				actionMap[commandEval.Priority] = append(actionMap[commandEval.Priority], action)
			}
		}
	}

	var keys []int
	for key := range actionMap {
		keys = append(keys, key)
	}

	sort.Ints(keys)
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

//since we've already sorted by priority and device, so the first element's child is the second and so on
func CreateChildRelationships(actions []base.ActionStructure) ([]base.ActionStructure, error) {

	color.Set(color.FgHiMagenta)
	base.Log("[reconciler] creating child relationships...")

	for i, action := range actions {

		base.Log("[reconciler] considering action %s against device %s...", action.Action, action.Device.Name)

		if i != len(actions)-1 {

			base.Log("[reconciler] creating relationship %s, %s -> %s, %s", action.Action, action.Device.Name, actions[i+1].Action, actions[i+1].Device.Name)

			action.Children = append(action.Children, &actions[i+1])
		}
	}

	color.Unset()
	return actions, nil
}
