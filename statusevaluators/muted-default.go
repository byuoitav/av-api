package statusevaluators

import (
	"log"

	"github.com/byuoitav/configuration-database-microservice/structs"
)

const MutedDefaultName = "STATUS_MutedDefault"
const MutedDefaultCommandName = "STATUS_Muted"

type MutedDefault struct {
}

func (p *MutedDefault) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

func (p *MutedDefault) GenerateCommands(devices []structs.Device) ([]StatusCommand, error) {
	return generateStandardStatusCommand(devices, MutedDefaultName, MutedDefaultCommandName)
}

func (p *MutedDefault) EvaluateResponse(label string, value interface{}, Source structs.Device, dest DestinationDevice) (string, interface{}, error) {
	log.Printf("Evaluating response: %s, %s in evaluator %v", label, value, MutedDefaultCommandName)
	return label, value, nil
}
