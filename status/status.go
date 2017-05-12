package status

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
	Volume int `json:"Volume"`
}
