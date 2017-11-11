package statusevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
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

		if device.HasRole("AudioOut") {

			audioDevices = append(audioDevices, device)
		}

		if device.HasRole("DSP") {

			dsps = append(dsps, device)
		}
	}

	//business as usual for audioDevices
	commands, count, err := generateStandardStatusCommand(audioDevices, INPUT_DSP, STATUS_INPUT_DSP)
	if err != nil {
		errorMessage := "Could not generate audio device status commands: " + err.Error()
		log.Printf(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	//validate DSP configuration
	if dsps == nil || len(dsps) != 1 {
		return []StatusCommand{}, 0, errors.New("Invalid DSP configuration detected")
	}

	dsp := dsps[0]

	//get switcher associated with DSP

	switchers, err := dbo.GetDevicesByBuildingAndRoomAndRole(dsp.Building.Shortname, dsp.Room.Name, "VideoSwitcher")
	if err != nil {
		errorMessage := "Could not get video switcher in building: " + dsp.Building.Shortname + ", room: " + dsp.Room.Name + "" + err.Error()
		log.Printf(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	//validate number of switchers
	if switchers == nil || len(switchers) != 1 {
		return []StatusCommand{}, 0, errors.New("Invalid video switcher configuration detected")
	}

	for _, port := range switchers[0].Ports {

		if port.Destination == dsp.Name { //found port configuration, issue command to switcher

			parameters := make(map[string]string)
			parameters["address"] = switchers[0].Address
			//split on ':' and take the second field
			realPorts := strings.Split(port.Name, ":")
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

func (p *InputDSP) EvaluateResponse(label string, value interface{}, source structs.Device, destination base.DestinationDevice) (string, interface{}, error) {
	for _, port := range destination.Ports {

		valueString, ok := value.(string)
		if ok && port.Name == valueString {
			value = port.Source
		}
	}

	return label, value, nil
}
