package statusevaluators

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// DefaultInputEvaluator is a constant variable for the name of the evaluator.
const DefaultInputEvaluator = "STATUS_InputDefault"

// DefaultInputCommand is a constant variable for the name of the command.
const DefaultInputCommand = "STATUS_Input"

// InputDefault implements the StatusEvaluator struct.
type InputDefault struct {
}

// GetDevices returns a list of devices in the given room.
func (p *InputDefault) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

// GenerateCommands generates a list of commands for the given devices.
func (p *InputDefault) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(devices, DefaultInputEvaluator, DefaultInputCommand)
}

// EvaluateResponse processes the response information that is given.
func (p *InputDefault) EvaluateResponse(label string, value interface{}, source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	log.L.Infof("[statusevals] Evaluating response: %s, %s in evaluator %v", label, value, DefaultInputEvaluator)

	//we need to remap the port value to the device name, for this case, that's just the device plugged into that port, as defined in the port mapping

	for _, port := range dest.Ports {

		valueString, ok := value.(string)
		if ok && port.ID == valueString {

			value = port.SourceDevice

		}

	}

	return label, value, nil
}
