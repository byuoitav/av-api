package init

/*
CheckRoomInitialization will check if the system is running locally (if it
should be mapped to a room). If yes, pull the room configuration and run the
init code.
*/
func CheckRoomInitialization() error {

  log.Printf("Initializing.")

  //Check if local
  if len(os.Getenv("LOCAL_ENVIRONMENT")) < 1  {
    log.Printf("Not a local instance of the API.")
    log.Printf("Done.")
    return nil
  }

  log.Printf("Getting room information.")

  /*
  It's not local, parse the hostname for the building room
  hostname must be in the format {{BuildingShortname}}-{{RoomIdentifier}}
  or buildling hyphen room. e.g. ITB-1001A
  */
  hostname := os.Getenv("HOSTNAME")
  splitValues := hostname.split("-")
  log.Printf("Room %v-%v", splitValues[0], splitValues[1])

  room, err := dbo.GetRoomByInfo(splitValues[0], splitValues[1])
  if err != nil {
    log.Printf("Error Retrieving room information.")
    return err
  }

  //There is no initializer, no need to run code
  if len(room.Configuration.InitKey) < 1 {
    return nil
  }

  //take our room and get the init key
  initMap := Init()
  if initializor, ok := initMap[room.Configuration.InitKey] ; ok {
   initializor.Initialize()
   return nil
  }

  return errors.New("No initializer for the key in configuration")
}

//Initializer is the interface programmed against to build a new roomInitializer
type RoomInitializer interface {

  /*
  Initizlize performs the actions necessary for the room on startup.
  This is called when the AV-API service is spun up.
  */
  func Initialize() error
}

//CommandMap is the map that contains the initializers
var InitializerMap = make(map[string]RoomInitializer)
var roomInitializerBuilt = false

//Init builds or returns the CommandMap
func Init() map[string]RoomInitializer {
  if !roomInitializerBuilt {
    //Add the new initializers here
    RoomInitializer["Default"] = &DefaultInitializer{}
  }

  return InitializerMap
}
