package status

import (
	"errors"
	"log"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

const PowerDefaultEvaluatorName = "STATUS_PowerDefault"
const PowerDefaultCommand = "STATUS_Power"

type PowerDefault struct {
}

//when querying power, we care about every device
func (p *PowerDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return room.Devices, nil
}

func (p *PowerDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return generateStandardStatusCommand(devices, PowerDefaultEvaluatorName, PowerDefaultCommand)
}

func (p *PowerDefault) EvaluateResponse(label string, value interface{}, Source accessors.Device, dest DestinationDevice) (string, interface{}, error) {
	log.Printf("Evaluating response: %s, %s in evaluator %v", label, value, PowerDefaultEvaluatorName)
	if value == nil {
		return label, value, errors.New("cannot process nil value")
	}

	return label, value, nil
}
