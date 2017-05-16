package status

import "github.com/byuoitav/configuration-database-microservice/accessors"

type InputDefault struct {
}

func (p *InputDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return []accessors.Device{}, nil
}

func (p *InputDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return []StatusCommand{}, nil
}

func (p *InputDefault) EvaluateResponse(label string, value interface{}) (string, interface{}, error) {
	return label, value, nil
}
