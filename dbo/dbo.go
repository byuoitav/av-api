package dbo

import (
	"bytes"
	"encoding/json"
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
	req, _ := http.NewRequest("GET", url, nil)

	err := setToken(req)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
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
	log.Printf("Done.")
	return nil
}

//PostData hits POST endpoints
func PostData(url string, structToAdd interface{}) ([]byte, error) {
	log.Printf("Posting data to URL: %s...", url)

	body, err := json.Marshal(structToAdd)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))

	err = setToken(req)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(response.Body)

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
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/commands"
	err = GetData(url, &commands)

	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	log.Printf("Done.")
	return
}

// GetRoomByInfo simply retrieves a device's information from the databse.
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
	var buildings []accessors.Building
	err := GetData(url, &buildings)

	return buildings, err
}

// GetRooms returns all the rooms in a given building
func GetRoomsByBuilding(building string) ([]accessors.Room, error) {
	log.Printf("getting all rooms from %v ...", building)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + building + "/rooms"
	var rooms []accessors.Room
	err := GetData(url, &rooms)
	return rooms, err
}

// AddBuilding monsters
func AddBuilding(buildingToAdd accessors.Building) (accessors.Building, error) {
	log.Printf("adding building %v to database", buildingToAdd.Shortname)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + buildingToAdd.Shortname

	result, err := PostData(url, buildingToAdd)
	if err != nil {
		return buildingToAdd, err
	}

	var building accessors.Building
	err = json.Unmarshal(result, &building)
	if err != nil {
		return building, err
	}

	return building, nil

}
