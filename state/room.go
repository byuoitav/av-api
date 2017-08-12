package state

import (
	"errors"
	"log"
	"strings"

	ar "github.com/byuoitav/av-api/actionreconcilers"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/configuration-database-microservice/structs"
)

func ReconcileActions(room structs.Room, actions []base.ActionStructure) (batches [][]base.ActionStructure, err error) {

	log.Printf("Reconciling actions...")

	//Initialize map of strings to commandevaluators
	reconcilers := ar.Init()

	curReconciler := reconcilers[room.Configuration.RoomKey]
	if curReconciler == nil {
		err = errors.New("No reconciler corresponding to key " + room.Configuration.RoomKey)
		return
	}

	actions, err = curReconciler.Reconcile(actions)
	if err != nil {
		return
	}

	return
}

/*
ReplaceIPAddressEndpoint is a simple helper
*/
func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}
