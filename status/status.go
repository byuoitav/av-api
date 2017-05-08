package status

import (
	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/configuration-database-microservice/accessors"
)

type PowerStatus struct {
	Power string `json:"power",omitempty`
}

type BlankedStatus struct {
	Blanked bool `json:"blanked",omitempty`
}

type MuteStatus struct {
	Muted bool `json:"muted",omitempty`
}

type VideoInput struct {
	Input string `json:"input",omitempty`
}

type AudioInput struct {
	Input string `json:"input",omitempty`
}

type AudioList struct {
	Inputs []AudioInput `json"inputs",omitempty`
}

type VideoList struct {
	Inputs []VideoInput `json:"inputs",omitemtpy`
}

//a status evaluator looks for all the commands labelled 'STATUS' for each device and decides if those are the statuses we want
type StatusEvaluator interface {

	//Identifies relevant devices
	GetDevices(base.PublicRoom) ([]accessors.Device, error)

	//Generates action list
	EvaluateCommands(devices []accessors.Device) ([]base.ActionStructure, error)
}

const FLAG = "STATUS"
