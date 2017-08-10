package statusevaluators

import (
	"log"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

const MutedDefaultName = "STATUS_MutedDefault"
const MutedDefaultCommandName = "STATUS_Muted"

type MutedDefault struct {
}

func (p *MutedDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return room.Devices, nil
}

func (p *MutedDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return generateStandardStatusCommand(devices, MutedDefaultName, MutedDefaultCommandName)
}

func (p *MutedDefault) EvaluateResponse(label string, value interface{}, Source accessors.Device, dest DestinationDevice) (string, interface{}, error) {
	log.Printf("Evaluating response: %s, %s in evaluator %v", label, value, MutedDefaultCommandName)
	return label, value, nil
}
