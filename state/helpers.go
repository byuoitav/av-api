package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/gateway"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
	"github.com/fatih/color"
)

//TIMEOUT is the duration constant to wait before timing out.
const TIMEOUT = 5

// const LOCAL_CHECK_INDEX = 21
// const GATEWAY_CHECK_INDEX = 5

//builds a Status object corresponding to a device and writes it to the channel
func issueCommands(commands []se.StatusCommand, channel chan []se.StatusResponse, control *sync.WaitGroup) {
	//final output
	outputs := []se.StatusResponse{}

	//iterate over list of StatusCommands
	//TODO:make sure devices can handle rapid-fire API requests
	for _, command := range commands {

		log.L.Infof("[state] issuing command: %s against device %s, destination device: %s, parameters: %v", command.Action.ID, command.Device.ID, command.DestinationDevice.Device.ID, command.Parameters)

		output := se.StatusResponse{
			Callback:          command.Callback,
			Generator:         command.Generator,
			SourceDevice:      command.Device,
			DestinationDevice: command.DestinationDevice,
		}
		statusResponseMap := make(map[string]interface{})

		//build url
		endpoint, err := ReplaceParameters(command.Action.Endpoint.Path, command.Parameters)
		if err != nil {
			msg := fmt.Sprintf("unable to replace parameters for %s: %s", command.Action.ID, err.Error())
			log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
			base.PublishError(msg, events.Error, command.Device.ID)
			continue
		}

		address := fmt.Sprintf("%s%s", command.Action.Microservice.Address, endpoint)

		url, err := gateway.SetStatusGateway(address, command.Device)
		if err != nil {
			msg := fmt.Sprintf("unable to set gateway for %s: %s", command.Action.ID, err.Error())
			log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
			base.PublishError(msg, events.Error, command.Device.ID)
			continue
		}

		log.L.Infof("%s", color.HiBlueString("[state] sending request to %s", url))
		timeout := time.Duration(TIMEOUT * time.Second)
		client := http.Client{Timeout: timeout}
		response, err := client.Get(url)
		if err != nil {
			msg := fmt.Sprintf("unable to complete request to %s for device %s: %s", url, command.Device.Name, err.Error())
			log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
			output.ErrorMessage = &msg //do we want to do this? why not just publish the error here?
			outputs = append(outputs, output)
			continue
		}

		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			msg := fmt.Sprintf("unable to read response from %s for device %s: %s", url, command.Device.Name, err.Error())
			log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
			output.ErrorMessage = &msg
			outputs = append(outputs, output)
			continue
		}

		//check to see if it returned a non 200 response, if so, we need to build the error.
		if response.StatusCode != 200 {
			msg := fmt.Sprintf("non-200 response code: %d, message: %s", response.StatusCode, string(body))
			log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
			base.PublishError(msg, events.Error, command.Device.ID)
			continue
		}

		log.L.Infof("[state] microservice returned: %s for action %s against device %s", string(body), command.Action.ID, command.Device.ID, string(body))

		var status map[string]interface{}
		err = json.Unmarshal(body, &status)
		if err != nil {
			msg := fmt.Sprintf("failed to unmarshal response: %s, microservice returned: %s", command.Device.Name, string(body))
			output.ErrorMessage = &msg
			outputs = append(outputs, output)
			log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
			base.PublishError(msg, events.Error, command.Device.ID)
			continue
		}

		log.L.Info("[state] copying data into output")
		for device, object := range status {
			statusResponseMap[device] = object
			//		base.Log("%s maps to %v", device, object) TODO make this visible with debugging mode
		}

		output.Status = statusResponseMap
		outputs = append(outputs, output) //add the full status response
	}

	//write output to channel
	log.L.Info("[state] writing output to channel...")
	for _, output := range outputs {
		log.L.Infof("outputs from device %v", output.SourceDevice.ID)
		for key, value := range output.Status {
			log.L.Infof("%s maps to %v", key, value)
		}
	}

	channel <- outputs
	log.L.Infof("%s", color.HiBlueString("[state] done acquiring statuses from  %s", commands[0].Device.ID))
	control.Done()
}

func processAudioDevice(device se.Status) (base.AudioDevice, error) {
	log.L.Infof("Adding audio device: %s", device.DestinationDevice.Name)
	log.L.Infof("Status map: %v", device.Status)

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
			log.L.Errorf("%s", color.HiRedString("[error] volume type assertion failed for %v", volume))
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

	log.L.Infof("Adding display: %s", device.DestinationDevice.Name)

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

//ExecuteCommand makes a GET request given a microservice and endpoint and publishes the results
//returns the state the microservice reports or nothing if the microservice doesn't respond
//publishes a state event or an error
//@pre the parameters have been filled, e.g. the endpoint does not contain ":"
func ExecuteCommand(action base.ActionStructure, command structs.Command, endpoint, requestor string) se.StatusResponse {
	client := &http.Client{
		Timeout: TIMEOUT * time.Second,
	}
	//set the gateway
	url, err := gateway.SetGateway(command.Microservice.Address+endpoint, action.Device)
	if err != nil {
		log.L.Warnf("Couldn't find gated device: %v", err.Error())
		msg := fmt.Sprintf("unable to reach gated device: %s: %s", action.Device.Name, err.Error())
		return se.StatusResponse{ErrorMessage: &msg}
	}

	log.L.Infof("%s", color.HiBlueString("[state] sending request to %s...", url))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		msg := err.Error()
		return se.StatusResponse{ErrorMessage: &msg}
	}

	if len(os.Getenv("LOCAL_ENVIRONMENT")) == 0 {
		token, err := bearertoken.GetToken()
		if err != nil {
			return se.StatusResponse{}
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
	}

	resp, err := client.Do(req)
	if err != nil { //record any errors
		msg := fmt.Sprintf("error sending request: %s", err.Error())
		log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
		PublishError(msg, action, requestor)
		return se.StatusResponse{ErrorMessage: &msg}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { //check the response code, if non-200, we need to record and report

		log.L.Errorf("%s", color.HiRedString("[error] non-200 response code: %v", resp.StatusCode))

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.L.Errorf("%s", color.HiRedString("[error] problem reading the response: %s", err.Error()))
		}

		log.L.Errorf("%s", color.HiRedString("[error] microservice returned: %s for action %s against device %s.", b, action.Action, action.Device.Name))
		PublishError(fmt.Sprintf("%s", b), action, requestor)

		return se.StatusResponse{}

	}

	//TODO: we need to find some way to check against the correct response value, just as a further validation
	for _, event := range action.EventLog {
		base.SendEvent(event)
	}

	log.L.Infof("%s", color.HiGreenString("[state] sent command %s to device %s.", action.Action, action.Device.Name))
	status := make(map[string]interface{})
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorString := fmt.Sprintf("could not read response body: %s", err.Error())
		PublishError(errorString, action, requestor)
	}

	err = json.Unmarshal(body, &status)
	if err != nil {
		message := fmt.Sprintf("could not unmarshal response struct: %s", err.Error())
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

//ReplaceParameters replaces parameters in the command endpoint
//@pre the endpoint's IP parameter has already been replaced
//@post the endpoint does not contain ':'
func ReplaceParameters(endpoint string, parameters map[string]string) (string, error) {

	if parameters == nil { //should I keep this check?
		return endpoint, nil
	}

	for k, v := range parameters {
		toReplace := ":" + k
		if !strings.Contains(endpoint, toReplace) {
			msg := fmt.Sprintf("%s not found", toReplace)
			log.L.Errorf("%s", color.HiRedString("[error] %s", msg))
			return "", errors.New(msg)
		}

		endpoint = strings.Replace(endpoint, toReplace, v, -1)
	}

	index := strings.IndexRune(endpoint, ':')

	if index >= 0 {

		if strings.Contains(endpoint[index+1:], ":") {
			errorString := fmt.Sprintf("not enough parameters provided for command: %s", endpoint) //TODO change this setup?
			return "", errors.New(errorString)
		}
	}

	return endpoint, nil
}

//PublishError creates an Event based on the error message and ActionStructure information, and then sends it to the event messaging system.
func PublishError(message string, action base.ActionStructure, requestor string) {

	log.L.Errorf("[error] publishing error: %s...", message)

	e := events.Event{
		TargetDevice: events.GenerateBasicDeviceInfo(action.Device.ID),
		AffectedRoom: events.GenerateBasicRoomInfo(action.Device.GetDeviceRoomID()),
		Key:          action.Action,
		Value:        message,
		User:         requestor,
	}

	e.AddToTags(events.Error, events.UserGenerated)

	base.SendEvent(e)
}
