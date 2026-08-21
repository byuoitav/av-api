package commandevaluators

import (
	"testing"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

func TestUnMuteDSPGeneratesUnmuteForVolumeOnlyDisplayAudioRequest(t *testing.T) {
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

	if count != 1 || len(actions) != 1 {
		t.Fatalf("expected one unmute action, got count %d and actions %d", count, len(actions))
	}

	action := actions[0]
	if action.Action != "UnMute" {
		t.Fatalf("expected UnMute action, got %q", action.Action)
	}

	if action.Device.ID != "JKB-1104-DSP1" {
		t.Fatalf("expected action to target DSP, got %q", action.Device.ID)
	}

	if action.DestinationDevice.Device.ID != "JKB-1104-D1" {
		t.Fatalf("expected destination to be D1, got %q", action.DestinationDevice.Device.ID)
	}

	if action.Parameters["input"] != "JKB1104Media" {
		t.Fatalf("expected DSP media input JKB1104Media, got %q", action.Parameters["input"])
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
