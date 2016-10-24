package dbo

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//GetData will run a get on the url, and attempt to fill the interface provided
//from the returned JSON.
func GetData(url string, structToFill interface{}) error {
	log.Printf("Getting data from URL: %s...", url)
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
	log.Printf("Done.")
	return nil
}

//GetAllRawCommands retrieves all the commands
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

//GetRoomByInfo simply retrieves a device's information from the databse.
func GetRoomByInfo(roomName string, buildingName string) (toReturn accessors.Room, err error) {
	log.Printf("Getting room %s in building %s...", roomName, buildingName)
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+buildingName+"/rooms/"+roomName, &toReturn)
	return
}

//GetDeviceByName simply retrieves a device's information from the databse.
func GetDeviceByName(roomName string, buildingName string, deviceName string) (toReturn accessors.Device, err error) {
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+buildingName+"/rooms/"+roomName+"/devices/"+deviceName, &toReturn)
	return
}

//GetDevicesByRoom will jut get the devices based on the room.
func GetDevicesByRoom(roomName string, buildingName string) (toReturn []accessors.Device, err error) {
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+buildingName+"/rooms/"+roomName+"/devices", &toReturn)
	return
}

//GetDevicesByBuildingAndRoomAndRole will get the devices with the given role from the DB
func GetDevicesByBuildingAndRoomAndRole(room string, building string, roleName string) (toReturn []accessors.Device, err error) {
	err = GetData(os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS")+"/buildings/"+building+"/rooms/"+room+"/devices/roles/"+roleName, &toReturn)
	return
}

//SetAudioInDB will set the audio levels in the database
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
