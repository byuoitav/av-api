package commandevaluators

import (
	"testing"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

func TestUnMuteDSPSkipsVolumeOnlyAudioRequest(t *testing.T) {
	volume := 50
	dbRoom := dspRoomWithDisplayAudio()
	request := base.PublicRoom{
		Building: "JKB",
		Room:     "1104",
		AudioDevices: []base.AudioDevice{
			{
				Device: base.Device{Name: "D1"},
				Volume: &volume,
			},
		},
	}

	actions, count, err := (&UnMuteDSP{}).Evaluate(dbRoom, request, "test")
	if err != nil {
		t.Fatalf("Evaluate returned error: %s", err)
	}

	if count != 0 || len(actions) != 0 {
		t.Fatalf("expected no unmute actions, got count %d and actions %d", count, len(actions))
	}
}

func TestUnMuteDSPSkipsVolumeRequestWhenExplicitlyMuted(t *testing.T) {
	volume := 50
	muted := true
	dbRoom := dspRoomWithDisplayAudio()
	request := base.PublicRoom{
		Building: "JKB",
		Room:     "1104",
		AudioDevices: []base.AudioDevice{
			{
				Device: base.Device{Name: "D1"},
				Muted:  &muted,
				Volume: &volume,
			},
		},
	}

	actions, count, err := (&UnMuteDSP{}).Evaluate(dbRoom, request, "test")
	if err != nil {
		t.Fatalf("Evaluate returned error: %s", err)
	}

	if count != 0 || len(actions) != 0 {
		t.Fatalf("expected no unmute actions, got count %d and actions %d", count, len(actions))
	}
}

func dspRoomWithDisplayAudio() structs.Room {
	return structs.Room{
		ID:   "JKB-1104",
		Name: "JKB-1104",
		Devices: []structs.Device{
			{
				ID:   "JKB-1104-D1",
				Name: "D1",
				Roles: []structs.Role{
					{ID: "AudioOut"},
					{ID: "VideoOut"},
					{ID: "Microphone"},
				},
			},
			{
				ID:   "JKB-1104-DSP1",
				Name: "DSP1",
				Roles: []structs.Role{
					{ID: "AudioOut"},
					{ID: "DSP"},
				},
				Ports: []structs.Port{
					{
						ID:                "JKB1104Media",
						SourceDevice:      "JKB-1104-D1",
						DestinationDevice: "JKB-1104-DSP1",
					},
				},
			},
		},
	}
}
