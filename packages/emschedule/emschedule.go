package emschedule

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"time"

	"github.com/byuoitav/av-api/packages/soap"
)

var Username string
var Password string

// IsRoomAvailable returns a bool representing whether or not a room is available according to the EMS scheduling system
func IsRoomAvailable(building string, room string) (bool, error) {
	roomID, err := GetRoomID(building, room)
	if err != nil {
		return false, err
	}

	now := time.Now()
	date := now
	startTime := now
	endTime := now.Add(30 * time.Minute) // Check a half hour time interval

	request := &RoomAvailabilityRequestSOAP{
		Username:    Username,
		Password:    Password,
		RoomID:      roomID,
		BookingDate: date,
		StartTime:   startTime,
		EndTime:     endTime,
	}

	encodedRequest, err := soap.Encode(&request)
	if err != nil {
		return false, err
	}

	response, err := soap.Request("https://emsweb.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	if err != nil {
		return false, err
	}

	availability := RoomAvailabilityResponseSOAP{}
	err = soap.Decode([]byte(response), &availability)
	if err != nil {
		return false, err
	}

	roomAvailability := RoomResponse{}
	err = xml.Unmarshal([]byte(availability.Result), &roomAvailability)
	if err != nil {
		return false, err
	}

	return roomAvailability.Response[0].Available, nil
}

// GetRoomID returns the ID of a building from its building code
func GetRoomID(building string, room string) (int, error) {
	buildingID, err := GetBuildingID(building)
	if err != nil {
		return -1, err
	}

	rooms, err := GetAllRooms(buildingID)
	if err != nil {
		return -1, err
	}

	// Some of the room names in the EMS API have asterisks following them for unknown reasons so we have to use a RegEx to ignore them
	re := regexp.MustCompile(`(` + building + " " + room + `)\w*`)

	for index := range rooms.Rooms {
		match := re.FindStringSubmatch(rooms.Rooms[index].Description) // Make the RegEx do magic

		if len(match) != 0 {
			return rooms.Rooms[index].ID, nil
		}
	}

	return -1, fmt.Errorf("Couldn't find a record of the supplied %s room in the %s building in the EMS database", room, building)
}

func GetAllRooms(buildingID int) (AllRooms, error) {
	var buildings []int
	buildings = append(buildings, buildingID)
	request := &AllRoomsRequestSOAP{
		Username:  Username,
		Password:  Password,
		Buildings: buildings,
	}

	encodedRequest, err := soap.Encode(&request)
	if err != nil {
		return AllRooms{}, err
	}

	response, err := soap.Request("https://emsweb.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	if err != nil {
		return AllRooms{}, err
	}

	allRooms := AllRoomsResponseSOAP{}
	err = soap.Decode([]byte(response), &allRooms)
	if err != nil {
		return AllRooms{}, err
	}

	rooms := AllRooms{}
	err = xml.Unmarshal([]byte(allRooms.Result), &rooms)
	if err != nil {
		return AllRooms{}, err
	}

	return rooms, nil
}

func GetBuildingID(buildingCode string) (int, error) {
	buildings, err := GetAllBuildings()
	if err != nil {
		return -1, nil
	}

	for index := range buildings.Buildings {
		if buildings.Buildings[index].BuildingCode == buildingCode {
			return buildings.Buildings[index].ID, nil
		}
	}

	return -1, fmt.Errorf("Couldn't find a record of the supplied %s building in the EMS database", buildingCode)
}

func GetAllBuildings() (AllBuildings, error) {
	request := &AllBuildingsRequestSOAP{
		Username: Username,
		Password: Password,
	}

	encodedRequest, err := soap.Encode(&request)
	if err != nil {
		return AllBuildings{}, err
	}

	response, err := soap.Request("https://emsweb.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	if err != nil {
		return AllBuildings{}, err
	}

	allBuilding := AllBuildingsResponseSOAP{}
	err = soap.Decode([]byte(response), &allBuilding)
	if err != nil {
		return AllBuildings{}, err
	}

	buildings := AllBuildings{}
	err = xml.Unmarshal([]byte(allBuilding.Result), &buildings)
	if err != nil {
		return AllBuildings{}, err
	}

	return buildings, nil
}
