package gateway

import (
	"errors"
	"fmt"
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

func SetGateway(action *base.ActionStructure) error {

	if structs.HasRole(action.Device, "GatedDevice") { //we need to add a gateway parameter to the action
		gateway, err := getDeviceGateway(action.Device)
		if err != nil {
			msg := fmt.Sprintf("gateway for %s not found: %s", action.Device.Name, err.Error())
			log.Printf("%s", color.HiRedString("[error] %s", msg))
		}

		action.Parameters["gateway"] = gateway
	}
	return nil

}

func SetStatusGateway(action *statusevaluators.StatusCommand) error {

	if structs.HasRole(action.Device, "GatedDevice") { //we need to add a gateway parameter to the action

		log.Printf("%s", color.HiYellowString("[gateway] identified gated device %s", action.Device.Name))

		gateway, err := getDeviceGateway(action.Device)
		if err != nil {
			return err
		}

		action.Parameters["gateway"] = gateway
	}

	return nil
}

//finds the IP of the device that controls the given device
func getDeviceGateway(d structs.Device) (string, error) {

	//get devices by building and room and role
	devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(d.Building.Shortname, d.Room.Name, "Gateway")
	if err != nil {
		return "", err
	}

	if len(devices) == 0 {
		msg := "no gateway devices found"
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return "", errors.New(msg)
	}

	for _, device := range devices {

		log.Printf("%s", color.HiYellowString("[gateway] found device %s", device.Name))

		for _, port := range device.Ports {

			if port.Destination == d.Name {

				return device.Address, nil
			}
		}
	}

	return "", errors.New("gateway not found")
}
