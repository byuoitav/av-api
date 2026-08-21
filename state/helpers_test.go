package state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/byuoitav/av-api/base"
	se "github.com/byuoitav/av-api/statusevaluators"
	"github.com/byuoitav/common/structs"
)

func TestIssueCommandsContinuesAfterNon200Response(t *testing.T) {
	var mu sync.Mutex
	seen := make(map[string]int)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen[r.URL.Path]++
		mu.Unlock()

		switch r.URL.Path {
		case "/blanked":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"unknown blanked state 'err_inactive'"}`))
		case "/power":
			_, _ = w.Write([]byte(`{"power":"off"}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	device := structs.Device{
		ID:      "TEST-RM-D1",
		Name:    "D1",
		Address: "127.0.0.1",
		Type: structs.DeviceType{
			Commands: []structs.Command{
				statusCommand("STATUS_BlankedDefault", server.URL, "/blanked"),
				statusCommand("STATUS_PowerDefault", server.URL, "/power"),
			},
		},
	}

	commands := []se.StatusCommand{
		statusTestCommand(device, "STATUS_BlankedDefault"),
		statusTestCommand(device, "STATUS_PowerDefault"),
	}

	channel := make(chan []se.StatusResponse, 1)
	var group sync.WaitGroup
	group.Add(1)
	issueCommands(context.Background(), commands, channel, &group)
	group.Wait()
	close(channel)

	var responses []se.StatusResponse
	for responseList := range channel {
		responses = append(responses, responseList...)
	}

	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if responses[0].ErrorMessage == nil {
		t.Fatal("expected first response to contain the non-200 error")
	}
	if responses[1].ErrorMessage != nil {
		t.Fatalf("expected second response to succeed, got error: %s", *responses[1].ErrorMessage)
	}
	if responses[1].Status["power"] != "off" {
		t.Fatalf("expected power response to be recorded, got %#v", responses[1].Status)
	}

	mu.Lock()
	defer mu.Unlock()
	if seen["/blanked"] != 1 {
		t.Fatalf("expected /blanked to be requested once, got %d", seen["/blanked"])
	}
	if seen["/power"] != 1 {
		t.Fatalf("expected /power to be requested once, got %d", seen["/power"])
	}
}

func statusCommand(id, microserviceURL, path string) structs.Command {
	return structs.Command{
		ID: id,
		Microservice: structs.Microservice{
			Address: microserviceURL,
		},
		Endpoint: structs.Endpoint{
			Path: path,
		},
	}
}

func statusTestCommand(device structs.Device, actionID string) se.StatusCommand {
	return se.StatusCommand{
		Action:            device.GetCommandByID(actionID),
		Device:            device,
		DestinationDevice: base.DestinationDevice{Device: device, Display: true},
		Generator:         actionID,
	}
}
