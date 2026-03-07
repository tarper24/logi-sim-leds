package assettocorsa

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"net"
	"sync"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

const (
	DefaultPort    = 9996
	DefaultAddress = "127.0.0.1"

	// Operation IDs for handshake
	OperationHandshake        = 0
	OperationSubscribeUpdate  = 1
	OperationSubscribeSpot    = 2
	OperationSubscribeDismiss = 3

	// Telemetry packet offsets
	RTCarInfoEngineRPMOffset = 88
)

// AssettoCorsa implements the GameInterface for Assetto Corsa
type AssettoCorsa struct {
	port           int
	address        string
	conn           *net.UDPConn
	running        bool
	connected      bool
	handshakeStage int
	mu             sync.RWMutex
	maxRPM         float32
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewAssettoCorsa creates a new Assetto Corsa game interface
func NewAssettoCorsa() *AssettoCorsa {
	return &AssettoCorsa{
		port:    DefaultPort,
		address: DefaultAddress,
		maxRPM:  1000, // Default max RPM
	}
}

// NewAssettoCorsaWithPort creates a new Assetto Corsa game interface with a custom port
func NewAssettoCorsaWithPort(port int) *AssettoCorsa {
	return &AssettoCorsa{
		port:    port,
		address: DefaultAddress,
		maxRPM:  1000,
	}
}

// GetName returns the game name
func (ac *AssettoCorsa) GetName() string {
	return "Assetto Corsa"
}

// GetPort returns the UDP port this game uses
func (ac *AssettoCorsa) GetPort() int {
	return ac.port
}

// IsRunning returns true if the game is currently sending data
func (ac *AssettoCorsa) IsRunning() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.running && ac.connected
}

// Start begins listening for telemetry data
func (ac *AssettoCorsa) Start(ctx context.Context, dataChan chan<- core.TelemetryData) error {
	ac.mu.Lock()
	if ac.running {
		ac.mu.Unlock()
		return fmt.Errorf("Assetto Corsa client already running")
	}

	// Create a cancellable context
	ac.ctx, ac.cancel = context.WithCancel(ctx)

	// Setup UDP connection
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ac.address, ac.port))
	if err != nil {
		ac.mu.Unlock()
		return fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		ac.mu.Unlock()
		return fmt.Errorf("failed to connect to UDP port %d: %w", ac.port, err)
	}

	ac.conn = conn
	ac.running = true
	ac.handshakeStage = 0
	ac.connected = false
	ac.mu.Unlock()

	slog.Info("connecting", "game", "Assetto Corsa", "address", ac.address, "port", ac.port)

	// Start connection process and listening in goroutines
	go ac.maintainConnection()
	go ac.listen(dataChan)

	return nil
}

// Stop stops listening for telemetry data
func (ac *AssettoCorsa) Stop() error {
	ac.mu.Lock()
	if !ac.running {
		ac.mu.Unlock()
		return nil
	}

	ac.running = false
	ac.connected = false
	ac.handshakeStage = 0
	conn := ac.conn
	ac.conn = nil
	cancel := ac.cancel
	ac.mu.Unlock() // release before network I/O to avoid deadlock with sendHandshakeRequest

	if conn != nil {
		// Send dismiss without holding the lock
		buf := make([]byte, 12)
		binary.LittleEndian.PutUint32(buf[8:12], OperationSubscribeDismiss)
		conn.Write(buf)
		conn.Close()
	}

	if cancel != nil {
		cancel()
	}

	return nil
}

// maintainConnection periodically attempts to establish connection
func (ac *AssettoCorsa) maintainConnection() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			ac.mu.RLock()
			connected := ac.connected
			running := ac.running
			ac.mu.RUnlock()

			if !running {
				return
			}

			if !connected {
				ac.mu.Lock()
				ac.handshakeStage = 0
				ac.mu.Unlock()
				ac.sendHandshakeRequest(OperationHandshake)
			}
		}
	}
}

// listen continuously receives UDP packets
func (ac *AssettoCorsa) listen(dataChan chan<- core.TelemetryData) {
	buffer := make([]byte, 2048)
	lastDataTime := time.Now()

	for {
		select {
		case <-ac.ctx.Done():
			return
		default:
			// Set read deadline to allow checking context
			ac.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

			n, err := ac.conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Check if we haven't received data in a while
					if time.Since(lastDataTime) > 5*time.Second {
						ac.mu.Lock()
						ac.connected = false
						ac.mu.Unlock()
					}
					continue
				}
				// Connection closed or other error
				return
			}

			lastDataTime = time.Now()

			ac.mu.Lock()
			stage := ac.handshakeStage
			ac.mu.Unlock()

			if stage == 0 {
				// Handshake response received
				ac.handleHandshakeResponse(buffer[:n])
			} else {
				// Telemetry data
				data := ac.parseRTCarInfo(buffer[:n])

				// Send to data channel — non-blocking, drop stale data if full
				select {
				case dataChan <- data:
				default:
				}
				// Check context separately
				select {
				case <-ac.ctx.Done():
					return
				default:
				}
			}
		}
	}
}

// sendHandshakeRequest sends a handshake or subscription request
func (ac *AssettoCorsa) sendHandshakeRequest(operationID uint32) error {
	buffer := make([]byte, 12)
	binary.LittleEndian.PutUint32(buffer[0:4], 0)
	binary.LittleEndian.PutUint32(buffer[4:8], 0)
	binary.LittleEndian.PutUint32(buffer[8:12], operationID)

	ac.mu.RLock()
	conn := ac.conn
	ac.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("connection not initialized")
	}

	_, err := conn.Write(buffer)
	return err
}

// handleHandshakeResponse processes the handshake response
func (ac *AssettoCorsa) handleHandshakeResponse(packet []byte) {
	if len(packet) < 312 {
		return
	}

	ac.mu.Lock()
	ac.handshakeStage = 1
	ac.connected = true
	ac.mu.Unlock()

	// Parse car name from handshake (UTF-16LE string at offset 0, 100 bytes)
	carName := parseUTF16String(packet[0:100])

	slog.Info("connected to car", "game", "Assetto Corsa", "car", carName)

	// Subscribe to updates
	ac.sendHandshakeRequest(OperationSubscribeUpdate)
}

// parseRTCarInfo parses the real-time car info packet
func (ac *AssettoCorsa) parseRTCarInfo(packet []byte) core.TelemetryData {
	// RTCarInfo structure (simplified for LED control)
	// Offset 88: engineRPM (float32)

	if len(packet) < 92 {
		return core.TelemetryData{
			Timestamp: time.Now(),
		}
	}

	// Read engine RPM at offset 88
	rpm := readFloat32LE(packet, RTCarInfoEngineRPMOffset)

	// Auto-detect max RPM: round up to next 100
	ac.mu.Lock()
	if rpm > ac.maxRPM {
		ac.maxRPM = float32(math.Ceil(float64(rpm)/100) * 100)
	}
	maxRPM := ac.maxRPM
	ac.mu.Unlock()

	// Could also extract gear at offset 96 if needed

	return core.TelemetryData{
		RPM:       rpm,
		MaxRPM:    maxRPM,
		Source:    "Assetto Corsa",
		Timestamp: time.Now(),
	}
}

// SetMaxRPM allows manually setting the maximum RPM
func (ac *AssettoCorsa) SetMaxRPM(maxRPM float32) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.maxRPM = maxRPM
	slog.Debug("max RPM set", "game", "Assetto Corsa", "rpm", maxRPM)
}

// Helper functions

func readFloat32LE(data []byte, offset int) float32 {
	bits := binary.LittleEndian.Uint32(data[offset : offset+4])
	return *(*float32)(unsafe.Pointer(&bits))
}

func parseUTF16String(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	// Convert bytes to uint16 slice
	u16s := make([]uint16, len(data)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = binary.LittleEndian.Uint16(data[i*2 : (i+1)*2])
	}

	// Find null terminator
	endIdx := -1
	for i, r := range u16s {
		if r == 0 {
			endIdx = i
			break
		}
	}

	if endIdx != -1 {
		u16s = u16s[:endIdx]
	}

	// Decode UTF-16
	runes := utf16.Decode(u16s)
	return string(runes)
}
