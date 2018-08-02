package statusevaluators

import (
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// InputVideoSwitcherEvaluator is a constant variable for the name of the evaluator.
const InputVideoSwitcherEvaluator = "STATUS_InputVideoSwitcher"

// InputVideoSwitcher implements the StatusEvaluator struct.
type InputVideoSwitcher struct {
}

// GetDevices returns a list of devices in the given room.
func (p *InputVideoSwitcher) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

// GenerateCommands generates a list of commands for the given devices.
func (p *InputVideoSwitcher) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {
	log.L.Info("[statusevals] Generating status commands from STATUS_Video_Switcher")

	//xuther: we could do this via another call to the database, but looping through is actually faster.
	log.L.Info("[statusevals] Looking for video switcher in room")

	var switcher structs.Device
	var command structs.Command
	statusCommands := []StatusCommand{}

	/*
		found := false
			for _, device := range devices {
				for _, role := range device.Roles {
					if role.ID == "VideoSwitcher" {
						log.L.Info("[statusevals] Found.")
						found = true
						break
					}
				}
				if found {
					found = false
					//check to see if it has the get input by output port command
					for _, c := range device.Type.Commands {
						if c.ID == "STATUS_Input" {
							found = true
							command = c
							break
						}
					}
					if found {
						switcher = device
						log.L.Info("[statusevals] Video switcher found")
					}
					break
				}
			}
			if !found {
				log.L.Info("[statusevals] No video switcher found in the room, generating standard commands")
				return generateStandardStatusCommand(devices, DefaultInputEvaluator, DefaultInputCommand)
			}
	*/

	var count int

	for _, device := range devices {
		log.L.Infof("[statusevals] Considering device: %v", device.ID)

		cont := false
		var destinationDevice base.DestinationDevice
		//for now assume that everything is going through the switcher, check to make sure it's a device we care about
		for _, role := range device.Roles {
			if role.ID == "AudioOut" {
				cont = true
				destinationDevice.AudioDevice = true
			}

			if role.ID == "VideoOut" {
				cont = true
				destinationDevice.Display = true
			}
		}
		if !cont {
			log.L.Info("[statusevals] Device is not an output device.")
			continue
		}
		log.L.Info("[statusevals] Device is an output device.")

		destinationDevice.Device = device
		parameters := make(map[string]string)
		parameters["address"] = switcher.Address

		log.L.Info("[statusevals] Looking for an output port that matches the goal device.")

		//find the outport for the device
		for _, p := range switcher.Ports {
			if p.DestinationDevice == device.ID {
				split := strings.Split(p.ID, ":")
				parameters["port"] = split[1]
				log.L.Infof("[statusevals] Found an output port on switcher %v for device %v. Port: %v", switcher.ID, device.ID, split[1])
				break
			}
		}
		if _, ok := parameters["port"]; !ok {
			log.L.Info("[statusevals] Could find no output port matching the device on the switcher, skipping.")
			continue
		}

		log.L.Info("[statusevals] Generating status command.")

		statusCommands = append(statusCommands, StatusCommand{
			Action:            command,
			Device:            switcher,
			Generator:         InputVideoSwitcherEvaluator,
			DestinationDevice: destinationDevice,
			Parameters:        parameters,
		})
		count++
	}
	log.L.Info("[statusevals] Done.")

	return statusCommands, count, nil
}

// EvaluateResponse processes the response information that is given.
func (p *InputVideoSwitcher) EvaluateResponse(label string, value interface{}, source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	log.L.Infof("[statusevals] Evaluating response: %s, %s in evaluator %v", label, value, BlankedDefaultEvaluator)

	//in this case we assume that there's a single video switcher, so first we get the video switcher in the room, then we match source and dest
	switcherList, err := db.GetDB().GetDevicesByRoomAndRole(source.GetDeviceRoomID(), "VideoSwitcher")
	if err != nil {
		log.L.Errorf("[statusevals] Error getting the video switcher: %v", err.Error())
		return "", nil, err
	}
	if len(switcherList) != 1 {
		msg := fmt.Sprintf("[statusevals] Invalid room for this evaluator, there are %v switchers in the room, expecting 1", len(switcherList))
		log.L.Error(msg)
		return "", nil, errors.New(msg)
	}

	//source and dest are in the value string
	bay, ok := value.(string)
	if !ok {
		errString := "[statusevals] Invalid response value for this evaluiator, expects a string"
		log.L.Error(errString)
		return "", nil, errors.New(errString)
	}

	for _, port := range switcherList[0].Ports {
		split := strings.Split(port.ID, ":")
		if strings.EqualFold(port.DestinationDevice, dest.ID) && bay == split[0] {
			log.L.Infof("[statusevals] Found a source device that matches the port returned: %v, %v", bay, port.SourceDevice)
			return label, port.SourceDevice, nil
		}
	}

	log.L.Infof("[statusevals] Couldn't find a mapping for entry port %v on video switcher %v", bay, switcherList[0].ID)

	return label, value, nil
}
