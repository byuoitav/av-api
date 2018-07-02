package statusevaluators

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
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
	valueString, ok := value.(string)
	if !ok {
		return "", nil, errors.New(fmt.Sprintf("incorrect type of response (%v). Expected %s; but got %s", valueString, reflect.TypeOf(""), reflect.TypeOf(value)))
	}

	var inputID string

	for _, port := range dest.Ports {
		if strings.EqualFold(port.ID, valueString) {
			inputID = port.SourceDevice
			break
		}
	}

	if len(inputID) == 0 {
		return "", nil, errors.New(fmt.Sprintf("missing port of device: %s", valueString))
	}

	// match the inputID from the port to a device in the db, and return that devices' name
	device, err := db.GetDB().GetDevice(inputID)
	return label, device.Name, err
}
