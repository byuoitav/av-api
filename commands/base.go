package commands

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

//ActionStructure is the internal struct we use to pass commands around once
//they've been evaluated.
type ActionStructure struct {
	Action         string            `json:"action"`
	Device         *accessors.Device `json:"device"`
	Parameter      string            `json:"parameter"`
	DeviceSpecific bool              `json:"deviceSpecific, omitempty"`
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

func getDevice(devs []accessors.Device, d string, room string, building string) (dev *accessors.Device, err error) {
	for i, curDevice := range devs {
		if checkDevicesEqual(&curDevice, d, room, building) {
			dev = &devs[i]
			return
		}
	}
	var device accessors.Device

	device, err = dbo.GetDeviceByName(room, building, d)
	if err != nil {
		return
	}
	dev = &device
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
func ExecuteActions(actions []ActionStructure) (err error) {

	return
}

//ReconcileActions checks for incompatable actions within the structure passed in.
func ReconcileActions(actions []ActionStructure) (err error) {
	log.Printf("Reconciling actions.")
	deviceActionMap := make(map[int][]ActionStructure)

	log.Printf("Generating device action set.")
	//generate a set of actions for each device.
	for _, a := range actions {
		if _, has := deviceActionMap[a.Device.ID]; has {
			deviceActionMap[a.Device.ID] = append(deviceActionMap[a.Device.ID], a)
		} else {
			deviceActionMap[a.Device.ID] = []ActionStructure{a}
		}
	}

	log.Printf("Checking for incompatable actions.")
	for devID, v := range deviceActionMap {
		//for each device, construct set of actions
		actions := make(map[string]bool)
		incompat := make(map[string]string)

		for _, action := range v {
			actions[action.Action] = true
			//for each device, construct set of incompatable actions
			incompatableActions := CommandMap[action.Action].GetIncompatableActions()
			for _, incompatAct := range incompatableActions {
				incompat[incompatAct] = action.Action
			}
		}

		//find intersection of sets.
		for k := range actions {
			for incompatableAction, baseAction := range incompat {
				if strings.EqualFold(k, incompatableAction) {
					errorString := incompatableAction + " is an incompatable action with " + baseAction + " for device with ID: " +
						string(devID)
					log.Printf("%s", errorString)
					err = errors.New(errorString)
					return
				}
			}
		}
	}

	log.Printf("Done.")
	return
}

func checkDevicesEqual(dev *accessors.Device, name string, room string, building string) bool {
	return strings.EqualFold(dev.Name, name) &&
		strings.EqualFold(dev.Room.Name, room) &&
		strings.EqualFold(dev.Building.Shortname, building)
}

func checkCommands(commands []accessors.Command, commandName string) bool {
	for _, c := range commands {
		if strings.EqualFold(c.Name, commandName) {
			return true
		}
	}
	return false
}

//Init adds the commands to the commandMap here.
func Init() *map[string]CommandEvaluation {
	if !commandMapInitialized {
		CommandMap["PowerOn"] = &PowerOn{}

		commandMapInitialized = true
	}

	return &CommandMap
}
