package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/helpers"
	"github.com/labstack/echo"
)

type fusionResponse struct {
	APIRooms []fusionRoom `json:"API_Rooms"`
}

type fusionRoom struct {
	RoomID   string
	RoomName string
	Symbols  []fusionSymbol
}

type fusionSymbol struct {
	ProcessorName string
	ConnectInfo   string
	SymbolID      string
	Available     bool
}

type room struct {
	Building string
	Room     string
	Hostname string
	Address  string
}

type roomWithAvailability struct {
	Building  string
	Room      string
	Hostname  string
	Address   string
	Available bool
}

// GetRooms returns a list of all rooms Crestron Fusion knows about
func GetRooms(c echo.Context) error {
	response, err := helpers.GetHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/") // MAKE SURE YOU HAVE THE TRAILING SLASH
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	return c.String(http.StatusOK, string(response))
}

func GetRoomByName(c echo.Context) error {
	// Get the room's ID from its name
	response, err := helpers.GetHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?search="+c.Param("room"))
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	rooms := fusionResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	// Get info about the room using its ID
	response, err = helpers.GetHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/"+rooms.APIRooms[0].RoomID)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	rooms = fusionResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	hostname := rooms.APIRooms[0].Symbols[0].ProcessorName
	address := rooms.APIRooms[0].Symbols[0].ConnectInfo
	building := strings.Split(c.Param("room"), "+")
	roomName := strings.Split(c.Param("room"), "+")

	roomResponse := room{Building: building[0], Room: roomName[1], Hostname: hostname, Address: address}

	jsonResponse, _ := json.Marshal(roomResponse)
	return c.String(http.StatusOK, string(jsonResponse))
}

// GetRoomByNameAndBuilding is almost identical to GetRoomByName with the addition of room availability checking (possible because of the supplying of a building in the API call)
func GetRoomByNameAndBuilding(c echo.Context) error {
	// Get the room's ID from its name
	response, err := helpers.GetHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?search="+c.Param("room"))
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	rooms := fusionResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	if len(rooms.APIRooms) == 0 { // Return an error if Fusion doesn't have record of the room specified
		return c.String(http.StatusNotFound, "Could not find the room specified")
	}

	// Get info about the room using its ID
	response, err = helpers.GetHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/"+rooms.APIRooms[0].RoomID)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	rooms = fusionResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	fmt.Printf("%+v\n", rooms)

	hostname := rooms.APIRooms[0].Symbols[0].ProcessorName
	address := rooms.APIRooms[0].Symbols[0].ConnectInfo
	availability, err := helpers.CheckAvailability(c.Param("building"), c.Param("room"))
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator. Error: "+err.Error())
	}

	roomResponse := roomWithAvailability{Building: c.Param("building"), Room: c.Param("room"), Hostname: hostname, Address: address, Available: availability}

	jsonResponse, _ := json.Marshal(roomResponse)
	return c.String(http.StatusOK, string(jsonResponse))
}
