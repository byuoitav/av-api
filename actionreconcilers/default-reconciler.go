package actionreconcilers

import (
	"log"

	"github.com/byuoitav/av-api/base"
)

//DefaultReconciler is the Default Reconciler
//Sorts by device, then by priority
type DefaultReconciler struct{}

//Reconcile fulfills the requirement to be a Reconciler.
func (d *DefaultReconciler) Reconcile(actions []base.ActionStructure) ([]base.ActionStructure, error) {

	log.Printf("Removing incompatible actions...")

	actionMap := make(map[int][]base.ActionStructure)
	for _, action := range actions {
		actionMap[action.Device.ID] = append(actionMap[action.Device.ID], action) //this should work every time, right?
	}

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

		output = append(output, actionList...)

	}

	return output, nil
}

func SortActionsByPriority(actions []base.ActionStructure) ([]base.ActionStructure, error) {

	var output []base.ActionStructure

	return output, nil
}
