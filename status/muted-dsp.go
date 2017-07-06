package status

import (
	"errors"
	"log"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

/* ASSUMPTIONS

a) a mic has only one port configuration with the DSP as a destination device

*/

const MUTED_DSP = "MutedDSP"
const MUTE_DSP_STATUS = "STATUS_Muted"

type MutedDSP struct{}

func (p *MutedDSP) GetDevices(room accessors.Room) ([]accessors.Device, error) {

	return room.Devices, nil
}

func (p *MutedDSP) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {

	//sort mics out of audio devices:w
	var audioDevices, mics, dsp []accessors.Device

	for _, device := range devices {

		if device.HasRole("AudioOut") {

			audioDevices = append(audioDevices, device)
		} else if device.HasRole("Microphone") {

			mics = append(mics, device)
		} else if device.HasRole("DSP") {

			dsp = append(dsp, device)
		} else {
			continue
		}
	}

	//business as ususal for audioDevices
	commands, err := generateStandardStatusCommand(audioDevices, MUTED_DSP, MUTE_DSP_STATUS)
	if err != nil {
		errorMessage := "Could not generate audio device status commands: " + err.Error()
		log.Printf(errorMessage)
		return []StatusCommand{}, errors.New(errorMessage)
	}

	micCommands, err := generateMicStatusCommands(mics, MUTED_DSP, MUTE_DSP_STATUS)
	if err != nil {
		errorMessage := "Could not generate microphone status commands: " + err.Error()
		log.Printf(errorMessage)
		return []StatusCommand{}, errors.New(errorMessage)
	}

	commands = append(commands, micCommands...)

	dspCommands, err := generateDSPStatusCommands(dsp, MUTED_DSP, MUTE_DSP_STATUS)
	if err != nil {
		return []StatusCommand{}, err
	}

	commands = append(commands, dspCommands...)

	return commands, nil

}

func (p *MutedDSP) EvaluateResponse(label string, value interface{}, source accessors.Device, destintation DestinationDevice) (string, interface{}, error) {

	log.Printf("Evaluating response: %s, %s in evaluator %v", label, value, BlankedDefaultName)

	if destintation.Device.HasRole("DSP") {

		for _, port := range destintation.Ports {

			valueString, ok := value.(string)
			if ok && port.Name == valueString {

				value = port.Source
			}
		}

	}

	return label, value, nil
}

func generateMicStatusCommands(mics []accessors.Device, evaluator string, command string) ([]StatusCommand, error) {

	var commands []StatusCommand

	for _, mic := range mics {

		//address DSP based on the (only) port a mic has
		port := mic.Ports[0]
		dsp, err := dbo.GetDeviceByName(mic.Building.Name, mic.Room.Name, port.Destination)
		if err != nil {
			return []StatusCommand{}, err
		}

		destinationDevice := DestinationDevice{
			Device:      dsp,
			AudioDevice: true,
		}

		statusCommand := mic.GetCommandByName(command)

		parameters := make(map[string]string)
		parameters["input"] = port.Name
		parameters["address"] = dsp.Address

		//issue status command to DSP
		commands = append(commands, StatusCommand{
			Action:            statusCommand,
			Device:            mic,
			Generator:         MUTED_DSP,
			DestinationDevice: destinationDevice,
			Parameters:        parameters,
		})
	}

	return commands, nil
}

func generateDSPStatusCommands(dsp []accessors.Device, evaluator string, command string) ([]StatusCommand, error) {

	var commands []StatusCommand

	return commands, nil
}
