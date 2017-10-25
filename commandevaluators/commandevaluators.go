package commandevaluators

import (
	"encoding/json"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
)

/*
CommandEvaluator is an interface that must be implemented for each command to be
evaluated.
*/
type CommandEvaluator interface {
	/*
		 	Evalute takes a public room struct, scans the struct and builds any needed
			actions based on the contents of the struct. It also returns the number of status
			that will be needed
	*/
	Evaluate(base.PublicRoom, string) ([]base.ActionStructure, int, error)
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

func getDevice(devs []structs.Device, d string, room string, building string) (dev structs.Device, err error) {
	for i, curDevice := range devs {
		if checkDevicesEqual(curDevice, d, room, building) {
			dev = devs[i]
			return
		}
	}
	var device structs.Device

	device, err = dbo.GetDeviceByName(building, room, d)
	if err != nil {
		return
	}
	dev = device
	return
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

//soft singleton command map
var EVALUATORS = map[string]CommandEvaluator{
	"PowerOnDefault":                &PowerOnDefault{},
	"StandbyDefault":                &StandbyDefault{},
	"ChangeVideoInputDefault":       &ChangeVideoInputDefault{},
	"ChangeAudioInputDefault":       &ChangeAudioInputDefault{},
	"ChangeVideoInputVideoSwitcher": &ChangeVideoInputVideoSwitcher{},
	"BlankDisplayDefault":           &BlankDisplayDefault{},
	"UnBlankDisplayDefault":         &UnBlankDisplayDefault{},
	"MuteDefault":                   &MuteDefault{},
	"UnMuteDefault":                 &UnMuteDefault{},
	"SetVolumeDefault":              &SetVolumeDefault{},
	"SetVolumeTecLite":              &SetVolumeTecLite{},
	"MuteDSP":                       &MuteDSP{},
	"UnmuteDSP":                     &UnMuteDSP{},
	"SetVolumeDSP":                  &SetVolumeDSP{},
}
