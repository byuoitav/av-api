package statusevaluators

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/byuoitav/common/status"

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

// GenerateCommands generates a list of commands for the given devices.
func (p *InputDefault) GenerateCommands(room structs.Room) ([]StatusCommand, int, error) {
	return generateStandardStatusCommand(room.Devices, DefaultInputEvaluator, DefaultInputCommand)
}

// EvaluateResponse processes the response information that is given.
func (p *InputDefault) EvaluateResponse(room structs.Room, label string, value interface{}, source structs.Device, dest base.DestinationDevice) (string, interface{}, error) {
	log.L.Infof("[statusevals] Evaluating response: %s, %s in evaluator %v", label, value, DefaultInputEvaluator)

	//we need to remap the port value to the device name, for this case, that's just the device plugged into that port, as defined in the port mapping
	valueString, ok := value.(string)
	if !ok {
		return "", nil, fmt.Errorf("incorrect type of response (%v). Expected %s; but got %s", valueString, reflect.TypeOf(""), reflect.TypeOf(value))
	}

	var inputID string

	for _, port := range dest.Ports {
		if strings.EqualFold(port.ID, valueString) {
			inputID = port.SourceDevice
			break
		}
	}

	if len(inputID) == 0 {
		return "", nil, fmt.Errorf("missing port of device: %s", valueString)
	}

	// match the inputID from the port to a device in the db, and return that devices' name
	device, err := db.GetDB().GetDevice(inputID)

	inputValue := device.Name

	if device.HasRole("STB-Stream-Player") {
		resp, err := http.Get(fmt.Sprintf("http://%s:8032/stream", device.Address))
		if err == nil {
			body, _ := ioutil.ReadAll(resp.Body)
			var input status.Input
			err = json.Unmarshal(body, &input)
			if err != nil {
			}
			inputValue = inputValue + "|" + input.Input
		}
	}

	return label, inputValue, err
}
