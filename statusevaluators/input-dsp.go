package statusevaluators

import (
	"errors"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/structs"
)

const INPUT_DSP = "STATUS_InputDSP"
const STATUS_INPUT_DSP = "STATUS_Input"

type InputDSP struct{}

func (p *InputDSP) GetDevices(room structs.Room) ([]structs.Device, error) {

	return room.Devices, nil
}

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
	commands, count, err := generateStandardStatusCommand(audioDevices, INPUT_DSP, STATUS_INPUT_DSP)
	if err != nil {
		errorMessage := "Could not generate audio device status commands: " + err.Error()
		base.Log(errorMessage)
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
		errorMessage := "Could not get video switcher in building: " + dsp.GetDeviceRoomID() + " " + err.Error()
		base.Log(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	//validate number of switchers
	if switchers == nil || len(switchers) != 1 {
		return []StatusCommand{}, 0, errors.New("Invalid video switcher configuration detected")
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

			command := switchers[0].GetCommandByName(STATUS_INPUT_DSP)

			statusCommand := StatusCommand{
				Action:            command,
				Device:            switchers[0],
				Parameters:        parameters,
				DestinationDevice: destinationDevice,
				Generator:         INPUT_DSP,
			}

			commands = append(commands, statusCommand)
		}
		count++
	}

	return commands, count, nil
}

func (p *InputDSP) EvaluateResponse(label string, value interface{}, source structs.Device, DestinationDevice base.DestinationDevice) (string, interface{}, error) {
	for _, port := range DestinationDevice.Ports {

		valueString, ok := value.(string)
		if ok && port.ID == valueString {
			value = port.SourceDevice
		}
	}

	return label, value, nil
}
