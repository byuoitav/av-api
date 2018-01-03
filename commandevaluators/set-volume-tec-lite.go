package commandevaluators

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/byuoitav/av-api/base"
)

type SetVolumeTecLite struct {
}

/*
Evaluate fulfils the requirements of the interface.

The Tec-Liet Evaluate calls the SetVoulmeDefault evaluate function - but re-maps the volume levels from 0-100 to 0-65 to be issued to the device.
*/
func (*SetVolumeTecLite) Evaluate(room base.PublicRoom, requestor string) ([]base.ActionStructure, int, error) {
	//call the default set volume to get the list of actions, then go through and remap them
	//for the new volme level
	defaultSetVolume := &SetVolumeDefault{}
	actions, count, err := defaultSetVolume.Evaluate(room, requestor)

	if err != nil {
		return actions, count, err
	}

	for i := range actions {
		oldLevel, err := strconv.Atoi(actions[i].Parameters["level"])
		if err != nil {
			err = errors.New(fmt.Sprintf("Could not parse parameter 'level' for an integer: %s", err.Error()))
			log.Printf("%s", err.Error())
			return actions, count, err
		}
		actions[i].Parameters["level"] = strconv.Itoa(calculateNewLevel(oldLevel, 65))
	}
	return actions, count, nil
}

func calculateNewLevel(level int, max int) int {
	return (level * (max) / 100)
}

//Validate validates that the volume set falls within the max and minimum values
func (*SetVolumeTecLite) Validate(action base.ActionStructure) error {
	maximum := 65
	minimum := 0
	return validateSetVolumeMaxMin(action, maximum, minimum)
}

//GetIncompatibleCommands returns a string array of commands incompatible with setting the volume
func (*SetVolumeTecLite) GetIncompatibleCommands() []string {
	return nil
}
