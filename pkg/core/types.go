package core

import (
	"context"
	"time"
)

// TelemetryData represents the normalized telemetry data from any game
type TelemetryData struct {
	RPM        float32   // Current engine RPM
	MaxRPM     float32   // Maximum engine RPM
	Speed      float32   // Speed in km/h
	Gear       int8      // Current gear (-1 for reverse, 0 for neutral)
	Timestamp  time.Time // When this data was received
}

// DeviceInterface represents a racing wheel or other output device
type DeviceInterface interface {
	// GetName returns the device name (e.g., "Logitech G29")
	GetName() string

	// GetID returns a unique identifier for the device
	GetID() string

	// Connect establishes connection to the device
	Connect() error

	// Disconnect closes the connection to the device
	Disconnect() error

	// IsConnected returns true if the device is currently connected
	IsConnected() bool

	// UpdateLEDs sets the LED state based on telemetry data
	UpdateLEDs(data TelemetryData) error

	// SetLEDMask directly sets the LED bitmask (for advanced control)
	SetLEDMask(mask uint8) error
}

// GameInterface represents a racing game telemetry source
type GameInterface interface {
	// GetName returns the game name (e.g., "BeamNG.drive")
	GetName() string

	// Start begins listening for telemetry data
	Start(ctx context.Context, dataChan chan<- TelemetryData) error

	// Stop stops listening for telemetry data
	Stop() error

	// IsRunning returns true if the game is currently sending data
	IsRunning() bool

	// GetPort returns the UDP port this game uses
	GetPort() int

	// SetMaxRPM manually overrides the maximum RPM used for LED calculation.
	// The game's auto-detect will still raise this value if actual RPM exceeds it.
	SetMaxRPM(rpm float32)
}

// DeviceDetector finds available racing wheel devices
type DeviceDetector interface {
	// Detect scans for available devices
	Detect() ([]DeviceInterface, error)

	// Watch continuously monitors for device connection/disconnection
	Watch(ctx context.Context, deviceChan chan<- DeviceEvent) error
}

// DeviceEvent represents a device connection or disconnection event
type DeviceEvent struct {
	Device    DeviceInterface
	Connected bool // true for connection, false for disconnection
	Timestamp time.Time
}
