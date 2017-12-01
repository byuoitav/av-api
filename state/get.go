package state

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/av-api/base"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/fatih/color"
)

func GenerateStatusCommands(room structs.Room, commandMap map[string]se.StatusEvaluator) ([]se.StatusCommand, int, error) {

	color.Set(color.FgHiCyan)
	log.Printf("[state] generating status commands...")
	color.Unset()

	var output []se.StatusCommand
	var count int

	for _, possibleEvaluator := range room.Configuration.Evaluators {

		if strings.HasPrefix(possibleEvaluator.EvaluatorKey, se.FLAG) {

			currentEvaluator := se.STATUS_EVALUATORS[possibleEvaluator.EvaluatorKey]

			//we can get the number of output devices here
			devices, err := currentEvaluator.GetDevices(room)
			if err != nil {
				return []se.StatusCommand{}, 0, err
			}

			//we get the number of commands here
			commands, c, err := currentEvaluator.GenerateCommands(devices)
			if err != nil {
				return []se.StatusCommand{}, 0, err
			}
			count += c
			output = append(output, commands...)
		}
	}

	return output, count, nil
}

func RunStatusCommands(commands []se.StatusCommand) (outputs []se.StatusResponse, err error) {

	log.Printf("%s", color.HiBlueString("[state] running status commands..."))

	if len(commands) == 0 {
		err = errors.New("no commands")
		return
	}

	//map device names to commands
	commandMap := make(map[string][]se.StatusCommand)

	log.Printf("%s", color.HiBlueString("[state] building device map..."))
	for _, command := range commands {

		log.Printf("[state] command: %s against device %s, destination device: %s, parameters: %v", command.Action.Name, command.Device.Name, command.DestinationDevice.Device.Name, command.Parameters)
		_, present := commandMap[command.Device.Name]
		if !present {
			commandMap[command.Device.Name] = []se.StatusCommand{command}
			//	log.Printf("Device %s identified", command.Device.Name)
		} else {
			commandMap[command.Device.Name] = append(commandMap[command.Device.Name], command)
		}

	}

	log.Printf("[state] creating channel")
	channel := make(chan []se.StatusResponse, len(commandMap))
	var group sync.WaitGroup

	for _, deviceCommands := range commandMap {
		group.Add(1)
		go issueCommands(deviceCommands, channel, &group)

		log.Printf("%s", color.HiBlueString("[state] commands to issue:"))
		for _, command := range deviceCommands {
			log.Printf("[state] command: %s against device %s, destination device: %s, parameters: %v", command.Action.Name, command.Device.Name, command.DestinationDevice.Device.Name, command.Parameters)
		}
	}

	log.Printf("[state] waiting for commands to issue...")
	group.Wait()
	log.Printf("[state] Done. Closing channel...")
	close(channel)

	for outputList := range channel {
		for _, output := range outputList {
			if output.ErrorMessage != nil {
				msg := fmt.Sprintf("problem querying status of device: %s with destination %s: %s", output.SourceDevice.Name, output.DestinationDevice.Name, *output.ErrorMessage)
				log.Printf("%s", color.HiRedString("[error] %s", msg))
				cause := eventinfrastructure.INTERNAL
				base.PublishError(msg, cause)
			}
			log.Printf("[state] appending status: %v of %s to output", output.Status, output.DestinationDevice.Name)
			outputs = append(outputs, output)
		}
	}
	return
}

func EvaluateResponses(responses []se.StatusResponse, count int) (base.PublicRoom, error) {

	log.Printf("%s", color.HiBlueString("[state] Evaluating responses..."))

	if len(responses) == 0 { //make sure things aren't broken
		msg := "no status responses found"
		return base.PublicRoom{}, errors.New(msg)
		log.Printf("%s", color.HiRedString("[error] %s", msg))
	}

	var AudioDevices []base.AudioDevice
	var Displays []base.Display
	doneCount := 0

	//we need to create our return channel
	returnChan := make(chan base.StatusPackage, len(responses))

	//make our array of Statuses by device
	responsesByDestinationDevice := make(map[string]se.Status)
	for _, resp := range responses {

		//we do thing the old fashioned way
		if resp.Callback == nil {
			for key, value := range resp.Status {
				log.Printf("[state] Checking generator: %s", resp.Generator)
				k, v, err := se.STATUS_EVALUATORS[resp.Generator].EvaluateResponse(key, value, resp.SourceDevice, resp.DestinationDevice)
				if err != nil {

					color.Set(color.FgHiRed, color.Bold)
					log.Printf("[state] problem procesing the response %v - %v with evaluator %v: %s",
						key, value, resp.Generator, err.Error())
					color.Unset()
					continue
				}

				if _, ok := responsesByDestinationDevice[resp.DestinationDevice.GetFullName()]; ok {
					responsesByDestinationDevice[resp.DestinationDevice.GetFullName()].Status[k] = v
					doneCount++
				} else {
					newMap := make(map[string]interface{})
					newMap[k] = v
					statusForDevice := se.Status{
						Status:            newMap,
						DestinationDevice: resp.DestinationDevice,
					}
					responsesByDestinationDevice[resp.DestinationDevice.GetFullName()] = statusForDevice
					log.Printf("[state] Adding Device %v to the map", resp.DestinationDevice.GetFullName())
					doneCount++
				}
			}
		} else {
			//we call the callback and then wait for it to come back to us
			for key, value := range resp.Status {
				resp.Callback(base.StatusPackage{key, value, resp.SourceDevice, resp.DestinationDevice}, returnChan)
			}
		}
	}

	//start a timer to give us our timeout
	timer := time.NewTimer(time.Second)
	done := false

	//now we wait for the timeout, or all of the responses
	for doneCount < count && !done {
		select {
		case <-timer.C:
			//get out
			done = true
			break

		//pull something out of the response channel
		case val := <-returnChan:
			if _, ok := responsesByDestinationDevice[val.Dest.GetFullName()]; ok {
				responsesByDestinationDevice[val.Dest.GetFullName()].Status[val.Key] = val.Value
				doneCount++
			} else {
				newMap := make(map[string]interface{})
				newMap[val.Key] = val.Value
				statusForDevice := se.Status{
					Status:            newMap,
					DestinationDevice: val.Dest,
				}
				responsesByDestinationDevice[val.Dest.GetFullName()] = statusForDevice
				log.Printf("[state] Adding Device %v to the map", val.Dest.GetFullName())
				doneCount++
			}
		}
	}

	//now we carry on

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
