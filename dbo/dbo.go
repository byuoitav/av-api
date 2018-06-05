package dbo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

// GetData will run a get on the url, and attempt to fill the interface provided from the returned JSON.
func GetData(url string, structToFill interface{}) error {
	log.L.Infof("[dbo] getting data from URL: %s...", url)
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
		color.Set(color.FgHiRed, color.Bold)
		log.L.Infof("Error on request: %s", err.Error())
		color.Unset()
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		errorBytes, err := ioutil.ReadAll(resp.Body)
		errorString := fmt.Sprintf("Error Code %v. Error String: %s", resp.StatusCode, errorBytes)
		if err != nil {
			return err
		}
		return errors.New(string(errorString))
	}

	err = json.Unmarshal(b, structToFill)
	if err != nil {
		return err
	}
	log.L.Infof("[dbo] done getting data from url: %s", url)
	return nil
}

func SendData(url string, structToAdd interface{}, structToFill interface{}, method string) error {
	body, err := json.Marshal(structToAdd)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))

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

//PostData hits POST endpoints
func PostData(url string, structToAdd interface{}, structToFill interface{}) error {
	log.L.Infof("[dbo Posting data to URL: %s...", url)
	return SendData(url, structToAdd, structToFill, "POST")

}

//PutData hits PUT endpoints
func PutData(url string, structToAdd interface{}, structToFill interface{}) error {
	log.L.Infof("[dbo] Putting data to URL: %v...", url)
	return SendData(url, structToAdd, structToFill, "PUT")
}

func setToken(request *http.Request) error {
	if len(os.Getenv("LOCAL_ENVIRONMENT")) == 0 {

		log.L.Info("[dbo] adding the bearer token for inter-service communication")

		token, err := bearertoken.GetToken()
		if err != nil {
			return err
		}

		request.Header.Set("Authorization", "Bearer "+token.Token)

	}

	return nil
}

//GetAllRawCommands retrieves all the commands
func GetAllRawCommands() (commands []structs.RawCommand, err error) {
	log.L.Info("[dbo] getting all commands.")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/commands"
	err = GetData(url, &commands)

	if err != nil {
		color.Set(color.FgHiRed, color.Bold)
		log.L.Infof("[error]: %s", err.Error())
		color.Unset()
		return
	}

	log.L.Info("[dbo] Done.")
	return
}

func AddRawCommand(toAdd structs.RawCommand) (structs.RawCommand, error) {
	log.L.Infof("[dbo] adding raw command: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/commands/" + toAdd.Name

	var toFill structs.RawCommand
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return structs.RawCommand{}, err
	}

	return toFill, nil
}

func GetRoomByInfo(buildingName string, roomName string) (toReturn structs.Room, err error) {

	log.L.Infof("[dbo] getting room %s in building %s...", roomName, buildingName)
	url := fmt.Sprintf("%s/buildings/%s/rooms/%s", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"), buildingName, roomName)
	err = GetData(url, &toReturn)
	return
}

func GetRoomById(roomId int) (*structs.Room, error) {
	url := fmt.Sprintf("%s/rooms/id/%d", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"), roomId)
	var room structs.Room
	err := GetData(url, &room)
	if err != nil {
		return &structs.Room{}, err
	}

	return &room, nil
}

// GetDeviceByName simply retrieves a device's information from the databse.
func GetDeviceByName(buildingName string, roomName string, deviceName string) (toReturn structs.Device, err error) {
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+buildingName+"/rooms/"+roomName+"/devices/"+deviceName, &toReturn)
	return
}

func GetDeviceById(id int) (toReturn structs.Device, err error) {

	url := fmt.Sprintf("%s/devices/%d", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"), id)

	err = GetData(url, &toReturn)

	return
}

// GetDevicesByRoom will jut get the devices based on the room.
func GetDevicesByRoom(buildingName string, roomName string) (toReturn []structs.Device, err error) {
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+buildingName+"/rooms/"+roomName+"/devices", &toReturn)
	return
}

func GetDevicesByRoomId(roomId int) ([]structs.Device, error) {

	url := fmt.Sprintf("%s/rooms/%d/devices", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"), roomId)

	var devices []structs.Device
	err := GetData(url, &devices)
	if err != nil {
		return []structs.Device{}, err
	}

	return devices, nil
}

// GetDevicesByBuildingAndRoomAndRole will get the devices with the given role from the DB
func GetDevicesByBuildingAndRoomAndRole(building string, room string, roleName string) (toReturn []structs.Device, err error) {

	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+building+"/rooms/"+room+"/devices/roles/"+roleName, &toReturn)
	if err != nil {
		log.L.Infof("%s", color.HiRedString("[error] problem getting device by role: %s", err.Error()))
	}
	return
}

func GetDevicesByRoomIdAndRoleId(roomId, roleId int) ([]structs.Device, error) {

	url := fmt.Sprintf("%s/rooms/%d/roles/%d", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"), roomId, roleId)

	var devices []structs.Device
	err := GetData(url, &devices)
	if err != nil {
		return []structs.Device{}, err
	}

	return devices, nil
}

// GetBuildings will return all buildings
func GetBuildings() ([]structs.Building, error) {
	log.L.Info("[dbo] getting all buildings...")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings"
	log.L.Infof("[dbo] url: %s", url)
	var buildings []structs.Building
	err := GetData(url, &buildings)

	return buildings, err
}

func GetRooms() ([]structs.Room, error) {
	url := fmt.Sprintf("%s/rooms", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"))
	var rooms []structs.Room
	err := GetData(url, &rooms)

	return rooms, err
}

// GetRooms returns all the rooms in a given building
func GetRoomsByBuilding(building string) ([]structs.Room, error) {

	log.L.Infof("[dbo] getting all rooms from %v ...", building)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms"
	var rooms []structs.Room
	err := GetData(url, &rooms)
	return rooms, err
}

// GetBuildingByShortname returns a building with a given shortname
func GetBuildingByShortname(building string) (structs.Building, error) {
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/shortname/" + building
	var output structs.Building
	err := GetData(url, &output)
	if err != nil {
		return output, err
	}
	return output, nil
}

// AddBuilding
func AddBuilding(buildingToAdd structs.Building) (structs.Building, error) {
	log.L.Infof("[dbo] adding building %v to database", buildingToAdd.Shortname)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + buildingToAdd.Shortname

	var buildingToFill structs.Building
	err := PostData(url, buildingToAdd, &buildingToFill)
	if err != nil {
		return structs.Building{}, err
	}

	return buildingToFill, nil
}

func AddRoom(building string, roomToAdd structs.Room) (structs.Room, error) {
	log.L.Infof("[dbo] adding room %v to building %v in database", roomToAdd.Name, building)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms/" + roomToAdd.Name

	var roomToFill structs.Room
	err := PostData(url, roomToAdd, &roomToFill)
	if err != nil {
		return structs.Room{}, err
	}

	return roomToFill, nil
}

func GetDeviceTypes() ([]structs.DeviceType, error) {
	log.L.Info("[dbo] getting all device types")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/types/"

	var DeviceTypes []structs.DeviceType
	err := GetData(url, &DeviceTypes)
	if err != nil {
		return []structs.DeviceType{}, err
	}

	return DeviceTypes, nil
}

func AddDeviceType(toAdd structs.DeviceType) (structs.DeviceType, error) {
	log.L.Infof("[dbo] adding device type: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/types/" + toAdd.Name

	var toFill structs.DeviceType
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return structs.DeviceType{}, err
	}

	return toFill, nil
}
func GetPowerStates() ([]structs.PowerState, error) {
	log.L.Info("[dbo] getting all power states")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/powerstates/"

	var powerStates []structs.PowerState
	err := GetData(url, &powerStates)
	if err != nil {
		return []structs.PowerState{}, err
	}

	return powerStates, nil
}

func AddPowerState(toAdd structs.PowerState) (structs.PowerState, error) {
	log.L.Infof("[dbo] adding power state: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/powerstates/" + toAdd.Name

	var toFill structs.PowerState
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return structs.PowerState{}, err
	}

	return toFill, nil
}

func GetMicroservices() ([]structs.Microservice, error) {
	log.L.Info("[dbo] getting all microservices")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/microservices"

	var microservices []structs.Microservice
	err := GetData(url, &microservices)
	if err != nil {
		return []structs.Microservice{}, err
	}

	return microservices, nil
}

func AddMicroservice(toAdd structs.Microservice) (structs.Microservice, error) {
	log.L.Infof("[dbo] adding microservice: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/microservices/" + toAdd.Name

	var toFill structs.Microservice
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return structs.Microservice{}, err
	}

	return toFill, nil
}

func GetEndpoints() ([]structs.Endpoint, error) {
	log.L.Info("[dbo] getting all endpoints")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/endpoints"

	var endpoints []structs.Endpoint
	err := GetData(url, &endpoints)
	if err != nil {
		return []structs.Endpoint{}, err
	}

	return endpoints, nil
}

func AddEndpoint(toAdd structs.Endpoint) (structs.Endpoint, error) {
	log.L.Infof("[dbo] adding endpoint: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/endpoints/" + toAdd.Name

	var toFill structs.Endpoint
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return structs.Endpoint{}, err
	}

	return toFill, nil
}

func GetPortsByClass(class string) ([]structs.DeviceTypePort, error) {
	log.L.Infof("[dbo] Getting ports for class %v", class)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + fmt.Sprintf("/classes/%v/ports", class)

	var ports []structs.DeviceTypePort
	err := GetData(url, &ports)
	return ports, err
}

func GetPorts() ([]structs.PortType, error) {
	log.L.Info("[dbo] getting all ports")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/ports"

	var ports []structs.PortType
	err := GetData(url, &ports)
	if err != nil {
		return []structs.PortType{}, err
	}

	return ports, nil
}

func AddPort(portToAdd structs.PortType) (structs.PortType, error) {
	log.L.Infof("[dbo] adding Port: %v to database", portToAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/ports/" + portToAdd.Name

	var portToFill structs.PortType
	err := PostData(url, portToAdd, &portToFill)
	if err != nil {
		return structs.PortType{}, err
	}

	return portToFill, nil
}

func GetDeviceRoleDefinitions() ([]structs.DeviceRoleDef, error) {
	log.L.Info("[dbo] getting device role definitions")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/roledefinitions"

	var definitions []structs.DeviceRoleDef
	err := GetData(url, &definitions)
	if err != nil {
		return []structs.DeviceRoleDef{}, err
	}

	return definitions, nil
}

func GetDeviceRoleDefinitionById(roleId int) (structs.DeviceRoleDef, error) {

	url := fmt.Sprintf("%s/devices/roledefinitions/%d", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"), roleId)

	var toFill structs.DeviceRoleDef
	err := GetData(url, &toFill)
	if err != nil {
		return structs.DeviceRoleDef{}, err
	}

	return toFill, nil
}

func AddRoleDefinition(toAdd structs.DeviceRoleDef) (structs.DeviceRoleDef, error) {
	log.L.Infof("[dbo] adding role definition: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/roledefinitions/" + toAdd.Name

	var toFill structs.DeviceRoleDef
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return structs.DeviceRoleDef{}, err
	}

	return toFill, nil
}

func GetRoomConfigurations() ([]structs.RoomConfiguration, error) {
	log.L.Info("[dbo] getting room configurations")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/configurations"

	var rcs []structs.RoomConfiguration
	err := GetData(url, &rcs)
	if err != nil {
		return []structs.RoomConfiguration{}, err
	}

	return rcs, nil

}

func GetRoomDesignations() ([]string, error) {
	log.L.Info("[dbo] getting room designations")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/rooms/designations"
	var toReturn []string
	err := GetData(url, &toReturn)
	if err != nil {
		log.L.Errorf("err: %v", err.Error())
		return toReturn, err
	}

	return toReturn, nil
}

func AddDevice(toAdd structs.Device) (structs.Device, error) {
	log.L.Infof("[dbo] adding device: %v to database", toAdd.Name)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + toAdd.Building.Shortname + "/rooms/" + toAdd.Room.Name + "/devices/" + toAdd.Name

	var toFill structs.Device
	err := PostData(url, toAdd, &toFill)
	if err != nil {
		return structs.Device{}, err
	}

	return toFill, nil
}

func GetDeviceClasses() ([]structs.DeviceClass, error) {
	log.L.Info("[dbo] getting all classes")
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/devices/classes"

	var classes []structs.DeviceClass
	err := GetData(url, &classes)

	return classes, err
}

func SetDeviceAttribute(attributeInfo structs.DeviceAttributeInfo) (structs.Device, error) {
	log.L.Infof("[dbo] Setting device attrbute %v to %v for device %v", attributeInfo.AttributeName, attributeInfo.AttributeValue, attributeInfo.AttributeValue)

	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + fmt.Sprintf("/devices/attribute")

	device := structs.Device{}
	err := PutData(url, attributeInfo, &device)
	if err != nil {
		log.L.Errorf("[error] %v", err.Error())
	} else {
		log.L.Info("[dbo] Done.")
	}

	return device, err
}
