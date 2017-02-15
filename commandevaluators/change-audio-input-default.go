package commandevaluators

import "github.com/byuoitav/av-api/base"

//ChangeAudioInputDefault f
type ChangeAudioInputDefault struct {
}

//Evaluate f
func (p *ChangeAudioInputDefault) Evaluate(room base.PublicRoom) (actions []base.ActionStructure, err error) {
	//RoomWideSetAudioInput
	if len(room.CurrentAudioInput) > 0 { // Check if the user sent a PUT body changing the current audio input
		var tempActions []base.ActionStructure

		//generate action
		tempActions, err = generateChangeInputByRole(
			"AudioOut",
			room.CurrentVideoInput,
			room.Room,
			room.Building,
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

		var action base.ActionStructure

		action, err = generateChangeInputByDevice(d.Device, room.Room, room.Building)
		if err != nil {
			return
		}
		actions = append(actions, action)
	}
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
