package statusevaluators

import (
	"errors"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// InputDSPEvaluator is a constant variable for the name of the evaluator.
const InputDSPEvaluator = "STATUS_InputDSP"

// InputDSPCommand is a constant variable for the name of the command.
const InputDSPCommand = "STATUS_Input"

// InputDSP implements the StatusEvaluator struct.
type InputDSP struct{}

// GenerateCommands generates a list of commands for the given devices.
func (p *InputDSP) GenerateCommands(room structs.Room) ([]StatusCommand, int, error) {

	audioDevices := FilterDevicesByRole(room.Devices, "AudioOut")
	dsps := FilterDevicesByRole(room.Devices, "DSP")

	//business as usual for audioDevices
	commands, count, err := generateStandardStatusCommand(audioDevices, InputDSPEvaluator, InputDSPCommand)
	if err != nil {
		errorMessage := "[statusevals] Could not generate audio device status commands: " + err.Error()
		log.L.Error(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	//validate DSP configuration
	if dsps == nil || len(dsps) != 1 {
		return []StatusCommand{}, 0, errors.New("Invalid DSP configuration detected")
	}

	dsp := dsps[0]

	//get switcher associated with DSP
	switchers := FilterDevicesByRole(room.Devices, "VideoSwitcher")

	//validate number of switchers
	if switchers == nil || len(switchers) != 1 {
		return []StatusCommand{}, 0, errors.New("[statusevals] Invalid video switcher configuration detected")
	}

	for _, port := range switchers[0].Ports {

		if port.DestinationDevice == dsp.ID { //found port configuration, issue command to switcher

			parameters := make(map[string]string)
			parameters["address"] = switchers[0].Address
			//split on ':' and take the second field
			realPorts := strings.Split(port.ID, ":")
			parameters["port"] = realPorts[1]

			destinationDevice := base.DestinationDevice{
				Device:      dsp,
				AudioDevice: true,
			}

			command := switchers[0].GetCommandByID(InputDSPCommand)

			statusCommand := StatusCommand{
				Action:            command,
				Device:            switchers[0],
				Parameters:        parameters,
				DestinationDevice: destinationDevice,
				Generator:         InputDSPEvaluator,
			}

			commands = append(commands, statusCommand)
		}
		count++
	}

	return commands, count, nil
}

// EvaluateResponse processes the response information that is given.
func (p *InputDSP) EvaluateResponse(room structs.Room, label string, value interface{}, source structs.Device, DestinationDevice base.DestinationDevice) (string, interface{}, error) {
	for _, port := range DestinationDevice.Ports {

		valueString, ok := value.(string)
		if ok && port.ID == valueString {
			value = port.SourceDevice
		}
	}

	return label, value, nil
}
