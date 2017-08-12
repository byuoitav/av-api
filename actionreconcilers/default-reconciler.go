package actionreconcilers

import (
	"log"

	"github.com/byuoitav/av-api/base"
)

//DefaultReconciler is the Default Reconciler
//Sorts by device, then by priority
type DefaultReconciler struct{}

//Reconcile fulfills the requirement to be a Reconciler.
func (d *DefaultReconciler) Reconcile(actions []base.ActionStructure) (actionsNew [][]base.ActionStructure, err error) {

	//map all possible commands to a reference to the command struct

	log.Printf("Reconciling actions.")

	//map a device ID to an array of actions specific to the device
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

	log.Printf("Checking for incompatible actions.")
	//	for devID, v := range deviceActionMap {
	//
	//	}

	log.Printf("Done.")
	actionsNew = append(actionsNew, actions)
	return
}
