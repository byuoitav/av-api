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

//AudioDevice represents an audio device
type AudioDevice struct {
	Name   string `json:"name"`
	Power  string `json:"power"`
	Input  string `json:"input"`
	Muted  *bool  `json:"muted"`
	Volume *int   `json:"volume"`
}

//Display represents a display
type Display struct {
	Name    string `json:"name"`
	Power   string `json:"power"`
	Input   string `json:"input"`
	Blanked *bool  `json:"blanked"`
}
