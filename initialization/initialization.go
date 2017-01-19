package init

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

  }

  return InitializerMap
}
