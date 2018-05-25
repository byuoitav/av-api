package init

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"
)

/*
CheckRoomInitialization will check if the system is running locally (if it
should be mapped to a room). If yes, pull the room configuration and run the
init code.
*/
func CheckRoomInitialization() error {

	base.Log("Initializing.")

	//Check if local
	if len(os.Getenv("LOCAL_ENVIRONMENT")) < 1 {
		base.Log("Not a local instance of the API.")
		base.Log("Done.")
		return nil
	}

	base.Log("Getting room information.")

	/*
	  It's not local, parse the hostname for the building room
	  hostname must be in the format {{BuildingShortname}}-{{RoomIdentifier}}
	  or buildling hyphen room. e.g. ITB-1001D
	*/

	hostname := os.Getenv("PI_HOSTNAME")

	splitValues := strings.Split(hostname, "-")
	roomID := fmt.Sprintf("%v-%v", splitValues[0], splitValues[1])
	base.Log("Room %v", roomID)

	attempts := 0

	room, err := db.GetDB().GetRoom(roomID)
	if err != nil {

		//If there was an error we want to attempt to connect multiple times - as the
		//configuration service may not be up.
		for attempts < 40 {
			base.Log("Attempting to connect to DB...")
			room, err = db.GetDB().GetRoom(roomID)
			if err != nil {
				base.Log("Error: %s", err.Error())
				attempts++
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
		if attempts > 30 && err != nil {
			base.Log("Error Retrieving room information.")
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

	return errors.New("No initializer for the key in configuration")
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
