package statusevaluators

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

const VolumeDefaultName = "STATUS_VolumeDefault"
const VolumeDefaultCommandName = "STATUS_Volume"

type VolumeDefault struct {
}

func (p *VolumeDefault) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

func (p *VolumeDefault) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(devices, VolumeDefaultName, VolumeDefaultCommandName)
}

func (p *VolumeDefault) EvaluateResponse(label string, value interface{}, Source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	base.Log("Evaluating response: %s, %s in evaluator %v", label, value, VolumeDefaultCommandName)
	return label, value, nil
}
