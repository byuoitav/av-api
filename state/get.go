package state

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/av-api/base"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

// GenerateStatusCommands determines the status commands for the type of room that the device is in.
func GenerateStatusCommands(room structs.Room, commandMap map[string]se.StatusEvaluator) ([]se.StatusCommand, int, error) {

	color.Set(color.FgHiCyan)
	base.Log("[state] generating status commands...")
	color.Unset()

	var output []se.StatusCommand
	var count int

	for _, possibleEvaluator := range room.Configuration.Evaluators {

		if strings.HasPrefix(possibleEvaluator.CodeKey, se.FLAG) {

			currentEvaluator := se.STATUS_EVALUATORS[possibleEvaluator.CodeKey]

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

// RunStatusCommands maps the device names to their commands, and then puts them in a channel to be run.
func RunStatusCommands(commands []se.StatusCommand) (outputs []se.StatusResponse, err error) {

	base.Log("%s", color.HiBlueString("[state] running status commands..."))

	if len(commands) == 0 {
		err = errors.New("no commands")
		return
	}

	//map device names to commands
	commandMap := make(map[string][]se.StatusCommand)

	base.Log("%s", color.HiBlueString("[state] building device map..."))
	for _, command := range commands {

		//base.Log("[state] command: %s against device %s, destination device: %s, parameters: %v", command.Action.Name, command.Device.Name, command.DestinationDevice.Device.Name, command.Parameters)
		_, present := commandMap[command.Device.ID]
		if !present {
			commandMap[command.Device.ID] = []se.StatusCommand{command}
			//	base.Log("Device %s identified", command.Device.Name)
		} else {
			commandMap[command.Device.ID] = append(commandMap[command.Device.ID], command)
		}

	}

	base.Log("[state] creating channel")
	channel := make(chan []se.StatusResponse, len(commandMap))
	var group sync.WaitGroup

	for _, deviceCommands := range commandMap {
		group.Add(1)
		go issueCommands(deviceCommands, channel, &group)

		base.Log("%s", color.HiBlueString("[state] commands to issue:"))

		/*
			for _, command := range deviceCommands {
				base.Log("[state] command: %s against device %s, destination device: %s, parameters: %v", command.Action.Name, command.Device.Name, command.DestinationDevice.Device.Name, command.Parameters)
			}
		*/
	}

	base.Log("[state] waiting for commands to issue...")
	group.Wait()
	base.Log("[state] Done. Closing channel...")
	close(channel)

	for outputList := range channel {
		for _, output := range outputList {
			if output.ErrorMessage != nil {
				msg := fmt.Sprintf("problem querying status of device: %s with destination %s: %s", output.SourceDevice.Name, output.DestinationDevice.Name, *output.ErrorMessage)
				base.Log("%s", color.HiRedString("[error] %s", msg))
				cause := events.INTERNAL
				base.PublishError(msg, cause)
			}
			//base.Log("[state] appending status: %v of %s to output", output.Status, output.DestinationDevice.Name)
			outputs = append(outputs, output)
		}
	}
	return
}

// EvaluateResponses organizes the responses that are received when the commands are issued.
func EvaluateResponses(responses []se.StatusResponse, count int) (base.PublicRoom, error) {

	base.Log("%s", color.HiBlueString("[state] Evaluating responses..."))

	if len(responses) == 0 { //make sure things aren't broken
		msg := "no status responses found"
		base.Log("%s", color.HiRedString("[error] %s", msg))
		return base.PublicRoom{}, errors.New(msg)
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
				base.Log("[state] Checking generator: %s", resp.Generator)
				k, v, err := se.STATUS_EVALUATORS[resp.Generator].EvaluateResponse(key, value, resp.SourceDevice, resp.DestinationDevice)
				if err != nil {

					base.Log("%s", color.HiRedString("[state] problem procesing the response %v - %v with evaluator %v: %s",
						key, value, resp.Generator, err.Error()))
					continue
				}

				if _, ok := responsesByDestinationDevice[resp.DestinationDevice.ID]; ok {
					responsesByDestinationDevice[resp.DestinationDevice.ID].Status[k] = v
					doneCount++
				} else {
					newMap := make(map[string]interface{})
					newMap[k] = v
					statusForDevice := se.Status{
						Status:            newMap,
						DestinationDevice: resp.DestinationDevice,
					}
					responsesByDestinationDevice[resp.DestinationDevice.ID] = statusForDevice
					base.Log("[state] adding device %v to the map", resp.DestinationDevice.ID)
					doneCount++
				}
			}
		} else {
			//we call the callback and then wait for it to come back to us
			for key, value := range resp.Status {
				resp.Callback(base.StatusPackage{Key: key, Value: value, Device: resp.SourceDevice, Dest: resp.DestinationDevice}, returnChan)
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
			if _, ok := responsesByDestinationDevice[val.Dest.ID]; ok {
				responsesByDestinationDevice[val.Dest.ID].Status[val.Key] = val.Value
				doneCount++
			} else {
				newMap := make(map[string]interface{})
				newMap[val.Key] = val.Value
				statusForDevice := se.Status{
					Status:            newMap,
					DestinationDevice: val.Dest,
				}
				responsesByDestinationDevice[val.Dest.ID] = statusForDevice
				base.Log("[state] adding device %v to the map", val.Dest.ID)
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
