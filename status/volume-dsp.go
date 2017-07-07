package status

import (
	"errors"
	"log"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

const VOLUME_DSP = "STATUS_VolumeDSP"
const STATUS_VOLUME_DSP = "STATUS_VolumeDSP"

type VolumeDSP struct{}

func (p *VolumeDSP) GetDevices(room accessors.Room) ([]accessors.Device, error) {

	return room.Devices, nil
}

func (p *VolumeDSP) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {

	var audioDevices, mics, dsp []accessors.Device

	for _, device := range devices {

		if device.HasRole("Microhphone") {

			mics = append(mics, device)
		} else if device.HasRole("DSP") {

			dsp = append(dsp, device)
		} else if device.HasRole("AudioOut") {

			audioDevices = append(audioDevices, device)
		} else {
			continue
		}
	}

	commands, err := generateStandardStatusCommand(audioDevices, VOLUME_DSP, STATUS_VOLUME_DSP)
	if err != nil {
		errorMessage := "Could not generate " + STATUS_VOLUME_DSP + "commands for audio devices: " + err.Error()
		log.Printf(errorMessage)
		return []StatusCommand{}, errors.New(errorMessage)
	}

	micCommands, err := generateMicStatusCommands(mics, VOLUME_DSP, STATUS_VOLUME_DSP)
	if err != nil {
		errorMessage := "Could not generate " + STATUS_VOLUME_DSP + "commands for microphones: " + err.Error()
		log.Printf(errorMessage)
		return []StatusCommand{}, errors.New(errorMessage)
	}

	commands = append(commands, micCommands...)

	dspCommands, err := generateDSPStatusCommands(dsp, VOLUME_DSP, STATUS_VOLUME_DSP)
	if err != nil {
		errorMessage := "Could not generate " + STATUS_VOLUME_DSP + "commands for DSP: " + err.Error()
		log.Printf(errorMessage)
		return []StatusCommand{}, errors.New(errorMessage)
	}

	commands = append(commands, dspCommands...)

	return commands, nil
}

func (p *VolumeDSP) EvaluateResponse(label string, value interface{}, source accessors.Device, destination DestinationDevice) (string, interface{}, error) {

	if destination.Device.HasRole("Microphone") {
		log.Printf("Evaluating response pertaining to microphone: %s", destination.Device.Name)
	}
	return label, value, nil
}
