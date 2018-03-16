package statusevaluators

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/configuration-database-microservice/structs"
)

const BlankedDefaultName = "STATUS_BlankedDefault"
const BlankedDefaultCommandName = "STATUS_Blanked"

type BlankedDefault struct {
}

func (p *BlankedDefault) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

func (p *BlankedDefault) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(devices, BlankedDefaultName, BlankedDefaultCommandName)
}
func (p *BlankedDefault) EvaluateResponse(label string, value interface{}, Source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	base.Log("Evaluating response: %s, %s in evaluator %v", label, value, BlankedDefaultName)
	return label, value, nil
}
