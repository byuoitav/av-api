package status

import (
	"log"

	"github.com/byuoitav/configuration-database-microservice/accessors"
)

const VolumeDefaultName = "STATUS_VolumeDefault"
const VolumeDefaultCommandName = "STATUS_Volume"

type VolumeDefault struct {
}

func (p *VolumeDefault) GetDevices(room accessors.Room) ([]accessors.Device, error) {
	return room.Devices, nil
}

func (p *VolumeDefault) GenerateCommands(devices []accessors.Device) ([]StatusCommand, error) {
	return generateStandardStatusCommand(devices, VolumeDefaultName, VolumeDefaultCommandName)
}

func (p *VolumeDefault) EvaluateResponse(label string, value interface{}) (string, interface{}, error) {
	log.Printf("Evaluating response: %s, %s in evaluator %v", label, value, VolumeDefaultCommandName)
	return label, value, nil
}
