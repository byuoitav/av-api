package init

import (
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/commandevaluators"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//DMPSInitializer implements the Initializer interface for rooms that have
// a DMPS as a control processor.
type DMPSInitializer struct {
}

/*
Initialize fulfils the interface requirements.

We need to 'register' each DMPS with the Crestron-Control-Microservice.
Then spawn a 'nagger' that will have the service check the sig file every 15 minutes
to see if it needs to be refreshed.
*/
func (i *DMPSInitializer) Initialize(room accessors.Room) error {
	log.Printf("Running the DMPSInitializer.")
	var actions []base.ActionStructure

	log.Printf("Checking for DMPS devices in the room.")
	for _, dev := range room.Devices {
		//Get all the DMPS
		if dev.Type == "DMPS" {
			//validate that it has the command
			log.Printf("Found DMPS, looking for UpdateSig command.")
			if ok, _ := commandevaluators.CheckCommands(dev.Commands, "UpdateSig"); ok {
				log.Printf("Command found.")
				//build an action so we can use the ExecuteActions to run it.
				actions = append(actions, base.ActionStructure{
					Action:              "UpdateSig",
					GeneratingEvaluator: "Initializer",
					Device:              dev,
					Parameters:          make(map[string]string),
					DeviceSpecific:      true,
				})
			} else {
				log.Printf("No updateSig command found for %v", dev.GetFullName())
			}
		}
	}

	log.Pritnf("Found %v DMPS.", len(actions))
	log.Printf("Registering...")

	status, err := commandevaluators.ExecuteActions(actions)
	if err != nil {
		return err
	}
	//check the status and report on it
	for _, curStatus := range status {
		if curStatus.Success {
			log.Printf("%v registered.", curStatus.Device)
		} else {
			log.Printf("Error registering %v. Error: %v", curStatus.Device, curStatus.Err)
		}
	}

	return nil
}
