package statusevaluators

import (
	"errors"
	"strconv"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/fatih/color"
)

/* ASSUMPTIONS

a) a mic has only one port configuration with the DSP as a destination device

*/

// MutedDSPEvaluator is a constant variable for the name of the evaluator.
const MutedDSPEvaluator = "STATUS_MutedDSP"

// MutedDSPCommand is a constant variable for the name of the command.
const MutedDSPCommand = "STATUS_MutedDSP"

// MutedDSP implements the StatusEvaluator struct.
type MutedDSP struct{}

// GenerateCommands generates a list of commands for the given devices.
func (p *MutedDSP) GenerateCommands(room structs.Room) ([]StatusCommand, int, error) {

	log.L.Info("[statusevals] Generating \"Muted\" status commands...")

	//sort mics out of audio devices

	audioDevices := FilterDevicesByRole(room.Devices, "AudioOut")
	dsp := FilterDevicesByRole(room.Devices, "DSP")
	mics := FilterDevicesByRole(room.Devices, "Microphone")

	//business as ususal for audioDevices
	commands, count, err := generateStandardStatusCommand(audioDevices, MutedDSPEvaluator, MutedDefaultCommand)
	if err != nil {
		errorMessage := "[statusevals] Could not generate audio device status commands: " + err.Error()
		log.L.Error(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	micCommands, c, err := generateMicStatusCommands(room, mics, MutedDSPEvaluator, MutedDSPCommand)
	if err != nil {
		errorMessage := "[statusevals] Could not generate microphone status commands: " + err.Error()
		log.L.Error(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	count += c
	commands = append(commands, micCommands...)

	dspCommands, c, err := generateDSPStatusCommands(room, dsp, MutedDSPEvaluator, MutedDSPCommand)
	if err != nil {
		return []StatusCommand{}, 0, err
	}

	count += c
	commands = append(commands, dspCommands...)

	log.L.Infof(color.HiYellowString("[STATUS-Muted-DSP] Generated %v commands", len(commands)))
	return commands, count, nil

}

// EvaluateResponse processes the response information that is given.
func (p *MutedDSP) EvaluateResponse(room structs.Room, label string, value interface{}, source structs.Device, destintation base.DestinationDevice) (string, interface{}, error) {

	return label, value, nil
}

func generateMicStatusCommands(room structs.Room, mics []structs.Device, evaluator string, command string) ([]StatusCommand, int, error) {

	log.L.Infof("[statusevals] Generating %s commands agains mics...", command)

	var commands []StatusCommand

	if len(mics) == 0 {
		errorMessage := "[statusevals] No mics"

		log.L.Error(errorMessage)
		return []StatusCommand{}, 0, nil
	}

	dsp := FilterDevicesByRole(room.Devices, "DSP")

	if len(dsp) != 1 {
		errorMessage := "[statusevals] Invalid number of DSP devices found in room: " + strconv.Itoa(len(dsp))
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	var count int

	for _, mic := range mics {

		log.L.Infof("[statusevals] Considering mic %s...", mic.Name)

		//find the only DSP the room has

		for _, port := range dsp[0].Ports {

			if port.SourceDevice == mic.ID {
				log.L.Infof("[statusevals] Port configuration identified for mic %s and DSP %s", mic.Name, dsp[0].Name)
				destinationDevice := base.DestinationDevice{
					Device:      mic,
					AudioDevice: true,
					Display:     mic.HasRole("VideoOut"),
				}

				statusCommand := dsp[0].GetCommandByID(command)

				parameters := make(map[string]string)
				parameters["input"] = strings.Replace(port.ID, "OUT", "", 1)
				parameters["address"] = dsp[0].Address

				//issue status command to DSP
				commands = append(commands, StatusCommand{
					Action:            statusCommand,
					Device:            dsp[0],
					Generator:         MutedDSPEvaluator,
					DestinationDevice: destinationDevice,
					Parameters:        parameters,
				})
				count++
			}
		}

	}

	return commands, count, nil
}

func generateDSPStatusCommands(room structs.Room, dsp []structs.Device, evaluator string, command string) ([]StatusCommand, int, error) {

	var commands []StatusCommand

	//validate the correct number of dsps
	if dsp == nil || len(dsp) != 1 {
		errorMessage := "[statusevals] Invalid number of DSP devices found in room: " + strconv.Itoa(len(dsp))
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	if !structs.HasRole(dsp[0], "AudioOut") {
		//we don't need to get the state of it
		return []StatusCommand{}, 0, nil
	}

	log.L.Infof("[statusevals] Generating DSP status command: %s against device: %s", command, dsp[0])

	parameters := make(map[string]string)
	parameters["address"] = dsp[0].Address

	statusCommand := dsp[0].GetCommandByID(command)

	destinationDevice := base.DestinationDevice{
		Device:      dsp[0],
		AudioDevice: true,
		Display:     dsp[0].HasRole("VideoOut"),
	}
	var count int

	//one command for each port that's not a mic
	for _, port := range dsp[0].Ports {
		device := FindDevice(room.Devices, port.SourceDevice)

		if !structs.HasRole(device, "Microphone") {

			parameters["input"] = port.ID
			commands = append(commands, StatusCommand{
				Action:            statusCommand,
				Device:            dsp[0],
				Generator:         evaluator,
				DestinationDevice: destinationDevice,
				Parameters:        parameters,
			})
		}
		count++
	}

	return commands, count, nil
}
