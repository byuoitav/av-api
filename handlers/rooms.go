package handlers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
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
	err := getData(databaseLocation+"/buildings/"+buildingName+"/rooms/"+roomName+"/devices/"+deviceName, &toReturn)
	return toReturn, err
}

func getDevicesByRoom(roomName string, buildingName string) ([]accessors.Device, error) {
	var toReturn []accessors.Device

	resp, err := http.Get(databaseLocation + "/buildings/" + buildingName + "/rooms/" + roomName + "/devices")

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
	CurrentVideoInput string        `json:"currentVideoInput"`
	CurrentAudioInput string        `json:"currentAudioInput"`
	Displays          []Display     `json:"displays"`
	AudioDevices      []AudioDevice `json:"audioDevices"`
}

//AudioDevice represents an audio device
type AudioDevice struct {
	Name   string `json:"name"`
	Power  string `json:"power"`
	Muted  bool   `json:"muted"`
	Volume int    `json:"volume"`
}

//Display represents a display
type Display struct {
	Name    string `json:"name"`
	Power   string `json:"power"`
	Blanked bool   `json:"blanked"`
}

func getDevicesByBuildingAndRoomAndRole(room string, building string, roleName string) ([]accessors.Device, error) {

	resp, err := http.Get(databaseLocation + "/buildings/" + building + "/rooms/" + room + "/devices/roles/" + roleName)
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

/*validateChagnePowerSuppliedValues will go through each of the output devices
(audio and video) and validate that they are
a) valid devices for the room and
b) valid power states for the device
*/
func validateSuppliedValuesPowerChange(roomInfo PublicRoom, room string, building string) ([]accessors.Device, bool, error) {
	toReturn := []accessors.Device{}

	if len(roomInfo.AudioDevices) <= 0 && len(roomInfo.Displays) <= 0 {
		return toReturn, false, nil
	}

	for _, device := range roomInfo.Displays {
		//validate that the device exists in the room
		fullDevice, valid, err := validateRoomDeviceByRole(device.Name, room, building, "VideoOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}
		if !valid {
			return []accessors.Device{}, false, errors.New("Invalid display " + device.Name + " specified.")
		}
		valid = false
		//validate that it is a valid powerstate.
		for _, val := range fullDevice.PowerStates {
			if strings.EqualFold(val, device.Power) {
				valid = true
				break
			}
		}
		if !valid {
			return []accessors.Device{}, false, errors.New("Invalid power state " + device.Power + " specified.")
		}
		toReturn = append(toReturn, fullDevice)
	}

	for _, device := range roomInfo.AudioDevices {
		fullDevice, valid, err := validateRoomDeviceByRole(device.Name, room, building, "AudioOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}
		if !valid {
			return []accessors.Device{}, false, errors.New("Invalid audio device " + device.Name + " specified.")
		}
		//validate that it is a valid powerstate.
		for _, val := range fullDevice.PowerStates {
			if strings.EqualFold(val, device.Power) {
				valid = true
				break
			}
		}
		if !valid {
			return []accessors.Device{}, false, errors.New("Invalid power state " + device.Power + " specified.")
		}
		toReturn = append(toReturn, fullDevice)
	}

	return toReturn, true, nil
}

/*
	validateRoomDeviceByRole validates that a room has a named device with the given role.
*/
func validateRoomDeviceByRole(deviceToCheck string, room string, building string, roleName string) (accessors.Device, bool, error) {
	if len(deviceToCheck) > 0 {
		log.Printf("Validating device %s in room %s with role %s...\n", deviceToCheck, building+" "+room, roleName)
		log.Printf("Getting all devices for role %s in room...\n", roleName)
		devices, err := getDevicesByBuildingAndRoomAndRole(room, building, roleName)
		if err != nil {
			log.Printf("Error %s\n", err.Error())
			return accessors.Device{}, false, err
		}

		if len(devices) < 1 {
			log.Printf("Room has no input devices.\n")
			return accessors.Device{}, false, errors.New("No " + roleName + " input devices in room.")
		}
		log.Printf("%v devices found.\n", len(devices))
		log.Printf("Checking for %s.\n", deviceToCheck)
		for _, val := range devices {
			log.Printf("%+v\n", val)

			if strings.EqualFold(deviceToCheck, val.Name) || strings.EqualFold(deviceToCheck, val.Type) {
				log.Printf("Device validated.\n")
				return val, true, nil
			}
		}
		log.Printf("Device not found. Invalid device.\n")
		return accessors.Device{}, false, errors.New("Invalid " + roleName + " devices sepecified.")
	}
	return accessors.Device{}, false, nil //there were no devices to check.
}

/*
PutRoomChanges is the handler to accept puts to /buildlings/:buildling/rooms/:room with the json payload with one or more of the fields:
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
	log.Printf("Putting room changes.\n")

	var roomInQuestion PublicRoom
	err := context.Bind(&roomInQuestion)
	if err != nil {
		return err
	}

	log.Printf("Room: %+v\n", roomInQuestion)

	//This is to say if we want to set audio input even if devices are both A/V outputs.
	//forceAudioChange := true
	//if we're setting both, default to video.
	if len(roomInQuestion.CurrentAudioInput) > 0 && len(roomInQuestion.CurrentVideoInput) > 0 {
		//forceAudioChange = false
	}

	log.Printf("Checking for input changes.\n")
	//TODO: Have logic here that checks what was passed in and only changes what is necessary.
	_, valid, err := validateRoomDeviceByRole(roomInQuestion.CurrentVideoInput, room, building, "VideoIn")
	if err != nil {
		return err
	} else if valid {
		log.Printf("Changing current input\n")
		err = changeCurrentVideoInput(roomInQuestion, room, building)
		if err != nil {
			log.Printf("Error: %s.\n", err.Error())
			return err
		}
		log.Printf("Done.\n")
	}

	log.Printf("Checking for power changes.\n")
	_, valid, err = validateSuppliedValuesPowerChange(roomInQuestion, room, building)
	if err != nil {
		return nil
	} else if valid {
		log.Printf("Changing power states.\n")
		err = changeCurrentPowerStateForMultipleDevices(roomInQuestion, room, building)
		if err != nil {
			log.Printf("Error: %s.\n", err.Error())
			return err
		}
	}
	log.Printf("Done.\n")

	return nil
}

func changeCurrentPowerStateForMultipleDevices(roomInfo PublicRoom, room string, building string) error {
	commandNames := make(map[string]string)
	commandNames["on"] = "PowerOn"
	commandNames["standby"] = "Standby"

	//Do the Displays
	for _, val := range roomInfo.Displays {
		log.Printf("Changing power states for display %s to %s.\n", val.Name, val.Power)
		device, err := getDeviceByName(room, building, val.Name)
		if err != nil {
			//TODO: Figure out reporting here.
			continue
		}

		curCommandName := commandNames[strings.ToLower(val.Power)]
		log.Printf("Checking commands for command %s\n", curCommandName)
		for _, command := range device.Commands {
			if strings.EqualFold(command.Name, curCommandName) {
				log.Printf("Command found.\n")
				endpointPath := ReplaceIPAddressEndpoint(command.Endpoint.Path, device.Address)
				log.Printf("sending Command\n")
				_, err = http.Get("http://" + command.Microservice + endpointPath)
				if err != nil {
					log.Printf("Error %s\n", err.Error())
					break
				}
				log.Printf("Command Sent.\n")
				break
			}
			log.Printf("Command not found.\n")
		}
	}

	//Do the Audio Devices
	for _, val := range roomInfo.AudioDevices {
		log.Printf("Changing power states for display %s to %s.\n", val.Name, val.Power)
		device, err := getDeviceByName(room, building, val.Name)
		if err != nil {
			//TODO: Figure out reporting here.
			continue
		}

		curCommandName := commandNames[val.Power]
		log.Printf("Checking commands for command %s\n", curCommandName)
		for _, command := range device.Commands {
			if strings.EqualFold(command.Name, curCommandName) {
				log.Printf("Command found.\n")
				endpointPath := ReplaceIPAddressEndpoint(command.Endpoint.Path, device.Address)
				log.Printf("sending Command\n")
				_, err = http.Get("http://" + command.Microservice + endpointPath)
				if err != nil {
					log.Printf("Error %s\n", err.Error())
					continue
				}
				log.Printf("Command Sent.\n")
			}
			log.Printf("Command not found.\n")
		}
	}

	return nil
}

func changeCurrentVideoInput(room PublicRoom, roomName string, buildingName string) error {
	//magic strings - we'll replace these in the endpoint path.
	portToMatch := ":port"
	commandName := "ChangeInput"
	log.Printf("Changing video input.\n")
	log.Printf("Getting all video out devices for room.\n")
	devices, err := getDevicesByBuildingAndRoomAndRole(roomName, buildingName, "VideoOut")
	if err != nil {
		return err
	}
	log.Printf("%v devices found\n", len(devices))

	log.Printf("Getting input device %s\n", room.CurrentVideoInput)
	inputDevice, err := getDeviceByName(roomName, buildingName, room.CurrentVideoInput)
	if err != nil {
		return err
	}

	for _, val := range devices {
		log.Printf("Checking for command %s\n", commandName)
		var curCommand accessors.Command

		for _, val := range val.Commands {
			if strings.EqualFold(val.Name, commandName) {
				log.Printf("Found.")
				curCommand = val
			}
		}
		if len(curCommand.Name) <= 0 {
			log.Printf("Command not found, continuing\n")
			continue
		}

		log.Printf("Checking output device ports for input device %s.", inputDevice.Name)
		//placeholder to be replaced by the value we get down below.
		var portValue string
		for _, val := range val.Ports {
			if strings.EqualFold(val.Source, inputDevice.Name) {
				log.Printf("Found, input device into port %s\n", val.Name)
				portValue = val.Name
			}
		}

		if len(portValue) <= 0 {
			//TODO: figure out error reporting here.
			log.Printf("Port not found, continuing\n")
			continue
		}

		endpointPath := ReplaceIPAddressEndpoint(curCommand.Endpoint.Path, val.Address)

		endpointPath = strings.Replace(endpointPath, portToMatch, portValue, -1)
		//something to get the current port
		log.Printf("Sending Command.\n")
		_, err = http.Get("http://" + curCommand.Microservice + endpointPath)
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			continue
		}
		log.Printf("Command send to device %s.\n", val.Name)
	}

	return nil
}

func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}
