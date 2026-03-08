package manager

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/config"
	"github.com/tarper24/logi-sim-leds/pkg/core"
)

// --- Mocks ---

type mockDevice struct {
	name      string
	id        string
	connected bool
	lastLEDs  core.TelemetryData
}

func (d *mockDevice) GetName() string        { return d.name }
func (d *mockDevice) GetID() string          { return d.id }
func (d *mockDevice) IsConnected() bool      { return d.connected }
func (d *mockDevice) Connect() error         { d.connected = true; return nil }
func (d *mockDevice) Disconnect() error      { d.connected = false; return nil }
func (d *mockDevice) SetLEDMask(mask uint8) error {
	if !d.connected {
		return fmt.Errorf("device not connected")
	}
	return nil
}
func (d *mockDevice) UpdateLEDs(data core.TelemetryData) error {
	if !d.connected {
		return fmt.Errorf("device not connected")
	}
	d.lastLEDs = data
	return nil
}

type mockGame struct {
	name    string
	port    int
	running bool
	maxRPM  float32
}

func (g *mockGame) GetName() string  { return g.name }
func (g *mockGame) GetPort() int     { return g.port }
func (g *mockGame) IsRunning() bool  { return g.running }
func (g *mockGame) SetMaxRPM(rpm float32) { g.maxRPM = rpm }
func (g *mockGame) Start(ctx context.Context, ch chan<- core.TelemetryData) error {
	g.running = true
	return nil
}
func (g *mockGame) Stop() error {
	g.running = false
	return nil
}

type mockDetector struct {
	devices []core.DeviceInterface
}

func (d *mockDetector) Detect() ([]core.DeviceInterface, error) {
	return d.devices, nil
}
func (d *mockDetector) Watch(ctx context.Context, ch chan<- core.DeviceEvent) error {
	<-ctx.Done()
	return ctx.Err()
}

// --- Helper ---

func newTestManager(games []core.GameInterface, detector core.DeviceDetector) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		ctx:              ctx,
		cancel:           cancel,
		games:            games,
		deviceDetector:   detector,
		telemetryChan:    make(chan core.TelemetryData, 100),
		deviceEventChan:  make(chan core.DeviceEvent, 10),
		enableAutoDetect: true,
		uiTelemetryChan:  make(chan core.TelemetryData, 100),
		uiDeviceChan:     make(chan string, 10),
		uiGameChan:       make(chan string, 10),
	}
}

// --- Tests ---

func TestNewManager(t *testing.T) {
	cfg := config.Default()
	m := NewManager(cfg)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if len(m.games) != 3 {
		t.Errorf("expected 3 games, got %d", len(m.games))
	}
	if !m.enableAutoDetect {
		t.Error("expected auto-detect to be enabled")
	}
	if m.telemetryChan == nil {
		t.Error("telemetryChan is nil")
	}
	if m.deviceEventChan == nil {
		t.Error("deviceEventChan is nil")
	}
	if m.uiTelemetryChan == nil {
		t.Error("uiTelemetryChan is nil")
	}
	if m.uiDeviceChan == nil {
		t.Error("uiDeviceChan is nil")
	}
	if m.uiGameChan == nil {
		t.Error("uiGameChan is nil")
	}
}

func TestDone(t *testing.T) {
	m := newTestManager(nil, &mockDetector{})
	done := m.Done()

	select {
	case <-done:
		t.Fatal("Done channel should not be closed yet")
	default:
	}

	m.cancel()

	select {
	case <-done:
		// expected
	case <-time.After(time.Second):
		t.Fatal("Done channel should be closed after cancel")
	}
}

func TestGetAvailableGames(t *testing.T) {
	games := []core.GameInterface{
		&mockGame{name: "GameA"},
		&mockGame{name: "GameB"},
	}
	m := newTestManager(games, &mockDetector{})
	names := m.GetAvailableGames()
	if len(names) != 2 {
		t.Fatalf("expected 2 games, got %d", len(names))
	}
	if names[0] != "GameA" || names[1] != "GameB" {
		t.Errorf("unexpected game names: %v", names)
	}
}

func TestGetStatus_NoDeviceNoGame(t *testing.T) {
	m := newTestManager(nil, &mockDetector{})
	status := m.GetStatus()
	if !strings.Contains(status, "None") {
		t.Errorf("expected status to contain 'None', got %q", status)
	}
	if !strings.Contains(status, "waiting for device") {
		t.Errorf("expected 'waiting for device' in status, got %q", status)
	}
	if !strings.Contains(status, "waiting for game data") {
		t.Errorf("expected 'waiting for game data' in status, got %q", status)
	}
}

func TestSetMaxRPM_NoActiveGame(t *testing.T) {
	m := newTestManager(nil, &mockDetector{})
	err := m.SetMaxRPM(8000)
	if err == nil {
		t.Fatal("expected error when no active game")
	}
	if !strings.Contains(err.Error(), "no active game") {
		t.Errorf("expected 'no active game' error, got %q", err.Error())
	}
}

func TestProcessTelemetry_SourceRouting(t *testing.T) {
	game := &mockGame{name: "TestGame"}
	games := []core.GameInterface{game}
	m := newTestManager(games, &mockDetector{})

	// Start telemetry processing
	go m.processTelemetry()

	// Send telemetry data
	m.telemetryChan <- core.TelemetryData{
		Source:    "TestGame",
		RPM:      5000,
		MaxRPM:   8000,
		Timestamp: time.Now(),
	}

	// Wait for game name to appear on UI channel
	select {
	case gameName := <-m.uiGameChan:
		if gameName != "TestGame" {
			t.Errorf("expected game name 'TestGame', got %q", gameName)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for game name on UI channel")
	}

	// Verify active game was set
	m.mu.RLock()
	ag := m.activeGame
	m.mu.RUnlock()
	if ag == nil {
		t.Fatal("expected activeGame to be set")
	}
	if ag.GetName() != "TestGame" {
		t.Errorf("expected activeGame 'TestGame', got %q", ag.GetName())
	}

	// Cleanup
	m.cancel()
}
