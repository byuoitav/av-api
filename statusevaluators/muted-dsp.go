package statusevaluators

import (
	"errors"
	"strconv"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

/* ASSUMPTIONS

a) a mic has only one port configuration with the DSP as a destination device

*/

const MUTED_DSP = "STATUS_MutedDSP"
const MUTE_DSP_STATUS = "STATUS_MutedDSP"

type MutedDSP struct{}

func (p *MutedDSP) GetDevices(room structs.Room) ([]structs.Device, error) {

	return room.Devices, nil
}

func (p *MutedDSP) GenerateCommands(devices []structs.Device) ([]StatusCommand, int, error) {

	base.Log("Generating \"Muted\" status commands...")

	//sort mics out of audio devices:w
	var audioDevices, mics, dsp []structs.Device

	for _, device := range devices {

		base.Log("Considering device: %s", device.Name)

		if device.HasRole("Microphone") {

			mics = append(mics, device)
		} else if device.HasRole("DSP") {

			dsp = append(dsp, device)
		} else if device.HasRole("AudioOut") {

			audioDevices = append(audioDevices, device)
		} else {
			continue
		}
	}

	//business as ususal for audioDevices
	commands, count, err := generateStandardStatusCommand(audioDevices, MUTED_DSP, MutedDefaultCommandName)
	if err != nil {
		errorMessage := "Could not generate audio device status commands: " + err.Error()
		base.Log(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	micCommands, c, err := generateMicStatusCommands(mics, MUTED_DSP, MUTE_DSP_STATUS)
	if err != nil {
		errorMessage := "Could not generate microphone status commands: " + err.Error()
		base.Log(errorMessage)
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	count += c
	commands = append(commands, micCommands...)

	dspCommands, c, err := generateDSPStatusCommands(dsp, MUTED_DSP, MUTE_DSP_STATUS)
	if err != nil {
		return []StatusCommand{}, 0, err
	}

	count += c
	commands = append(commands, dspCommands...)
	/*
		for _, command := range commands {

			base.Log("action: %v", command.Action)
			base.Log("Device: %v", command.Device)
			base.Log("Destination device: %v", command.DestinationDevice)
			base.Log("Parameters: %v", command.Parameters)

		}
	*/
	base.Log(color.HiYellowString("[STATUS-Muted-DSP] Generated %v commands", len(commands)))
	return commands, count, nil

}

func (p *MutedDSP) EvaluateResponse(label string, value interface{}, source structs.Device, destintation base.DestinationDevice) (string, interface{}, error) {

	return label, value, nil
}

func generateMicStatusCommands(mics []structs.Device, evaluator string, command string) ([]StatusCommand, int, error) {

	base.Log("Generating %s commands agains mics...", command)

	var commands []StatusCommand

	if len(mics) == 0 {
		errorMessage := "No mics"

		base.Log(errorMessage)
		return []StatusCommand{}, 0, nil
	}

	dsp, err := dbo.GetDevicesByBuildingAndRoomAndRole(mics[0].Building.Shortname, mics[0].Room.Name, "DSP")
	if err != nil {
		return []StatusCommand{}, 0, err
	}

	if len(dsp) != 1 {
		errorMessage := "Invalid number of DSP devices found in room: " + strconv.Itoa(len(dsp))
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	var count int

	for _, mic := range mics {

		base.Log("Considering mic %s...", mic.Name)

		//find the only DSP the room has

		for _, port := range dsp[0].Ports {

			if port.Source == mic.Name {
				base.Log("Port configuration identified for mic %s and DSP %s", mic.Name, dsp[0].Name)
				destinationDevice := base.DestinationDevice{
					Device:      mic,
					AudioDevice: true,
				}

				statusCommand := dsp[0].GetCommandByName(command)

				parameters := make(map[string]string)
				parameters["input"] = port.Name
				parameters["address"] = dsp[0].Address

				//issue status command to DSP
				commands = append(commands, StatusCommand{
					Action:            statusCommand,
					Device:            dsp[0],
					Generator:         MUTED_DSP,
					DestinationDevice: destinationDevice,
					Parameters:        parameters,
				})
				count++
			}
		}

	}

	return commands, count, nil
}

func generateDSPStatusCommands(dsp []structs.Device, evaluator string, command string) ([]StatusCommand, int, error) {

	var commands []StatusCommand

	//validate the correct number of dsps
	if dsp == nil || len(dsp) != 1 {
		errorMessage := "Invalide number of DSP devices found in room: " + strconv.Itoa(len(dsp))
		return []StatusCommand{}, 0, errors.New(errorMessage)
	}

	base.Log("Generating DSP status command: %s against device: %s", command, dsp[0])

	parameters := make(map[string]string)
	parameters["address"] = dsp[0].Address

	statusCommand := dsp[0].GetCommandByName(command)

	destinationDevice := base.DestinationDevice{
		Device:      dsp[0],
		AudioDevice: true,
	}
	var count int

	//one command for each port that's not a mic
	for _, port := range dsp[0].Ports {

		device, err := dbo.GetDeviceByName(dsp[0].Building.Shortname, dsp[0].Room.Name, port.Source)
		if err != nil {
			return []StatusCommand{}, 0, err
		}

		if !device.HasRole("Microphone") {

			parameters["input"] = port.Name
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
