package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/core"
	"github.com/tarper24/logi-sim-leds/pkg/devices/logitech"
	"github.com/tarper24/logi-sim-leds/pkg/games/assettocorsa"
	"github.com/tarper24/logi-sim-leds/pkg/games/beamng"
	"github.com/tarper24/logi-sim-leds/pkg/games/dirt"
)

// Manager orchestrates the connection between games and devices
type Manager struct {
	ctx              context.Context
	cancel           context.CancelFunc
	games            []core.GameInterface
	deviceDetector   core.DeviceDetector
	activeDevice     core.DeviceInterface
	activeGame       core.GameInterface
	telemetryChan    chan core.TelemetryData
	deviceEventChan  chan core.DeviceEvent
	mu               sync.RWMutex
	enableAutoDetect bool
	// UI update channels
	uiTelemetryChan chan core.TelemetryData
	uiDeviceChan    chan string
	uiGameChan      chan string
}

// NewManager creates a new manager instance
func NewManager(enableAutoDetect bool) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize all supported games
	games := []core.GameInterface{
		beamng.NewBeamNG(),
		assettocorsa.NewAssettoCorsa(),
		dirt.NewDirt(),
	}

	return &Manager{
		ctx:              ctx,
		cancel:           cancel,
		games:            games,
		deviceDetector:   logitech.NewDetector(),
		telemetryChan:    make(chan core.TelemetryData, 100),
		deviceEventChan:  make(chan core.DeviceEvent, 10),
		enableAutoDetect: enableAutoDetect,
		uiTelemetryChan:  make(chan core.TelemetryData, 100),
		uiDeviceChan:     make(chan string, 10),
		uiGameChan:       make(chan string, 10),
	}
}

// Start begins the manager's operation
func (m *Manager) Start() error {
	fmt.Println("Starting logi-sim-leds manager...")

	// Detect initial devices
	if err := m.detectAndConnectDevice(); err != nil {
		fmt.Printf("Warning: No devices found initially: %v\n", err)
	}

	// Start device monitoring if auto-detect is enabled
	if m.enableAutoDetect {
		go m.monitorDevices()
	}

	// Start all game listeners
	for _, game := range m.games {
		if err := game.Start(m.ctx, m.telemetryChan); err != nil {
			fmt.Printf("Warning: Failed to start %s: %v\n", game.GetName(), err)
		}
	}

	// Start telemetry processing
	go m.processTelemetry()

	fmt.Println("Manager started successfully")
	return nil
}

// Stop stops the manager's operation
func (m *Manager) Stop() error {
	fmt.Println("Stopping logi-sim-leds manager...")

	// Cancel context to stop all goroutines
	m.cancel()

	// Stop all games
	for _, game := range m.games {
		if err := game.Stop(); err != nil {
			fmt.Printf("Warning: Error stopping %s: %v\n", game.GetName(), err)
		}
	}

	// Disconnect active device
	m.mu.Lock()
	if m.activeDevice != nil {
		if err := m.activeDevice.Disconnect(); err != nil {
			fmt.Printf("Warning: Error disconnecting device: %v\n", err)
		}
		m.activeDevice = nil
	}
	m.mu.Unlock()

	fmt.Println("Manager stopped")
	return nil
}

// detectAndConnectDevice scans for devices and connects to the first one found
func (m *Manager) detectAndConnectDevice() error {
	devices, err := m.deviceDetector.Detect()
	if err != nil {
		return fmt.Errorf("device detection failed: %w", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no compatible devices found")
	}

	// Connect to the first device found
	device := devices[0]
	if err := device.Connect(); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", device.GetName(), err)
	}

	m.mu.Lock()
	m.activeDevice = device
	m.mu.Unlock()

	fmt.Printf("Connected to device: %s\n", device.GetName())

	// Notify UI of device connection
	select {
	case m.uiDeviceChan <- device.GetName():
	default:
	}

	return nil
}

// monitorDevices watches for device connection/disconnection events
func (m *Manager) monitorDevices() {
	// Start device watcher
	go m.deviceDetector.Watch(m.ctx, m.deviceEventChan)

	for {
		select {
		case <-m.ctx.Done():
			return
		case event := <-m.deviceEventChan:
			m.handleDeviceEvent(event)
		}
	}
}

// handleDeviceEvent processes device connection/disconnection events
func (m *Manager) handleDeviceEvent(event core.DeviceEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if event.Connected {
		// Device connected
		fmt.Printf("Device connected: %s\n", event.Device.GetName())

		// If we don't have an active device, connect to this one
		if m.activeDevice == nil {
			if err := event.Device.Connect(); err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", event.Device.GetName(), err)
				return
			}
			m.activeDevice = event.Device
			fmt.Printf("Activated device: %s\n", event.Device.GetName())

			// Notify UI of device change
			select {
			case m.uiDeviceChan <- event.Device.GetName():
			default:
			}
		}
	} else {
		// Device disconnected
		fmt.Printf("Device disconnected: %s\n", event.Device.GetName())

		// If this was our active device, clear it
		if m.activeDevice != nil && m.activeDevice.GetID() == event.Device.GetID() {
			m.activeDevice.Disconnect()
			m.activeDevice = nil
			fmt.Println("Active device disconnected, waiting for new device...")

			// Notify UI of device disconnection
			select {
			case m.uiDeviceChan <- "":
			default:
			}

			// Try to reconnect to another device
			go func() {
				time.Sleep(1 * time.Second)
				if err := m.detectAndConnectDevice(); err != nil {
					fmt.Printf("No devices available: %v\n", err)
				}
			}()
		}
	}
}

// processTelemetry receives telemetry data and updates the active device
func (m *Manager) processTelemetry() {
	var lastGameName string

	for {
		select {
		case <-m.ctx.Done():
			return
		case data := <-m.telemetryChan:
			m.mu.Lock()
			device := m.activeDevice
			m.mu.Unlock()

			// Determine which game is sending data
			gameName := m.identifyGame()
			if gameName != lastGameName && gameName != "" {
				fmt.Printf("Receiving telemetry from: %s\n", gameName)
				lastGameName = gameName

				m.mu.Lock()
				m.activeGame = m.getGameByName(gameName)
				m.mu.Unlock()

				// Notify UI of game change
				select {
				case m.uiGameChan <- gameName:
				default:
				}
			}

			// Send telemetry to UI (non-blocking)
			select {
			case m.uiTelemetryChan <- data:
			default:
			}

			if device == nil {
				continue // No device connected
			}

			// Update device LEDs
			if err := device.UpdateLEDs(data); err != nil {
				fmt.Printf("LED update error: %v\n", err)
			}
		}
	}
}

// identifyGame attempts to identify which game is sending telemetry
func (m *Manager) identifyGame() string {
	// Check which games are currently running
	for _, game := range m.games {
		if game.IsRunning() {
			return game.GetName()
		}
	}
	return ""
}

// getGameByName returns a game interface by name
func (m *Manager) getGameByName(name string) core.GameInterface {
	for _, game := range m.games {
		if game.GetName() == name {
			return game
		}
	}
	return nil
}

// GetActiveDevice returns the currently active device
func (m *Manager) GetActiveDevice() core.DeviceInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeDevice
}

// GetActiveGame returns the currently active game
func (m *Manager) GetActiveGame() core.GameInterface {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeGame
}

// GetStatus returns the current status of the manager
func (m *Manager) GetStatus() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var status string

	if m.activeDevice != nil {
		status += fmt.Sprintf("Device: %s (connected)\n", m.activeDevice.GetName())
	} else {
		status += "Device: None (waiting for device...)\n"
	}

	if m.activeGame != nil {
		status += fmt.Sprintf("Game: %s (active)\n", m.activeGame.GetName())
	} else {
		status += "Game: None (waiting for game data...)\n"
	}

	return status
}

// GetUITelemetryChan returns the telemetry channel for UI updates
func (m *Manager) GetUITelemetryChan() <-chan core.TelemetryData {
	return m.uiTelemetryChan
}

// GetUIDeviceChan returns the device change channel for UI updates
func (m *Manager) GetUIDeviceChan() <-chan string {
	return m.uiDeviceChan
}

// GetUIGameChan returns the game change channel for UI updates
func (m *Manager) GetUIGameChan() <-chan string {
	return m.uiGameChan
}

// GetAvailableDevices returns a list of all currently available devices
func (m *Manager) GetAvailableDevices() []string {
	devices, err := m.deviceDetector.Detect()
	if err != nil {
		return []string{}
	}

	names := make([]string, len(devices))
	for i, device := range devices {
		names[i] = device.GetName()
	}
	return names
}

// GetAvailableGames returns a list of all supported games
func (m *Manager) GetAvailableGames() []string {
	names := make([]string, len(m.games))
	for i, game := range m.games {
		names[i] = game.GetName()
	}
	return names
}

// SetActiveDevice manually switches to a specific device by name
func (m *Manager) SetActiveDevice(name string) error {
	devices, err := m.deviceDetector.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect devices: %w", err)
	}

	// Find the device with the matching name
	var targetDevice core.DeviceInterface
	for _, device := range devices {
		if device.GetName() == name {
			targetDevice = device
			break
		}
	}

	if targetDevice == nil {
		return fmt.Errorf("device not found: %s", name)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Disconnect current device if any
	if m.activeDevice != nil {
		m.activeDevice.Disconnect()
	}

	// Connect to new device
	if err := targetDevice.Connect(); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", name, err)
	}

	m.activeDevice = targetDevice
	fmt.Printf("Manually switched to device: %s\n", name)

	// Notify UI
	select {
	case m.uiDeviceChan <- name:
	default:
	}

	return nil
}

// SetMaxRPM sets the max RPM on the active game. The game's auto-detect will
// still raise this value if actual RPM exceeds it.
func (m *Manager) SetMaxRPM(rpm float32) error {
	m.mu.RLock()
	game := m.activeGame
	m.mu.RUnlock()

	if game == nil {
		return fmt.Errorf("no active game")
	}

	game.SetMaxRPM(rpm)
	fmt.Printf("Set max RPM to: %.0f\n", rpm)
	return nil
}
