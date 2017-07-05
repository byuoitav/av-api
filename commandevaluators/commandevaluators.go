package commandevaluators

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
	"github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

//CommandExecutionReporting is a struct we use to keep track of command execution
//for reporting to the user.
type CommandExecutionReporting struct {
	Success bool   `json:"success"`
	Action  string `json:"action"`
	Device  string `json:"device"`
	Err     string `json:"error,omitempty"`
}

/*
CommandEvaluator is an interface that must be implemented for each command to be
evaluated.
*/
type CommandEvaluator interface {
	/*
		 	Evalute takes a public room struct, scans the struct and builds any needed
			actions based on the contents of the struct.
	*/
	Evaluate(base.PublicRoom) ([]base.ActionStructure, error)
	/*
		  Validate takes an action structure (for the command) and validates
			that the device and parameter are valid for the command.
	*/
	Validate(base.ActionStructure) error
	/*
			   GetIncompatableActions returns a list of commands that are incompatable
		     with this one (i.e. 'standby' and 'power on', or 'mute' and 'volume up')
	*/
	GetIncompatibleCommands() []string
}

//CommandMap is a singleton that
//maps known commands to their evaluation structure. init will return a pointer to this.
var CommandMap = make(map[string]CommandEvaluator)
var commandMapInitialized = false

func getDevice(devs []accessors.Device, d string, room string, building string) (dev accessors.Device, err error) {
	for i, curDevice := range devs {
		if checkDevicesEqual(curDevice, d, room, building) {
			dev = devs[i]
			return
		}
	}
	var device accessors.Device

	device, err = dbo.GetDeviceByName(building, room, d)
	if err != nil {
		return
	}
	dev = device
	return
}

//ExecuteActions carries out the actions defined in the struct
func ExecuteActions(actions []base.ActionStructure) ([]CommandExecutionReporting, error) {

	status := []CommandExecutionReporting{}

	for _, a := range actions {
		if a.Overridden {
			log.Printf("Action %s on device %s have been overriden. Continuing.",
				a.Action, a.Device.Name)
			continue
		}

		has, cmd := CheckCommands(a.Device.Commands, a.Action)
		if !has {
			errorStr := "There was an error retrieving the command " + a.Action + " for device " + a.Device.GetFullName()
			log.Printf("%s", errorStr)
			err := errors.New(errorStr)
			return []CommandExecutionReporting{}, err
		}

		//replace the address
		endpoint := ReplaceIPAddressEndpoint(cmd.Endpoint.Path, a.Device.Address)

		//go through and replace the parameters with the parameters in the actions
		for k, v := range a.Parameters {
			toReplace := ":" + k
			if !strings.Contains(endpoint, toReplace) {
				errorString := "The parameter " + toReplace + " was not found in the command " +
					cmd.Name + " for device " + a.Device.GetFullName() + "."

				log.Printf("%s", errorString)

				err := errors.New(errorString)
				return []CommandExecutionReporting{}, err
			}

			endpoint = strings.Replace(endpoint, toReplace, v, -1)
		}

		if strings.Contains(endpoint, ":") {
			errorString := "Not enough parameters provided for command " +
				cmd.Name + " for device " + a.Device.GetFullName() + "." + " After evaluation " +
				"endpoint was " + endpoint + "."

			log.Printf("%s", errorString)

			err := errors.New(errorString)
			return []CommandExecutionReporting{}, err
		}

		//Execute the command.
		client := &http.Client{}
		req, err := http.NewRequest("GET", cmd.Microservice+endpoint, nil)
		if err != nil {
			return []CommandExecutionReporting{}, err
		}

		if len(os.Getenv("LOCAL_ENVIRONMENT")) == 0 {
			token, err := bearertoken.GetToken()
			if err != nil {
				return []CommandExecutionReporting{}, err
			}
			req.Header.Set("Authorization", "Bearer "+token.Token)
		}

		resp, err := client.Do(req)
		defer resp.Body.Close()

		//if error, record it
		if err != nil {
			base.SendEvent(
				eventinfrastructure.ERROR,
				eventinfrastructure.USERINPUT,
				a.Device.GetFullName(),
				a.Device.Room.Name,
				a.Device.Building.Shortname,
				cmd.Name,
				err.Error(),
				true)

			log.Printf("ERROR: %s. Continuing.", err.Error())

			status = append(status, CommandExecutionReporting{
				Success: false,
				Action:  a.Action,
				Device:  a.Device.Name,
				Err:     err.Error(),
			})

			continue
		} else if resp.StatusCode != 200 { //check the response code, if non-200, we need to record and report

			//check the response code
			log.Printf("Probalem with the request, response code; %v", resp.StatusCode)

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("There was a problem reading the response: %v", err.Error())

				base.SendEvent(
					eventinfrastructure.ERROR,
					eventinfrastructure.USERINPUT,
					a.Device.GetFullName(),
					a.Device.Room.Name,
					a.Device.Building.Shortname,
					cmd.Name,
					err.Error(),
					true)

				status = append(status, CommandExecutionReporting{
					Success: false,
					Action:  a.Action,
					Device:  a.Device.Name,
					Err:     err.Error(),
				})
				continue
			}

			log.Printf("microservice returned: %v", b)

			//now we report the event
			base.SendEvent(
				eventinfrastructure.ERROR,
				eventinfrastructure.USERINPUT,
				a.Device.GetFullName(),
				a.Device.Room.Name,
				a.Device.Building.Shortname,
				cmd.Name,
				fmt.Sprintf("%s", b),
				true)

			status = append(status, CommandExecutionReporting{
				Success: false,
				Action:  a.Action,
				Device:  a.Device.Name,
				Err:     fmt.Sprintf("%s", b),
			})
			continue
		} else {

			//TODO: we need to find some way to check against the correct response value, just as a further validation

			//Vals := getKeyValueFromCommmand(a)

			for _, event := range a.EventLog {

				base.SendEvent(
					event.Type,
					event.EventCause,
					event.Device,
					a.Device.Room.Name,
					a.Device.Building.Shortname,
					event.EventInfoKey,
					event.EventInfoValue,
					false,
				)
			}

			log.Printf("Successfully sent command %s to device %s.", a.Action, a.Device.Name)

			status = append(status, CommandExecutionReporting{
				Success: true,
				Action:  a.Action,
				Device:  a.Device.Name,
				Err:     "",
			})
		}
	}
	return status, nil
}

func getKeyValueFromCommmand(action base.ActionStructure) []string {
	switch action.Action {
	case "PowerOn":
		return []string{"power", "on"}
	case "Standby":
		return []string{"power", "standby"}
	case "ChangeInput":
		b, _ := json.Marshal(action.Parameters)
		return []string{"input", string(b)}
	case "SetVolume":
		return []string{"volume", action.Parameters["level"]}
	case "BlankDisplay":
		return []string{"blanked", "true"}
	case "UnblankDisplay":
		return []string{"blanked", "false"}
	case "Mute":
		return []string{"Muted", "true"}
	case "UnMute":
		return []string{"Muted", "false"}
	}
	return []string{}
}

/*
ReplaceIPAddressEndpoint is a simple helper
*/
func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}

//Init adds the commands to the commandMap here.
func Init() map[string]CommandEvaluator {
	if !commandMapInitialized {
		CommandMap["PowerOnDefault"] = &PowerOnDefault{}
		CommandMap["StandbyDefault"] = &StandbyDefault{}
		CommandMap["ChangeVideoInputDefault"] = &ChangeVideoInputDefault{}
		CommandMap["ChangeAudioInputDefault"] = &ChangeAudioInputDefault{}
		CommandMap["ChangeVideoInputVideoSwitcher"] = &ChangeVideoInputVideoSwitcher{}
		CommandMap["BlankDisplayDefault"] = &BlankDisplayDefault{}
		CommandMap["UnBlankDisplayDefault"] = &UnBlankDisplayDefault{}
		CommandMap["MuteDefault"] = &MuteDefault{}
		CommandMap["UnMuteDefault"] = &UnMuteDefault{}
		CommandMap["SetVolumeDefault"] = &SetVolumeDefault{}
		CommandMap["SetVolumeDMPS"] = &SetVolumeDMPS{}
		CommandMap["SetVolumeTecLite"] = &SetVolumeTecLite{}
		CommandMap["ChangeVideoInputDMPS"] = &ChangeVideoInputDMPS{}

		commandMapInitialized = true
	}

	return CommandMap
}
