package fusion

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// HTTP is used to quickly make calls to the Crestron Fusion API and request a JSON response
func HTTP(requestType string, url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(requestType, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// GetRecordCount returns an int representing how many records Fusion has (for circumventing pagination)
func GetRecordCount() (int, error) {
	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?pagesize=1")
	if err != nil {
		return -1, err
	}

	count := recordCount{}
	json.Unmarshal(response, &count)

	return count.Count, nil
}

// GetRoomID gets a room's Fusion ID from its building and room name
func GetRoomID(building string, room string) (string, error) {
	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?search="+building+"+"+room)
	if err != nil {
		return "", err
	}

	rooms := RoomsResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return "", err
	}

	if len(rooms.Rooms) == 0 { // Return an error if Fusion doesn't have record of the room specified
		return "", errors.New("Could not find room " + room + " in the " + building + " building in the Fusion database")
	} else if len(rooms.Rooms) > 1 {
		return "", errors.New("Your search \"" + building + " " + room + "\" returned multiple results from the Fusion database")
	}

	return rooms.Rooms[0].RoomID, nil
}

func GetRoomSymbolID(roomID string) (string, error) {
	return "", nil
}

// IsRoomAvailable returns a bool representing whether or not a room is available according to the Fusion "SYSTEM_POWER" symbol
func IsRoomAvailable(roomID string) (bool, error) {
	symbol, err := GetRoomSymbolID(roomID)
	if err != nil {
		return false, err
	}

	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/SignalValues/"+symbol+"/SYSTEM_POWER")
	if err != nil {
		return false, err
	}

	availability := RoomsResponse{}
	json.Unmarshal([]byte(response), &availability)

	if len(availability.Rooms) == 0 { // Return a false positive if Fusion doesn't have the "POWER_ON" symbol for the given room
		return true, nil
	}

	if availability.Rooms[0].Available { // If the system is currently powered on (Fusion does things backwards from how we want)
		return false, nil
	}

	return true, nil
}

// GetRooms returns all known rooms from the Crestron Fusion database
func GetRooms() (RoomsResponse, error) {
	count, err := GetRecordCount()
	if err != nil {
		return RoomsResponse{}, err
	}

	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?pagesize="+strconv.Itoa(count))
	if err != nil {
		return RoomsResponse{}, err
	}

	rooms := RoomsResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return RoomsResponse{}, err
	}

	return rooms, nil
}

// GetRoomByName gets information about a room from only its name (EG: ASB+A203)
func GetRoomByName(roomName string) (Room, error) {
	roomParse := strings.Split(roomName, "-")
	if len(roomParse) != 2 {
		return Room{}, errors.New("Please supply a room name in the format of 'BLDG-ROOM' similar to 'ASB-A203'")
	}

	room, err := GetRoomByNameAndBuilding(roomParse[0], roomParse[1])
	if err != nil {
		return Room{}, err
	}

	return room, nil
}

// GetRoomByNameAndBuilding gets information about a room using its supplied building and room number
func GetRoomByNameAndBuilding(building string, room string) (Room, error) {
	roomID, err := GetRoomID(building, room)
	if err != nil {
		return Room{}, err
	}

	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/"+roomID)
	if err != nil {
		return Room{}, err
	}

	rooms := RoomsResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return Room{}, err
	}

	hostname := rooms.Rooms[0].Symbols[0].ProcessorName
	address := rooms.Rooms[0].Symbols[0].ConnectInfo
	availability, err := IsRoomAvailable(roomID)
	if err != nil {
		return Room{}, err
	}

	roomResponse := Room{
		RoomID:    roomID,
		RoomName:  building + "-" + room,
		Building:  building,
		Room:      room,
		Hostname:  hostname,
		Address:   address,
		Available: availability,
		Symbols:   rooms.Rooms[0].Symbols,
	}

	return roomResponse, nil
}
