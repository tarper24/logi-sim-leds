package beamng

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"net"
	"sync"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

const (
	DefaultPort    = 4444
	DefaultAddress = "127.0.0.1"

	// OutGauge packet structure offsets
	OutGaugeRPMOffset = 16
)

// BeamNG implements the GameInterface for BeamNG.drive using OutGauge protocol
type BeamNG struct {
	port      int
	address   string
	conn      *net.UDPConn
	running   bool
	mu        sync.RWMutex
	maxRPM    float32
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewBeamNG creates a new BeamNG game interface
func NewBeamNG() *BeamNG {
	return &BeamNG{
		port:    DefaultPort,
		address: DefaultAddress,
		maxRPM:  1000, // Default max RPM
	}
}

// NewBeamNGWithPort creates a new BeamNG game interface with a custom port
func NewBeamNGWithPort(port int) *BeamNG {
	return &BeamNG{
		port:    port,
		address: DefaultAddress,
		maxRPM:  1000,
	}
}

// GetName returns the game name
func (b *BeamNG) GetName() string {
	return "BeamNG.drive"
}

// GetPort returns the UDP port this game uses
func (b *BeamNG) GetPort() int {
	return b.port
}

// IsRunning returns true if the game is currently sending data
func (b *BeamNG) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// Start begins listening for telemetry data
func (b *BeamNG) Start(ctx context.Context, dataChan chan<- core.TelemetryData) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return fmt.Errorf("BeamNG client already running")
	}

	// Create a cancellable context
	b.ctx, b.cancel = context.WithCancel(ctx)

	// Setup UDP listener
	addr := &net.UDPAddr{
		Port: b.port,
		IP:   net.ParseIP(b.address),
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		b.mu.Unlock()
		return fmt.Errorf("failed to listen on UDP port %d: %w", b.port, err)
	}

	b.conn = conn
	b.running = true
	b.mu.Unlock()

	slog.Info("listening", "game", "BeamNG", "address", b.address, "port", b.port)

	// Start listening in a goroutine
	go b.listen(dataChan)

	return nil
}

// Stop stops listening for telemetry data
func (b *BeamNG) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return nil
	}

	b.running = false

	if b.cancel != nil {
		b.cancel()
	}

	if b.conn != nil {
		_ = b.conn.Close()
		b.conn = nil
	}

	return nil
}

// listen continuously receives UDP packets
func (b *BeamNG) listen(dataChan chan<- core.TelemetryData) {
	buffer := make([]byte, 1024)
	lastDataTime := time.Now()

	for {
		select {
		case <-b.ctx.Done():
			return
		default:
			// Set read deadline to allow checking context
			_ = b.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

			n, _, err := b.conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Check if we haven't received data in a while
					if time.Since(lastDataTime) > 5*time.Second {
						b.mu.Lock()
						b.running = false
						b.mu.Unlock()
					}
					continue
				}
				// Connection closed or other error
				return
			}

			if n < 20 {
				// Packet too small
				continue
			}

			lastDataTime = time.Now()
			b.mu.Lock()
			wasRunning := b.running
			b.running = true
			b.mu.Unlock()

			if !wasRunning {
				slog.Info("connected and receiving data", "game", "BeamNG")
			}

			// Parse OutGauge packet
			data := b.parseOutGauge(buffer[:n])

			// Send to data channel — non-blocking, drop stale data if full
			select {
			case dataChan <- data:
			default:
			}
			// Check context separately
			select {
			case <-b.ctx.Done():
				return
			default:
			}
		}
	}
}

// parseOutGauge parses an OutGauge UDP packet
func (b *BeamNG) parseOutGauge(packet []byte) core.TelemetryData {
	// OutGauge packet structure:
	// Offset 16-19: RPM (float32, little endian)

	if len(packet) < 20 {
		return core.TelemetryData{
			Timestamp: time.Now(),
		}
	}

	// Read RPM as float32 at offset 16
	rpm := math.Float32frombits(binary.LittleEndian.Uint32(packet[OutGaugeRPMOffset : OutGaugeRPMOffset+4]))

	// Auto-detect max RPM: if current RPM exceeds known max, update via SetMaxRPM
	// which rounds up to next 100 so the threshold doesn't chase the peak.
	b.mu.RLock()
	currentMax := b.maxRPM
	b.mu.RUnlock()
	if rpm > currentMax {
		b.SetMaxRPM(rpm)
	}

	b.mu.RLock()
	maxRPM := b.maxRPM
	b.mu.RUnlock()

	// For more complete parsing, we could extract:
	// - Speed (offset 4-7: float32)
	// - Gear (offset 20-21)
	// - Other telemetry data
	// But for LED control, RPM is the primary concern

	return core.TelemetryData{
		RPM:       rpm,
		MaxRPM:    maxRPM,
		Source:    "BeamNG.drive",
		Timestamp: time.Now(),
	}
}

// roundedMaxRPM rounds rpm up to the next 100 so the LED threshold doesn't
// chase the peak and trigger the flash zone prematurely.
func roundedMaxRPM(rpm float32) float32 {
	return float32(math.Ceil(float64(rpm)/100) * 100)
}

// SetMaxRPM allows manually setting the maximum RPM
func (b *BeamNG) SetMaxRPM(maxRPM float32) {
	rounded := roundedMaxRPM(maxRPM)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maxRPM = rounded
	slog.Debug("max RPM set", "game", "BeamNG", "rpm", rounded)
}
