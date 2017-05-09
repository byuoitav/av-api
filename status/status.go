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

//represents output from a device, use Error field to flag errors
type Status struct {
	Device accessors.Device       `json:"device"`
	Status map[string]interface{} `json:"status"`
	Error  bool                   `json:"error"`
}

//StatusCommand contains information to issue a status command against a device
type StatusCommand struct {
	Action            accessors.Command `json:"action"`
	Device            accessors.Device  `json:"device"`
	DestinationDevice DestinationDevice `json:"destination"`
	Parameters        map[string]string `json:"parameters"`
}

//DestinationDevice represents the device a status command is issued to
type DestinationDevice struct {
	Device      accessors.Device `json:"device"`
	AudioDevice bool             `json:"audio"`
	VideoDevice bool             `json:"video"`
}

type StatusEvaluator interface {

	//Identifies relevant devices
	GetDevices(room accessors.Room) ([]accessors.Device, error)

	//Generates action list
	GenerateCommands(devices []accessors.Device) ([]StatusCommand, error)
}

const FLAG = "STATUS"

var DEFAULT_MAP = map[string]StatusEvaluator{
	"STATUS_PowerDefault":   &PowerDefault{},
	"STATUS_BlankedDefault": &BlankedDefault{},
	"STATUS_MutedDefault":   &MutedDefault{},
	"STATUS_InputDefault":   &InputDefault{},
	"STATUS_VolumeDefault":  &VolumeDefault{},
}
