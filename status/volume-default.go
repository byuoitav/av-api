package status

import "github.com/byuoitav/configuration-database-microservice/accessors"

type VolumeDefault struct {
}

func (p *VolumeDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return []accessors.Device{}, nil
}

func (p *VolumeDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return []StatusCommand{}, nil
}
