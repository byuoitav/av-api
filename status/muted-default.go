package status

import "github.com/byuoitav/configuration-database-microservice/accessors"

type MutedDefault struct {
}

func (p *MutedDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return []accessors.Device{}, nil
}

func (p *MutedDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return []StatusCommand{}, nil
}
