package actionReconcilers

/*
Reconciler is an interface that builds a reconciler for a room configuration.
The purpose of a reconciler is to
*/
type actionReconciler interface {
	/*
	   Reconcile takes a slice of ActionStructure objects, and returns an ordered list
	   of the same.

	   It is the purpose of the reconcile function to allow control of the interplay of
	   commands within a room (order of execution, mutually exclusive commands, etc.)

	   The ActionStructure elements will be evaluated (executed) in the order returned
	   from Reconcile.
	*/
	Reconcile([]ActionStructure) ([]ActionStructure, error)
}

//reconcilerMap is a singleton that maps known keys to their reconciler struct.
var reconcilerMap = make(map[string]actionReconciler)
var reconcilerMapInitialized = false

//Init adds the commands to the commandMap here.
func Init() *map[string]actionReconciler {
	if !reconcilerMapInitialized {

		reconcilerMapInitialized = true
	}

	return &reconcilerMap
}
