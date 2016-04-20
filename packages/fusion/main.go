package fusion

import (
	"encoding/json"
	"errors"
	"fmt"
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
func GetRoomID(building string, room string) (int, error) {
	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?search="+building+"+"+room)
	if err != nil {
		return -1, err
	}

	rooms := Response{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return -1, err
	}

	fmt.Printf("%+v", rooms)

	if len(rooms.Rooms) == 0 { // Return an error if Fusion doesn't have record of the room specified
		return -1, errors.New("Could not find room " + room + " in the " + building + " building in the Fusion database")
	} else if len(rooms.Rooms) > 1 {
		return -1, errors.New("Your search \"" + building + " " + room + "\" returned multiple results from the Fusion database")
	}

	return 1, nil // TODO: Return the actual ID
}

func RoomAvailable(symbol string) (bool, error) {
	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/SignalValues/"+symbol+"/SYSTEM_POWER")
	if err != nil {
		return false, err
	}

	availability := roomResponseFusion{}
	json.Unmarshal([]byte(response), &availability)

	if len(availability.Response) == 0 { // Return a false positive if Fusion doesn't have the "POWER_ON" symbol for the given room
		return true, nil
	}

	if availability.Response[0].Available { // If the system is currently powered on
		return false, nil
	}

	return true, nil
}

// GetRooms returns all known rooms from the Crestron Fusion database
func GetRooms() (Response, error) {
	count, err := GetRecordCount()
	if err != nil {
		return Response{}, err
	}

	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?pagesize="+strconv.Itoa(count))
	if err != nil {
		return Response{}, err
	}

	rooms := Response{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return Response{}, err
	}

	return rooms, nil
}

// GetRoomByName gets information about a room from only its name (EG: ASB+A203)
func GetRoomByName(roomName string) (Room, error) {
	roomParse := strings.Split("+", roomName)
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

	// Get info about the room using its ID
	response, err = HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/"+strconv.Itoa(roomID))
	if err != nil {
		return Room{}, err
	}

	rooms = Response{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return Room{}, err
	}

	hostname := rooms.Rooms[0].Symbols[0].ProcessorName
	address := rooms.Rooms[0].Symbols[0].ConnectInfo

	roomResponse := room{Building: building, Room: room, Hostname: hostname, Address: address}

	return roomResponse, nil
}
