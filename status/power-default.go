package status

import (
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

type PowerDefault struct {
}

//when querying power, we care about every device
func (p *PowerDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return []accessors.Device{}, nil
}

func (p *PowerDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return []StatusCommand{}, nil
}
