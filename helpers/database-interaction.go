package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

func getData(url string, structToFill interface{}) error {
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

//GetRoomByInfo simply retrieves a device's information from the databse.
func GetRoomByInfo(roomName string, buildingName string) (accessors.Room, error) {
	log.Printf("Getting room %s in building %s...", roomName, buildingName)
	url := os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS") + "/buildings/" + buildingName + "/rooms/" + roomName
	var toReturn accessors.Room
	err := getData(url, &toReturn)
	return toReturn, err
}

//GetDeviceByName simply retrieves a device's information from the databse.
func GetDeviceByName(roomName string, buildingName string, deviceName string) (accessors.Device, error) {
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
