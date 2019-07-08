package statusevaluators

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// MutedDefaultEvaluator is a constant variable for the name of the evaluator.
const MutedDefaultEvaluator = "STATUS_MutedDefault"

// MutedDefaultCommand is a constant variable for the name of the command.
const MutedDefaultCommand = "STATUS_Muted"

// MutedDefault implements the StatusEvaluator struct.
type MutedDefault struct {
}

// GenerateCommands generates a list of commands for the given devices.
func (p *MutedDefault) GenerateCommands(room structs.Room) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(room.Devices, MutedDefaultEvaluator, MutedDefaultCommand)
}

// EvaluateResponse processes the response information that is given.
func (p *MutedDefault) EvaluateResponse(room structs.Room, label string, value interface{}, Source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	log.L.Infof("[statusevals] Evaluating response: %s, %s in evaluator %v", label, value, MutedDefaultCommand)
	return label, value, nil
}
