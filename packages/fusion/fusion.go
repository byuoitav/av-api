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

	count := FusionRecordCount{}
	json.Unmarshal(response, &count)

	return count.TotalRecords, nil
}

// TranslateFusionSignalTypes takes Fusion-returned integer values and returns a human-readable string describing the signal's type
func TranslateFusionSignalTypes(signalType int) (string, error) {
	knownSignals := []string{"analog", "digital", "serial"}

	if signalType-1 < len(knownSignals) {
		return knownSignals[signalType-1], nil
	}

	return "", errors.New("Unknown signal type: " + string(signalType))
}

// IsRoomAvailable returns a bool representing whether or not a room is available according to the Fusion "SYSTEM_POWER" symbol
func IsRoomAvailable(symbolID string) (bool, error) {
	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/SignalValues/"+symbolID+"/SYSTEM_POWER")
	if err != nil {
		return false, err
	}

	availability := FusionAllRooms{}
	json.Unmarshal([]byte(response), &availability)

	if len(availability.APIRooms) == 0 { // Return a false positive if Fusion doesn't have the "POWER_ON" symbol for the given room
		return true, nil
	}

	if availability.APIRooms[0].Available { // If the system is currently powered on (Fusion does things backwards from how we want)
		return false, nil
	}

	return true, nil
}

// GetRoomID gets a room's Fusion ID from its building and room name
func GetRoomID(building string, room string) (string, error) {
	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?search="+building+"+"+room)
	if err != nil {
		return "", err
	}

	rooms := FusionAllRooms{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return "", err
	}

	if len(rooms.APIRooms) == 0 { // Return an error if Fusion doesn't have record of the room specified
		return "", errors.New("Could not find room " + room + " in the " + building + " building in the Fusion database")
	} else if len(rooms.APIRooms) > 1 {
		return "", errors.New("Your search \"" + building + " " + room + "\" returned multiple results from the Fusion database")
	}

	return rooms.APIRooms[0].RoomID, nil
}

// GetAllRooms returns all known rooms from the Crestron Fusion database
func GetAllRooms() (AllRooms, error) {
	count, err := GetRecordCount()
	if err != nil {
		return AllRooms{}, err
	}

	response, err := HTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?pagesize="+strconv.Itoa(count))
	if err != nil {
		return AllRooms{}, err
	}

	fusionRooms := FusionAllRooms{}
	err = json.Unmarshal(response, &fusionRooms)
	if err != nil {
		return AllRooms{}, err
	}

	rooms := AllRooms{}

	for i := 0; i < len(fusionRooms.APIRooms); i++ {
		fusionRoom := fusionRooms.APIRooms[i]

		room := SlimRoom{
			Name: fusionRoom.RoomName,
			ID:   fusionRoom.RoomID,
		}

		rooms.Rooms = append(rooms.Rooms, room)
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

	rooms := FusionAllRooms{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return Room{}, err
	}

	sampleSymbol := rooms.APIRooms[0].Symbols[0]

	hostname := sampleSymbol.ProcessorName
	address := sampleSymbol.ConnectInfo

	roomResponse := Room{
		Name:     building + "-" + room,
		ID:       roomID,
		Building: building,
		Room:     room,
		Hostname: hostname,
		Address:  address,
		Symbol:   sampleSymbol.SymbolID,
	}

	for i := range rooms.APIRooms[0].Symbols[0].Signals {
		fusionSignal := rooms.APIRooms[0].Symbols[0].Signals[i]
		signalType, err := TranslateFusionSignalTypes(fusionSignal.AttributeType)
		if err != nil {
			return Room{}, err
		}

		signal := Signal{
			Name:  fusionSignal.AttributeName,
			ID:    fusionSignal.AttributeID,
			Type:  signalType,
			Value: fusionSignal.RawValue,
		}

		roomResponse.Signals = append(roomResponse.Signals, signal)
	}

	return roomResponse, nil
}

func GetAllSignalsByRoomAndBuilding(building string, room string) (SlimRoom, error) {
	fullRoom, err := GetRoomByNameAndBuilding(building, room)
	if err != nil {
		return SlimRoom{}, err
	}

	slimRoom := SlimRoom{
		Name: fullRoom.Name,
		ID:   fullRoom.ID,
	}

	for i := range fullRoom.Signals {
		signal := Signal{
			Name:  fullRoom.Signals[i].Name,
			ID:    fullRoom.Signals[i].ID,
			Type:  fullRoom.Signals[i].Type,
			Value: fullRoom.Signals[i].Value,
		}

		slimRoom.Signals = append(slimRoom.Signals, signal)
	}

	return slimRoom, nil
}

func GetSignalByRoomAndBuilding(building string, room string, signalName string) (SlimRoom, error) {
	fullRoom, err := GetRoomByNameAndBuilding(building, room)
	if err != nil {
		return SlimRoom{}, err
	}

	slimRoom := SlimRoom{
		Name: fullRoom.Name,
		ID:   fullRoom.ID,
	}

	for i := range fullRoom.Signals {
		if strings.ToLower(fullRoom.Signals[i].Name) == strings.ToLower(signalName) {
			signal := Signal{
				Name:  fullRoom.Signals[i].Name,
				ID:    fullRoom.Signals[i].ID,
				Type:  fullRoom.Signals[i].Type,
				Value: fullRoom.Signals[i].Value,
			}

			slimRoom.Signals = append(slimRoom.Signals, signal)
		}
	}

	return slimRoom, nil
}
