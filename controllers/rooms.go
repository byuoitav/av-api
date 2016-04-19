package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/byuoitav/av-api/helpers"
	"github.com/labstack/echo"
)

type fusionResponse struct {
	APIRooms   []fusionRoom `json:"API_Rooms"`
	Pagination string       `json:"Message"`
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
}

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
	// response, err := helpers.RequestHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/") // MAKE SURE YOU HAVE THE TRAILING SLASH
	// if err != nil {
	// 	return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	// }
	//
	// rooms := fusionResponse{}
	// err = json.Unmarshal(response, &rooms)

	currentPage := 1
	lastPage := 1

	var toReturn []FusionRoomInfo

	for currentPage <= lastPage {
		reqAddress := address + "?page=" + strconv.Itoa(currentPage)
		fmt.Printf("\nRequestAddress %s \n", reqAddress)
		req, err := http.NewRequest("GET", reqAddress, nil)
		req.Header.Add("Content-Type", "application/json")
		check(err)

		resp, err := client.Do(req)
		check(err)

		var response = FusionRoomResponse{}
		bits, err := ioutil.ReadAll(resp.Body)
		check(err)

		fmt.Printf("\nResponse: %s\n", bits)

		err = json.Unmarshal(bits, &response)
		check(err)

		myExp := regexp.MustCompile(`Page ([0-9]+) of ([0-9]+)`)

		match := myExp.FindStringSubmatch(response.Message)

		toReturn = append(toReturn, response.APIRooms...)

		currentPage, err = strconv.Atoi(match[1])
		check(err)
		lastPage, err = strconv.Atoi(match[2])
		check(err)
		fmt.Printf("\nDownloaded page %v of %v\n", currentPage, lastPage)

		currentPage++
	}

	return c.JSON(http.StatusOK, rooms)
}

// GetRoomByName get a room from Fusion using only its name
func GetRoomByName(c echo.Context) error {
	// Get the room's ID from its name
	response, err := helpers.RequestHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?search="+c.Param("building")+"+"+c.Param("room"))
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	rooms := fusionResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	// Get info about the room using its ID
	response, err = helpers.RequestHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/"+rooms.APIRooms[0].RoomID)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	rooms = fusionResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	hostname := rooms.APIRooms[0].Symbols[0].ProcessorName
	address := rooms.APIRooms[0].Symbols[0].ConnectInfo
	building := strings.Split(c.Param("room"), "+")
	roomName := strings.Split(c.Param("room"), "+")

	roomResponse := room{Building: building[0], Room: roomName[1], Hostname: hostname, Address: address}

	return c.JSON(http.StatusOK, roomResponse)
}

// GetRoomByNameAndBuilding is almost identical to GetRoomByName with the addition of room availability checking (possible because of the supplying of a building in the API call)
func GetRoomByNameAndBuilding(c echo.Context) error {
	// Get the room's ID from its name
	response, err := helpers.RequestHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/?search="+c.Param("building")+"+"+c.Param("room"))
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	rooms := fusionResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	if len(rooms.APIRooms) == 0 { // Return an error if Fusion doesn't have record of the room specified
		return c.String(http.StatusNotFound, "An error was encountered. Please contact your system administrator.\nError: Could not find room "+c.Param("room")+" in the "+c.Param("building")+" building in the Fusion database")
	} else if len(rooms.APIRooms) > 1 {
		return c.String(http.StatusNotFound, "Error: Your search \""+c.Param("building")+" "+c.Param("room")+"\" returned multiple results from the Fusion database")
	}

	// Get info about the room using its ID
	response, err = helpers.RequestHTTP("GET", "http://lazyeye.byu.edu/fusion/apiservice/rooms/"+rooms.APIRooms[0].RoomID)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	rooms = fusionResponse{}
	err = json.Unmarshal(response, &rooms)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	hostname := rooms.APIRooms[0].Symbols[0].ProcessorName
	address := rooms.APIRooms[0].Symbols[0].ConnectInfo
	health, err := helpers.GetHealth(address)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	availability, err := helpers.CheckAvailability(c.Param("building"), c.Param("room"), rooms.APIRooms[0].Symbols[0].SymbolID)
	if err != nil {
		return c.String(http.StatusBadRequest, "An error was encountered. Please contact your system administrator.\nError: "+err.Error())
	}

	roomResponse := roomWithAvailability{Building: c.Param("building"), Room: c.Param("room"), Hostname: hostname, Address: address, Health: health, Available: availability}

	return c.JSON(http.StatusOK, roomResponse)
}
