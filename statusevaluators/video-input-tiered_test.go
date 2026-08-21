package statusevaluators

import (
	"testing"
	"time"

	"github.com/byuoitav/av-api/base"
	"github.com/byuoitav/common/structs"
)

func TestStartAggregatorExitsWithNoCallbacks(t *testing.T) {
	restoreTimeouts := useTieredSwitcherTimeouts(20*time.Millisecond, 100*time.Millisecond)
	defer restoreTimeouts()

	callback := &TieredSwitcherCallback{
		InChan:              make(chan base.StatusPackage, 1),
		ExpectedActionCount: 1,
	}
	callback.SetDevices([]structs.Device{testOutputDevice()})

	done := make(chan struct{})
	go func() {
		callback.StartAggregator()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("StartAggregator did not exit after the overall timeout")
	}
}

func TestStartAggregatorExitsAfterCallbackSettleTimeout(t *testing.T) {
	restoreTimeouts := useTieredSwitcherTimeouts(20*time.Millisecond, 500*time.Millisecond)
	defer restoreTimeouts()

	outputDevice := testOutputDevice()
	inputDevice := testInputDevice()
	outputDevice.Ports = []structs.Port{
		{
			ID:                "HDMI1",
			SourceDevice:      inputDevice.ID,
			DestinationDevice: outputDevice.ID,
		},
	}

	callback := &TieredSwitcherCallback{
		InChan:              make(chan base.StatusPackage, 1),
		ExpectedActionCount: 2,
	}
	callback.SetDevices([]structs.Device{inputDevice, outputDevice})

	out := make(chan base.StatusPackage, 1)
	done := make(chan struct{})
	go func() {
		callback.StartAggregator()
		close(done)
	}()

	if err := callback.Callback(base.StatusPackage{
		Device: outputDevice,
		Value:  "HDMI1",
	}, out); err != nil {
		t.Fatalf("Callback returned error: %v", err)
	}

	select {
	case val := <-out:
		if val.Key != "input" {
			t.Fatalf("expected input key, got %q", val.Key)
		}
		if val.Value != inputDevice.Name {
			t.Fatalf("expected input value %q, got %v", inputDevice.Name, val.Value)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("StartAggregator did not publish input paths after settle timeout")
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("StartAggregator did not exit after publishing input paths")
	}
}

func TestStartAggregatorIgnoresUnexpectedCallbackValueType(t *testing.T) {
	restoreTimeouts := useTieredSwitcherTimeouts(20*time.Millisecond, 500*time.Millisecond)
	defer restoreTimeouts()

	callback := &TieredSwitcherCallback{
		InChan:              make(chan base.StatusPackage, 1),
		ExpectedActionCount: 1,
	}
	callback.SetDevices([]structs.Device{testOutputDevice()})

	out := make(chan base.StatusPackage, 1)
	done := make(chan struct{})
	go func() {
		callback.StartAggregator()
		close(done)
	}()

	if err := callback.Callback(base.StatusPackage{
		Device: testOutputDevice(),
		Value:  123,
	}, out); err != nil {
		t.Fatalf("Callback returned error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("StartAggregator did not exit after receiving an unexpected callback value type")
	}
}

func TestCallbackDoesNotBlockAfterAggregatorExits(t *testing.T) {
	restoreTimeouts := useTieredSwitcherTimeouts(20*time.Millisecond, 50*time.Millisecond)
	defer restoreTimeouts()

	callback := &TieredSwitcherCallback{
		InChan:              make(chan base.StatusPackage),
		ExpectedActionCount: 1,
	}
	callback.SetDevices([]structs.Device{testOutputDevice()})

	done := make(chan struct{})
	go func() {
		callback.StartAggregator()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("StartAggregator did not exit after the overall timeout")
	}

	callbackDone := make(chan struct{})
	go func() {
		_ = callback.Callback(base.StatusPackage{
			Device: testOutputDevice(),
			Value:  "HDMI1",
		}, make(chan base.StatusPackage))
		close(callbackDone)
	}()

	select {
	case <-callbackDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Callback blocked after aggregator exit")
	}
}

func useTieredSwitcherTimeouts(settle, overall time.Duration) func() {
	oldSettle := tieredSwitcherSettleWindow
	oldOverall := tieredSwitcherOverallTimeout
	tieredSwitcherSettleWindow = settle
	tieredSwitcherOverallTimeout = overall

	return func() {
		tieredSwitcherSettleWindow = oldSettle
		tieredSwitcherOverallTimeout = oldOverall
	}
}

func testInputDevice() structs.Device {
	return structs.Device{
		ID:   "ITB-1001-CP1",
		Name: "Laptop",
		Type: structs.DeviceType{
			Input: true,
		},
	}
}

func testOutputDevice() structs.Device {
	return structs.Device{
		ID:   "ITB-1001-D1",
		Name: "Display",
		Type: structs.DeviceType{
			Output: true,
		},
		Roles: []structs.Role{
			{ID: "VideoOut"},
		},
	}
}
