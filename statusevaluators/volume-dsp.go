package statusevaluators

import (
	"errors"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

const VOLUME_DSP = "STATUS_VolumeDSP"
const STATUS_VOLUME_DSP = "STATUS_VolumeDSP"

type VolumeDSP struct{}

func (p *VolumeDSP) GetDevices(room structs.Room) ([]structs.Device, error) {

	return room.Devices, nil
}

func (p *VolumeDSP) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {

	var audioDevices, mics, dsp []structs.Device

	for _, device := range devices {

		base.Log("Considering device: %s", device.ID)

		if structs.HasRole(device, "Microphone") {

			base.Log("Appending %s to mic array...", device.Name)
			mics = append(mics, device)
		} else if structs.HasRole(device, "DSP") {

			base.Log("Appending %s to DSP array...", device.Name)
			dsp = append(dsp, device)
		} else if structs.HasRole(device, "AudioOut") {

			base.Log("Appending %s to audio devices array...", device.Name)
			audioDevices = append(audioDevices, device)
		} else {
			continue
		}
	}

	commands, count, err := generateStandardStatusCommand(audioDevices, VOLUME_DSP, VolumeDefaultCommandName)
	if err != nil {
		errorMessage := "Could not generate " + STATUS_VOLUME_DSP + "commands for audio devices: " + err.Error()
		base.Log(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	micCommands, c, err := generateMicStatusCommands(mics, VOLUME_DSP, STATUS_VOLUME_DSP)
	if err != nil {
		errorMessage := "Could not generate " + STATUS_VOLUME_DSP + "commands for microphones: " + err.Error()
		base.Log(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	count += c
	commands = append(commands, micCommands...)

	dspCommands, c, err := generateDSPStatusCommands(dsp, VOLUME_DSP, STATUS_VOLUME_DSP)
	if err != nil {
		errorMessage := "Could not generate " + STATUS_VOLUME_DSP + "commands for DSP: " + err.Error()
		base.Log(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	count += c
	commands = append(commands, dspCommands...)

	return commands, count, nil
}

func (p *VolumeDSP) EvaluateResponse(label string, value interface{}, source structs.Device, destination base.DestinationDevice) (string, interface{}, error) {

	const SCALE_FACTOR = 3
	const MINIMUM = 45
	if structs.HasRole(destination.Device, "Microphone") {

		intValue, ok := value.(int)
		if ok {

			return label, (intValue - MINIMUM) * SCALE_FACTOR, nil

		}
	}
	return label, value, nil
}
