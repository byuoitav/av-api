package actionreconcilers

import (
	"bytes"
	"log"
	"sort"

	"github.com/byuoitav/av-api/base"
	"github.com/fatih/color"
)

//DefaultReconciler is the Default Reconciler
//Sorts by device, then by priority
type DefaultReconciler struct{}

//Reconcile fulfills the requirement to be a Reconciler.
func (d *DefaultReconciler) Reconcile(actions []base.ActionStructure) ([]base.ActionStructure, error) {

	log.Printf("[reconciler] Removing incompatible actions...")
	var buffer bytes.Buffer

	actionMap := make(map[int][]base.ActionStructure)
	for _, action := range actions {

		buffer.WriteString(action.Device.Name + " ")
		actionMap[action.Device.ID] = append(actionMap[action.Device.ID], action) //this should work every time, right?
	}

	// DEBUGGING ================================================================================================

	color.Set(color.FgHiMagenta)
	log.Printf("[reconciler] devices identified: %s", buffer.String())
	color.Unset()

	buffer.Reset()
	for deviceID, actions := range actionMap {

		for index, action := range actions {

			buffer.WriteString(action.Action)
			if index != len(actions)-1 {
				buffer.WriteString(", ")
			}

		}

		color.Set(color.FgHiMagenta)
		log.Printf("[reconciler] found device with ID: %v and commands: %s", deviceID, buffer.String())
		color.Unset()
	}
	// DEBUGGING ================================================================================================

	output := []base.ActionStructure{
		base.ActionStructure{
			Action:              "Start",
			GeneratingEvaluator: "DefaultReconciler",
			Overridden:          true,
		},
	}

	for device, actionList := range actionMap {

		actionList, err := StandardReconcile(device, actionList)
		if err != nil {
			return []base.ActionStructure{}, err
		}

		actionList, err = SortActionsByPriority(actionList)
		if err != nil {
			return []base.ActionStructure{}, err
		}

		output[0].Children = append(output[0].Children, &actionList[0])
		output = append(output, actionList...)

	}

	return output, nil
}

func SortActionsByPriority(actions []base.ActionStructure) (output []base.ActionStructure, err error) {

	actionMap := make(map[int][]base.ActionStructure)

	for _, action := range actions {

		for _, command := range action.Device.Commands {

			if command.Name == action.Action {

				actionMap[command.Priority] = append(actionMap[command.Priority], action)
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
