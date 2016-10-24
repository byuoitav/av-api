package helpers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//PublicRoom is the struct that is returned (or put) as part of the public API
type PublicRoom struct {
	Building          string        `json:"building, omitempty"`
	Room              string        `json:"room, omitempty"`
	CurrentVideoInput string        `json:"currentVideoInput"`
	CurrentAudioInput string        `json:"currentAudioInput"`
	Power             string        `json:"power"`
	Blanked           *bool         `json:"blanked"`
	Displays          []Display     `json:"displays"`
	AudioDevices      []AudioDevice `json:"audioDevices"`
}

//AudioDevice represents an audio device
type AudioDevice struct {
	Name   string `json:"name"`
	Power  string `json:"power"`
	Input  string `json:"input"`
	Muted  *bool  `json:"muted"`
	Volume *int   `json:"volume"`
}

//Display represents a display
type Display struct {
	Name    string `json:"name"`
	Power   string `json:"power"`
	Input   string `json:"input"`
	Blanked *bool  `json:"blanked"`
}

//EditRoomStateNew is just a placeholder
func EditRoomStateNew(roomInQuestion PublicRoom) error {

	log.Printf("Room: %v\n", roomInQuestion)

	//Evaluate commands
	evaluateCommands(roomInQuestion)
	return nil
}

/*
	Note that is is important to add a command to this list and set the rules surounding that command (functionally mapping) property -> command
	here.
*/
func evaluateCommands(roomInQuestion PublicRoom) (actions []ActionStructure, err error) {

	//getAllCommands
	log.Printf("Getting command orders.")
	commands, err := GetAllRawCommands()

	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	//order commands by priority
	commands = orderCommands(commands)
	fmt.Printf("%+v", commands)
	//Switch on each command.

	for _, c := range commands {
		switch c.Name {

		}
	}

	return
}

func orderCommands(commands []accessors.RawCommand) []accessors.RawCommand {
	sorter := accessors.CommandSorterByPriority{Commands: commands}
	sort.Sort(&sorter)
	return sorter.Commands
}

//EditRoomState actually carries out the room
func EditRoomState(roomInQuestion PublicRoom, building string, room string) error {

	log.Printf("Room: %+v\n", roomInQuestion)

	//This is to say if we want to set audio input even if devices are both A/V outputs.
	//forceAudioChange := true
	//if we're setting both, default to video.
	if len(roomInQuestion.CurrentAudioInput) > 0 && len(roomInQuestion.CurrentVideoInput) > 0 {
		//forceAudioChange = false
	}

	log.Printf("Checking for power changes.\n")
	_, valid, err := validateSuppliedValuesPowerChange(&roomInQuestion, room, building)
	if err != nil {
		log.Printf("Error: %s.\n", err.Error())
		return err
	} else if valid {
		log.Printf("Changing power states.\n")
		err = changeCurrentPowerStateForMultipleDevices(roomInQuestion, room, building)
		if err != nil {
			log.Printf("Error: %s.\n", err.Error())
			return err
		}
	}
	log.Printf("Done.\n")

	log.Printf("Checking for input changes.\n")
	//TODO: Have logic here that checks what was passed in and only changes what is necessary.
	valid, err = validateSuppliedVideoChange(roomInQuestion, room, building)
	if err != nil {
		log.Printf("Error: %s.\n", err.Error())
		return err
	} else if valid {
		log.Printf("Changing current input\n")
		err = changeCurrentVideoInput(roomInQuestion, room, building)
		if err != nil {
			log.Printf("Error: %s.\n", err.Error())
			return err
		}
	}
	log.Printf("Done.\n")

	log.Printf("Checking Audio-specific states.\n")
	devices, valid, err := validateSuppliedAudioStateChange(roomInQuestion, room, building)
	if err != nil {
		log.Printf("Error: %s.\n", err.Error())
		return err
	} else if valid {
		err = changeAudioStateForMultipleDevices(roomInQuestion, room, building, devices)
	}
	log.Printf("Done.\n")

	//Check Video Specific states.
	log.Printf("Chacking video-specific states.\n")
	devices, valid, err = validateSuppliedVideoStateChange(&roomInQuestion, room, building)
	if err != nil {
		return err
	} else if valid {
		err = changeVideoStateForMultipleDevices(roomInQuestion, room, building, devices)
	}
	log.Printf("Done.\n")

	return nil
}

func changeVideoStateForMultipleDevices(roomInfo PublicRoom, room string, building string, devices []accessors.Device) error {

	for _, desired := range roomInfo.Displays {
		log.Printf("Setting video state for %s\n", desired.Name)

		var current accessors.Device

		for _, val2 := range devices {
			if strings.EqualFold(val2.Name, desired.Name) {
				current = val2
				break
			}
		}

		if desired.Blanked != nil && len(current.Name) > 0 {
			log.Printf("Setting video to blanked.\n")
			command := ""
			if *desired.Blanked {
				command = "BlankScreen"
			} else {
				command = "UnblankScreen"
			}
			log.Printf("Getting command")
			for _, comm := range current.Commands {
				if strings.EqualFold(comm.Name, command) {
					log.Printf("Command found.")
					address := "http://" +
						comm.Microservice +
						ReplaceIPAddressEndpoint(comm.Endpoint.Path, current.Address)

					log.Printf("Sending Command...")
					_, err := http.Get(address)
					if err != nil {
						return err
					}
					log.Printf("Command Sent")
					break
				}
			}
		}
	}
	return nil
}

func changeAudioStateForMultipleDevices(roomInfo PublicRoom, room string, building string, devices []accessors.Device) error {
	// Get command for set volume devices[0].Command
	for _, desired := range roomInfo.AudioDevices {
		log.Printf("Setting audio states for %s\n", desired.Name)

		var current accessors.Device
		//get accessorDevice for room Device
		for _, val2 := range devices {
			if strings.EqualFold(val2.Name, desired.Name) {
				current = val2
				break
			}
		}

		if desired.Volume != nil {
			if desired.Muted != nil {
				return errors.New("Cannot set Muted and Volume in the same call.")
			}
			b := false
			desired.Muted = &b
		}

		if desired.Muted != nil {
			if *desired.Muted && !*current.Muted {
				//get the muted command
				log.Printf("Setting muted.")
				for _, comm := range current.Commands {
					if strings.EqualFold(comm.Name, "Mute") {
						_, err := http.Get("http://" + comm.Microservice + ReplaceIPAddressEndpoint(comm.Endpoint.Path, current.Address))
						if err != nil {
							log.Printf("Error Muting device %s: %s\n", desired.Name, err.Error())
						} else {
							//set the new volume in the DB.
							*current.Muted = *desired.Muted
							err = setAudioInDB(building, room, current)
							if err != nil {
								return err
							}
						}
					}
				}
			} else if !*desired.Muted && *current.Muted {
				for _, comm := range current.Commands {
					if strings.EqualFold(comm.Name, "UnMute") {
						_, err := http.Get("http://" + comm.Microservice + ReplaceIPAddressEndpoint(comm.Endpoint.Path, current.Address))
						if err != nil {
							log.Printf("Error UnMuting device %s: %s\n", desired.Name, err.Error())
						} else {
							//set the new volume in the DB.
							*current.Muted = *desired.Muted
							err = setAudioInDB(building, room, current)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}

		//after setting the muted.
		if desired.Volume != nil {
			log.Printf("Setting volume.")
			for _, comm := range current.Commands {
				if strings.EqualFold(comm.Name, "SetVolume") {
					endpoint := comm.Endpoint.Path
					endpoint = ReplaceIPAddressEndpoint(endpoint, current.Address)
					//this is our difference
					if strings.Contains(endpoint, "difference") {
						log.Printf("Current: %v, Desired: %v\n", *current.Volume, *desired.Volume)
						difference := *desired.Volume - *current.Volume
						log.Printf("Difference: %v\n", difference)
						endpoint = strings.Replace(endpoint, ":difference", strconv.Itoa(difference), -1)
						log.Printf("New Endpoint: %s\n", endpoint)
					} else {
						endpoint = strings.Replace(endpoint, ":level", strconv.Itoa(*desired.Volume), -1)
					}
					_, err := http.Get("http://" + comm.Microservice + endpoint)
					if err != nil {
						log.Printf("Error Setting volume for device %s: %s. May need to calibrate device.\n", desired.Name, err.Error())
					} else {
						//set the new volume in the DB.
						*current.Muted = false
						*current.Volume = *desired.Volume
						err = setAudioInDB(building, room, current)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func changeCurrentPowerStateForMultipleDevices(roomInfo PublicRoom, room string, building string) error {
	commandNames := make(map[string]string)
	commandNames["on"] = "PowerOn"
	commandNames["standby"] = "Standby"

	//Do the Displays
	for _, val := range roomInfo.Displays {
		log.Printf("Changing power states for display %s to %s.\n", val.Name, val.Power)
		device, err := GetDeviceByName(room, building, val.Name)
		if err != nil {
			//TODO: Figure out reporting here.
			continue
		}

		curCommandName := commandNames[strings.ToLower(val.Power)]
		log.Printf("Checking commands for command %s\n", curCommandName)
		for _, command := range device.Commands {
			if strings.EqualFold(command.Name, curCommandName) {
				log.Printf("Command found.\n")
				endpointPath := ReplaceIPAddressEndpoint(command.Endpoint.Path, device.Address)
				log.Printf("sending Command\n")
				_, err = http.Get("http://" + command.Microservice + endpointPath)
				if err != nil {
					log.Printf("Error %s\n", err.Error())
					break
				}
				log.Printf("Command Sent.\n")
				break
			}
		}
	}

	//Do the Audio Devices
	for _, val := range roomInfo.AudioDevices {
		log.Printf("Changing power states for AudioDevices %s to %s.\n", val.Name, val.Power)
		device, err := GetDeviceByName(room, building, val.Name)
		if err != nil {
			//TODO: Figure out reporting here.
			continue
		}

		curCommandName := commandNames[val.Power]
		log.Printf("Checking commands for command %s\n", curCommandName)
		for _, command := range device.Commands {
			if strings.EqualFold(command.Name, curCommandName) {
				log.Printf("Command found.\n")
				endpointPath := ReplaceIPAddressEndpoint(command.Endpoint.Path, device.Address)
				log.Printf("sending Command\n")
				_, err = http.Get("http://" + command.Microservice + endpointPath)
				if err != nil {
					log.Printf("Error %s\n", err.Error())
					continue
				}
				log.Printf("Command Sent.\n")

				//now we need to update the database so it definately says that it's not muted (assuming it was turned off).
				m := false
				device.Muted = &m
				setAudioInDB(building, room, device)
			}
		}
	}

	return nil
}

type changeInputTuple struct {
	Device accessors.Device
	Input  string
}

/*
First we'll esablish what devices need to be set to which input (allowing for
device specific input settings as well as room-wide settings). We do this by
populating a map - subsequently we send the appropriate commands.
*/
func changeCurrentVideoInput(room PublicRoom, roomName string, buildingName string) error {
	//magic strings - we'll replace these in the endpoint path.
	portToMatch := ":port"
	commandName := "ChangeInput"
	log.Printf("Changing video input.\n")

	//we correlate which device in the room.Displays go to which input.
	deviceInputCorrelation := []changeInputTuple{}

	devices, err := getDevicesByBuildingAndRoomAndRole(roomName, buildingName, "VideoOut")
	if err != nil {
		return err
	}
	log.Printf("%v devices found\n", len(devices))

	//We're gonna go through and see if devices have an indivudual input set. If
	//not we set them to go to the default device, if specified.
	for _, outDev1 := range devices {
		found := false
		for _, outDev2 := range room.Displays {
			if strings.EqualFold(outDev1.Name, outDev2.Name) && len(outDev2.Input) > 0 {
				log.Printf("Found a specific input %s specified for device %s.", outDev2.Input, outDev1.Name)
				found = true
				deviceInputCorrelation = append(deviceInputCorrelation, changeInputTuple{Device: outDev1, Input: outDev2.Input})
				//break from inner loop
				break
			}
		}
		if !found && len(room.CurrentVideoInput) > 0 {
			log.Printf("No sepcific input specified for %s. As a default was speficied will set to default %s", outDev1.Name, room.CurrentVideoInput)
			deviceInputCorrelation = append(deviceInputCorrelation, changeInputTuple{Device: outDev1, Input: room.CurrentVideoInput})
		}
	}

	for _, tuple := range deviceInputCorrelation {
		device := tuple.Device
		input := tuple.Input

		log.Printf("Setting input for %s to %s.", device.Name, input)

		log.Printf("Checking for command %s.", commandName)
		var curCommand accessors.Command

		for _, command := range device.Commands {
			if strings.EqualFold(command.Name, commandName) {
				log.Printf("Found.")
				curCommand = command
				break
			}
		}
		if len(curCommand.Name) <= 0 {
			log.Printf("Command not found, continuing.")
			continue
		}

		log.Printf("Checking output device ports for input device %s.", input)
		//placeholder to be replaced by the value we get down below.
		var portValue string
		for _, port := range device.Ports {
			if strings.EqualFold(port.Source, input) {
				log.Printf("Found, input device into port %s.", port.Name)
				portValue = port.Name
			}
		}

		if len(portValue) <= 0 {
			//TODO: figure out error reporting here.
			log.Printf("Port not found, continuing.")
			continue
		}

		endpointPath := ReplaceIPAddressEndpoint(curCommand.Endpoint.Path, device.Address)

		endpointPath = strings.Replace(endpointPath, portToMatch, portValue, -1)
		//something to get the current port
		log.Printf("Sending Command.\n")
		_, err = http.Get("http://" + curCommand.Microservice + endpointPath)
		if err != nil {
			log.Printf("Error: %s.", err.Error())
			continue
		}
		log.Printf("Command sent to device %s.", device.Name)
	}
	log.Printf("Done setting video inputs.")

	return nil
}

/*
ReplaceIPAddressEndpoint is a simple helper
*/
func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}
