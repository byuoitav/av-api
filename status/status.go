package status

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
