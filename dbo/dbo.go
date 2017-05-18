package dbo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

// GetData will run a get on the url, and attempt to fill the interface provided from the returned JSON.
func GetData(url string, structToFill interface{}) error {
	log.Printf("Getting data from URL: %s...", url)
	// Make an HTTP client so we can add custom headers (currently used for adding in the Bearer token for inter-microservice communication)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	err = setToken(req)
	if err != nil {
		return err
	}

	if req == nil {
		fmt.Printf("Alert! req is nil!")
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error on request: %s", err.Error())
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		errorString, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(errorString))
	}

	err = json.Unmarshal(b, structToFill)
	if err != nil {
		return err
	}
	log.Printf("Done.")
	return nil
}

//PostData hits POST endpoints
func PostData(url string, structToAdd interface{}, structToFill interface{}) error {
	log.Printf("Posting data to URL: %s...", url)

	body, err := json.Marshal(structToAdd)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))

	req.Header.Set("Content-Type", "application/json")

	err = setToken(req)
	if err != nil {
		return err
	}

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		errorString, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return errors.New(string(errorString))
	}

	jsonArray, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonArray, structToFill)
	if err != nil {
		return err
	}

	return nil
}

func setToken(request *http.Request) error {
	if len(os.Getenv("LOCAL_ENVIRONMENT")) == 0 {

		log.Printf("Adding the bearer token for inter-service communication")

		token, err := bearertoken.GetToken()
		if err != nil {
			return err
		}

		request.Header.Set("Authorization", "Bearer "+token.Token)

	}

	return nil
}

// GetAllRawCommands retrieves all the commands
func GetAllRawCommands() (commands []accessors.RawCommand, err error) {
	log.Printf("Getting all commands.")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/commands"
	err = GetData(url, &commands)

	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	log.Printf("Done.")
	return
}

func AddRawCommand(toAdd accessors.RawCommand) (accessors.RawCommand, error) {
	log.Printf("adding raw command: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/commands/" + toAdd.Name

	var toFill accessors.RawCommand
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return accessors.RawCommand{}, err
	}

	return toFill, nil
}

func GetRoomByInfo(buildingName string, roomName string) (toReturn accessors.Room, err error) {
	log.Printf("Getting room %s in building %s...", roomName, buildingName)
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+buildingName+"/rooms/"+roomName, &toReturn)
	return
}

// GetDeviceByName simply retrieves a device's information from the databse.
func GetDeviceByName(buildingName string, roomName string, deviceName string) (toReturn accessors.Device, err error) {
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+buildingName+"/rooms/"+roomName+"/devices/"+deviceName, &toReturn)
	return
}

// GetDevicesByRoom will jut get the devices based on the room.
func GetDevicesByRoom(buildingName string, roomName string) (toReturn []accessors.Device, err error) {
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+buildingName+"/rooms/"+roomName+"/devices", &toReturn)
	return
}

// GetDevicesByBuildingAndRoomAndRole will get the devices with the given role from the DB
func GetDevicesByBuildingAndRoomAndRole(building string, room string, roleName string) (toReturn []accessors.Device, err error) {
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+building+"/rooms/"+room+"/devices/roles/"+roleName, &toReturn)
	if err != nil {
		log.Printf("Error getting device by role: %s", err.Error())
	}
	return
}

// SetAudioInDB will set the audio levels in the database
func SetAudioInDB(building string, room string, device accessors.Device) error {
	log.Printf("Updating audio levels in DB.")

	if device.Volume != nil {
		url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms/" + room + "/devices/" + device.Name + "/attributes/volume/" + strconv.Itoa(*device.Volume)
		request, err := http.NewRequest("PUT", url, nil)
		client := &http.Client{}
		_, err = client.Do(request)

		if err != nil {
			return err
		}
	}

	if device.Muted != nil {
		url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms/" + room + "/devices/" + device.Name + "/attributes/muted/" + strconv.FormatBool(*device.Muted)
		request, err := http.NewRequest("PUT", url, nil)
		client := &http.Client{}
		_, err = client.Do(request)

		if err != nil {
			return err
		}
	}

	return nil
}

// GetBuildings will return all buildings
func GetBuildings() ([]accessors.Building, error) {
	log.Printf("getting all buildings...")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings"
	log.Printf(url)
	var buildings []accessors.Building
	err := GetData(url, &buildings)

	return buildings, err
}

// GetRooms returns all the rooms in a given building
func GetRoomsByBuilding(building string) ([]accessors.Room, error) {

	log.Printf("dbo.GetRoomsByBuilding called")

	log.Printf("getting all rooms from %v ...", building)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms"
	var rooms []accessors.Room
	err := GetData(url, &rooms)
	return rooms, err
}

// GetBuildingByShortname returns a building with a given shortname
func GetBuildingByShortname(building string) (accessors.Building, error) {
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/shortname/" + building
	var output accessors.Building
	err := GetData(url, &output)
	if err != nil {
		return output, err
	}
	return output, nil
}

// AddBuilding
func AddBuilding(buildingToAdd accessors.Building) (accessors.Building, error) {
	log.Printf("adding building %v to database", buildingToAdd.Shortname)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + buildingToAdd.Shortname

	var buildingToFill accessors.Building
	err := PostData(url, buildingToAdd, &buildingToFill)
	if err != nil {
		return accessors.Building{}, err
	}

	return buildingToFill, nil
}

func AddRoom(building string, roomToAdd accessors.Room) (accessors.Room, error) {
	log.Printf("adding room %v to building %v in database", roomToAdd.Name, building)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms/" + roomToAdd.Name

	var roomToFill accessors.Room
	err := PostData(url, roomToAdd, &roomToFill)
	if err != nil {
		return accessors.Room{}, err
	}

	return roomToFill, nil
}

func GetDeviceTypes() ([]accessors.DeviceType, error) {
	log.Printf("getting all device types")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/types/"

	var DeviceTypes []accessors.DeviceType
	err := GetData(url, &DeviceTypes)
	if err != nil {
		return []accessors.DeviceType{}, err
	}

	return DeviceTypes, nil
}

func AddDeviceType(toAdd accessors.DeviceType) (accessors.DeviceType, error) {
	log.Printf("adding device type: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/types/" + toAdd.Name

	var toFill accessors.DeviceType
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return accessors.DeviceType{}, err
	}

	return toFill, nil
}
func GetPowerStates() ([]accessors.PowerState, error) {
	log.Printf("getting all power states")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/powerstates/"

	var powerStates []accessors.PowerState
	err := GetData(url, &powerStates)
	if err != nil {
		return []accessors.PowerState{}, err
	}

	return powerStates, nil
}

func AddPowerState(toAdd accessors.PowerState) (accessors.PowerState, error) {
	log.Printf("adding power state: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/powerstates/" + toAdd.Name

	var toFill accessors.PowerState
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return accessors.PowerState{}, err
	}

	return toFill, nil
}

func GetMicroservices() ([]accessors.Microservice, error) {
	log.Printf("getting all microservices")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/microservices"

	var microservices []accessors.Microservice
	err := GetData(url, &microservices)
	if err != nil {
		return []accessors.Microservice{}, err
	}

	return microservices, nil
}

func AddMicroservice(toAdd accessors.Microservice) (accessors.Microservice, error) {
	log.Printf("adding microservice: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/microservices/" + toAdd.Name

	var toFill accessors.Microservice
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return accessors.Microservice{}, err
	}

	return toFill, nil
}

func GetEndpoints() ([]accessors.Endpoint, error) {
	log.Printf("getting all endpoints")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/endpoints"

	var endpoints []accessors.Endpoint
	err := GetData(url, &endpoints)
	if err != nil {
		return []accessors.Endpoint{}, err
	}

	return endpoints, nil
}

func AddEndpoint(toAdd accessors.Endpoint) (accessors.Endpoint, error) {
	log.Printf("adding endpoint: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/endpoints/" + toAdd.Name

	var toFill accessors.Endpoint
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return accessors.Endpoint{}, err
	}

	return toFill, nil
}

func GetPorts() ([]accessors.PortType, error) {
	log.Printf("getting all ports")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/ports"

	var ports []accessors.PortType
	err := GetData(url, &ports)
	if err != nil {
		return []accessors.PortType{}, err
	}

	return ports, nil
}

func AddPort(portToAdd accessors.PortType) (accessors.PortType, error) {
	log.Printf("adding Port: %v to database", portToAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/ports/" + portToAdd.Name

	var portToFill accessors.PortType
	err := PostData(url, portToAdd, &portToFill)
	if err != nil {
		return accessors.PortType{}, err
	}

	return portToFill, nil
}

func GetDeviceRoleDefinitions() ([]accessors.DeviceRoleDef, error) {
	log.Printf("getting device role definitions")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/roledefinitions"

	var definitions []accessors.DeviceRoleDef
	err := GetData(url, &definitions)
	if err != nil {
		return []accessors.DeviceRoleDef{}, err
	}

	return definitions, nil
}

func AddRoleDefinition(toAdd accessors.DeviceRoleDef) (accessors.DeviceRoleDef, error) {
	log.Printf("adding role definition: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/roledefinitions/" + toAdd.Name

	var toFill accessors.DeviceRoleDef
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return accessors.DeviceRoleDef{}, err
	}

	return toFill, nil
}

func GetRoomConfigurations() ([]accessors.RoomConfiguration, error) {
	log.Printf("getting room configurations")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/configurations"

	var rcs []accessors.RoomConfiguration
	err := GetData(url, &rcs)
	if err != nil {
		return []accessors.RoomConfiguration{}, err
	}

	return rcs, nil

}

func AddDevice(toAdd accessors.Device) (accessors.Device, error) {
	log.Printf("adding device: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + toAdd.Building.Shortname + "/rooms/" + toAdd.Room.Name + "/devices/" + toAdd.Name

	var toFill accessors.Device
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return accessors.Device{}, err
	}

	return toFill, nil
}
