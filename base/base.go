package base

//PublicRoom is the struct that is returned (or put) as part of the public API
type PublicRoom struct {
	Building          string        `json:"building, omitempty"`
	Room              string        `json:"room, omitempty"`
	CurrentVideoInput string        `json:"currentVideoInput"`
	CurrentAudioInput string        `json:"currentAudioInput"`
	Power             string        `json:"power"`
	Blanked           *bool         `json:"blanked"`
	Displays          []Display     `json:"displays"`
	AudioDevices      []AudioDevice `json:"audioDevices"`
}

//Device is a struct for inheriting
type Device struct {
	Name  string `json:"name"`
	Power string `json:"power"`
	Input string `json:"input"`
}

//AudioDevice represents an audio device
type AudioDevice struct {
	Device
	Muted  *bool `json:"muted"`
	Volume *int  `json:"volume"`
}

//Display represents a display
type Display struct {
	Device
	Blanked *bool `json:"blanked"`
}
