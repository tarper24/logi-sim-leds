package main

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

func TestResolveLogPath(t *testing.T) {
	path := resolveLogPath()
	if !strings.HasSuffix(path, "logi-sim-leds.log") {
		t.Errorf("expected path ending in logi-sim-leds.log, got %q", path)
	}
}

func TestStartUILoop_TelemetryUpdate(t *testing.T) {
	done := make(chan struct{})
	telemetryChan := make(chan core.TelemetryData, 1)
	deviceChan := make(chan string)
	gameChan := make(chan string)

	var mu sync.Mutex
	var got core.TelemetryData
	called := false

	updater := uiUpdater{
		onTelemetry: func(data core.TelemetryData) {
			mu.Lock()
			got = data
			called = true
			mu.Unlock()
		},
		onDeviceChange: func(string) {},
		onGameChange:   func(string) {},
		getDevices:     func() []string { return nil },
	}

	go startUILoop(done, telemetryChan, deviceChan, gameChan, updater)

	sent := core.TelemetryData{RPM: 5000, MaxRPM: 8000, Source: "test"}
	telemetryChan <- sent
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !called {
		t.Fatal("onTelemetry was not called")
	}
	if got.RPM != sent.RPM || got.MaxRPM != sent.MaxRPM || got.Source != sent.Source {
		t.Errorf("telemetry mismatch: got %+v, want %+v", got, sent)
	}
	close(done)
}

func TestStartUILoop_DeviceChange(t *testing.T) {
	done := make(chan struct{})
	telemetryChan := make(chan core.TelemetryData)
	deviceChan := make(chan string, 1)
	gameChan := make(chan string)

	var mu sync.Mutex
	deviceCalled := false
	getDevicesCalled := false
	var gotDevice string

	updater := uiUpdater{
		onTelemetry: func(core.TelemetryData) {},
		onDeviceChange: func(name string) {
			mu.Lock()
			deviceCalled = true
			gotDevice = name
			mu.Unlock()
		},
		onGameChange: func(string) {},
		getDevices: func() []string {
			mu.Lock()
			getDevicesCalled = true
			mu.Unlock()
			return nil
		},
	}

	go startUILoop(done, telemetryChan, deviceChan, gameChan, updater)

	deviceChan <- "G29"
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !deviceCalled {
		t.Fatal("onDeviceChange was not called")
	}
	if gotDevice != "G29" {
		t.Errorf("got device %q, want %q", gotDevice, "G29")
	}
	if !getDevicesCalled {
		t.Fatal("getDevices was not called")
	}
	close(done)
}

func TestStartUILoop_DeviceChange_Empty(t *testing.T) {
	done := make(chan struct{})
	telemetryChan := make(chan core.TelemetryData)
	deviceChan := make(chan string, 1)
	gameChan := make(chan string)

	var mu sync.Mutex
	deviceCalled := false
	getDevicesCalled := false

	updater := uiUpdater{
		onTelemetry: func(core.TelemetryData) {},
		onDeviceChange: func(string) {
			mu.Lock()
			deviceCalled = true
			mu.Unlock()
		},
		onGameChange: func(string) {},
		getDevices: func() []string {
			mu.Lock()
			getDevicesCalled = true
			mu.Unlock()
			return nil
		},
	}

	go startUILoop(done, telemetryChan, deviceChan, gameChan, updater)

	deviceChan <- ""
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if deviceCalled {
		t.Fatal("onDeviceChange should NOT be called for empty string")
	}
	if !getDevicesCalled {
		t.Fatal("getDevices should still be called")
	}
	close(done)
}

func TestStartUILoop_GameChange(t *testing.T) {
	done := make(chan struct{})
	telemetryChan := make(chan core.TelemetryData)
	deviceChan := make(chan string)
	gameChan := make(chan string, 1)

	var mu sync.Mutex
	called := false
	var gotGame string

	updater := uiUpdater{
		onTelemetry:    func(core.TelemetryData) {},
		onDeviceChange: func(string) {},
		onGameChange: func(name string) {
			mu.Lock()
			called = true
			gotGame = name
			mu.Unlock()
		},
		getDevices: func() []string { return nil },
	}

	go startUILoop(done, telemetryChan, deviceChan, gameChan, updater)

	gameChan <- "BeamNG"
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !called {
		t.Fatal("onGameChange was not called")
	}
	if gotGame != "BeamNG" {
		t.Errorf("got game %q, want %q", gotGame, "BeamNG")
	}
	close(done)
}

func TestStartUILoop_DoneExits(t *testing.T) {
	done := make(chan struct{})
	telemetryChan := make(chan core.TelemetryData)
	deviceChan := make(chan string)
	gameChan := make(chan string)

	updater := uiUpdater{
		onTelemetry:    func(core.TelemetryData) {},
		onDeviceChange: func(string) {},
		onGameChange:   func(string) {},
		getDevices:     func() []string { return nil },
	}

	exited := make(chan struct{})
	go func() {
		startUILoop(done, telemetryChan, deviceChan, gameChan, updater)
		close(exited)
	}()

	close(done)

	select {
	case <-exited:
		// success
	case <-time.After(time.Second):
		t.Fatal("startUILoop did not exit after done was closed")
	}
}

func TestVersion(t *testing.T) {
	if version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %q", version)
	}
}
