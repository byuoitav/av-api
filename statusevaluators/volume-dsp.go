package statusevaluators

import (
	"errors"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

// VolumeDSPEvaluator is a constant variable for the name of the evaluator.
const VolumeDSPEvaluator = "STATUS_VolumeDSP"

// VolumeDSPCommand is a constant variable for the name of the command.
const VolumeDSPCommand = "STATUS_VolumeDSP"

// VolumeDSP implements the StatusEvaluator struct.
type VolumeDSP struct{}

// GenerateCommands generates a list of commands for the given devices.
func (p *VolumeDSP) GenerateCommands(room structs.Room) ([]StatusCommand, int, error) {

	audioDevices := FilterDevicesByRole(room.Devices, "AudioOut")
	dsp := FilterDevicesByRole(room.Devices, "DSP")
	mics := FilterDevicesByRole(room.Devices, "Microphone")

	commands, count, err := generateStandardStatusCommand(audioDevices, VolumeDSPEvaluator, VolumeDefaultCommand)
	if err != nil {
		errorMessage := "[statusevals] Could not generate " + VolumeDSPCommand + "commands for audio devices: " + err.Error()
		log.L.Error(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	micCommands, c, err := generateMicStatusCommands(room, mics, VolumeDSPEvaluator, VolumeDSPCommand)
	if err != nil {
		errorMessage := "[statusevals] Could not generate " + VolumeDSPCommand + "commands for microphones: " + err.Error()
		log.L.Error(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	count += c
	commands = append(commands, micCommands...)

	dspCommands, c, err := generateDSPStatusCommands(room, dsp, VolumeDSPEvaluator, VolumeDSPCommand)
	if err != nil {
		errorMessage := "[statusevals] Could not generate " + VolumeDSPCommand + "commands for DSP: " + err.Error()
		log.L.Error(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	count += c
	commands = append(commands, dspCommands...)

	return commands, count, nil
}

// EvaluateResponse processes the response information that is given.
func (p *VolumeDSP) EvaluateResponse(room structs.Room, label string, value interface{}, source structs.Device, destination base.DestinationDevice) (string, interface{}, error) {

	const ScaleFactor = 3
	const MINIMUM = 45
	if structs.HasRole(destination.Device, "Microphone") {

		intValue, ok := value.(int)
		if ok {

			return label, (intValue - MINIMUM) * ScaleFactor, nil

		}
	}
	return label, value, nil
}
