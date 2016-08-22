package controllers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/fusion"
	"github.com/byuoitav/configuration-database-microservice/accessors"
	"github.com/byuoitav/hateoas"
	"github.com/labstack/echo"
)

var databaseLocation = "http://localhost:8006"

func isRoomAvailable(room fusion.Room) (fusion.Room, error) {
	available, err := helpers.IsRoomAvailable(room)
	if err != nil {
		return fusion.Room{}, err
	}

	room.Available = available

	return room, nil
}

// GetAllRooms returns a list of all rooms Crestron Fusion knows about
func GetAllRooms(context echo.Context) error {
	allRooms, err := fusion.GetAllRooms()
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	// Add HATEOAS links
	for i := range allRooms.Rooms {
		links, err := hateoas.AddLinks(context.Path(), []string{strings.Replace(allRooms.Rooms[i].Name, " ", "-", -1)})
		if err != nil {
			return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allRooms.Rooms[i].Links = links
	}

	return context.JSON(http.StatusOK, allRooms)
}

// GetRoomByName get a room from Fusion using only its name
func GetRoomByName(context echo.Context) error {
	room, err := fusion.GetRoomByName(context.Param("room"))
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	links, err := hateoas.AddLinks(context.Path(), []string{context.Param("building"), context.Param("room")})
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Links = links

	health, err := helpers.GetHealth(room.Address)
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	room.Health = health

	room, err = isRoomAvailable(room)
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	return context.JSON(http.StatusOK, room)
}

// GetAllRoomsByBuilding pulls room information from fusion by building designator
func GetAllRoomsByBuilding(context echo.Context) error {
	allRooms, err := fusion.GetAllRooms()
	if err != nil {
		return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
	}

	// Remove rooms that are not in the asked-for building
	for i := len(allRooms.Rooms) - 1; i >= 0; i-- {
		roomBuilding := strings.Split(allRooms.Rooms[i].Name, " ")

		if roomBuilding[0] != context.Param("building") {
			allRooms.Rooms = append(allRooms.Rooms[:i], allRooms.Rooms[i+1:]...)
		}
	}

	// Add HATEOAS links
	for i := range allRooms.Rooms {
		room := strings.Split(allRooms.Rooms[i].Name, " ")

		links, err := hateoas.AddLinks(context.Path(), []string{context.Param("building"), room[1]})
		if err != nil {
			return context.JSON(http.StatusBadRequest, helpers.ReturnError(err))
		}

		allRooms.Rooms[i].Links = links
	}

	return context.JSON(http.StatusOK, allRooms)
}

// GetRoomByNameAndBuilding is almost identical to GetRoomByName
func GetRoomByNameAndBuilding(context echo.Context) error {
	//room, err := fusion.GetRoomByNameAndBuilding(context.Param("building"), context.Param("room"))
	return nil
}

func getData(url string, structToFill interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, structToFill)
	if err != nil {
		return err
	}
	return nil
}

func getRoomByInfo(roomName string, buildingName string) (accessors.Room, error) {
	url := databaseLocation + "/buildings/" + buildingName + "/rooms/" + roomName
	var toReturn accessors.Room
	err := getData(url, &toReturn)
	return toReturn, err
}

func getDeviceByName(roomName string, buildingName string, deviceName string) (accessors.Device, error) {
	var toReturn accessors.Device
	err := getData(databaseLocation+"/buildings/"+buildingName+"/ "+roomName+"/devices/"+deviceName, &toReturn)
	return toReturn, err
}

func getDevicesByRoom(roomName string, buildingName string) ([]accessors.Device, error) {
	var toReturn []accessors.Device

	resp, err := http.Get(databaseLocation + "/buildings/" + buildingName + "/ " + roomName + "/devices")

	if err != nil {
		return toReturn, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return toReturn, err
	}

	err = json.Unmarshal(b, &toReturn)
	if err != nil {
		return toReturn, err
	}

	return toReturn, nil
}

//PublicRoom is the struct that is returned (or put) as part of the public API
type PublicRoom struct {
	CurrentVideoInput string
	CurrentAudioInput string
	Displays          []Display
	AudioDevices      []AudioDevice
}

//AudioDevice represents an audio device
type AudioDevice struct {
	Name   string
	Power  string
	Muted  bool
	Volume int
}

//Display represents a display
type Display struct {
	Name    string
	Power   string
	Blanked bool
}

func getDevicesByBuildingAndRoomAndRole(room string, building string, roleName string) ([]accessors.Device, error) {

	resp, err := http.Get(databaseLocation + "/buildings/" + building + "/rooms/" + room + "/devices/role/" + roleName)
	if err != nil {
		return []accessors.Device{}, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []accessors.Device{}, err
	}

	var devices []accessors.Device
	err = json.Unmarshal(b, &devices)
	if err != nil {
		return []accessors.Device{}, err
	}

	return devices, nil
}

func validateChangeInputSuppliedValue(deviceToCheck string, room string, building string, roleName string) (bool, error) {

	if len(deviceToCheck) > 0 {
		devices, err := getDevicesByBuildingAndRoomAndRole(room, building, roleName)
		if err != nil {
			return false, err
		}
		if len(devices) < 1 {
			return false, errors.New("No " + roleName + " input devices in room.")
		}

		for _, val := range devices {
			if strings.EqualFold(deviceToCheck, val.Name) || strings.EqualFold(deviceToCheck, val.Type) {
				return true, nil
			}
		}
	}
	return false, errors.New("Invalid " + roleName + " devices sepecified.")
}

/*PutRoomChanges is the handler to accept puts to /buildlings/:buildling/rooms/:room
	with the json payload with one or more of the fields:
	{
    "currentInput": "computer",
    "displays": [{
        "name": "dp1",
        "power": "on",
        "blanked": false
    }],
		"audioDeivices": [{
		"muted": false,
		"volume": 50
		}]
	}
	Or the 'PublicRoom' struct defined in this package.
}
*/
func PutRoomChanges(context echo.Context) error {
	building, room := context.Param("building"), context.Param("room")

	var roomInQuestion PublicRoom
	err := context.Bind(&roomInQuestion)
	if err != nil {
		return err
	}

	//This is to say if we want to set audio input even if devices are both A/V outputs.
	//forceAudioChange := true
	//if we're setting both, default to video.
	if len(roomInQuestion.CurrentAudioInput) > 0 && len(roomInQuestion.CurrentVideoInput) > 0 {
		//forceAudioChange = false
	}

	//TODO: Have logic here that checks what was passed in and only changes what is necessary.
	has, err := validateChangeInputSuppliedValue(roomInQuestion.CurrentAudioInput, room, building, "VideoIn")
	if err != nil {
		return err
	} else if has {
		return changeCurrentVideoInput(roomInQuestion, room, building)
	}

	//has, err = validateChangePowerSuppliedValues()

	return nil
}

func changeCurrentVideoInput(room PublicRoom, roomName string, buildingName string) error {

	//magic strings - we'll replace these in the endpoint path.
	portToMatch := ":port"
	commandName := "ChangeInput"

	devices, err := getDevicesByBuildingAndRoomAndRole(roomName, buildingName, "VideoOut")
	if err != nil {
		return err
	}

	inputDevice, err := getDeviceByName(roomName, buildingName, room.CurrentVideoInput)
	if err != nil {
		return err
	}

	for _, val := range devices {
		var curCommand accessors.Command
		for _, val := range val.Commands {
			if strings.EqualFold(val.Name, commandName) {
				curCommand = val
			}
		}

		//placeholder to be replaced by the value we get down below.
		var portValue string
		for _, val := range val.Ports {
			if strings.EqualFold(val.Source, inputDevice.Name) {
				portValue = val.Name
			}
		}
		if len(portValue) <= 0 {
			//TODO: figure out error reporting here.
			continue
		}

		endpointPath := ReplaceIPAddressEndpoint(curCommand.Endpoint.Path, val.Address)

		endpointPath = strings.Replace(endpointPath, portToMatch, portValue, -1)
		//something to get the current port

		_, err = http.Get("http://" + curCommand.Microservice + endpointPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}
