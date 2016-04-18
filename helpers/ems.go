package helpers

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
)

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

func getallBuildings() allBuildings {
	request := &allBuildingsRequest{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD")}
	encodedRequest, err := SoapEncode(&request)
	CheckErr(err)

	response := SoapRequest("https://emsweb-dev.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	allBuildingsContainer := allBuildingsResponse{}
	err = SoapDecode([]byte(response), &allBuildingsContainer)
	CheckErr(err)

	buildings := allBuildings{}
	err = xml.Unmarshal([]byte(allBuildingsContainer.Result), &buildings)
	CheckErr(err)

	// fmt.Printf("%v\n", buildings)

	return buildings
}

func getBuildingID(buildingCode string) (int, error) {
	buildings := getallBuildings()

	for index := range buildings.Buildings {
		if buildings.Buildings[index].BuildingCode == buildingCode {
			return buildings.Buildings[index].ID, nil
		}
	}

	return -1, fmt.Errorf("Couldn't find a record of the supplied %s building", buildingCode)
}

func getAllRooms(buildingID int) allRooms {
	var buildings []int
	buildings = append(buildings, buildingID)
	request := &allRoomsRequest{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD"), Buildings: buildings}
	encodedRequest, err := SoapEncode(&request)
	CheckErr(err)

	response := SoapRequest("https://emsweb-dev.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	allRoomsContainer := allRoomsResponse{}
	err = SoapDecode([]byte(response), &allRoomsContainer)
	CheckErr(err)

	rooms := allRooms{}
	err = xml.Unmarshal([]byte(allRoomsContainer.Result), &rooms)
	CheckErr(err)

	// fmt.Printf("%v\n", rooms)

	return rooms
}

// GetRoomID returns the ID of a building from its building code
func GetRoomID(building string, room string) (int, error) {
	buildingID, err := getBuildingID(building)
	CheckErr(err)

	rooms := getAllRooms(buildingID)

	// Some of the room names in the EMS API have asterisks following them for unknown reasons so we have to use a RegEx to ignore them
	re := regexp.MustCompile(`(` + building + " " + room + `)\w*`)

	for index := range rooms.Rooms {
		match := re.FindStringSubmatch(rooms.Rooms[index].Description) // Make the RegEx do magic

		if len(match) != 0 {
			return rooms.Rooms[index].ID, nil
		}
	}

	return -1, fmt.Errorf("Couldn't find a record of the supplied %s room in the %s building", room, building)
}
