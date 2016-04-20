package fusion

type FusionRecordCount struct {
	TotalRecords int `json:"TotalRecords"`
}

type FusionAvailability struct {
	RawValue bool
}

// FusionAllRooms is a struct for receiving responses from Fusion
type FusionAllRooms struct {
	APIRooms []FusionRoom `json:"API_Rooms"`
}

// FusionRoom is a struct representing a room in Fusion's data
type FusionRoom struct {
	RoomName  string
	RoomID    string
	Hostname  string
	Address   string
	Building  string
	Room      string
	Available bool
	Symbols   []FusionSymbol
}

type FusionSymbol struct {
	ProcessorName string
	ConnectInfo   string
	SymbolID      string
	Signals       []FusionSignal
}

type FusionSignal struct {
	AttributeID string
	RawValue    string
	SymbolID    string
}

// AllRooms is a clean struct for returning room data
type AllRooms struct {
	Rooms []Room `json:"rooms"`
}

// Room is a clean struct representing a room
type Room struct {
	Name      string `json:"name"`
	ID        string
	Hostname  string `json:"hostname,omitempty"`
	Address   string `json:"address,omitempty"`
	Building  string `json:"building,omitempty"`
	Room      string `json:"room,omitempty"`
	Available bool   `json:"available,omitempty"`
}
