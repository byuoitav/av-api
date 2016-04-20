package emschedule

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/byuoitav/av-api/packages/soap"
)

func IsRoomAvailable(building string, room string) (bool, error) {
	roomID, err := GetRoomID(building, room)
	if err != nil {
		return false, err
	}

	now := time.Now()
	date := now
	startTime := now
	endTime := now.Add(30 * time.Minute) // Check a half hour time interval

	request := &roomAvailabilityRequestEMS{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD"), RoomID: roomID, BookingDate: date, StartTime: startTime, EndTime: endTime}
	encodedRequest, err := soap.Encode(&request)
	if err != nil {
		return false, err
	}

	response, err := soap.Request("https://emsweb.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	if err != nil {
		return false, err
	}

	availability := roomAvailabilityResponseEMS{}
	err = soap.Decode([]byte(response), &availability)
	if err != nil {
		return false, err
	}

	roomAvailability := roomResponseEMS{}
	err = xml.Unmarshal([]byte(availability.Result), &roomAvailability)
	if err != nil {
		return false, err
	}

	return roomAvailability.Response[0].Available, nil
}

func getallBuildings() (allBuildings, error) {
	request := &allBuildingsRequest{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD")}
	encodedRequest, err := soap.Encode(&request)
	if err != nil {
		return allBuildings{}, err
	}

	response, err := soap.Request("https://emsweb.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	if err != nil {
		return allBuildings{}, err
	}

	allBuildingsContainer := allBuildingsResponse{}
	err = soap.Decode([]byte(response), &allBuildingsContainer)
	if err != nil {
		return allBuildings{}, err
	}

	buildings := allBuildings{}
	err = xml.Unmarshal([]byte(allBuildingsContainer.Result), &buildings)
	if err != nil {
		return allBuildings{}, err
	}

	return buildings, nil
}

func getBuildingID(buildingCode string) (int, error) {
	buildings, err := getallBuildings()
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

func getAllRooms(buildingID int) (allRooms, error) {
	var buildings []int
	buildings = append(buildings, buildingID)
	request := &allRoomsRequest{Username: os.Getenv("EMS_API_USERNAME"), Password: os.Getenv("EMS_API_PASSWORD"), Buildings: buildings}
	encodedRequest, err := soap.Encode(&request)
	if err != nil {
		return allRooms{}, err
	}

	response, err := soap.Request("https://emsweb.byu.edu/EMSAPI/Service.asmx", encodedRequest)
	if err != nil {
		return allRooms{}, err
	}

	allRoomsContainer := allRoomsResponse{}
	err = soap.Decode([]byte(response), &allRoomsContainer)
	if err != nil {
		return allRooms{}, err
	}

	rooms := allRooms{}
	err = xml.Unmarshal([]byte(allRoomsContainer.Result), &rooms)
	if err != nil {
		return allRooms{}, err
	}

	return rooms, nil
}

// GetRoomID returns the ID of a building from its building code
func GetRoomID(building string, room string) (int, error) {
	buildingID, err := getBuildingID(building)
	if err != nil {
		return -1, err
	}

	rooms, err := getAllRooms(buildingID)
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
