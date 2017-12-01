package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/debug"
	"github.com/byuoitav/av-api/gateway"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/configuration-database-microservice/structs"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/fatih/color"
)

const TIMEOUT = 5
const CHECK_INDEX = 5

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
			Callback:          command.Callback,
			Generator:         command.Generator,
			SourceDevice:      command.Device,
			DestinationDevice: command.DestinationDevice,
		}
		statusResponseMap := make(map[string]interface{})

		if err := gateway.SetStatusGateway(&command); err != nil {
			msg := fmt.Sprintf("unable to set gateway for %s: %s", command.Action.Microservice, err.Error())
			log.Printf("%s", color.HiRedString("[error] %s", msg))
			base.PublishError(msg, ei.INTERNAL)
		}

		//build url
		url, err := ReplaceParameters(command.Action.Microservice+command.Action.Endpoint.Path, command.Parameters)
		if err != nil {
			msg := fmt.Sprintf("unable to replace paramaters for %s: %s, aborting command...", command.Action.Name, err.Error())
			log.Printf("%s", color.HiRedString("[error] %s", msg))
			base.PublishError(msg, ei.INTERNAL)
			continue
		}

		log.Printf("[state] sending requqest to %s", url)
		timeout := time.Duration(TIMEOUT * time.Second)
		client := http.Client{Timeout: timeout}
		response, err := client.Get(url)
		if err != nil {
			msg := fmt.Sprintf("unable to complete request to %s for device %s: %s", url, command.Device.Name, err.Error())
			log.Printf("%s", color.HiRedString("[error] %s", msg))
			output.ErrorMessage = &msg //do we want to do this? why not just publish the error here?
			outputs = append(outputs, output)
			continue
		}

		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			msg := fmt.Sprintf("unable to read response from %s for device %s: %s", url, command.Device.Name, err.Error())
			log.Printf("%s", color.HiRedString("[error] %s", msg))
			output.ErrorMessage = &msg
			outputs = append(outputs, output)
			continue
		}

		//check to see if it returned a non 200 response, if so, we need to build the error.
		if response.StatusCode != 200 {
			msg := fmt.Sprintf("non-200 response code: %d, message: %s", response.StatusCode, string(body))
			log.Printf("%s", color.HiRedString("[error] %s", msg))
			base.PublishError(msg, ei.INTERNAL)
			continue
		}

		log.Printf("[state] microservice returned: %s", string(body))

		var status map[string]interface{}
		err = json.Unmarshal(body, &status)
		if err != nil {
			msg := fmt.Sprintf("failed to unmarshal response: %s, microservice returned: %s", command.Device.Name, string(body))
			output.ErrorMessage = &msg
			outputs = append(outputs, output)
			log.Printf("%s", color.HiRedString("[error] %s", msg))
			base.PublishError(msg, ei.INTERNAL)
			continue
		}

		log.Printf("[state] copying data into output")
		for device, object := range status {
			statusResponseMap[device] = object
			//		log.Printf("%s maps to %v", device, object) TODO make this visible with debugging mode
		}

		output.Status = statusResponseMap
		outputs = append(outputs, output) //add the full status response
	}

	//write output to channel
	log.Printf("[state] writing output to channel...")
	if debug.DEBUG {
		for _, output := range outputs {
			log.Printf("outputs from device %v", output.SourceDevice.GetFullName())
			for key, value := range output.Status {
				log.Printf("%s maps to %v", key, value)
			}
		}
	}

	channel <- outputs
	log.Printf("%s", color.HiBlueString("[state] done acquiring statuses from  %s", commands[0].Device.GetFullName()))
	control.Done()
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

//make a GET request given a microservice and endpoint and publishes the results
//returns the state the microservice reports or nothing if the microservice doesn't respond
//publishes a state event or an error
//@pre the parameters have been filled, e.g. the endpoint does not contain ":"
func ExecuteCommand(action base.ActionStructure, command structs.Command, endpoint, requestor string) se.StatusResponse {

	log.Printf("[state] Sending request to %s%s...", command.Microservice, endpoint)

	client := &http.Client{
		Timeout: TIMEOUT * time.Second,
	}
	req, err := http.NewRequest("GET", command.Microservice+endpoint, nil)
	if err != nil {
		return se.StatusResponse{}
	}

	if len(os.Getenv("LOCAL_ENVIRONMENT")) == 0 {
		token, err := bearertoken.GetToken()
		if err != nil {
			return se.StatusResponse{}
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
	}

	resp, err := client.Do(req)

	//if error, record it
	if err != nil {

		errorMessage := fmt.Sprintf("Problem sending request: %s", err.Error())
		log.Printf(errorMessage)
		PublishError(errorMessage, action, requestor)
		return se.StatusResponse{}

	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { //check the response code, if non-200, we need to record and report

		color.Set(color.FgHiRed, color.Bold)
		log.Printf("[error] non-200 response code: %v", resp.StatusCode)
		color.Unset()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {

			color.Set(color.FgHiRed, color.Bold)
			log.Printf("[error] Problem reading the response: %s", err.Error())
			PublishError(err.Error(), action, requestor)
			color.Unset()
		}

		log.Printf("microservice returned: %s", b)
		PublishError(fmt.Sprintf("%s", b), action, requestor)

		return se.StatusResponse{}

	}

	//TODO: we need to find some way to check against the correct response value, just as a further validation

	for _, event := range action.EventLog {

		base.SendEvent(
			event.Type,
			event.EventCause,
			event.Device,
			action.Device.Room.Name,
			action.Device.Building.Shortname,
			event.EventInfoKey,
			event.EventInfoValue,
			event.Requestor,
			false,
		)
	}

	color.Set(color.FgHiGreen, color.Bold)
	log.Printf("[state] Successfully sent command %s to device %s.", action.Action, action.Device.Name)
	color.Unset()

	log.Printf("[state] Unmarshalling status...")

	status := make(map[string]interface{})
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorString := fmt.Sprintf("Could not read response body: %s", err.Error())
		PublishError(errorString, action, requestor)
	}

	err = json.Unmarshal(body, &status)
	if err != nil {
		message := fmt.Sprint("Could not unmarshal response struct: %s", err.Error())
		PublishError(message, action, requestor)
	}
	response := se.StatusResponse{
		SourceDevice:      action.Device,
		DestinationDevice: action.DestinationDevice,
		Generator:         SET_STATE_STATUS_EVALUATORS[action.GeneratingEvaluator],
		Status:            status,
		Callback:          action.Callback,
	}

	return response

}

/*
ReplaceIPAddressEndpoint is a simple helper
*/
func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}

//@pre the endpoint's IP parameter has already been replaced
//replaces parameters in the command endpoint
//@post the endpoint does not contain ':'
func ReplaceParameters(endpoint string, parameters map[string]string) (string, error) {

	log.Printf("[state] replacing formal parameters with actual parameters in %s...", endpoint)

	for k, v := range parameters {
		toReplace := ":" + k
		if !strings.Contains(endpoint, toReplace) {
			msg := fmt.Sprintf("%s not found", toReplace)
			log.Printf("%s", color.HiRedString("[error] %s", msg))
			return "", errors.New(msg)
		}

		endpoint = strings.Replace(endpoint, toReplace, v, -1)
	}

	if strings.Contains(endpoint[CHECK_INDEX:], ":") {
		errorString := "not enough parameters provided for command"
		log.Printf(errorString)
		return "", errors.New(errorString)
	}

	return endpoint, nil
}

func PublishError(message string, action base.ActionStructure, requestor string) {

	log.Printf("[error] Publishing error: %s...", message)
	base.SendEvent(
		ei.ERROR,
		ei.USERINPUT,
		action.Device.GetFullName(),
		action.Device.Room.Name,
		action.Device.Building.Shortname,
		action.Action,
		message,
		requestor,
		true)

}
