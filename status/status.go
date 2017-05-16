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

type Input struct {
	Input string `json:"input"`
}

type AudioList struct {
	Inputs []Input `json"inputs"`
}

type VideoList struct {
	Inputs []Input `json:"inputs"`
}

type Volume struct {
	Volume int `json:"volume"`
}

//represents output from a device, use Error field to flag errors
type Status struct {
	DestinationDevice DestinationDevice      `json:"device"`
	Responses         []StatusResponse       `json:"responses"`
	Status            map[string]interface{} `json:"status"`
	ErrorMessage      *string                `json:"error"`
}

//represents a status response, including the generator that created the command that returned the status
type StatusResponse struct {
	Generator string                 `json:"generator"`
	Status    map[string]interface{} `json:"status"`
}

//StatusCommand contains information to issue a status command against a device
type StatusCommand struct {
	Action            accessors.Command `json:"action"`
	Device            accessors.Device  `json:"device"`
	Generator         string            `json:"generator"`
	DestinationDevice DestinationDevice `json:"destination"`
	Parameters        map[string]string `json:"parameters"`
}

//DestinationDevice represents the device a status command is issued to
type DestinationDevice struct {
	accessors.Device
	AudioDevice bool `json:"audio"`
	Display     bool `json:"video"`
}

type StatusEvaluator interface {

	//Identifies relevant devices
	GetDevices(room accessors.Room) ([]accessors.Device, error)

	//Generates action list
	GenerateCommands(devices []accessors.Device) ([]StatusCommand, error)

	//Evaluate Response
	EvaluateResponse(label string, value interface{}) (string, interface{}, error)
}

const FLAG = "STATUS"

var DEFAULT_MAP = map[string]StatusEvaluator{
	"STATUS_PowerDefault":   &PowerDefault{},
	"STATUS_BlankedDefault": &BlankedDefault{},
	"STATUS_MutedDefault":   &MutedDefault{},
	"STATUS_InputDefault":   &InputDefault{},
	"STATUS_VolumeDefault":  &VolumeDefault{},
}
