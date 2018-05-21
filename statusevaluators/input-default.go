package statusevaluators

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

const DEFAULT_INPUT_EVALUATOR = "STATUS_InputDefault"
const DEFAULT_INPUT_COMMAND = "STATUS_Input"

type InputDefault struct {
}

func (p *InputDefault) GetDevices(room structs.Room) ([]structs.Device, error) {
	return room.Devices, nil
}

func (p *InputDefault) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(devices, DEFAULT_INPUT_EVALUATOR, DEFAULT_INPUT_COMMAND)
}

func (p *InputDefault) EvaluateResponse(label string, value interface{}, source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	base.Log("Evaluating response: %s, %s in evaluator %v", label, value, BlankedDefaultName)

	//we need to remap the port value to the device name, for this case, that's just the device plugged into that port, as defined in the port mapping

	for _, port := range dest.Ports {

		valueString, ok := value.(string)
		if ok && port.ID == valueString {

			value = port.SourceDevice

		}

	}

	return label, value, nil
}
