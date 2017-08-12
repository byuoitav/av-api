package state

import (
	"errors"
	"log"
	"strings"
	"sync"

	"github.com/byuoitav/av-api/base"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

func GenerateStatusCommands(room structs.Room, commandMap map[string]se.StatusEvaluator) ([]se.StatusCommand, error) {

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

func EvaluateResponses(responses []se.StatusResponse) (base.PublicRoom, error) {

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
