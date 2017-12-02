package gateway

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

func SetGateway(path string, device structs.Device) (string, error) {

	parameter := ":gateway"

	if structs.HasRole(device, "GatedDevice") && strings.Contains(path, parameter) { //we need to add a gateway parameter to the action
		gateway, err := getDeviceGateway(device)
		if err != nil {
			return "", err
		}

		return strings.Replace(path, parameter, gateway, -1), nil
	}

	return path, nil //if the condition failed, just pass through

}

func SetStatusGateway(url string, device structs.Device) (string, error) {

	parameter := ":gateway"

	if structs.HasRole(device, "GatedDevice") && strings.Contains(url, parameter) { //we need to add a gateway parameter to the action

		log.Printf("%s", color.HiYellowString("[gateway] identified gated device %s", device.Name))

		gateway, err := getDeviceGateway(device)
		if err != nil {
			return "", err
		}

		return strings.Replace(url, parameter, gateway, -1), nil
	}

	return url, nil
}

//finds the IP of the device that controls the given device
func getDeviceGateway(d structs.Device) (string, error) {

	//get devices by building and room and role
	devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(d.Building.Shortname, d.Room.Name, "Gateway")
	if err != nil {
		return "", err
	}

	if len(devices) == 0 {
		return "", errors.New(fmt.Sprintf("no gateway devices found in room %s-%s", d.Building.Name, d.Room.Name))
	}

	for _, device := range devices {

		for _, port := range device.Ports {

			if port.Destination == d.Name {

				return device.Address, nil
			}
		}
	}

	return "", errors.New("gateway not found")
}
