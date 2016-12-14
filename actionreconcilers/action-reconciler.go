package actionreconcilers

import "github.com/byuoitav/av-api/base"

/*
ActionReconciler is an interface that builds a reconciler for a room configuration.
The purpose of a reconciler is to
*/
type ActionReconciler interface {
	/*
	   Reconcile takes a slice of ActionStructure objects, and returns an ordered list
	   of the same.

	   It is the purpose of the reconcile function to allow control of the interplay of
	   commands within a room (order of execution, mutually exclusive commands, etc.)

	   The ActionStructure elements will be evaluated (executed) in the order returned
	   from Reconcile.
	*/
	Reconcile([]base.ActionStructure) ([]base.ActionStructure, error)
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
