package init

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

/*
CheckRoomInitialization will check if the system is running locally (if it
should be mapped to a room). If yes, pull the room configuration and run the
init code.
*/
func CheckRoomInitialization() error {

	log.L.Info("[init] Initializing.")

	//Check if local
	if len(os.Getenv("ROOM_SYSTEM")) < 1 {
		log.L.Info("[init] Not a local instance of the API.")
		log.L.Info("[init] Done.")
		return nil
	}

	log.L.Info("[init] Getting room information.")

	/*
	  It's not local, parse the hostname for the building room
	  hostname must be in the format {{BuildingShortname}}-{{RoomIdentifier}}
	  or buildling hyphen room. e.g. ITB-1001D
	*/

	hostname := os.Getenv("SYSTEM_ID")
	if len(hostname) == 0 {
		log.L.Fatal("SYSTEM_ID is not set.")
	}

	splitValues := strings.Split(hostname, "-")
	roomID := fmt.Sprintf("%v-%v", splitValues[0], splitValues[1])
	log.L.Infof("[init] Room %v", roomID)

	attempts := 0

	room, err := db.GetDB().GetRoom(roomID)
	if err != nil {

		//If there was an error we want to attempt to connect multiple times - as the
		//configuration service may not be up.
		for attempts < 40 {
			log.L.Info("[init] Attempting to connect to DB...")
			room, err = db.GetDB().GetRoom(roomID)
			if err != nil {
				log.L.Errorf("[init] Error: %s", err.Error())
				attempts++
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
		if attempts > 30 && err != nil {
			log.L.Error("[init] Error Retrieving room information.")
			return err
		}
	}

	//There is no initializer, no need to run code
	if len(room.Configuration.Description) < 1 {
		return nil
	}

	//take our room and get the init key
	initMap := getMap()
	if initializor, ok := initMap[room.Configuration.Description]; ok {
		initializor.Initialize(room)
		return nil
	}

	msg := fmt.Sprintf("[init] No initializer for the key in configuration")
	log.L.Error(msg)
	return errors.New(msg)
}

//RoomInitializer is the interface programmed against to build a new roomInitializer
type RoomInitializer interface {

	/*
	  Initizlize performs the actions necessary for the room on startup.
	  This is called when the AV-API service is spun up.
	*/
	Initialize(structs.Room) error
}

//InitializerMap is the map that contains the initializers
var InitializerMap = make(map[string]RoomInitializer)
var roomInitializerBuilt = false

//Init builds or returns the CommandMap
func getMap() map[string]RoomInitializer {
	if !roomInitializerBuilt {
		//Add the new initializers here
		InitializerMap["Default"] = &DefaultInitializer{}
	}

	return InitializerMap
}
