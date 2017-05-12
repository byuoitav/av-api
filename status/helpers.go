package status

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

func GetRoomStatus(building string, roomName string) (base.PublicRoom, error) {

	//get room from database
	room, err := dbo.GetRoomByInfo(building, roomName)
	if err != nil {
		return base.PublicRoom{}, err
	}

	log.Printf("Generating commands...")
	commands, err := generateStatusCommands(room, DEFAULT_MAP)
	if err != nil {
		return base.PublicRoom{}, err
	}

	log.Printf("Running commands...")
	responses, err := runStatusCommands(commands)
	if err != nil {
		return base.PublicRoom{}, err
	}

	log.Printf("Evaluating Responses...")
	roomStatus, err := evaluateResponses(responses)
	if err != nil {
		return base.PublicRoom{}, err
	}

	roomStatus.Building = building
	roomStatus.Room = roomName

	return roomStatus, nil
}

func generateStatusCommands(room accessors.Room, commandMap map[string]StatusEvaluator) ([]StatusCommand, error) {

	var outputs []StatusCommand

	//iterate over each status evaluator
	for _, command := range room.Configuration.Evaluators {

		if strings.HasPrefix(command.EvaluatorKey, FLAG) {

			evaluator := DEFAULT_MAP[command.EvaluatorKey]

			//Idenify relevant devices
			devices, err := evaluator.GetDevices(room)
			if err != nil {
				return []StatusCommand{}, err
			}

			//Generate actions by iterating over the commands of each device
			commands, err := evaluator.GenerateCommands(devices)
			if err != nil {
				return []StatusCommand{}, err
			}

			//log.Printf("Appending commands: %v to action list", commands)
			outputs = append(outputs, commands...)
		}
	}

	//log.Printf("Final command output: %v", outputs)
	return outputs, nil
}

func runStatusCommands(commands []StatusCommand) (outputs []Status, err error) {
	//log.Printf("Commands: %v", commands)
	if len(commands) == 0 {
		err = errors.New("No commands")
		return
	}

	//map device names to commands
	commandMap := make(map[string][]StatusCommand)

	log.Printf("Building device map")
	for _, command := range commands {

		_, present := commandMap[command.Device.Name]
		if !present {
			commandMap[command.Device.Name] = []StatusCommand{command}
			log.Printf("Device %s identified", command.Device.Name)
		} else {
			commandMap[command.Device.Name] = append(commandMap[command.Device.Name], command)
		}

	}

	//make a channel with the same number of 'slots' as devices
	log.Printf("Creating channel")
	channel := make(chan Status, len(commandMap))
	var group sync.WaitGroup

	for device, deviceCommands := range commandMap {

		//spin up new go routine
		log.Printf("Starting new goroutine for device %s", device)
		group.Add(1)
		go issueCommands(deviceCommands, channel, &group)
	}

	log.Printf("Waiting for WaitGroup")
	group.Wait()
	log.Printf("Done waiting")

	log.Printf("Closing channel...")
	close(channel)

	for output := range channel {
		if output.ErrorMessage != nil {
			log.Printf("Error querying status with destination: %s", output.DestinationDevice.Device.Name)
			event := eventinfrastructure.Event{Event: "Status Retrieval",
				Success:  false,
				Building: output.DestinationDevice.Device.Building.Name,
				Room:     output.DestinationDevice.Device.Room.Name,
				Device:   output.DestinationDevice.Device.Name,
			}
			base.Publish(event)
		}
		log.Printf("Appending results of %s to output", output.DestinationDevice.Device.Name)
		outputs = append(outputs, output)
	}
	return
}

//builds a Status object corresponding to a device and writes it to the channel
func issueCommands(commands []StatusCommand, channel chan Status, control *sync.WaitGroup) {

	//add task to waitgroup

	//final output
	output := Status{DestinationDevice: commands[0].DestinationDevice}
	statuses := make(map[string]interface{})

	//iterate over list of StatusCommands
	//TODO:make sure devices can handle rapid-fire API requests
	for _, command := range commands {

		//build url
		url := command.Action.Microservice + command.Action.Endpoint.Path
		for formal, actual := range command.Parameters {
			toReplace := ":" + formal
			if !strings.Contains(url, toReplace) {
				errorMessage := "Could not find parameter " + toReplace + " issuing the command " + command.Action.Name
				output.ErrorMessage = &errorMessage
				log.Printf(errorMessage)
			} else {
				url = strings.Replace(url, toReplace, actual, -1)
			}
		}

		log.Printf("Sending requqest to %s", url)
		response, err := http.Get(url)
		if err != nil {
			errorMessage := err.Error()
			output.ErrorMessage = &errorMessage
			log.Printf("Error getting response from %s", command.Device.Name)
			continue
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			errorMessage := err.Error()
			output.ErrorMessage = &errorMessage
			log.Printf("Error reading response from %s", command.Device.Name)
			continue
		}
		log.Printf("Microservice returned: %s", body)

		var status map[string]interface{}
		err = json.Unmarshal(body, &status)
		if err != nil {
			errorMessage := err.Error()
			output.ErrorMessage = &errorMessage
			log.Printf("Error unmarshalling response from %s", command.Device.Name)
			continue
		}

		for device, object := range status {
			statuses[device] = object
		}
	}

	//write output to channel
	channel <- output
	log.Printf("Done acquiring status for %s", output.DestinationDevice.Device.Name)
	control.Done()
}

func evaluateResponses(responses []Status) (base.PublicRoom, error) {

	var AudioDevices []base.AudioDevice
	var Displays []base.Display

	for _, device := range responses {

		if device.DestinationDevice.AudioDevice {

			var audioDevice base.AudioDevice

			for _, response := range device.Status {

				data, ok := response.(int)
				if ok {
					audioDevice.Volume = &data
				}

				other := response.(bool)
				audioDevice.Muted = &other
			}

			power, ok := device.Status["power"]
			powerString, ok := power.(string)
			if ok {
				audioDevice.Power = powerString
			}

			input, ok := device.Status["input"]
			inputString, ok := input.(string)
			if ok {
				audioDevice.Input = inputString
			}

			audioDevice.Name = device.DestinationDevice.Device.Name

			log.Printf("Appending device: %s to AudioDevice array", audioDevice.Name)
			AudioDevices = append(AudioDevices, audioDevice)
		}

		if device.DestinationDevice.Display {
			var display base.Display

			for _, response := range device.Status {

				data, ok := response.(bool)
				if ok {
					display.Blanked = &data
				}

			}

			power, ok := device.Status["power"]
			powerString, ok := power.(string)
			if ok {
				display.Power = powerString
			}

			input, ok := device.Status["input"]
			inputString, ok := input.(string)
			if ok {
				display.Input = inputString
			}

			display.Name = device.DestinationDevice.Device.Name

			log.Printf("Appending device: %s to Dispaly array", display.Name)
			Displays = append(Displays, display)

		}

	}

	return base.PublicRoom{Displays: Displays, AudioDevices: AudioDevices}, nil
}
