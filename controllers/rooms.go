package controllers

import (
	"encoding/json"
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
}

type room struct {
	Building  string
	Room      string
	Hostname  string
	Address   string
	Available bool
}

func GetRooms(c echo.Context) error {
	response := helpers.GetHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/")
	return c.String(http.StatusOK, string(response)) // MAKE SURE YOU HAVE THE TRAILING SLASH
}

func GetRoomByName(c echo.Context) error {
	// Get the room's ID from its name
	response := helpers.GetHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?search="+c.Param("room"))
	rooms := fusionResponse{}
	err := json.Unmarshal(response, &rooms)
	helpers.CheckErr(err)

	// Get info about the room using its ID
	response = helpers.GetHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/"+rooms.APIRooms[0].RoomID)
	rooms = fusionResponse{}
	err = json.Unmarshal(response, &rooms)
	helpers.CheckErr(err)

	hostname := rooms.APIRooms[0].Symbols[0].ProcessorName
	address := rooms.APIRooms[0].Symbols[0].ConnectInfo
	building := strings.Split(c.Param("room"), "+")
	roomName := strings.Split(c.Param("room"), "+")

	roomResponse := room{Building: building[0], Room: roomName[1], Hostname: hostname, Address: address, Available: helpers.CheckAvailability()}

	jsonResponse, _ := json.Marshal(roomResponse)
	return c.String(http.StatusOK, string(jsonResponse))
}
