package base

import (
	"github.com/byuoitav/common/structs"
	ei "github.com/byuoitav/common/v2/events"
)

//PublicRoom is the struct that is returned (or put) as part of the public API
type PublicRoom struct {
	Building          string        `json:"-"`
	Room              string        `json:"-"`
	CurrentVideoInput string        `json:"currentVideoInput,omitempty"`
	CurrentAudioInput string        `json:"currentAudioInput,omitempty"`
	Power             string        `json:"power,omitempty"`
	Blanked           *bool         `json:"blanked,omitempty"`
	Muted             *bool         `json:"muted,omitempty"`
	Volume            *int          `json:"volume,omitempty"`
	Displays          []Display     `json:"displays,omitempty"`
	AudioDevices      []AudioDevice `json:"audioDevices,omitempty"`
}

//Device is a struct for inheriting
type Device struct {
	Name  string `json:"name,omitempty"`
	Power string `json:"power,omitempty"`
	Input string `json:"input,omitempty"`
}

//AudioDevice represents an audio device
type AudioDevice struct {
	Device
	Muted  *bool `json:"muted,omitempty"`
	Volume *int  `json:"volume,omitempty"`
}

//Display represents a display
type Display struct {
	Device
	Blanked *bool `json:"blanked,omitempty"`
}

//ActionStructure is the internal struct we use to pass commands around once
//they've been evaluated.
//also contains a list of Events to be published
type ActionStructure struct {
	Action              string             `json:"action"`
	GeneratingEvaluator string             `json:"generatingEvaluator"`
	Device              structs.Device     `json:"device"`
	DestinationDevice   DestinationDevice  `json:"destination_device"`
	Parameters          map[string]string  `json:"parameters"`
	DeviceSpecific      bool               `json:"deviceSpecific,omitempty"`
	Overridden          bool               `json:"overridden"`
	EventLog            []ei.Event         `json:"events"`
	Children            []*ActionStructure `json:"children"`
	Callback            func(StatusPackage, chan<- StatusPackage) error
}

// DestinationDevice represents the device that is being acted upon.
type DestinationDevice struct {
	structs.Device
	AudioDevice bool `json:"audio"`
	Display     bool `json:"video"`
}

// StatusPackage contains the callback information for the action.
type StatusPackage struct {
	Key    string
	Value  interface{}
	Device structs.Device
	Dest   DestinationDevice
}

//Equals checks if the action structures are equal
func (a *ActionStructure) Equals(b ActionStructure) bool {
	return a.Action == b.Action &&
		a.Device.ID == b.Device.ID &&
		a.Device.Address == b.Device.Address &&
		a.DeviceSpecific == b.DeviceSpecific &&
		a.Overridden == b.Overridden && CheckStringMapsEqual(a.Parameters, b.Parameters)
}

//ActionByPriority implements the sort.Interface for []ActionStructure
type ActionByPriority []ActionStructure

func (abp ActionByPriority) Len() int { return len(abp) }

func (abp ActionByPriority) Swap(i, j int) { abp[i], abp[j] = abp[j], abp[i] }

func (abp ActionByPriority) Less(i, j int) bool {
	var ipri int
	var jpri int
	//we've gotta go through and get the priorities
	for _, command := range abp[i].Device.Type.Commands {
		if command.ID == abp[i].Action {
			ipri = command.Priority
			break
		}
	}
	for _, command := range abp[j].Device.Type.Commands {
		if command.ID == abp[j].Action {
			jpri = command.Priority
			break
		}
	}
	return ipri < jpri
}

//CheckStringMapsEqual just takes two map[string]string and compares them.
func CheckStringMapsEqual(a map[string]string, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if b[k] != v {
			return false
		}
	}

	return true
}

//CheckStringSliceEqual is a simple helper to check if two string slices contain
//the same elements.
func CheckStringSliceEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
