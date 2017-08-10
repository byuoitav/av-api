package state

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/av-api/base"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/accessors"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

func GenerateStatusCommands(room accessors.Room, commandMap map[string]se.StatusEvaluator) ([]se.StatusCommand, error) {

	log.Printf("Generating commands...")

	var output []se.StatusCommand

	for _, possibleEvaluator := range room.Configuration.Evaluators {

		if strings.HasPrefix(possibleEvaluator.EvaluatorKey, se.FLAG) {

			currentEvaluator := se.STATUS_EVALUATORS[possibleEvaluator.EvaluatorKey]

			devices, err := currentEvaluator.GetDevices(room)
			if err != nil {
				return []se.StatusCommand{}, err
			}

			commands, err := currentEvaluator.GenerateCommands(devices)
			if err != nil {
				return []se.StatusCommand{}, err
			}

			output = append(output, commands...)
		}
	}

	log.Printf("All commands generated: \n\n")
	for _, command := range output {
		log.Printf("Command: %s against device %s, destination device: %s, parameters: %v", command.Action.Name, command.Device.Name, command.DestinationDevice.Device.Name, command.Parameters)
	}
	return output, nil
}

func RunStatusCommands(commands []se.StatusCommand) (outputs []se.StatusResponse, err error) {

	log.Printf("Running commands...")

	if len(commands) == 0 {
		err = errors.New("No commands")
		return
	}

	//map device names to commands
	commandMap := make(map[string][]se.StatusCommand)

	log.Printf("Building device map\n\n")
	for _, command := range commands {

		log.Printf("Command: %s against device %s, destination device: %s, parameters: %v", command.Action.Name, command.Device.Name, command.DestinationDevice.Device.Name, command.Parameters)
		_, present := commandMap[command.Device.Name]
		if !present {
			commandMap[command.Device.Name] = []se.StatusCommand{command}
			//	log.Printf("Device %s identified", command.Device.Name)
		} else {
			commandMap[command.Device.Name] = append(commandMap[command.Device.Name], command)
		}

	}

	log.Printf("Creating channel")
	channel := make(chan []se.StatusResponse, len(commandMap))
	var group sync.WaitGroup

	for _, deviceCommands := range commandMap {
		group.Add(1)
		go issueCommands(deviceCommands, channel, &group)

		log.Printf("Commands getting issued: \n\n")

		for _, command := range deviceCommands {
			log.Printf("Command: %s against device %s, destination device: %s, parameters: %v", command.Action.Name, command.Device.Name, command.DestinationDevice.Device.Name, command.Parameters)
		}
	}

	log.Printf("Waiting for commands to issue...")
	group.Wait()
	log.Printf("Done. Closing channel...")
	close(channel)

	for outputList := range channel {
		for _, output := range outputList {
			if output.ErrorMessage != nil {
				log.Printf("Error querying status of device: %s: %s", output.SourceDevice.Name, *output.ErrorMessage)
				cause := eventinfrastructure.INTERNAL
				message := *output.ErrorMessage
				message = "Error querying status for destination: " + output.DestinationDevice.Name + ": " + message
				base.PublishError(message, cause)
			}
			log.Printf("Appending status: %v of %s to output", output.Status, output.DestinationDevice.Name)
			outputs = append(outputs, output)
		}
	}
	return
}

//builds a Status object corresponding to a device and writes it to the channel
func issueCommands(commands []se.StatusCommand, channel chan []se.StatusResponse, control *sync.WaitGroup) {

	log.Printf("Issuing commands...\n\n")

	//final output
	outputs := []se.StatusResponse{}

	//iterate over list of StatusCommands
	//TODO:make sure devices can handle rapid-fire API requests
	for _, command := range commands {

		log.Printf("Command: %s against device %s, destination device: %s, parameters: %v", command.Action.Name, command.Device.Name, command.DestinationDevice.Device.Name, command.Parameters)

		output := se.StatusResponse{
			Generator:         command.Generator,
			SourceDevice:      command.Device,
			DestinationDevice: command.DestinationDevice,
		}
		statusResponseMap := make(map[string]interface{})

		//build url
		url := command.Action.Microservice + command.Action.Endpoint.Path
		for formal, actual := range command.Parameters {

			log.Printf("Formal: %s, actual: %s", formal, actual)

			toReplace := ":" + formal
			if !strings.Contains(url, toReplace) {
				errorMessage := "Could not find parameter " + toReplace + " issuing the command " + command.Action.Name
				output.ErrorMessage = &errorMessage
				outputs = append(outputs, output)
				log.Printf(errorMessage)
			} else {
				url = strings.Replace(url, toReplace, actual, -1)
			}
		}

		log.Printf("Sending requqest to %s", url)
		timeout := time.Duration(TIMEOUT * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		response, err := client.Get(url)
		if err != nil {
			errorMessage := err.Error()
			output.ErrorMessage = &errorMessage
			outputs = append(outputs, output)
			log.Printf("Error getting response from %s", command.Device.Name)
			continue
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			errorMessage := err.Error()
			output.ErrorMessage = &errorMessage
			outputs = append(outputs, output)
			log.Printf("Error reading response from %s", command.Device.Name)
			continue
		}

		//check to see if it returned a non 200 response, if so, we need to build the error.
		if response.StatusCode != 200 {
			log.Printf("Non 200 recieved: logging. Message recieved: %s", body)
			errorMessage := "Error with the request: %v" + string(body)
			base.PublishError(errorMessage, eventinfrastructure.AUTOGENERATED)
			continue
		}

		log.Printf("microservice returned: %s", body)

		var status map[string]interface{}
		err = json.Unmarshal(body, &status)
		if err != nil {
			errorMessage := "microservice returned: " + string(body)
			output.ErrorMessage = &errorMessage
			outputs = append(outputs, output)
			log.Printf("Error unmarshalling response from %s", command.Device.Name)
			continue
		}

		log.Printf("Copying data into output")
		for device, object := range status {
			statusResponseMap[device] = object
			log.Printf("%s maps to %v", device, object)
		}

		output.Status = statusResponseMap
		//add the full status response
		outputs = append(outputs, output)
	}

	//write output to channel
	log.Printf("Writing output to channel...")
	for _, output := range outputs {
		log.Printf("outputs from device %v", output.SourceDevice.GetFullName())
		for key, value := range output.Status {
			log.Printf("%s maps to %v", key, value)
		}
	}

	channel <- outputs
	log.Printf("Done acquiring statuses from  %s", commands[0].Device.GetFullName())
	control.Done()
}

func evaluateResponses(responses []se.StatusResponse) (base.PublicRoom, error) {

	log.Printf("Evaluating responses...")

	var AudioDevices []base.AudioDevice
	var Displays []base.Display

	//make our array of Statuses by device
	responsesByDestinationDevice := make(map[string]se.Status)
	for _, resp := range responses {
		for key, value := range resp.Status {
			log.Printf("Checking generator: %s", resp.Generator)
			k, v, err := se.STATUS_EVALUATORS[resp.Generator].EvaluateResponse(key, value, resp.SourceDevice, resp.DestinationDevice)
			if err != nil {
				//log an error
				log.Printf("There was a problem procesing the response %v - %v with evaluator %v: %s",
					key, value, resp.Generator, err.Error())
				continue
			}
			if _, ok := responsesByDestinationDevice[resp.DestinationDevice.GetFullName()]; ok {
				responsesByDestinationDevice[resp.DestinationDevice.GetFullName()].Status[k] = v
			} else {
				newMap := make(map[string]interface{})
				newMap[k] = v
				statusForDevice := se.Status{
					Status:            newMap,
					DestinationDevice: resp.DestinationDevice,
				}
				responsesByDestinationDevice[resp.DestinationDevice.GetFullName()] = statusForDevice
				log.Printf("Adding Device %v to the map", resp.DestinationDevice.GetFullName())
			}
		}
	}
	for _, v := range responsesByDestinationDevice {
		if v.DestinationDevice.AudioDevice {
			audioDevice, err := processAudioDevice(v)
			if err == nil {
				AudioDevices = append(AudioDevices, audioDevice)
			}
		}
		if v.DestinationDevice.Display {

			display, err := processDisplay(v)
			if err == nil {
				Displays = append(Displays, display)
			}
		}
	}

	return base.PublicRoom{Displays: Displays, AudioDevices: AudioDevices}, nil
}

func processAudioDevice(device se.Status) (base.AudioDevice, error) {

	log.Printf("Adding audio device: %s", device.DestinationDevice.Name)
	log.Printf("Status map: %v", device.Status)

	var audioDevice base.AudioDevice

	muted, ok := device.Status["muted"]
	mutedBool, ok := muted.(bool)
	if ok {
		audioDevice.Muted = &mutedBool
	}

	volume, ok := device.Status["volume"]
	if ok {
		//Default unmarshals to a float 64 - so we have to coerce it to an int
		var volumeInt int
		if volFloat, ok := volume.(float64); ok {
			volumeInt = int(volFloat)
		} else {
			volumeInt, ok = volume.(int)
		}

		//volumeint should be set now
		if ok {
			audioDevice.Volume = &volumeInt
		} else {
			log.Printf("Volume type assertion failed for %v", volume)
		}
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

	audioDevice.Name = device.DestinationDevice.Name
	return audioDevice, nil
}

func processDisplay(device se.Status) (base.Display, error) {

	log.Printf("Adding display: %s", device.DestinationDevice.Name)

	var display base.Display

	blanked, ok := device.Status["blanked"]
	blankedBool, ok := blanked.(bool)
	if ok {
		display.Blanked = &blankedBool
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

	display.Name = device.DestinationDevice.Name

	return display, nil
}
