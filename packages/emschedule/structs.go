package emschedule

import "time"

// SOAP structs for calls

type RoomAvailabilityRequestSOAP struct {
	XMLName     struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailability"`
	Username    string   `xml:"UserName"`
	Password    string
	RoomID      int
	BookingDate time.Time
	StartTime   time.Time
	EndTime     time.Time
}

type RoomAvailabilityResponseSOAP struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailabilityResponse"`
	Result  string   `xml:"GetRoomAvailabilityResult"`
}

type AllBuildingsRequestSOAP struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

type AllBuildingsResponseSOAP struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

type AllRoomsRequestSOAP struct {
	XMLName   struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRooms"`
	Username  string   `xml:"UserName"`
	Password  string
	Buildings []int `xml:"int"`
}

type AllRoomsResponseSOAP struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomsResponse"`
	Result  string   `xml:"GetRoomsResult"`
}

type BuildingRequestSOAP struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

type BuildingResponseSOAP struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

// Structs for holding data from EMS

type RoomResponse struct {
	Response []RoomAvailability `xml:"Data"`
}

type RoomAvailability struct {
	Available bool
}

type EMSchedulingAllBuildings struct {
	Buildings []EMSchedulingBuilding `xml:"Data"`
}

type EMSchedulingBuilding struct {
	BuildingCode string `xml:"BuildingCode"`
	ID           int    `xml:"ID"`
	Description  string `xml:"Description"`
}

type EMSchedulingAllRooms struct {
	Rooms []EMSchedulingRoom `xml:"Data"`
}

type EMSchedulingRoom struct {
	Room        string
	ID          int    `xml:"ID"`
	Description string `xml:"Description"`
	Available   bool
}

// Clean structs for returning data

type AllBuildings struct {
	Buildings []Building `xml:"Data"`
}

type Building struct {
	BuildingCode string `xml:"BuildingCode"`
	ID           int    `xml:"ID"`
	Description  string `xml:"Description"`
}

type AllRooms struct {
	Rooms []Room `xml:"Data"`
}

type Room struct {
	Room        string
	ID          int    `xml:"ID"`
	Description string `xml:"Description"`
	Available   bool
}
