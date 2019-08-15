package commandevaluators

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

//ChangeAudioInputDefault implements the CommandEvaluation struct.
type ChangeAudioInputDefault struct {
}

//Evaluate verifies the information for a ChangeAudioInputDefault object and generates a list of actions based on the command.
func (p *ChangeAudioInputDefault) Evaluate(dbRoom structs.Room, room base.PublicRoom, requestor string) (actions []base.ActionStructure, count int, err error) {
	count = 0

	if len(room.CurrentAudioInput) > 0 { // Check if the user sent a PUT body changing the current audio input

		var tempActions []base.ActionStructure

		//generate action
		tempActions, err = generateChangeInputByRole(
			dbRoom,
			"AudioOut",
			room.CurrentVideoInput,
			room.Room,
			room.Building,
			"ChangeAudioInputDefault",
			requestor,
		)

		if err != nil {
			return
		}

		actions = append(actions, tempActions...)
	}

	//AudioDevice
	for _, d := range room.AudioDevices { // Loop through the audio devices array (potentially) passed in the user's PUT body
		if len(d.Input) < 1 {
			continue
		}

		var action []base.ActionStructure

		action, err = generateChangeInputByDevice(dbRoom, d.Device, room.Room, room.Building, "ChangeAudioInputDefault", requestor)
		if err != nil {
			return
		}
		actions = append(actions, action...)
	}
	count = len(actions)
	return
}

//Validate fulfills the Fulfill requirement on the command interface
func (p *ChangeAudioInputDefault) Validate(action base.ActionStructure) (err error) {
	return nil
}

//GetIncompatibleCommands keeps track of actions that are incompatable (on the same device)
func (p *ChangeAudioInputDefault) GetIncompatibleCommands() (incompatableActions []string) {
	return
}
