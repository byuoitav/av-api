package status

import (
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

type PowerStatus struct {
	Power string `json:"power"`
}

type BlankedStatus struct {
	Blanked bool `json:"blanked"`
}

type MuteStatus struct {
	Muted bool `json:"muted"`
}

type VideoInput struct {
	Input string `json:"input"`
}

type AudioInput struct {
	Input string `json:"input"`
}

type AudioList struct {
	Inputs []AudioInput `json"inputs"`
}

type VideoList struct {
	Inputs []VideoInput `json:"inputs"`
}

type Volume struct {
	Volume int `json:"volume"`
}

//StatusCommand
type StatusCommand struct {
	Action     string            `json:"action"`
	Device     accessors.Device  `json:"device"`
	Parameters map[string]string `json:"parameters"`
}

//a status evaluator looks for all the commands labelled 'STATUS' for each device and decides if those are the statuses we want
type StatusEvaluator interface {

	//Identifies relevant devices
	GetDevices(room accessors.Room) ([]accessors.Device, error)

	//Generates action list
	GenerateCommands(devices []accessors.Device) ([]StatusCommand, error)
}

const FLAG = "STATUS"
