package statusevaluators

import (
	"errors"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// InputDSPEvaluator is a constant variable for the name of the evaluator.
const InputDSPEvaluator = "STATUS_InputDSP"

// InputDSPCommand is a constant variable for the name of the command.
const InputDSPCommand = "STATUS_Input"

// InputDSP implements the StatusEvaluator struct.
type InputDSP struct{}

// GetDevices returns a list of devices in the given room.
func (p *InputDSP) GetDevices(room structs.Room) ([]structs.Device, error) {

	return room.Devices, nil
}

// GenerateCommands generates a list of commands for the given devices.
func (p *InputDSP) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {

	var audioDevices, dsps []structs.Device
	for _, device := range devices {

		if structs.HasRole(device, "AudioOut") {

			audioDevices = append(audioDevices, device)
		}

		if structs.HasRole(device, "DSP") {

			dsps = append(dsps, device)
		}
	}

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

	switchers, err := db.GetDB().GetDevicesByRoomAndRole(dsp.GetDeviceRoomID(), "VideoSwitcher")
	if err != nil {
		errorMessage := "[statusevals] Could not get video switcher in building: " + dsp.GetDeviceRoomID() + " " + err.Error()
		log.L.Error(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

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

			command := switchers[0].GetCommandByName(InputDSPCommand)

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
func (p *InputDSP) EvaluateResponse(label string, value interface{}, source structs.Device, DestinationDevice base.DestinationDevice) (string, interface{}, error) {
	for _, port := range DestinationDevice.Ports {

		valueString, ok := value.(string)
		if ok && port.ID == valueString {
			value = port.SourceDevice
		}
	}

	return label, value, nil
}
