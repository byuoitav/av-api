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
func ExecuteActions(actions [][]base.ActionStructure) (base.PublicRoom, error) {

	var status base.PublicRoom
	for _, b := range actions {

		for _, a := range b {

			if a.Overridden {
				log.Printf("Action %s on device %s have been overriden. Continuing.",
					a.Action, a.Device.Name)
				continue
			}

			has, cmd := CheckCommands(a.Device.Commands, a.Action)
			if !has {
				errorStr := fmt.Sprintf("Error retrieving the command %s for device %s.", a.Action, a.Device.GetFullName())
				log.Printf(errorStr)
				return base.PublicRoom{}, errors.New(errorStr)
			}

			//replace the address
			endpoint := ReplaceIPAddressEndpoint(cmd.Endpoint.Path, a.Device.Address)

			//go through and replace the parameters with the parameters in the actions
			for k, v := range a.Parameters {
				toReplace := ":" + k
				if !strings.Contains(endpoint, toReplace) {
					errorString := fmt.Sprintf("The parameter %s was not found in the command %s for device %s.", toReplace, cmd.Name, a.Device.GetFullName())
					log.Printf(errorString)
					return base.PublicRoom{}, errors.New(errorString)
				}

				endpoint = strings.Replace(endpoint, toReplace, v, -1)
			}

			if strings.Contains(endpoint, ":") {
				errorString := "Not enough parameters provided for command " +
					cmd.Name + " for device " + a.Device.GetFullName() + "." + " After evaluation " +
					"endpoint was " + endpoint + "."

				log.Printf("%s", errorString)

				return base.PublicRoom{}, errors.New(errorString)
			}

			//Execute the command.
			client := &http.Client{}
			req, err := http.NewRequest("GET", cmd.Microservice+endpoint, nil)
			if err != nil {
				return base.PublicRoom{}, err
			}

			if len(os.Getenv("LOCAL_ENVIRONMENT")) == 0 {
				token, err := bearertoken.GetToken()
				if err != nil {
					return base.PublicRoom{}, err
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

				continue

			} else if resp.StatusCode != 200 { //check the response code, if non-200, we need to record and report

				//check the response code
				log.Printf("Problem with the request, response code; %v", resp.StatusCode)

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

			}
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

//Singleton command map
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
		CommandMap["MuteDSP"] = &MuteDSP{}
		CommandMap["UnmuteDSP"] = &UnMuteDSP{}
		CommandMap["SetVolumeDSP"] = &SetVolumeDSP{}

		commandMapInitialized = true
	}

	return CommandMap
}
