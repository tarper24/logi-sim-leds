package core_test

import (
	"testing"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/core"
	"github.com/tarper24/logi-sim-leds/pkg/devices/logitech"
	"github.com/tarper24/logi-sim-leds/pkg/games/assettocorsa"
	"github.com/tarper24/logi-sim-leds/pkg/games/beamng"
	"github.com/tarper24/logi-sim-leds/pkg/games/dirt"
)

// Compile-time interface satisfaction checks
var _ core.DeviceInterface = (*logitech.G29)(nil)
var _ core.DeviceInterface = (*logitech.G920)(nil)
var _ core.DeviceInterface = (*logitech.G923)(nil)
var _ core.GameInterface = (*beamng.BeamNG)(nil)
var _ core.GameInterface = (*assettocorsa.AssettoCorsa)(nil)
var _ core.GameInterface = (*dirt.Dirt)(nil)
var _ core.DeviceDetector = (*logitech.Detector)(nil)

func TestDeviceInterfaceConformance(t *testing.T) {
	// The compile-time checks above ensure these types satisfy DeviceInterface.
	// This test exists so the check is visible in test output.
	t.Log("G29, G920, G923 all satisfy core.DeviceInterface")
}

func TestGameInterfaceConformance(t *testing.T) {
	t.Log("BeamNG, AssettoCorsa, Dirt all satisfy core.GameInterface")
}

func TestDeviceDetectorConformance(t *testing.T) {
	t.Log("logitech.Detector satisfies core.DeviceDetector")
}

func TestTelemetryDataFields(t *testing.T) {
	now := time.Now()
	td := core.TelemetryData{
		RPM:       5500.0,
		MaxRPM:    8000.0,
		Speed:     120.5,
		Gear:      3,
		Source:    "test",
		Timestamp: now,
	}

	if td.RPM != 5500.0 {
		t.Errorf("RPM = %v, want 5500", td.RPM)
	}
	if td.MaxRPM != 8000.0 {
		t.Errorf("MaxRPM = %v, want 8000", td.MaxRPM)
	}
	if td.Speed != 120.5 {
		t.Errorf("Speed = %v, want 120.5", td.Speed)
	}
	if td.Gear != 3 {
		t.Errorf("Gear = %v, want 3", td.Gear)
	}
	if td.Source != "test" {
		t.Errorf("Source = %v, want test", td.Source)
	}
	if !td.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", td.Timestamp, now)
	}
}

func TestDeviceEventFields(t *testing.T) {
	now := time.Now()
	ev := core.DeviceEvent{
		Device:    nil,
		Connected: true,
		Timestamp: now,
	}

	if !ev.Connected {
		t.Error("Connected = false, want true")
	}
	if !ev.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", ev.Timestamp, now)
	}
	if ev.Device != nil {
		t.Error("Device should be nil")
	}
}
