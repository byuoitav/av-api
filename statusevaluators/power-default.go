package statusevaluators

import (
	"errors"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// PowerDefaultEvaluator is a constant variable for the name of the evaluator.
const PowerDefaultEvaluator = "STATUS_PowerDefault"

// PowerDefaultCommand is a constant variable for the name of the command.
const PowerDefaultCommand = "STATUS_Power"

// PowerDefault implements the StatusEvaluator struct.
type PowerDefault struct {
}

// GenerateCommands generates a list of commands for the given devices.
func (p *PowerDefault) GenerateCommands(room structs.Room) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(room.Devices, PowerDefaultEvaluator, PowerDefaultCommand)
}

// EvaluateResponse processes the response information that is given
func (p *PowerDefault) EvaluateResponse(room structs.Room, label string, value interface{}, Source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	log.L.Infof("[statusevals] Evaluating response: %s, %s in evaluator %v", label, value, PowerDefaultEvaluator)
	if value == nil {
		return label, value, errors.New("cannot process nil value")
	}

	return label, value, nil
}
