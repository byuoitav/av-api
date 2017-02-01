package init

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/commandevaluators"
	"github.com/byuoitav/av-api/dbo"
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
	err := runUpdateOnDMPSInRoom(room, false)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return err
	}

	//Creat the nagger.
	go nagger(room)

	return nil
}

func nagger(room accessors.Room) {
	log.Printf("Starting nagger.")
	updateInterval := 15 * time.Minute

	roomName := room.Name
	buildingShortname := room.Building.Shortname

	ticker := time.NewTicker(updateInterval)
	for {
		select {
		case <-ticker.C:
			nag(roomName, buildingShortname)
		}
	}
}

func nag(roomName string, building string) error {
	log.Printf("NAGGER: Nagging the DMPS service.")
	//Get the devices
	room, err := dbo.GetRoomByInfo(building, roomName)
	if err != nil {
		log.Printf("NAGGER: Could not get devices for room.")
		return err
	}

	//get all the DMPS from the DB>
	err = runUpdateOnDMPSInRoom(room, true)
	if err != nil {
		return err
	}
	return nil
}

func logString(message string, nagger bool) {
	if nagger {
		log.Printf("NAGGER: %v", message)
	} else {
		log.Printf("%v", message)
	}
	return
}

func runUpdateOnDMPSInRoom(room accessors.Room, nagger bool) error {
	var actions []base.ActionStructure

	logString("Checking for DMPS devices in the room.", nagger)
	for _, dev := range room.Devices {
		//Get all the DMPS
		if dev.Type == "DMPS" {
			//validate that it has the command
			logString("Found DMPS, looking for UpdateSig command.", nagger)
			if ok, _ := commandevaluators.CheckCommands(dev.Commands, "UpdateSig"); ok {
				logString("Command found.", nagger)
				//build an action so we can use the ExecuteActions to run it.
				actions = append(actions, base.ActionStructure{
					Action:              "UpdateSig",
					GeneratingEvaluator: "Initializer",
					Device:              dev,
					Parameters:          make(map[string]string),
					DeviceSpecific:      true,
				})
			} else {
				logString("No updateSig command found for "+dev.GetFullName(), nagger)
			}
		}
	}

	logString(fmt.Sprintf("Found %v DMPS.", len(actions)), nagger)

	//Checking for the microservice to ensure that it's up.
	if len(actions) > 0 {
		has, cmd := commandevaluators.CheckCommands(actions[0].Device.Commands, "UpdateSig")
		if has {
			var err error
			//Try and connect to the DMPS
			for i := 0; i < 30; i++ {
				logString(fmt.Sprintf("Attempting to connect to %v", cmd.Microservice+"/health"), nagger)
				_, err = http.Get(cmd.Microservice + "/health")
				if err == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}
			if err != nil {
				log.Printf("Could not connect to the Crestron-Control-Microservice at %v to register DMPS.", cmd.Microservice+"/health")
				return err
			}
		}
	}

	logString("Registering...", nagger)

	status, err := commandevaluators.ExecuteActions(actions)
	if err != nil {
		return err
	}
	//check the status and report on it
	for _, curStatus := range status {
		if curStatus.Success {
			logString(fmt.Sprintf("%v registered.", curStatus.Device), nagger)
		} else {
			logString(fmt.Sprintf("Error registering %v. Error: %v", curStatus.Device, curStatus.Err), nagger)
		}
	}
	return nil
}
