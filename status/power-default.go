package status

import (
	"log"
	"strings"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

type PowerDefault struct {
}

//when querying power, we care about every device
func (p *PowerDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return room.Devices, nil
}

func (p *PowerDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	var output []StatusCommand

	//iterate over each device
	for _, device := range devices {

		for _, command := range device.Commands {

			if strings.HasPrefix(command.Name, FLAG) {

				log.Printf("Adding command: %s to action list", command.Name)
				output = append(output, StatusCommand{
					Action: command.Name,
					Device: device,
				})

			}

		}

	}
	return output, nil
}
