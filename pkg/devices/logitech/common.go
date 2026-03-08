package logitech

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/karalabe/hid"
	"github.com/tarper24/logi-sim-leds/pkg/core"
)

const (
	LogitechVendorID = 0x046D

	// LED thresholds (as fraction of max RPM)
	LED1Threshold = 0.45
	LED2Threshold = 0.55
	LED3Threshold = 0.625
	LED4Threshold = 0.71
	LED5Threshold = 0.85
	FlashThreshold = 0.93

	// LED command prefix
	LEDCommandPrefix = 0xf8
	LEDCommandType   = 0x12

	// Flash interval
	FlashInterval = 100 * time.Millisecond
)

// LEDConfig holds configurable LED threshold and flash values.
type LEDConfig struct {
	LED1Threshold  float64
	LED2Threshold  float64
	LED3Threshold  float64
	LED4Threshold  float64
	LED5Threshold  float64
	FlashThreshold float64
	FlashInterval  time.Duration
}

// DefaultLEDConfig returns the default LED configuration (thresholds as percentages).
func DefaultLEDConfig() LEDConfig {
	return LEDConfig{
		LED1Threshold:  LED1Threshold * 100,
		LED2Threshold:  LED2Threshold * 100,
		LED3Threshold:  LED3Threshold * 100,
		LED4Threshold:  LED4Threshold * 100,
		LED5Threshold:  LED5Threshold * 100,
		FlashThreshold: FlashThreshold * 100,
		FlashInterval:  FlashInterval,
	}
}

// LogitechWheel is a base implementation for Logitech racing wheels
type LogitechWheel struct {
	name              string
	productID         uint16
	device            *hid.Device
	connected         bool
	mu                sync.RWMutex
	previousLEDMask   uint8
	flashTimer        *time.Timer
	ledsOn            bool
	shouldFlash       bool
	ledCfg            LEDConfig
}

// NewLogitechWheel creates a new Logitech wheel instance with default LED config.
func NewLogitechWheel(name string, productID uint16) *LogitechWheel {
	return NewLogitechWheelWithConfig(name, productID, DefaultLEDConfig())
}

// NewLogitechWheelWithConfig creates a new Logitech wheel with custom LED config.
func NewLogitechWheelWithConfig(name string, productID uint16, ledCfg LEDConfig) *LogitechWheel {
	return &LogitechWheel{
		name:      name,
		productID: productID,
		connected: false,
		ledCfg:    ledCfg,
	}
}

// GetName returns the device name
func (w *LogitechWheel) GetName() string {
	return w.name
}

// GetID returns a unique identifier for the device
func (w *LogitechWheel) GetID() string {
	return fmt.Sprintf("logitech_%s_%04x", w.name, w.productID)
}

// Connect establishes connection to the device
func (w *LogitechWheel) Connect() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.connected {
		return nil
	}

	devices := hid.Enumerate(LogitechVendorID, w.productID)
	if len(devices) == 0 {
		return fmt.Errorf("device not found: %s", w.name)
	}

	// Probe each HID interface with a test LED command (all off).
	// The correct interface varies by wheel model and OS — this avoids
	// hard-coding a usage page or interface number.
	for i := range devices {
		d := &devices[i]
		device, err := d.Open()
		if err != nil {
			continue
		}

		w.device = device
		w.connected = true
		if err := w.setLEDMaskInternal(0); err != nil {
			_ = device.Close()
			w.device = nil
			w.connected = false
			continue
		}

		slog.Info("connected to device", "name", w.name, "usagePage", fmt.Sprintf("0x%04X", d.UsagePage), "usage", fmt.Sprintf("0x%04X", d.Usage), "interface", d.Interface)
		return nil
	}

	return fmt.Errorf("no working HID interface found for %s", w.name)
}

// Disconnect closes the connection to the device
func (w *LogitechWheel) Disconnect() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.connected {
		return nil
	}

	// Stop flash timer
	if w.flashTimer != nil {
		w.flashTimer.Stop()
		w.flashTimer = nil
	}

	// Turn off all LEDs
	if w.device != nil {
		_ = w.setLEDMaskInternal(0)
		_ = w.device.Close()
		w.device = nil
	}

	w.connected = false
	return nil
}

// IsConnected returns true if the device is currently connected
func (w *LogitechWheel) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connected
}

// UpdateLEDs sets the LED state based on telemetry data
func (w *LogitechWheel) UpdateLEDs(data core.TelemetryData) error {
	if !w.IsConnected() {
		return fmt.Errorf("device not connected")
	}

	if data.MaxRPM <= 0 {
		return nil // No valid max RPM yet
	}

	rpmFrac := float64(data.RPM / data.MaxRPM * 100)

	// Calculate LED mask based on configurable thresholds
	var ledMask uint8 = 0
	if rpmFrac > w.ledCfg.LED1Threshold {
		ledMask |= 0x01
	}
	if rpmFrac > w.ledCfg.LED2Threshold {
		ledMask |= 0x02
	}
	if rpmFrac > w.ledCfg.LED3Threshold {
		ledMask |= 0x04
	}
	if rpmFrac > w.ledCfg.LED4Threshold {
		ledMask |= 0x08
	}
	if rpmFrac > w.ledCfg.LED5Threshold {
		ledMask |= 0x10
	}

	// Handle flashing at high RPM
	w.mu.Lock()
	defer w.mu.Unlock()

	if rpmFrac > w.ledCfg.FlashThreshold {
		if !w.shouldFlash {
			w.shouldFlash = true
			w.startFlashing()
		}
	} else {
		if w.shouldFlash {
			w.shouldFlash = false
			w.stopFlashing()
		}
		// Only update if mask changed (to reduce HID traffic)
		if ledMask != w.previousLEDMask {
			w.previousLEDMask = ledMask
			return w.setLEDMaskInternal(ledMask)
		}
	}

	return nil
}

// SetLEDMask directly sets the LED bitmask
func (w *LogitechWheel) SetLEDMask(mask uint8) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.connected {
		return fmt.Errorf("device not connected")
	}

	w.stopFlashing()
	w.previousLEDMask = mask
	return w.setLEDMaskInternal(mask)
}

// setLEDMaskInternal sends the LED command to the device (must be called with lock held)
func (w *LogitechWheel) setLEDMaskInternal(mask uint8) error {
	if w.device == nil {
		return fmt.Errorf("device not initialized")
	}

	// The karalabe/hid library automatically prepends a 0x00 report ID byte on
	// Windows, so we send only the 7-byte payload here.
	command := []byte{LEDCommandPrefix, LEDCommandType, mask, 0x00, 0x00, 0x00, 0x00}
	_, err := w.device.Write(command)
	return err
}

// startFlashing begins LED flashing (must be called with lock held)
func (w *LogitechWheel) startFlashing() {
	if w.flashTimer != nil {
		return
	}

	w.ledsOn = false
	w.flashTimer = time.AfterFunc(0, w.flashLEDs)
}

// stopFlashing stops LED flashing (must be called with lock held)
func (w *LogitechWheel) stopFlashing() {
	if w.flashTimer != nil {
		w.flashTimer.Stop()
		w.flashTimer = nil
	}
}

// flashLEDs toggles the LEDs on and off
func (w *LogitechWheel) flashLEDs() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.shouldFlash || !w.connected {
		return
	}

	if w.ledsOn {
		_ = w.setLEDMaskInternal(0x00) // All off
	} else {
		_ = w.setLEDMaskInternal(0x1F) // All on
	}

	w.ledsOn = !w.ledsOn

	// Schedule next flash
	if w.shouldFlash {
		w.flashTimer = time.AfterFunc(w.ledCfg.FlashInterval, w.flashLEDs)
	}
}
