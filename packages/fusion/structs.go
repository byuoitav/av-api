package fusion

// Exported structs

// RoomsResponse is a struct for receiving responses from Fusion
type RoomsResponse struct {
	Rooms []Room `json:"API_Rooms"`
}

type Availability struct {
	Available bool `json:"RawValue"`
}

// Room is a clean struct representing a room populated with information from Fusion
type Room struct {
	RoomName  string
	RoomID    string
	Hostname  string   `json:",omitempty"`
	Address   string   `json:",omitempty"`
	Building  string   `json:",omitempty"`
	Room      string   `json:",omitempty"`
	Available bool     `json:",omitempty"`
	Symbols   []Symbol `json:",omitempty"`
}

type Symbol struct {
	ProcessorName string
	ConnectInfo   string
	SymbolID      string
	Signals       []Signal
}

type Signal struct {
	AttributeID string
	RawValue    string
	SymbolID    string
}

// Unexported structs

type recordCount struct {
	Count int `json:"TotalRecords"`
}
