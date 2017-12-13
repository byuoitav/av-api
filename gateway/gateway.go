package gateway

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

/*
	SetGateway will set a gateway for the device in question. It will also recursively assign gateways
	We make the assumption that there will only be one gateway per device.

	TODO: Figure out how to assign gateways to specific commands
*/
func SetGateway(url string, device structs.Device) (string, error) {
	if structs.HasRole(device, "GatedDevice") { //we need to add a gateway parameter to the action
		log.Printf(color.BlueString("[gateway-processing]Device %v is a gated device, looking for gateway", device.GetFullName()))
		parseString := `http:\/\/(.+?)\/(.*)`

		gateway, port, err := getDeviceGateway(device)
		if err != nil {
			return "", err
		}
		log.Printf(color.BlueString("[gateway-processing]Found a gateway %v connectd via port %v", gateway.GetFullName(), port))

		newpath, err := processPort(gateway, port)
		if err != nil {
			return "", err
		}
		log.Printf(color.BlueString("[gateway-processing] Generated a new path: %v", newpath))

		//now we need to parse the url and plug the values into the new string
		regex := regexp.MustCompile(parseString)
		vals := regex.FindAllStringSubmatch(url, -1)
		if len(vals) == 0 {
			msg := fmt.Sprintf("[gateway-processing]Invalid path, could not parse path for gateway replacement %v", url)
			log.Printf(color.HiRedString(msg))
			return "", errors.New(msg)
		}

		//now we go through and replace
		newpath = strings.Replace(newpath, ":address", vals[0][1], -1)
		newpath = strings.Replace(newpath, ":path", vals[0][2], -1)
		newpath = strings.Replace(newpath, ":gateway", gateway.Address, -1)

		log.Printf(color.BlueString("[gateway-processing] Processed path: %v", newpath))

		return SetGateway(newpath, gateway)
	}

	return url, nil
}

func SetStatusGateway(url string, device structs.Device) (string, error) {
	return SetGateway(url, device)
}

//finds the address of the device that controls the given device, including the port connecting the two
func getDeviceGateway(d structs.Device) (structs.Device, string, error) {

	//get devices by building and room and role
	devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(d.Building.Shortname, d.Room.Name, "Gateway")
	if err != nil {
		return structs.Device{}, "", err
	}

	if len(devices) == 0 {
		return structs.Device{}, "", errors.New(fmt.Sprintf("no gateway devices found in room %s-%s", d.Building.Name, d.Room.Name))
	}

	for _, device := range devices {

		for _, port := range device.Ports {

			if port.Destination == d.Name {

				return device, port.Name, nil
			}
		}
	}

	return structs.Device{}, "", errors.New("gateway not found")
}

func processPort(gateway structs.Device, port string) (string, error) {
	params := make(map[string]string)

	//check for parameters
	if strings.Contains(port, ":") {
		splits := strings.Split(port, ":")
		port = splits[0]
		i := 0
		for _, v := range splits[1:] {

			//now we process the raw parameters. I can't think of a good way to do this
			//TODO: JB 12/11/17:  revisit this
			params[":"+strconv.Itoa(i)] = v
		}
	}

	//check for a command that corresponds to the port
	command := gateway.GetCommandByName(port)
	if len(command.Name) == 0 {
		//there was an error
		msg := fmt.Sprintf("There was no command for the gateway device %v that corresponds to port %v", gateway.GetFullName(), port)
		log.Printf(color.HiRedString(msg))
		return "", errors.New(msg)
	}
	//for now we assume that those numbered parameters are only valid for the endpoint, otherwise we run into port issues
	path := command.Endpoint.Path

	//replace params
	for k, v := range params {
		path = strings.Replace(path, k, v, -1)
	}

	//we have the command, so we can build the command,
	path = command.Microservice + path

	return path, nil
}
