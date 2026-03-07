package dirt

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
	"unsafe"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

const (
	DefaultPort    = 20777
	DefaultAddress = "127.0.0.1"

	// Codemasters telemetry packet offsets
	// Based on DiRT Rally, DiRT 4, DiRT Rally 2.0, F1 series
	EngineRPMOffset = 37 * 4  // Offset 148 bytes (37 floats * 4 bytes)
	MaxRPMOffset    = 62 * 4  // Offset 248 bytes (62 floats * 4 bytes)
)

// Dirt implements the GameInterface for Codemasters Dirt games
// Also compatible with DiRT Rally, DiRT 4, DiRT Rally 2.0, and F1 series
type Dirt struct {
	port      int
	address   string
	conn      *net.UDPConn
	running   bool
	mu        sync.RWMutex
	maxRPM    float32
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewDirt creates a new Dirt game interface
func NewDirt() *Dirt {
	return &Dirt{
		port:    DefaultPort,
		address: DefaultAddress,
		maxRPM:  1000, // Default max RPM (will be updated from telemetry)
	}
}

// NewDirtWithPort creates a new Dirt game interface with a custom port
func NewDirtWithPort(port int) *Dirt {
	return &Dirt{
		port:    port,
		address: DefaultAddress,
		maxRPM:  1000,
	}
}

// GetName returns the game name
func (d *Dirt) GetName() string {
	return "Dirt/Codemasters"
}

// GetPort returns the UDP port this game uses
func (d *Dirt) GetPort() int {
	return d.port
}

// IsRunning returns true if the game is currently sending data
func (d *Dirt) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}

// Start begins listening for telemetry data
func (d *Dirt) Start(ctx context.Context, dataChan chan<- core.TelemetryData) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return fmt.Errorf("Dirt client already running")
	}

	// Create a cancellable context
	d.ctx, d.cancel = context.WithCancel(ctx)

	// Setup UDP listener
	addr := &net.UDPAddr{
		Port: d.port,
		IP:   net.ParseIP(d.address),
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		d.mu.Unlock()
		return fmt.Errorf("failed to listen on UDP port %d: %w", d.port, err)
	}

	d.conn = conn
	d.running = true
	d.mu.Unlock()

	slog.Info("listening", "game", "Dirt", "address", d.address, "port", d.port)

	// Start listening in a goroutine
	go d.listen(dataChan)

	return nil
}

// Stop stops listening for telemetry data
func (d *Dirt) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	d.running = false

	if d.cancel != nil {
		d.cancel()
	}

	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
	}

	return nil
}

// listen continuously receives UDP packets
func (d *Dirt) listen(dataChan chan<- core.TelemetryData) {
	buffer := make([]byte, 2048)
	lastDataTime := time.Now()

	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			// Set read deadline to allow checking context
			d.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

			n, _, err := d.conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Check if we haven't received data in a while
					if time.Since(lastDataTime) > 5*time.Second {
						d.mu.Lock()
						d.running = false
						d.mu.Unlock()
					}
					continue
				}
				// Connection closed or other error
				return
			}

			if n < 256 {
				// Packet too small
				continue
			}

			lastDataTime = time.Now()
			d.mu.Lock()
			wasRunning := d.running
			d.running = true
			d.mu.Unlock()

			if !wasRunning {
				slog.Info("connected and receiving data", "game", "Dirt")
			}

			// Parse Codemasters telemetry packet
			data := d.parseCodemastersPacket(buffer[:n])

			// Send to data channel — non-blocking, drop stale data if full
			select {
			case dataChan <- data:
			default:
			}
			// Check context separately
			select {
			case <-d.ctx.Done():
				return
			default:
			}
		}
	}
}

// parseCodemastersPacket parses a Codemasters telemetry packet
func (d *Dirt) parseCodemastersPacket(packet []byte) core.TelemetryData {
	// Codemasters telemetry packet structure:
	// Many floats, we need:
	// Offset 148 (37*4): Engine RPM (multiplied by 10 in packet)
	// Offset 248 (62*4): Max RPM (multiplied by 10 in packet)

	if len(packet) < 252 {
		return core.TelemetryData{
			Timestamp: time.Now(),
		}
	}

	// Read engine RPM at offset 148 (37th float)
	// The RPM is multiplied by 10 in the packet
	engineRPM := readFloat32LE(packet, EngineRPMOffset) * 10.0

	// Read max RPM at offset 248 (62nd float)
	// The max RPM is also multiplied by 10 in the packet
	maxRPM := readFloat32LE(packet, MaxRPMOffset) * 10.0

	// Update our max RPM if we got a valid value
	d.mu.Lock()
	if maxRPM > 0 {
		d.maxRPM = maxRPM
	}
	currentMax := d.maxRPM
	d.mu.Unlock()

	return core.TelemetryData{
		RPM:       engineRPM,
		MaxRPM:    currentMax,
		Source:    "Dirt/Codemasters",
		Timestamp: time.Now(),
	}
}

// SetMaxRPM allows manually setting the maximum RPM
func (d *Dirt) SetMaxRPM(maxRPM float32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.maxRPM = maxRPM
	slog.Debug("max RPM set", "game", "Dirt", "rpm", maxRPM)
}

// Helper functions

func readFloat32LE(data []byte, offset int) float32 {
	bits := binary.LittleEndian.Uint32(data[offset : offset+4])
	return *(*float32)(unsafe.Pointer(&bits))
}
