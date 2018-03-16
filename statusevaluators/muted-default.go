package statusevaluators

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/configuration-database-microservice/structs"
)

const MutedDefaultName = "STATUS_MutedDefault"
const MutedDefaultCommandName = "STATUS_Muted"

type MutedDefault struct {
}

func (p *MutedDefault) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

func (p *MutedDefault) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(devices, MutedDefaultName, MutedDefaultCommandName)
}

func (p *MutedDefault) EvaluateResponse(label string, value interface{}, Source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	base.Log("Evaluating response: %s, %s in evaluator %v", label, value, MutedDefaultCommandName)
	return label, value, nil
}
