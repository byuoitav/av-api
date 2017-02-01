package commandevaluators

import (
	"errors"
	"log"
	"strings"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/av-api/dbo"
)

type UnMuteDefault struct {
}

func (p *UnMuteDefault) Evaluate(room base.PublicRoom) ([]base.ActionStructure, error) {
	log.Printf("Evaluating UnMute command.")

	actions := []base.ActionStructure{}

	//check if request is a roomwide unmute
	if room.Muted != nil && !*room.Muted {

		log.Printf("Room-wide UnMute request recieved. Retrieving all devices")

		devices, err := dbo.GetDevicesByBuildingAndRoomAndRole(room.Building, room.Room, "AudioOut")
		if err != nil {
			return []base.ActionStructure{}, err
		}

		log.Printf("UnMuting alll devices in room.")

		for i := range devices {

			if devices[i].Output {

				log.Printf("Adding device %+v", devices[i].Name)

				actions = append(actions, base.ActionStructure{
					Action:              "UnMute",
					GeneratingEvaluator: "UnMuteDefault",
					Device:              devices[i],
					DeviceSpecific:      false,
				})

			}

		}

	}

	//check specific devices
	log.Printf("Evaluating individual audio devices for unmuting.")

	for _, audioDevice := range room.AudioDevices {

		log.Printf("Adding device %+v", audioDevice.Name)

		if audioDevice.Muted != nil && !*audioDevice.Muted {

			device, err := dbo.GetDeviceByName(room.Building, room.Room, audioDevice.Name)
			if err != nil {
				return []base.ActionStructure{}, err
			}

			actions = append(actions, base.ActionStructure{
				Action:              "UnMute",
				GeneratingEvaluator: "UnMuteDefault",
				Device:              device,
				DeviceSpecific:      true,
			})

		}

	}

	log.Printf("%v actions generated.", len(actions))
	log.Printf("Evalutation complete.")

	return actions, nil

}

func (p *UnMuteDefault) Validate(action base.ActionStructure) error {

	log.Printf("Validating action for command \"UnMute\"")

	ok, _ := CheckCommands(action.Device.Commands, "UnMute")

	if !ok || !strings.EqualFold(action.Action, "UnMute") {
		log.Printf("ERROR. %s is an invalid command for %s", action.Action, action.Device.Name)
		return errors.New(action.Action + " is an invalid command for" + action.Device.Name)
	}

	log.Printf("Done.")
	return nil
}

func (p *UnMuteDefault) GetIncompatibleCommands() (incompatibleActions []string) {

	incompatibleActions = []string{
		"Mute",
	}

	return
}
