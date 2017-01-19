package init

//Initializer is the interface programmed against to build a new roomInitializer
type RoomInitializer interface {

  func Initialize() error
}
