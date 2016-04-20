package fusion

// Exported structs

// RoomsResponse is a clean struct for parsing responses from Fusion and for sending our own responses
type RoomsResponse struct {
	Rooms []Room `json:"API_Rooms"`
}

type SignalsResponse struct {
	Signals []Availability `json:"API_Signals"`
}

type Availability struct {
	Available bool `json:"RawValue"`
}

// Room is a clean struct representing a room populated with information from Fusion
type Room struct {
	RoomID    string
	RoomName  string
	Symbols   []Symbol
	Building  string
	Room      string
	Hostname  string
	Address   string
	Available bool
}

type Symbol struct {
	ProcessorName string
	ConnectInfo   string
	SymbolID      string
}

// Unexported structs

type recordCount struct {
	Count int `json:"TotalRecords"`
}
