package state

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

//make a GET request given a microservice and endpoint and publishes the results
//returns the state the microservice reports or nothing if the microservice doesn't respond
//publishes a state event or an error
//@pre the parameters have been filled, e.g. the endpoint does not contain ":"
func ExecuteCommand(action base.ActionStructure, command structs.Command, endpoint string) interface{} {

	log.Printf("Sending request to %s%s...", command.Microservice, endpoint)

	client := &http.Client{
		Timeout: TIMEOUT * time.Second,
	}
	req, err := http.NewRequest("GET", command.Microservice+endpoint, nil)
	if err != nil {
		return nil
	}

	if len(os.Getenv("LOCAL_ENVIRONMENT")) == 0 {
		token, err := bearertoken.GetToken()
		if err != nil {
			return nil
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	//if error, record it
	if err != nil {

		errorMessage := fmt.Sprintf("Problem sending request: %s", err.Error())
		log.Printf(errorMessage)
		PublishError(errorMessage, action, command)
		return nil

	} else if resp.StatusCode != 200 { //check the response code, if non-200, we need to record and report

		log.Printf("[error] non-200 response code: %s", resp.StatusCode)

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {

			log.Printf("Problem reading the response: %v", err.Error())
			PublishError(err.Error(), action, command)
		}

		log.Printf("microservice returned: %v", b)
		PublishError(fmt.Sprintf("%s", b), action, command)

		return nil

	} else {

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
				false,
			)
		}

		log.Printf("Successfully sent command %s to device %s.", action.Action, action.Device.Name)
		//unmarshal status
		return resp.Body

	}

}

//@pre the endpoint's IP parameter has already been replaced
//replaces parameters in the command endpoint
//@post the endpoint does not contain ':'
func ReplaceParameters(endpoint string, parameters map[string]string) (string, error) {

	log.Printf("Replacing formal parameters with actual parameters...")

	for k, v := range parameters {
		toReplace := ":" + k
		if !strings.Contains(endpoint, toReplace) {
			errorString := "parameter not found"
			log.Printf(errorString)
			return "", errors.New(errorString)
		}

		endpoint = strings.Replace(endpoint, toReplace, v, -1)
	}

	if strings.Contains(endpoint, ":") {
		errorString := "not enough parameters provided for command"
		log.Printf(errorString)
		return "", errors.New(errorString)
	}

	return endpoint, nil
}

func PublishError(message string, action base.ActionStructure, command structs.Command) {

	log.Printf("Publishing error: %s...", message)
	base.SendEvent(
		eventinfrastructure.ERROR,
		eventinfrastructure.USERINPUT,
		action.Device.GetFullName(),
		action.Device.Room.Name,
		action.Device.Building.Shortname,
		command.Name,
		message,
		true)

}
