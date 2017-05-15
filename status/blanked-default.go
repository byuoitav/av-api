package status

import "github.com/byuoitav/configuration-database-microservice/accessors"

type BlankedDefault struct {
}

func (p *BlankedDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return []accessors.Device{}, nil
}

func (p *BlankedDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return []StatusCommand{}, nil
}
