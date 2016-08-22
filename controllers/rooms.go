package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/fusion"
	"github.com/byuoitav/configuration-database-microservice/accessors"
	"github.com/byuoitav/hateoas"
	"github.com/labstack/echo"
)

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

func getRoomByInfo(roomName string, buildingName string) (accessors.Room, error) {
	resp, err := http.Get("http://localhost:8006/buildings/" + buildingName + "/rooms/" + roomName)
	var toReturn accessors.Room
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

func getDevicesByRoom(roomName string, buildingName string) ([]accessors.Device, error) {
	var toReturn []accessors.Device

	resp, err := http.Get("http://localhost:8006/buildings/ITB/1110/devices")

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
	CurrentInput string
	Displays     []Display
	AudioDevices []AudioDevice
}

//AudioDevice represents an audio device
type AudioDevice struct {
	Muted  bool
	Volume int
}

//Display represents a display
type Display struct {
	Name    string
	Power   bool
	Blanked bool
}

func PutRoomChanges(context echo.Context) error {
	building, room := context.Param("building"), context.Param("room")

	var roomInQuestion PublicRoom
	err := context.Bind(&roomInQuestion)
	if err != nil {
		return err
	}
	//TODO: Have logic here that checks what was passed in and only changes what is necessary.
	if !strings.EqualFold(roomInQuestion.CurrentInput, "") {
		return changeCurrentInput(roomInQuestion, room, building)
	}
	return nil
}

func changeCurrentInput(room PublicRoom, roomName string, buildingName string) error {

	//magic strings - we'll replace these in the endpoint path.
	portToMatch := ":port"
	commandName := "ChangeInput"

	devices, err := getDevicesByRoom(roomName, buildingName)

	if err != nil {
		return err
	}

	//Get the output device. Assume FOR NOW it's a TV and will always be there.
	var outputDevice accessors.Device
	var inputDevice accessors.Device
	for _, val := range devices {
		if val.Output && val.Type == 1 {
			outputDevice = val
		}
		if strings.EqualFold(val.Name, room.CurrentInput) && val.Input {
			inputDevice = val
		}
	}
	var curCommand accessors.Command
	for _, val := range outputDevice.Commands {
		if strings.EqualFold(val.Name, commandName) {
			curCommand = val
		}
	}

	portValue := "Hdmi1"

	for _, val := range outputDevice.Ports {
		if strings.EqualFold(val.Source, inputDevice.Name) {
			portValue = val.Name
		}
	}

	endpointPath := ReplaceIPAddressEndpoint(curCommand.Endpoint.Path, outputDevice.Address)

	endpointPath = strings.Replace(endpointPath, portToMatch, portValue, -1)
	//something to get the current port

	_, err = http.Get("http://" + curCommand.Microservice + endpointPath)
	if err != nil {
		return err
	}

	return nil
}

func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}
