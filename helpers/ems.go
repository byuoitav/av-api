package helpers

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"time"
)

type AllBuildingsRequest struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

type AllBuildingsResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

type AllRoomsRequest struct {
	XMLName   struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRooms"`
	Username  string   `xml:"UserName"`
	Password  string
	Buildings []int `xml:"int"`
}

type AllRoomsResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomsResponse"`
	Result  string   `xml:"GetRoomsResult"`
}

// BuildingRequest represents an EMS API request for one building (by building ID)
type BuildingRequest struct {
	XMLName  struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildings"`
	Username string   `xml:"UserName"`
	Password string
}

// BuildingResponse represents the EMS API's response to a request for one building (by building ID)
type BuildingResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetBuildingsResponse"`
	Result  string   `xml:"GetBuildingsResult"`
}

type RoomAvailabilityRequest struct {
	XMLName     struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailability"`
	Username    string   `xml:"UserName"`
	Password    string
	RoomID      int
	BookingDate time.Time
	StartTime   time.Time
	EndTime     time.Time
}

type RoomAvailabilityResponse struct {
	XMLName struct{} `xml:"http://DEA.EMS.API.Web.Service/ GetRoomAvailabilityResponse"`
	Result  string   `xml:"GetRoomAvailabilityResult"`
}

// AllBuildings is a clean struct representing all buildings returned by the EMS API
type AllBuildings struct {
	Buildings []Building `xml:"Data"`
}

// Building is a clean struct representing a single building returned by the EMS API
type Building struct {
	BuildingCode string `xml:"BuildingCode"`
	ID           int    `xml:"ID"`
	Description  string `xml:"Description"`
}

// AllRooms is a clean struct representing all rooms returned for a building by the EMS API
type AllRooms struct {
	Rooms []Room `xml:"Data"`
}

// Room is a clean struct representing a single room returned by the EMS API
type Room struct {
	Room        string
	ID          int    `xml:"ID"`
	Description string `xml:"Description"`
}

// GetAllBuildings retrieves a list of all buildings listed by the EMS API
func GetAllBuildings() AllBuildings {
	request := &AllBuildingsRequest{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD")}
	encodedRequest, err := SoapEncode(&request)
	CheckErr(err)

	response := SoapRequest("https://emsweb-dev.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	allBuildings := AllBuildingsResponse{}
	err = SoapDecode([]byte(response), &allBuildings)
	CheckErr(err)

	buildings := AllBuildings{}
	err = xml.Unmarshal([]byte(allBuildings.Result), &buildings)
	CheckErr(err)

	return buildings
}

// GetBuildingID returns the ID of a building from its building code
func GetBuildingID(buildingCode string) (int, error) {
	buildings := GetAllBuildings()

	// fmt.Printf("All Buildings: %v\n", buildings)

	for index := range buildings.Buildings {
		// fmt.Printf("Building ID: %v\n", buildings.Buildings[index].ID)

		if buildings.Buildings[index].BuildingCode == buildingCode {
			return buildings.Buildings[index].ID, nil
		}
	}

	return -1, fmt.Errorf("Couldn't find a record of the supplied %s building", buildingCode)
}

// GetAllRooms retrieves a list of all rooms listed by the EMS API as belonging to the specified building
func GetAllRooms(buildingID int) AllRooms {
	var buildings []int
	buildings = append(buildings, buildingID)
	request := &AllRoomsRequest{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD"), Buildings: buildings}
	encodedRequest, err := SoapEncode(&request)
	CheckErr(err)

	// fmt.Printf("Request: %s\n", encodedRequest)

	response := SoapRequest("https://emsweb-dev.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	allRooms := AllRoomsResponse{}
	err = SoapDecode([]byte(response), &allRooms)
	CheckErr(err)

	// fmt.Printf("All Rooms: %s\n", allRooms)

	rooms := AllRooms{}
	err = xml.Unmarshal([]byte(allRooms.Result), &rooms)
	CheckErr(err)

	return rooms
}

// GetRoomID returns the ID of a building from its building code
func GetRoomID(building string, room string) (int, error) {
	buildingID, err := GetBuildingID(building)
	CheckErr(err)

	// fmt.Printf("Found Building ID: %v\n", buildingID)

	rooms := GetAllRooms(buildingID)

	// fmt.Printf("Rooms: %v\n", rooms)

	// Some of the room names in the EMS API have asterisks following them for unknown reasons so we have to use a REGEX to ignore them
	re := regexp.MustCompile(`(` + building + " " + room + `)\w*`)

	for index := range rooms.Rooms {
		// fmt.Printf("Room Info: %s\n", rooms.Rooms[index].Room)
		match := re.FindStringSubmatch(rooms.Rooms[index].Room) // Make the RegEx do magic

		if len(match) != 0 {
			return rooms.Rooms[index].ID, nil
		}
	}

	return -1, fmt.Errorf("Couldn't find a record of the supplied %s room in the %s building", room, building)
}
