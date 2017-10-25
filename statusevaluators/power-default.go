package statusevaluators

import (
	"errors"
	"log"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/configuration-database-microservice/structs"
)

const PowerDefaultEvaluatorName = "STATUS_PowerDefault"
const PowerDefaultCommand = "STATUS_Power"

type PowerDefault struct {
}

//when querying power, we care about every device
func (p *PowerDefault) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

func (p *PowerDefault) GenerateCommands(devices []structs.Device) ([]StatusCommand, error) {
	return generateStandardStatusCommand(devices, PowerDefaultEvaluatorName, PowerDefaultCommand)
}

func (p *PowerDefault) EvaluateResponse(label string, value interface{}, Source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	log.Printf("Evaluating response: %s, %s in evaluator %v", label, value, PowerDefaultEvaluatorName)
	if value == nil {
		return label, value, errors.New("cannot process nil value")
	}

	return label, value, nil
}
