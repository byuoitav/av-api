package commandevaluators

import (
	"testing"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

func TestUnMuteDefaultGeneratesUnmuteForVolumeOnlyAudioRequest(t *testing.T) {
	volume := 50
	dbRoom := defaultRoomWithDisplayAudio()
	request := base.PublicRoom{
		Building: "ITB",
		Room:     "JAKE",
		AudioDevices: []base.AudioDevice{
			{
				Device: base.Device{Name: "D1"},
				Volume: &volume,
			},
		},
	}

	actions, count, err := (&UnMuteDefault{}).Evaluate(dbRoom, request, "test")
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

	if action.Device.ID != "ITB-JAKE-D1" {
		t.Fatalf("expected action to target D1, got %q", action.Device.ID)
	}
}

func TestUnMuteDefaultSkipsVolumeRequestWhenExplicitlyMuted(t *testing.T) {
	volume := 50
	muted := true
	dbRoom := defaultRoomWithDisplayAudio()
	request := base.PublicRoom{
		Building: "ITB",
		Room:     "JAKE",
		AudioDevices: []base.AudioDevice{
			{
				Device: base.Device{Name: "D1"},
				Muted:  &muted,
				Volume: &volume,
			},
		},
	}

	actions, count, err := (&UnMuteDefault{}).Evaluate(dbRoom, request, "test")
	if err != nil {
		t.Fatalf("Evaluate returned error: %s", err)
	}

	if count != 0 || len(actions) != 0 {
		t.Fatalf("expected no unmute actions, got count %d and actions %d", count, len(actions))
	}
}

func defaultRoomWithDisplayAudio() structs.Room {
	return structs.Room{
		ID:   "ITB-JAKE",
		Name: "ITB-JAKE",
		Devices: []structs.Device{
			{
				ID:   "ITB-JAKE-D1",
				Name: "D1",
				Type: structs.DeviceType{
					Output: true,
					Commands: []structs.Command{
						{ID: "UnMute"},
					},
				},
				Roles: []structs.Role{
					{ID: "AudioOut"},
					{ID: "VideoOut"},
				},
			},
		},
	}
}
