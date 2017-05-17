package status

import (
	"log"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

const InputDefaultName = "STATUS_InputDefault"
const InputDefaultCommandName = "STATUS_Input"

type InputDefault struct {
}

func (p *InputDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return room.Devices, nil
}

func (p *InputDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return generateStandardStatusCommand(devices, InputDefaultName, InputDefaultCommandName)
}

func (p *InputDefault) EvaluateResponse(label string, value interface{}, Source accessors.Device, dest DestinationDevice) (string, interface{}, error) {
	log.Printf("Evaluating response: %s, %s in evaluator %v", label, value, BlankedDefaultName)

	//we need to remap the port value to the device name, for this case, that's just the device plugged into that port, as defined in the port mapping

	return label, value, nil
}
