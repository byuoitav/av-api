package state

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/byuoitav/av-api/base"
)

func TestSetRoomStateRunnerExecutesQueuedRequestsInOrder(t *testing.T) {
	originalSetRoomStateWithContext := setRoomStateWithContext
	defer func() {
		setRoomStateWithContext = originalSetRoomStateWithContext
	}()

	started := make(chan string, 3)
	releaseFirst := make(chan struct{})
	var mu sync.Mutex
	var executed []string

	setRoomStateWithContext = func(ctx context.Context, target base.PublicRoom, requestor string) (base.PublicRoom, error) {
		started <- target.CurrentVideoInput
		if target.CurrentVideoInput == "first" {
			<-releaseFirst
		}

		mu.Lock()
		executed = append(executed, target.CurrentVideoInput)
		mu.Unlock()

		return target, nil
	}

	runner := &setRoomStateRunner{}
	first := newSetRoomStateTestJob("first")
	second := newSetRoomStateTestJob("second")
	third := newSetRoomStateTestJob("third")

	runner.submit(first)
	if got := <-started; got != "first" {
		t.Fatalf("expected first request to start first, got %q", got)
	}

	runner.submit(second)
	runner.submit(third)
	close(releaseFirst)

	waitForSetRoomStateTestJob(t, first)
	waitForSetRoomStateTestJob(t, second)
	waitForSetRoomStateTestJob(t, third)

	mu.Lock()
	defer mu.Unlock()
	want := []string{"first", "second", "third"}
	if !reflect.DeepEqual(executed, want) {
		t.Fatalf("expected queued requests to execute in order %v, got %v", want, executed)
	}
}

func newSetRoomStateTestJob(input string) *setRoomStateJob {
	ctx, cancel := context.WithCancel(context.Background())
	return &setRoomStateJob{
		ctx:       ctx,
		cancel:    cancel,
		target:    base.PublicRoom{Building: "B", Room: "R", CurrentVideoInput: input},
		requestor: "test",
		done:      make(chan setRoomStateResult, 1),
	}
}

func waitForSetRoomStateTestJob(t *testing.T, job *setRoomStateJob) setRoomStateResult {
	t.Helper()

	select {
	case result := <-job.done:
		return result
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %q", job.target.CurrentVideoInput)
		return setRoomStateResult{}
	}
}
