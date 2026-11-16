package commandevaluators

import (
	"testing"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

func TestAudioVolumeAndMuteAreIndependent(t *testing.T) {
	muted, unmuted := true, false
	zero, volume := 0, 50
	displayRoom := dspRoomWithDisplayAudio()
	displayRoom.Devices[0].Roles = displayRoom.Devices[0].Roles[:2]
	mediaRoom := dspRoomWithDisplayAudio()
	mediaRoom.Devices[0].Roles = mediaRoom.Devices[0].Roles[:2]
	mediaRoom.Devices[1].Ports[0].SourceDevice = "D1"

	for _, config := range []struct {
		name       string
		room       structs.Room
		building   string
		roomName   string
		deviceName string
		targetID   string
		input      string
		volume     CommandEvaluator
		mute       CommandEvaluator
		unmute     CommandEvaluator
	}{
		{"default display", defaultRoomWithDisplayAudio(), "ITB", "JAKE", "D1", "ITB-JAKE-D1", "", &SetVolumeDefault{}, &MuteDefault{}, &UnMuteDefault{}},
		{"DSP microphone", dspRoomWithDisplayAudio(), "JKB", "1104", "D1", "JKB-1104-DSP1", "JKB1104Media", &SetVolumeDSP{}, &MuteDSP{}, &UnMuteDSP{}},
		{"DSP display", displayRoom, "JKB", "1104", "D1", "JKB-1104-D1", "", &SetVolumeDSP{}, &MuteDSP{}, &UnMuteDSP{}},
		{"DSP media", mediaRoom, "JKB", "1104", "DSP1", "JKB-1104-DSP1", "JKB1104Media", &SetVolumeDSP{}, &MuteDSP{}, &UnMuteDSP{}},
	} {
		t.Run(config.name, func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				volume *int
				muted  *bool
			}{
				{"no audio fields", nil, nil},
				{"volume only", &volume, nil},
				{"zero volume only", &zero, nil},
				{"mute only", nil, &muted},
				{"unmute only", nil, &unmuted},
				{"volume with mute", &volume, &muted},
				{"volume with unmute", &volume, &unmuted},
			} {
				t.Run(tc.name, func(t *testing.T) {
					request := base.PublicRoom{
						Building: config.building,
						Room:     config.roomName,
						AudioDevices: []base.AudioDevice{{
							Device: base.Device{Name: config.deviceName},
							Volume: tc.volume,
							Muted:  tc.muted,
						}},
					}
					for _, command := range []struct {
						evaluator CommandEvaluator
						action    string
						key       string
						value     string
						requested bool
					}{
						{config.volume, "SetVolume", "volume", "50", tc.volume != nil},
						{config.mute, "Mute", "muted", "true", tc.muted != nil && *tc.muted},
						{config.unmute, "UnMute", "muted", "false", tc.muted != nil && !*tc.muted},
					} {
						actions, count, err := command.evaluator.Evaluate(config.room, request, "test")
						if err != nil {
							t.Fatalf("%s: %v", command.action, err)
						}
						want := 0
						if command.requested {
							want = 1
						}
						if count != want || len(actions) != want {
							t.Fatalf("%s: expected %d actions, got count %d and actions %d", command.action, want, count, len(actions))
						}
						for _, action := range actions {
							if action.Action != command.action || action.Device.ID != config.targetID || action.DestinationDevice.Device.Name != config.deviceName {
								t.Fatalf("unexpected action or routing: %+v", action)
							}
							if config.input != "" && action.Parameters["input"] != config.input {
								t.Errorf("expected input %q, got %q", config.input, action.Parameters["input"])
							}
							value := command.value
							if command.action == "SetVolume" && *tc.volume == 0 {
								value = "0"
							}
							if len(action.EventLog) != 1 || action.EventLog[0].Key != command.key || action.EventLog[0].Value != value || action.EventLog[0].User != "test" {
								t.Errorf("unexpected %s events: %+v", command.action, action.EventLog)
							}
							if command.action == "SetVolume" && action.Parameters["level"] != value {
								t.Errorf("expected volume %s, got %q", value, action.Parameters["level"])
							}
						}
					}
				})
			}
		})
	}
}
