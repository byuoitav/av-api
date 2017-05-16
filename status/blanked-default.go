package status

import (
	"log"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

const BlankedDefaultName = "STATUS_BlankedDefault"
const BlankedDefaultCommandName = "STATUS_Blanked"

type BlankedDefault struct {
}

func (p *BlankedDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return room.Devices, nil
}

func (p *BlankedDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return generateStandardStatusCommand(devices, BlankedDefaultName, BlankedDefaultCommandName)
}
func (p *BlankedDefault) EvaluateResponse(label string, value interface{}) (string, interface{}, error) {
	log.Printf("Evaluating response: %s, %s in evaluator %v", label, value, BlankedDefaultName)
	return label, value, nil
}
