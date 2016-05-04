package fusion

import "github.com/byuoitav/hateoas"

// Structs for holding Fusion data

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
	AttributeID   string
	AttributeName string
	AttributeType int // TODO: Map the int to a string (analog, digital, etc.) in the clean struct
	RawValue      string
}

// Clean structs for returning data

// AllRooms is a clean struct for returning room data
type AllRooms struct {
	Links []hateoas.Link `json:"links,omitempty"`
	Rooms []SlimRoom     `json:"rooms"`
}

// SlimRoom is a clean struct representing a room
type SlimRoom struct {
	Links   []hateoas.Link `json:"links,omitempty"`
	Name    string         `json:"name"`
	ID      string
	Signals []Signal `json:"signals,omitempty"`
}

// Room is a clean struct representing a room
type Room struct {
	Links     []hateoas.Link `json:"links,omitempty"`
	Name      string         `json:"name"`
	ID        string
	Hostname  string   `json:"hostname,omitempty"`
	Address   string   `json:"address,omitempty"`
	Building  string   `json:"building,omitempty"`
	Room      string   `json:"room,omitempty"`
	Symbol    string   `json:"symbol,omitempty"`
	Available bool     `json:"available"`
	Health    Health   `json:"health"`
	Signals   []Signal `json:"signals"`
}

type Signal struct {
	Links []hateoas.Link `json:"links,omitempty"`
	Name  string         `json:"name"`
	ID    string
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Health represents the results of various health checks run on each box
type Health struct {
	PingIn  bool
	PingOut bool
}
