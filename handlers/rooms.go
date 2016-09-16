package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + buildingName + "/rooms/" + roomName
	var toReturn accessors.Room
	err := getData(url, &toReturn)
	return toReturn, err
}

func getDeviceByName(roomName string, buildingName string, deviceName string) (accessors.Device, error) {
	var toReturn accessors.Device
	err := getData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+buildingName+"/rooms/"+roomName+"/devices/"+deviceName, &toReturn)
	return toReturn, err
}

func getDevicesByRoom(roomName string, buildingName string) ([]accessors.Device, error) {
	var toReturn []accessors.Device

	resp, err := http.Get(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + buildingName + "/rooms/" + roomName + "/devices")

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
	Muted  *bool  `json:"muted"`
	Volume *int   `json:"volume"`
}

//Display represents a display
type Display struct {
	Name    string `json:"name"`
	Power   string `json:"power"`
	Blanked bool   `json:"blanked"`
}

func getDevicesByBuildingAndRoomAndRole(room string, building string, roleName string) ([]accessors.Device, error) {

	resp, err := http.Get(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms/" + room + "/devices/roles/" + roleName)
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

func setAudioInDB(building string, room string, device accessors.Device) error {
	log.Printf("Updating audio levels in DB.")

	if device.Volume != nil {
		url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms/" + room + "/devices/" + device.Name + "/attributes/volume/" + strconv.Itoa(*device.Volume)
		fmt.Printf(url + "\n")
		request, err := http.NewRequest("PUT", url, nil)
		client := &http.Client{}
		_, err = client.Do(request)

		if err != nil {
			return err
		}
	}

	if device.Muted != nil {
		url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms/" + room + "/devices/" + device.Name + "/attributes/muted/" + strconv.FormatBool(*device.Muted)
		fmt.Printf(url + "\n")
		request, err := http.NewRequest("PUT", url, nil)
		client := &http.Client{}
		_, err = client.Do(request)

		if err != nil {
			return err
		}
	}

	return nil
}

func validateSuppliedAudioStateChange(roomInfo PublicRoom, room string, building string) ([]accessors.Device, bool, error) {
	toReturn := []accessors.Device{}

	//validate that the list of devices are valid audio devices

	for _, device := range roomInfo.AudioDevices {
		fullDevice, valid, err := validateRoomDeviceByRole(device.Name, room, building, "AudioOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}
		if !valid {
			log.Printf("Invalid device %s specified.", device.Name)
			return []accessors.Device{}, false, errors.New("Invalid audio device " + device.Name + " specified.")
		}

		toReturn = append(toReturn, fullDevice)
	}

	return toReturn, true, nil
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
	needChange := false

	for _, device := range roomInfo.Displays {
		//validate that the device exists in the room
		if device.Power == "" {
			log.Printf("No power state specified for device %s.", device.Name)
			continue
		}
		needChange = true

		fullDevice, valid, err := validateRoomDeviceByRole(device.Name, room, building, "VideoOut")
		if err != nil {
			return []accessors.Device{}, false, err
		}
		if !valid {
			log.Printf("Invalid device %s specified.", device.Name)
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
			log.Printf("Invalid power state %s specified.", device.Power)
			return []accessors.Device{}, false, errors.New("Invalid power state " + device.Power + " specified.")
		}
		toReturn = append(toReturn, fullDevice)
	}

	for _, device := range roomInfo.AudioDevices {
		if device.Power == "" {
			log.Printf("No power state specified for device %s.", device.Name)
			continue
		}
		needChange = true

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

	return toReturn, needChange, nil
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
			log.Printf("Room has no %s devices.\n", roleName)
			return accessors.Device{}, false, errors.New("No " + roleName + " devices in room.")
		}
		log.Printf("%v devices found.\n", len(devices))
		log.Printf("Checking for %s.\n", deviceToCheck)
		for _, val := range devices {
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
		"audioDevices": [{
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
	}
	log.Printf("Done.\n")

	log.Printf("Checking for power changes.\n")
	_, valid, err = validateSuppliedValuesPowerChange(roomInQuestion, room, building)
	if err != nil {
		return err
	} else if valid {
		log.Printf("Changing power states.\n")
		err = changeCurrentPowerStateForMultipleDevices(roomInQuestion, room, building)
		if err != nil {
			log.Printf("Error: %s.\n", err.Error())
			return err
		}
	}
	log.Printf("Done.\n")

	log.Printf("Checking Audio-specific states.\n")
	devices, valid, err := validateSuppliedAudioStateChange(roomInQuestion, room, building)
	if err != nil {
		return err
	} else if valid {
		err = changeAudioStateForMultipleDevices(roomInQuestion, room, building, devices)
	}
	log.Printf("Done.\n")

	return nil
}

func changeAudioStateForMultipleDevices(roomInfo PublicRoom, room string, building string, devices []accessors.Device) error {
	// Get command for set volume devices[0].Command
	for _, desired := range roomInfo.AudioDevices {
		log.Printf("Setting audio states for %s\n", desired.Name)

		var current accessors.Device
		//get accessorDevice for room Device
		for _, val2 := range devices {
			if strings.EqualFold(val2.Name, desired.Name) {
				current = val2
				break
			}
		}

		if desired.Volume != nil {
			if desired.Muted != nil {
				return errors.New("Cannot set Muted and Volume in the same call.")
			}
			b := false
			desired.Muted = &b
		}

		if desired.Muted != nil {
			fmt.Printf("Desired: %v, Current: %v\n", *desired.Muted, *current.Muted)
			if *desired.Muted && !*current.Muted {
				//get the muted command
				log.Printf("Setting muted.")
				for _, comm := range current.Commands {
					if strings.EqualFold(comm.Name, "Mute") {
						_, err := http.Get("http://" + comm.Microservice + ReplaceIPAddressEndpoint(comm.Endpoint.Path, current.Address))
						if err != nil {
							log.Printf("Error Muting device %s: %s\n", desired.Name, err.Error())
						} else {
							//set the new volume in the DB.
							*current.Muted = *desired.Muted
							err = setAudioInDB(building, room, current)
							if err != nil {
								return err
							}
						}
					}
				}
			} else if !*desired.Muted && *current.Muted {
				for _, comm := range current.Commands {
					if strings.EqualFold(comm.Name, "UnMute") {
						_, err := http.Get("http://" + comm.Microservice + ReplaceIPAddressEndpoint(comm.Endpoint.Path, current.Address))
						if err != nil {
							log.Printf("Error UnMuting device %s: %s\n", desired.Name, err.Error())
						} else {
							//set the new volume in the DB.
							*current.Muted = *desired.Muted
							err = setAudioInDB(building, room, current)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}

		//after setting the muted.
		if desired.Volume != nil {
			log.Printf("Setting volume.")
			for _, comm := range current.Commands {
				if strings.EqualFold(comm.Name, "SetVolume") {
					endpoint := comm.Endpoint.Path
					endpoint = ReplaceIPAddressEndpoint(endpoint, current.Address)
					//this is our difference
					if strings.Contains(endpoint, "difference") {
						fmt.Printf("Current: %v, Desired: %v\n", *current.Volume, *desired.Volume)
						difference := *desired.Volume - *current.Volume
						fmt.Printf("Difference: %v\n", difference)
						endpoint = strings.Replace(endpoint, ":difference", strconv.Itoa(difference), -1)
						fmt.Printf("New Endpoint: %s\n", endpoint)
					} else {
						endpoint = strings.Replace(endpoint, ":level", strconv.Itoa(*desired.Volume), -1)
					}
					_, err := http.Get("http://" + comm.Microservice + endpoint)
					if err != nil {
						log.Printf("Error Setting volume for device %s: %s. May need to calibrate device.\n", desired.Name, err.Error())
					} else {
						//set the new volume in the DB.
						*current.Muted = false
						*current.Volume = *desired.Volume
						err = setAudioInDB(building, room, current)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
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
		fmt.Printf("%s\n", "http://"+curCommand.Microservice+endpointPath)
		_, err = http.Get("http://" + curCommand.Microservice + endpointPath)
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			continue
		}
		log.Printf("Command send to device %s.\n", val.Name)
	}

	return nil
}

/*
ReplaceIPAddressEndpoint is a simple helper
*/
func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}
