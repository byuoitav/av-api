package statusevaluators

import (
	"errors"
	"log"
	"strconv"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

/* ASSUMPTIONS

a) a mic has only one port configuration with the DSP as a destination device

*/

const MUTED_DSP = "STATUS_MutedDSP"
const MUTE_DSP_STATUS = "STATUS_MutedDSP"

type MutedDSP struct{}

func (p *MutedDSP) GetDevices(room accessors.Room) ([]accessors.Device, error) {

	return room.Devices, nil
}

func (p *MutedDSP) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {

	log.Printf("Generating \"Muted\" status commands...")

	//sort mics out of audio devices:w
	var audioDevices, mics, dsp []accessors.Device

	for _, device := range devices {

		log.Printf("Considering device: %s", device.Name)

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
	commands, err := generateStandardStatusCommand(audioDevices, MUTED_DSP, MutedDefaultCommandName)
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

	for _, command := range commands {

		log.Printf("action: %v", command.Action)
		log.Printf("Device: %v", command.Device)
		log.Printf("Destination device: %v", command.DestinationDevice)
		log.Printf("Parameters: %v", command.Parameters)

	}
	return commands, nil

}

func (p *MutedDSP) EvaluateResponse(label string, value interface{}, source accessors.Device, destintation DestinationDevice) (string, interface{}, error) {

	return label, value, nil
}

func generateMicStatusCommands(mics []accessors.Device, evaluator string, command string) ([]StatusCommand, error) {

	log.Printf("Generating %s commands agains mics...", command)

	var commands []StatusCommand

	if len(mics) == 0 {
		errorMessage := "No mics"
		return []StatusCommand{}, errors.New(errorMessage)
	}

	dsp, err := dbo.GetDevicesByBuildingAndRoomAndRole(mics[0].Building.Shortname, mics[0].Room.Name, "DSP")
	if err != nil {
		return []StatusCommand{}, err
	}

	if len(dsp) != 1 {
		errorMessage := "Invalid number of DSP devices found in room: " + strconv.Itoa(len(dsp))
		return []StatusCommand{}, errors.New(errorMessage)
	}

	for _, mic := range mics {

		log.Printf("Considering mic %s...", mic.Name)

		//find the only DSP the room has

		for _, port := range dsp[0].Ports {

			if port.Source == mic.Name {
				log.Printf("Port configuration identified for mic %s and DSP %s", mic.Name, dsp[0].Name)
				destinationDevice := DestinationDevice{
					Device:      mic,
					AudioDevice: true,
				}

				statusCommand := mic.GetCommandByName(command)

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
			}
		}

	}

	return commands, nil
}

func generateDSPStatusCommands(dsp []accessors.Device, evaluator string, command string) ([]StatusCommand, error) {

	var commands []StatusCommand

	//validate the correct number of dsps
	if dsp == nil || len(dsp) != 1 {
		errorMessage := "Invalide number of DSP devices found in room: " + strconv.Itoa(len(dsp))
		return []StatusCommand{}, errors.New(errorMessage)
	}

	log.Printf("Generating DSP status command: %s against device: %s", command, dsp[0])

	parameters := make(map[string]string)
	parameters["address"] = dsp[0].Address

	statusCommand := dsp[0].GetCommandByName(command)

	destinationDevice := DestinationDevice{
		Device:      dsp[0],
		AudioDevice: true,
	}

	//one command for each port that's not a mic
	for _, port := range dsp[0].Ports {

		device, err := dbo.GetDeviceByName(dsp[0].Building.Shortname, dsp[0].Room.Name, port.Source)
		if err != nil {
			return []StatusCommand{}, err
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
	}

	return commands, nil
}
