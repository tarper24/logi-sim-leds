package logitech

import (
	"context"
	"time"

	"github.com/karalabe/hid"
	"github.com/tarper24/logi-sim-leds/pkg/core"
)

// Detector implements device detection for Logitech wheels
type Detector struct {
	supportedDevices map[uint16]func() core.DeviceInterface
}

// NewDetector creates a new Logitech device detector
func NewDetector() *Detector {
	return &Detector{
		supportedDevices: map[uint16]func() core.DeviceInterface{
			G29ProductID:  func() core.DeviceInterface { return NewG29() },
			G920ProductID: func() core.DeviceInterface { return NewG920() },
			G923ProductID: func() core.DeviceInterface { return NewG923() },
		},
	}
}

// Detect scans for available Logitech devices
func (d *Detector) Detect() ([]core.DeviceInterface, error) {
	var devices []core.DeviceInterface

	for productID, factory := range d.supportedDevices {
		// Check if device exists
		infos := hid.Enumerate(LogitechVendorID, productID)
		if len(infos) > 0 {
			device := factory()
			devices = append(devices, device)
		}
	}

	return devices, nil
}

// Watch continuously monitors for device connection/disconnection
func (d *Detector) Watch(ctx context.Context, deviceChan chan<- core.DeviceEvent) error {
	// Track currently connected devices
	connectedDevices := make(map[string]core.DeviceInterface)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check for newly connected devices
			currentDevices := make(map[string]bool)

			for productID, factory := range d.supportedDevices {
				infos := hid.Enumerate(LogitechVendorID, productID)
				if len(infos) > 0 {
					device := factory()
					deviceID := device.GetID()
					currentDevices[deviceID] = true

					// Check if this is a new device
					if _, exists := connectedDevices[deviceID]; !exists {
						connectedDevices[deviceID] = device
						deviceChan <- core.DeviceEvent{
							Device:    device,
							Connected: true,
							Timestamp: time.Now(),
						}
					}
				}
			}

			// Check for disconnected devices
			for deviceID, device := range connectedDevices {
				if !currentDevices[deviceID] {
					delete(connectedDevices, deviceID)
					deviceChan <- core.DeviceEvent{
						Device:    device,
						Connected: false,
						Timestamp: time.Now(),
					}
				}
			}
		}
	}
}
