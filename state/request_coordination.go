package state

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/byuoitav/av-api/base"
)

var ErrSuperseded = errors.New("room state request superseded by a newer request")

const setRoomStateExecutionTimeout = 2 * time.Minute

var setRoomStateWithContext = SetRoomStateWithContext

type roomStateCacheEntry struct {
	status    base.PublicRoom
	err       error
	expiresAt time.Time
}

type roomStateInflight struct {
	done   chan struct{}
	status base.PublicRoom
	err    error
}

var roomStateRequests = struct {
	sync.Mutex
	cache   map[string]roomStateCacheEntry
	running map[string]*roomStateInflight
}{
	cache:   make(map[string]roomStateCacheEntry),
	running: make(map[string]*roomStateInflight),
}

func GetRoomStateShared(ctx context.Context, building string, roomName string, cacheTTL time.Duration, timeout time.Duration) (base.PublicRoom, error) {
	key := roomKey(building, roomName)
	now := time.Now()

	roomStateRequests.Lock()
	if cached, ok := roomStateRequests.cache[key]; ok && now.Before(cached.expiresAt) {
		roomStateRequests.Unlock()
		return cached.status, cached.err
	}

	if running, ok := roomStateRequests.running[key]; ok {
		roomStateRequests.Unlock()
		return waitForRoomState(ctx, running)
	}

	running := &roomStateInflight{
		done: make(chan struct{}),
	}
	roomStateRequests.running[key] = running
	roomStateRequests.Unlock()

	go func() {
		runCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		status, err := GetRoomStateWithContext(runCtx, building, roomName)

		roomStateRequests.Lock()
		running.status = status
		running.err = err
		delete(roomStateRequests.running, key)
		if err == nil && cacheTTL > 0 {
			roomStateRequests.cache[key] = roomStateCacheEntry{
				status:    status,
				err:       err,
				expiresAt: time.Now().Add(cacheTTL),
			}
		}
		close(running.done)
		roomStateRequests.Unlock()
	}()

	return waitForRoomState(ctx, running)
}

func waitForRoomState(ctx context.Context, running *roomStateInflight) (base.PublicRoom, error) {
	select {
	case <-running.done:
		return running.status, running.err
	case <-ctx.Done():
		return base.PublicRoom{}, ctx.Err()
	}
}

type setRoomStateResult struct {
	status base.PublicRoom
	err    error
}

type setRoomStateJob struct {
	ctx       context.Context
	cancel    context.CancelFunc
	target    base.PublicRoom
	requestor string
	done      chan setRoomStateResult
	once      sync.Once
}

func (j *setRoomStateJob) finish(status base.PublicRoom, err error) {
	j.once.Do(func() {
		j.done <- setRoomStateResult{status: status, err: err}
		close(j.done)
	})
}

type setRoomStateRunner struct {
	mu      sync.Mutex
	active  *setRoomStateJob
	queued  []*setRoomStateJob
	running bool
}

var setRoomStateRequests = struct {
	sync.Mutex
	rooms map[string]*setRoomStateRunner
}{
	rooms: make(map[string]*setRoomStateRunner),
}

func SetRoomStateLatest(ctx context.Context, target base.PublicRoom, requestor string) (base.PublicRoom, error) {
	key := roomKey(target.Building, target.Room)
	runner := getSetRoomStateRunner(key)

	jobCtx, cancel := context.WithTimeout(context.Background(), setRoomStateExecutionTimeout)
	job := &setRoomStateJob{
		ctx:       jobCtx,
		cancel:    cancel,
		target:    target,
		requestor: requestor,
		done:      make(chan setRoomStateResult, 1),
	}

	runner.submit(job)

	select {
	case result := <-job.done:
		cancel()
		return result.status, result.err
	case <-ctx.Done():
		return base.PublicRoom{}, ctx.Err()
	}
}

func getSetRoomStateRunner(key string) *setRoomStateRunner {
	setRoomStateRequests.Lock()
	defer setRoomStateRequests.Unlock()

	runner, ok := setRoomStateRequests.rooms[key]
	if !ok {
		runner = &setRoomStateRunner{}
		setRoomStateRequests.rooms[key] = runner
	}

	return runner
}

func (r *setRoomStateRunner) submit(job *setRoomStateJob) {
	r.mu.Lock()
	defer r.mu.Unlock()

	invalidateRoomStateCache(roomKey(job.target.Building, job.target.Room))

	r.queued = append(r.queued, job)
	if !r.running {
		r.running = true
		go r.run()
	}
}

func (r *setRoomStateRunner) run() {
	for {
		r.mu.Lock()
		var job *setRoomStateJob
		if len(r.queued) > 0 {
			job = r.queued[0]
			r.queued[0] = nil
			r.queued = r.queued[1:]
		}
		r.active = job
		if job == nil {
			r.active = nil
			r.running = false
			r.mu.Unlock()
			return
		}
		r.mu.Unlock()

		status, err := setRoomStateWithContext(job.ctx, job.target, job.requestor)
		if errors.Is(err, context.Canceled) {
			err = ErrSuperseded
		}
		job.finish(status, err)

		r.mu.Lock()
		if r.active == job {
			r.active = nil
		}
		r.mu.Unlock()
	}
}

func roomKey(building string, roomName string) string {
	return fmt.Sprintf("%s-%s", building, roomName)
}

func invalidateRoomStateCache(key string) {
	roomStateRequests.Lock()
	delete(roomStateRequests.cache, key)
	roomStateRequests.Unlock()
}
