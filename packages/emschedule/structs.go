package emschedule

import "time"

type roomAvailabilityRequestEMS struct {
	XMLName     struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailability"`
	Username    string   `xml:"UserName"`
	Password    string
	RoomID      int
	BookingDate time.Time
	StartTime   time.Time
	EndTime     time.Time
}

type roomAvailabilityResponseEMS struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailabilityResponse"`
	Result  string   `xml:"GetRoomAvailabilityResult"`
}

type roomResponseEMS struct {
	Response []roomAvailabilityEMS `xml:"Data"`
}

type roomAvailabilityEMS struct {
	Available bool
}

type allBuildingsRequest struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

type allBuildingsResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

type allRoomsRequest struct {
	XMLName   struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRooms"`
	Username  string   `xml:"UserName"`
	Password  string
	Buildings []int `xml:"int"`
}

type allRoomsResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomsResponse"`
	Result  string   `xml:"GetRoomsResult"`
}

type buildingRequest struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

type buildingResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

type allBuildings struct {
	Buildings []building `xml:"Data"`
}

type building struct {
	BuildingCode string `xml:"BuildingCode"`
	ID           int    `xml:"ID"`
	Description  string `xml:"Description"`
}

type allRooms struct {
	Rooms []room `xml:"Data"`
}

type room struct {
	Room        string
	ID          int    `xml:"ID"`
	Description string `xml:"Description"`
	Available   bool
}
