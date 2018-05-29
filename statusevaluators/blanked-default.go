package statusevaluators

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// BlankedDefaultEvaluator is a constant variable for the name of the evaluator.
const BlankedDefaultEvaluator = "STATUS_BlankedDefault"

// BlankedDefaultCommand is a constant variable for the name of the command.
const BlankedDefaultCommand = "STATUS_Blanked"

// BlankedDefault implements the StatusEvaluator struct.
type BlankedDefault struct {
}

// GetDevices returns a list of devices in the given room.
func (p *BlankedDefault) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

// GenerateCommands generates a list of commands for the given devices.
func (p *BlankedDefault) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(devices, BlankedDefaultEvaluator, BlankedDefaultCommand)
}

// EvaluateResponse is supposed to evaluate the response...but it seems like it's just logging a statement...
func (p *BlankedDefault) EvaluateResponse(label string, value interface{}, Source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	log.L.Infof("[statusevals] Evaluating response: %s, %s in evaluator %v", label, value, BlankedDefaultEvaluator)
	return label, value, nil
}
