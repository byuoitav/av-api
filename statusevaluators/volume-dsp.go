package statusevaluators

import (
	"errors"
	"log"

	"github.com/byuoitav/configuration-database-microservice/structs"
)

const VOLUME_DSP = "STATUS_VolumeDSP"
const STATUS_VOLUME_DSP = "STATUS_VolumeDSP"

type VolumeDSP struct{}

func (p *VolumeDSP) GetDevices(room structs.Room) ([]structs.Device, error) {

	return room.Devices, nil
}

func (p *VolumeDSP) GenerateCommands(devices []structs.Device) ([]StatusCommand, error) {

	var audioDevices, mics, dsp []structs.Device

	for _, device := range devices {

		log.Printf("Considering device: %s", device.GetFullName())

		if device.HasRole("Microphone") {

			log.Printf("Appending %s to mic array...", device.Name)
			mics = append(mics, device)
		} else if device.HasRole("DSP") {

			log.Printf("Appending %s to DSP array...", device.Name)
			dsp = append(dsp, device)
		} else if device.HasRole("AudioOut") {

			log.Printf("Appending %s to audio devices array...", device.Name)
			audioDevices = append(audioDevices, device)
		} else {
			continue
		}
	}

	commands, err := generateStandardStatusCommand(audioDevices, VOLUME_DSP, VolumeDefaultCommandName)
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

func (p *VolumeDSP) EvaluateResponse(label string, value interface{}, source structs.Device, destination DestinationDevice) (string, interface{}, error) {

	const SCALE_FACTOR = 3
	const MINIMUM = 45
	if destination.Device.HasRole("Microphone") {

		intValue, ok := value.(int)
		if ok {

			return label, (intValue - MINIMUM) * SCALE_FACTOR, nil

		}
	}
	return label, value, nil
}
