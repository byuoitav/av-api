package status

import (
	"errors"
	"log"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

const INPUT_DSP = "InputDSP"
const STATUS_INPUT_DSP = "STATUS_Input"

type InputDSP struct{}

func (p *InputDSP) GetDevices(room accessors.Room) ([]accessors.Device, error) {

	return room.Devices, nil
}

func (p *InputDSP) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {

	var audioDevices, dsps []accessors.Device
	for _, device := range devices {

		if device.HasRole("AudioOut") {

			audioDevices = append(audioDevices, device)
		}

		if device.HasRole("DSP") {

			dsps = append(dsps, device)
		}
	}

	//business as usual for audioDevices
	commands, err := generateStandardStatusCommand(audioDevices, INPUT_DSP, STATUS_INPUT_DSP)
	if err != nil {
		errorMessage := "Could not generate audio device status commands: " + err.Error()
		log.Printf(errorMessage)
		return []StatusCommand{}, errors.New(errorMessage)
	}

	//validate DSP configuration
	if dsps == nil || len(dsps) != 1 {
		return []StatusCommand{}, errors.New("Invalid DSP configuration detected")
	}

	dsp := dsps[0]

	//get switcher associated with DSP

	switchers, err := dbo.GetDevicesByBuildingAndRoomAndRole(dsp.Building.Name, dsp.Room.Name, "VideoSwitcher")
	if err != nil {
		errorMessage := "Could not get video switcher in building: " + dsp.Building.Name + ", room: " + dsp.Room.Name + "" + err.Error()
		log.Printf(errorMessage)
		return []StatusCommand{}, errors.New(errorMessage)
	}

	//validate number of switchers
	if switchers == nil || len(switchers) != 1 {
		return []StatusCommand{}, errors.New("Invalid video switcher configuration detected")
	}

	for _, port := range switchers[0].Ports {

		if port.Destination == dsp.Name { //found port configuration, issue command to switcher

			parameters := make(map[string]string)
			parameters["address"] = switchers[0].Address
			parameters["port"] = port.Name

			destinationDevice := DestinationDevice{
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
	}

	return commands, nil
}

func (p *InputDSP) EvaluateResponse(label string, value interface{}, source accessors.Device, destination DestinationDevice) (string, interface{}, error) {
	return label, value, nil
}
