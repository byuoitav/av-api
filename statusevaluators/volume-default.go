package statusevaluators

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// VolumeDefaultEvaluator is a constant variable for the name of the evaluator.
const VolumeDefaultEvaluator = "STATUS_VolumeDefault"

// VolumeDefaultCommand is a constant variable for the name of the command.
const VolumeDefaultCommand = "STATUS_Volume"

// VolumeDefault implements the StatusEvaluator struct.
type VolumeDefault struct {
}

// GetDevices returns a list of devices in the given room.
func (p *VolumeDefault) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

// GenerateCommands generates a list of commands for the given devices.
func (p *VolumeDefault) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(devices, VolumeDefaultEvaluator, VolumeDefaultCommand)
}

// EvaluateResponse processes the response information that is given.
func (p *VolumeDefault) EvaluateResponse(label string, value interface{}, Source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	log.L.Infof("[statusevals] Evaluating response: %s, %s in evaluator %v", label, value, VolumeDefaultCommand)
	return label, value, nil
}
