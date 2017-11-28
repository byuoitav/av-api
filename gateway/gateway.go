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
		gateway, err := getDeviceGateway(action.Device)
		if err != nil {
			msg := fmt.Sprintf("gateway for %s not found: %s", action.Device.Name, err.Error())
			log.Printf("%s", color.HiRedString("[error] %s", msg))
		}

		action.Parameters["gateway"] = gateway
	}
	return nil

}

//finds the IP of the device that controls the given device
func getDeviceGateway(d structs.Device) (string, error) {

	for _, port := range d.Ports { //range over all ports

		device, err := dbo.GetDeviceByName(d.Building.Name, d.Room.Name, port.Source)
		if err != nil {
			return "", errors.New(fmt.Sprintf("unable to get source device from port: %s", err.Error()))
		}

		if structs.HasRole(device, "Gateway") {
			return device.Address, nil
		}
	}

	return "", errors.New("no gateway found")
}
