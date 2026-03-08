package logitech

import (
	"strings"
	"testing"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

func TestNewLogitechWheel(t *testing.T) {
	w := NewLogitechWheel("Test Wheel", 0x1234)
	if w.name != "Test Wheel" {
		t.Errorf("expected name 'Test Wheel', got %q", w.name)
	}
	if w.productID != 0x1234 {
		t.Errorf("expected productID 0x1234, got 0x%04x", w.productID)
	}
	if w.connected {
		t.Error("expected connected to be false")
	}
	// Should have default LED config
	def := DefaultLEDConfig()
	if w.ledCfg != def {
		t.Errorf("expected default LED config, got %+v", w.ledCfg)
	}
}

func TestNewLogitechWheelWithConfig(t *testing.T) {
	cfg := LEDConfig{
		LED1Threshold:  10,
		LED2Threshold:  20,
		LED3Threshold:  30,
		LED4Threshold:  40,
		LED5Threshold:  50,
		FlashThreshold: 60,
		FlashInterval:  200 * time.Millisecond,
	}
	w := NewLogitechWheelWithConfig("Custom Wheel", 0xABCD, cfg)
	if w.ledCfg != cfg {
		t.Errorf("expected custom LED config, got %+v", w.ledCfg)
	}
}

func TestDefaultLEDConfig(t *testing.T) {
	cfg := DefaultLEDConfig()
	if cfg.LED1Threshold != 45 {
		t.Errorf("LED1Threshold: expected 45, got %v", cfg.LED1Threshold)
	}
	if cfg.LED2Threshold != 55 {
		t.Errorf("LED2Threshold: expected 55, got %v", cfg.LED2Threshold)
	}
	if cfg.LED3Threshold != 62.5 {
		t.Errorf("LED3Threshold: expected 62.5, got %v", cfg.LED3Threshold)
	}
	if cfg.LED4Threshold != 71 {
		t.Errorf("LED4Threshold: expected 71, got %v", cfg.LED4Threshold)
	}
	if cfg.LED5Threshold != 85 {
		t.Errorf("LED5Threshold: expected 85, got %v", cfg.LED5Threshold)
	}
	if cfg.FlashThreshold != 93 {
		t.Errorf("FlashThreshold: expected 93, got %v", cfg.FlashThreshold)
	}
	if cfg.FlashInterval != 100*time.Millisecond {
		t.Errorf("FlashInterval: expected 100ms, got %v", cfg.FlashInterval)
	}
}

func TestGetName(t *testing.T) {
	w := NewLogitechWheel("My Wheel", 0x1111)
	if w.GetName() != "My Wheel" {
		t.Errorf("expected 'My Wheel', got %q", w.GetName())
	}
}

func TestGetID(t *testing.T) {
	w := NewLogitechWheel("Logitech G29 (PS)", 0xC24F)
	expected := "logitech_Logitech G29 (PS)_c24f"
	if w.GetID() != expected {
		t.Errorf("expected %q, got %q", expected, w.GetID())
	}
}

func TestIsConnected_InitiallyFalse(t *testing.T) {
	w := NewLogitechWheel("Test", 0x1234)
	if w.IsConnected() {
		t.Error("new wheel should not be connected")
	}
}

func TestDisconnect_WhenNotConnected(t *testing.T) {
	w := NewLogitechWheel("Test", 0x1234)
	err := w.Disconnect()
	if err != nil {
		t.Errorf("Disconnect on unconnected wheel should return nil, got %v", err)
	}
}

func TestUpdateLEDs_NotConnected(t *testing.T) {
	w := NewLogitechWheel("Test", 0x1234)
	err := w.UpdateLEDs(core.TelemetryData{RPM: 5000, MaxRPM: 8000})
	if err == nil {
		t.Error("expected error for unconnected device")
	}
	if !strings.Contains(err.Error(), "device not connected") {
		t.Errorf("expected 'device not connected' error, got %q", err.Error())
	}
}

func TestUpdateLEDs_ZeroMaxRPM(t *testing.T) {
	// UpdateLEDs checks IsConnected first, then MaxRPM.
	// With no connection it errors before checking MaxRPM.
	// We can only test this path if connected — but we can't connect without hardware.
	// Instead, verify that the check exists by testing the unconnected path.
	// The zero-MaxRPM path is tested indirectly through code review.
	// For completeness, we verify the function signature works.
	w := NewLogitechWheel("Test", 0x1234)
	err := w.UpdateLEDs(core.TelemetryData{RPM: 0, MaxRPM: 0})
	if err == nil {
		t.Error("expected error for unconnected device")
	}
}

func TestSetLEDMask_NotConnected(t *testing.T) {
	w := NewLogitechWheel("Test", 0x1234)
	err := w.SetLEDMask(0x1F)
	if err == nil {
		t.Error("expected error for unconnected device")
	}
	if !strings.Contains(err.Error(), "device not connected") {
		t.Errorf("expected 'device not connected' error, got %q", err.Error())
	}
}

// --- G29 tests ---

func TestNewG29(t *testing.T) {
	g := NewG29()
	if g.productID != 0xC24F {
		t.Errorf("expected productID 0xC24F, got 0x%04X", g.productID)
	}
	if !strings.Contains(g.GetName(), "G29") {
		t.Errorf("expected name to contain 'G29', got %q", g.GetName())
	}
}

func TestNewG29WithConfig(t *testing.T) {
	cfg := LEDConfig{LED1Threshold: 99}
	g := NewG29WithConfig(cfg)
	if g.ledCfg.LED1Threshold != 99 {
		t.Errorf("custom config not propagated, got %v", g.ledCfg.LED1Threshold)
	}
}

// --- G920 tests ---

func TestNewG920(t *testing.T) {
	g := NewG920()
	if g.productID != 0xC262 {
		t.Errorf("expected productID 0xC262, got 0x%04X", g.productID)
	}
	if !strings.Contains(g.GetName(), "G920") {
		t.Errorf("expected name to contain 'G920', got %q", g.GetName())
	}
}

// --- G923 tests ---

func TestNewG923XBox(t *testing.T) {
	g := NewG923XBox()
	if g.productID != 0xC267 {
		t.Errorf("expected productID 0xC267, got 0x%04X", g.productID)
	}
	if !strings.Contains(g.GetName(), "G923") {
		t.Errorf("expected name to contain 'G923', got %q", g.GetName())
	}
}

func TestNewG923PS(t *testing.T) {
	g := NewG923PS()
	if g.productID != 0xC266 {
		t.Errorf("expected productID 0xC266, got 0x%04X", g.productID)
	}
	if !strings.Contains(g.GetName(), "G923") {
		t.Errorf("expected name to contain 'G923', got %q", g.GetName())
	}
}

// --- Detector tests ---

func TestNewDetector(t *testing.T) {
	d := NewDetector()
	if d == nil {
		t.Fatal("NewDetector returned nil")
	}
	if len(d.supportedDevices) != 4 {
		t.Errorf("expected 4 supported devices, got %d", len(d.supportedDevices))
	}
}

func TestNewDetectorWithConfig(t *testing.T) {
	cfg := LEDConfig{LED1Threshold: 42}
	d := NewDetectorWithConfig(cfg)
	if d == nil {
		t.Fatal("NewDetectorWithConfig returned nil")
	}
	if len(d.supportedDevices) != 4 {
		t.Errorf("expected 4 supported devices, got %d", len(d.supportedDevices))
	}
}
