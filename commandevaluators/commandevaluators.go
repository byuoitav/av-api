package commandevaluators

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
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
	GetIncompatableCommands() []string
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
func ExecuteActions(actions []base.ActionStructure) (status []CommandExecutionReporting, err error) {
	for _, a := range actions {
		if a.Overridden {
			log.Printf("Action %s on device %s have been overriden. Continuing.",
				a.Action, a.Device.Name)
			continue
		}

		has, cmd := checkCommands(a.Device.Commands, a.Action)
		if !has {
			errorStr := "There was an error retrieving the command " + a.Action +
				" for device " + a.Device.Name
			log.Printf("%s", errorStr)
			err = errors.New(errorStr)
			return
		}

		//replace the address
		endpoint := ReplaceIPAddressEndpoint(cmd.Endpoint.Path, a.Device.Address)

		//go through and replace the parameters with the parameters in the actions
		for k, v := range a.Parameters {
			toReplace := ":" + k
			if !strings.Contains(endpoint, toReplace) {
				errorString := "The parameter " + toReplace + " was not found in the command " +
					cmd.Name + " for device " + a.Device.Name + "."

				log.Printf("%s", errorString)

				err = errors.New(errorString)
				return
			}

			endpoint = strings.Replace(endpoint, toReplace, v, -1)
		}

		if strings.Contains(endpoint, ":") {
			errorString := "Not enough parameters provided for command " +
				cmd.Name + " for device " + a.Device.Name + "." + " After evaluation " +
				"endpoint was " + endpoint + "."

			log.Printf("%s", errorString)

			err = errors.New(errorString)
			return
		}

		//Execute the command.
		client := &http.Client{}

		token, er := bearertoken.GetToken()
		if er != nil {
			err = er
			return
		}

		req, _ := http.NewRequest("GET", cmd.Microservice+endpoint, nil)
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, er := client.Do(req)

		toReport := base.Event{
			Event:    a.Action,
			Building: a.Device.Building.Shortname,
			Room:     a.Device.Room.Name,
			Device:   a.Device.Name,
		}

		//if error, record it
		if er != nil {
			log.Printf("ERROR: %s. Continuing.", er.Error())

			toReport.ResponseCode = resp.StatusCode
			toReport.Success = true
			base.ReportToELK(toReport)

			status = append(status, CommandExecutionReporting{
				Success: false,
				Action:  a.Action,
				Device:  a.Device.Name,
				Err:     er.Error(),
			})
		} else {
			log.Printf("Successfully sent command %s to device %s.", a.Action, a.Device.Name)

			toReport.Success = false
			base.ReportToELK(toReport)

			status = append(status, CommandExecutionReporting{
				Success: true,
				Action:  a.Action,
				Device:  a.Device.Name,
				Err:     "",
			})
		}

	}

	return
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
		CommandMap["ChangeInputDefault"] = &ChangeInputDefault{}
		CommandMap["BlankDisplayDefault"] = &BlankDisplayDefault{}

		commandMapInitialized = true
	}

	return CommandMap
}
