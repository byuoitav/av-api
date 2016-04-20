package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/byuoitav/av-api/helpers"
	"github.com/byuoitav/av-api/packages/fusion"
	"github.com/labstack/echo"
)

type room struct {
	Building string
	Room     string
	Hostname string
	Address  string
	Health   helpers.Health
}

type roomWithAvailability struct {
	Building  string
	Room      string
	Hostname  string
	Address   string
	Health    helpers.Health
	Available bool
}

// GetRooms returns a list of all rooms Crestron Fusion knows about
func GetRooms(c echo.Context) error {
	response, err := fusion.GetRooms()
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

// GetRoomByName get a room from Fusion using only its name
func GetRoomByName(c echo.Context) error {
	response, err := fusion.GetRoomByName()
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	return c.JSON(http.StatusOK, response)
}

// GetRoomByNameAndBuilding is almost identical to GetRoomByName with the addition of room availability checking (possible because of the supplying of a building in the API call)
func GetRoomByNameAndBuilding(c echo.Context) error {
	// Get the room's ID from its name
	response, err := helpers.RequestHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?search="+c.Param("building")+"+"+c.Param("room"))
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	rooms := fusion.Response{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	if len(rooms.Rooms) == 0 { // Return an error if Fusion doesn't have record of the room specified
		return c.String(http.StatusNotFound, "An error was encountered. Please contact your system administrator.\nError: Could not find room "+c.Param("room")+" in the "+c.Param("building")+" building in the Fusion database")
	} else if len(rooms.Rooms) > 1 {
		return c.String(http.StatusNotFound, "Error: Your search \""+c.Param("building")+" "+c.Param("room")+"\" returned multiple results from the Fusion database")
	}

	// Get info about the room using its ID
	response, err = helpers.RequestHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/"+rooms.Rooms[0].RoomID)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	rooms = fusion.Response{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	hostname := rooms.Rooms[0].Symbols[0].ProcessorName
	address := rooms.Rooms[0].Symbols[0].ConnectInfo
	health, err := helpers.GetHealth(address)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	availability, err := helpers.CheckAvailability(c.Param("building"), c.Param("room"), rooms.Rooms[0].Symbols[0].SymbolID)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	roomResponse := roomWithAvailability{Building: c.Param("building"), Room: c.Param("room"), Hostname: hostname, Address: address, Health: health, Available: availability}

	return c.JSON(http.StatusOK, roomResponse)
}
