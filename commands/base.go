package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//ActionStructure is the internal struct we use to pass commands around once
//they've been evaluated.
type ActionStructure struct {
	Action         string           `json:"action"`
	Device         accessors.Device `json:"device"`
	Parameters     []string         `json:"parameters"`
	DeviceSpecific bool             `json:"deviceSpecific, omitempty"`
	Overridden     bool             `json:"overridden"`
}

//CommandExecutionReporting is a struct we use to keep track of command execution
//for reporting to the user.
type CommandExecutionReporting struct {
	Success bool   `json:"success"`
	Action  string `json:"action"`
	Device  string `json:"device"`
	Err     string `json:"error"`
}

/*
CommandEvaluation is an interface that must be implemented for each command to be
evaluated.
*/
type CommandEvaluation interface {
	/*
		 	Evalute takes a public room struct, scans the struct and builds any needed
			actions based on the contents of the struct.
	*/
	Evaluate(base.PublicRoom) ([]ActionStructure, error)
	/*
		  Validate takes a set of action structures (for the command) and validates
			that the device and parameter are valid for the comamnd.
	*/
	Validate([]ActionStructure) error
	/*
			   GetIncompatableActions returns A list of commands that are incompatable
		     with this one (i.e. 'standby' and 'power on', or 'mute' and 'volume up')
	*/
	GetIncompatableActions() []string
}

//CommandMap is a singleton that
//maps known commands to their evaluation structure. init will return a pointer to this.
var CommandMap = make(map[string]CommandEvaluation)
var commandMapInitialized = false

func getDevice(devs []accessors.Device, d string, room string, building string) (dev accessors.Device, err error) {
	for i, curDevice := range devs {
		if checkDevicesEqual(curDevice, d, room, building) {
			dev = devs[i]
			return
		}
	}
	var device accessors.Device

	device, err = dbo.GetDeviceByName(room, building, d)
	if err != nil {
		return
	}
	dev = device
	return
}

//Checks an action list to see if it has a device (by name, room, and building) already in it,
//if so, it returns the index of the device, if not -1.
func checkActionListForDevice(a []ActionStructure, d string, room string, building string) (index int) {
	for i, curDevice := range a {
		if checkDevicesEqual(curDevice.Device, d, room, building) {
			return i
		}
	}
	return -1
}

//ExecuteActions carries out the actions defined in the struct
func ExecuteActions(actions []ActionStructure) (status []CommandExecutionReporting, err error) {
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
		for i := range a.Parameters {
			indx := strings.Index(endpoint, ":")
			if indx == -1 {
				errorString := "Not enough parameter locations in endpoint string for command " +
					cmd.Name + " for device " + a.Device.Name + ". Expected " + string(len(a.Parameters))

				log.Printf("%s", errorString)

				err = errors.New(errorString)
				return
			}
			end := strings.Index(endpoint[:indx], "/")
			if end == -1 {
				endpoint = endpoint[:indx] + a.Parameters[i]
			} else {
				endpoint = endpoint[:indx] + a.Parameters[i] + endpoint[end:]
			}
		}

		//Execute the command.
		_, er := http.Get(cmd.Microservice + endpoint)

		//iff error, record it
		if er != nil {
			log.Printf("ERROR: %s. Continuing.", er.Error())
			status = append(status, CommandExecutionReporting{
				Success: false,
				Action:  a.Action,
				Device:  a.Device.Name,
				Err:     er.Error(),
			})
		} else {
			log.Printf("Successfully sent command %s to device %s.", a.Action, a.Device.Name)
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

//ReconcileActions checks for incompatable actions within the structure passed in.
func ReconcileActions(actions *[]ActionStructure) (err error) {
	log.Printf("Reconciling actions.")
	deviceActionMap := make(map[int][]ActionStructure)

	log.Printf("Generating device action set.")
	//generate a set of actions for each device.
	for _, a := range *actions {
		if _, has := deviceActionMap[a.Device.ID]; has {
			deviceActionMap[a.Device.ID] = append(deviceActionMap[a.Device.ID], a)
		} else {
			deviceActionMap[a.Device.ID] = []ActionStructure{a}
		}
	}

	log.Printf("Checking for incompatable actions.")
	for devID, v := range deviceActionMap {
		//for each device, construct set of actions
		actionsForEvaluation := make(map[string]ActionStructure)
		incompat := make(map[string]ActionStructure)

		for i := 0; i < len(v); i++ {
			actionsForEvaluation[v[i].Action] = v[i]
			//for each device, construct set of incompatable actions
			//Value is the action that generated the incompatable action.
			incompatableActions := CommandMap[v[i].Action].GetIncompatableActions()
			for _, incompatAct := range incompatableActions {
				incompat[incompatAct] = v[i]
			}
		}

		//find intersection of sets.

		//baseAction is the actionStructure generating the action (for cur action)
		//incompatableBaseAction is the actionStructure that generated the incompatable action.
		for curAction, baseAction := range actionsForEvaluation {
			fmt.Printf("%v: %+v\n", curAction, baseAction)
			if baseAction.Overridden {
				continue
			}

			for incompatableAction, incompatableBaseAction := range incompat {
				if incompatableBaseAction.Overridden {
					continue
				}

				if strings.EqualFold(curAction, incompatableAction) {
					log.Printf("%s is incompatable with %s.", incompatableAction, incompatableBaseAction.Action)
					// if one of them is room wide and the other is not override the room-wide
					// action.

					if !baseAction.DeviceSpecific && incompatableBaseAction.DeviceSpecific {
						log.Printf("%s is a device specific command. Overriding %s in favor of device-specific command %s.",
							incompatableBaseAction.Action, baseAction.Action, incompatableBaseAction.Action)
						baseAction.Overridden = true

					} else if baseAction.DeviceSpecific && !incompatableBaseAction.DeviceSpecific {
						log.Printf("%s is a device specific command. Overriding %s in favor of device-specific command %s.",
							baseAction.Action, incompatableBaseAction.Action, baseAction.Action)
						/*
							We have to mark it as incompatable in three places, incompat (so it doesn't cause problems for other commands),
							actionsForEvaluation for the same reason, and in actions so the action won't get sent. We were using pointers, but
							for simplicity and readability in code, we pulled them out.

							Don't judge. :D
						*/
						//markAsOverridden(incompatableBaseAction, &incompat, &actionsForEvaluation, actions)
						incompatableBaseAction.Overridden = true
					} else {
						errorString := incompatableAction + " is an incompatable action with " + incompatableBaseAction.Action + " for device with ID: " +
							string(devID)
						log.Printf("%s", errorString)
						err = errors.New(errorString)
						return
					}
				}
			}
		}
	}

	b, _ := json.Marshal(&actions)
	fmt.Printf("%s", b)
	log.Printf("Done.")
	return
}

func checkDevicesEqual(dev accessors.Device, name string, room string, building string) bool {
	return strings.EqualFold(dev.Name, name) &&
		strings.EqualFold(dev.Room.Name, room) &&
		strings.EqualFold(dev.Building.Shortname, building)
}

func checkCommands(commands []accessors.Command, commandName string) (bool, accessors.Command) {
	for _, c := range commands {
		if strings.EqualFold(c.Name, commandName) {
			return true, c
		}
	}
	return false, accessors.Command{}
}

//Init adds the commands to the commandMap here.
func Init() *map[string]CommandEvaluation {
	if !commandMapInitialized {
		CommandMap["PowerOn"] = &PowerOn{}
		CommandMap["Standby"] = &Standby{}

		commandMapInitialized = true
	}

	return &CommandMap
}

func markAsOverridden(action ActionStructure, structs ...*[]ActionStructure) {
	for i := 0; i < len(structs); i++ {
		for j := 0; j < len(structs[i]); j++ {
			if structs[i][j].equals(action) {
				structs[i][j].Overridden = true
			}
		}
	}
}

/*
ReplaceIPAddressEndpoint is a simple helper
*/
func ReplaceIPAddressEndpoint(path string, address string) string {
	//magic strings
	toReplace := ":address"

	return strings.Replace(path, toReplace, address, -1)

}

func (a *ActionStructure) equals(b ActionStructure) bool {
	return a.Action == b.Action &&
		a.Device.ID == b.Device.ID &&
		a.Device.Address == b.Device.Address &&
		a.DeviceSpecific == b.DeviceSpecific &&
		a.Overridden == b.Overridden && checkStringSliceEqual(a.Parameters, b.Parameters)
}

func checkStringSliceEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
