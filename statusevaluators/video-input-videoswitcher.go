package statusevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
)

const INPUT_STATUS_VIDEO_SWITCHER_EVALUATOR = "STATUS_InputVideoSwitcher"

type InputVideoSwitcher struct {
}

func (p *InputVideoSwitcher) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

func (p *InputVideoSwitcher) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {
	log.Printf("Generating status commands from STATUS_Video_Switcher")

	//first thing is to get the video switcher in the room
	//xuther: we could do this via another call to the database, but looping through is actually faster.
	log.Printf("Looking for video switcher in room")

	found := false
	var switcher structs.Device
	var command structs.Command
	statusCommands := []StatusCommand{}

	for _, device := range devices {
		for _, role := range device.Roles {
			if role == "VideoSwitcher" {
				log.Printf("Found.")
				found = true
				break
			}
		}
		if found {
			found = false
			//check to see if it has the get input by output port command
			for _, c := range device.Commands {
				if c.Name == "STATUS_Input" {
					found = true
					command = c
					break
				}
			}
			if found {
				switcher = device
				log.Printf("Video switcher found")
			}
			break
		}
	}
	if !found {
		log.Printf("No video switcher found in the room, generating standard commands")
		return generateStandardStatusCommand(devices, DEFAULT_INPUT_EVALUATOR, DEFAULT_INPUT_COMMAND)
	}

	var count int

	//this isn't going to be standard
	for _, device := range devices {
		log.Printf("Considering device: %v", device.GetFullName())

		cont := false
		var destinationDevice base.DestinationDevice
		//for now assume that everything is going through the switcher, check to make sure it's a device we care about
		for _, role := range device.Roles {
			if role == "AudioOut" {
				cont = true
				destinationDevice.AudioDevice = true
			}

			if role == "VideoOut" {
				cont = true
				destinationDevice.Display = true
			}
		}
		if !cont {
			log.Printf("Device is not an output device.")
			continue
		}
		log.Printf("Device is an output device.")

		destinationDevice.Device = device
		parameters := make(map[string]string)
		parameters["address"] = switcher.Address

		log.Printf("Looking for an output port that matches the goal device.")

		//find the outport for the device
		for _, p := range switcher.Ports {
			if p.Destination == device.Name {
				split := strings.Split(p.Name, ":")
				parameters["port"] = split[1]
				log.Printf("Found an output port on switcher %v for device %v. Port: %v", switcher.GetFullName(), device.GetFullName(), split[1])
				break
			}
		}
		if _, ok := parameters["port"]; !ok {
			log.Printf("Could find no output port matching the device on the switcher, skipping.")
			continue
		}

		log.Printf("Generating status command.")

		statusCommands = append(statusCommands, StatusCommand{
			Action:            command,
			Device:            switcher,
			Generator:         INPUT_STATUS_VIDEO_SWITCHER_EVALUATOR,
			DestinationDevice: destinationDevice,
			Parameters:        parameters,
		})
		count++
	}
	log.Printf("Done.")

	return statusCommands, count, nil
}

func (p *InputVideoSwitcher) EvaluateResponse(label string, value interface{}, source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	log.Printf("Evaluating response: %s, %s in evaluator %v", label, value, BlankedDefaultName)

	//in this case we assume that there's a single video switcher, so first we get the video switcher in the room, then we match source and dest
	switcherList, err := dbo.GetDevicesByBuildingAndRoomAndRole(source.Building.Shortname, source.Room.Name, "VideoSwitcher")
	if err != nil {
		log.Printf("Error getting the video switcher: %v", err.Error())
		return "", nil, err
	}
	if len(switcherList) != 1 {
		log.Printf("Invalid room for this evaluator, there are %v switchers in the room, expecting 1", len(switcherList))
		return "", nil, errors.New("Invalid room for this evaluator, there is more than one video switcher in the room.")
	}

	//source and dest are in the value string
	bay, ok := value.(string)
	if !ok {
		errString := "Invalid response value for this evaluiator, expects a string"
		log.Printf(errString)
		return "", nil, errors.New(errString)
	}

	for _, port := range switcherList[0].Ports {
		split := strings.Split(port.Name, ":")
		if strings.EqualFold(port.Destination, dest.Name) && bay == split[0] {
			log.Printf("Found a source device that matches the port returned: %v, %v", bay, port.Source)
			return label, port.Source, nil
		}
	}

	log.Printf("Couldn't find a mapping for entry port %v on video switcher %v", bay, switcherList[0].GetFullName())

	return label, value, nil
}
