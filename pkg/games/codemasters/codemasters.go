package codemasters

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"net"
	"sync"
	"time"
	"unsafe"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

const (
	DefaultPort    = 20777
	DefaultAddress = "127.0.0.1"

	// Legacy Dirt/Codemasters packet offsets (all floats, no version header)
	EngineRPMOffset = 37 * 4 // Offset 148 bytes
	MaxRPMOffset    = 62 * 4 // Offset 248 bytes

	// F1 series packetId values
	f1PacketIDCarTelemetry = 6

	// F1 car telemetry: engineRPM is always at offset 16 within the per-car entry (uint16)
	f1CarEntryRPMOffset = 16
)

// Codemasters handles UDP telemetry from Dirt Rally, Dirt 4, Dirt Rally 2.0,
// and the F1 series (2018–present). The protocol version is auto-detected per
// packet by inspecting the first two bytes:
//
//   - Legacy Dirt format: no version header; bytes 0-1 are the low bytes of a
//     float32 game-time field, which never falls in the range [2018, 2030].
//   - F1 format: bytes 0-1 are packetFormat (uint16 LE year, e.g. 2024).
type Codemasters struct {
	port     int
	address  string
	conn     *net.UDPConn
	running  bool
	mu       sync.RWMutex
	maxRPM   float32
	gameName string // last detected game name, used for Source and GetName
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewCodemasters creates a new Codemasters game interface.
func NewCodemasters() *Codemasters {
	return &Codemasters{
		port:     DefaultPort,
		address:  DefaultAddress,
		maxRPM:   1000,
		gameName: "Dirt/Codemasters",
	}
}

// NewCodemastersWithPort creates a new Codemasters game interface with a custom port.
func NewCodemastersWithPort(port int) *Codemasters {
	return &Codemasters{
		port:     port,
		address:  DefaultAddress,
		maxRPM:   1000,
		gameName: "Dirt/Codemasters",
	}
}

// GetName returns the detected game name (dynamic once a packet is received).
func (c *Codemasters) GetName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gameName
}

// GetPort returns the UDP port this game uses.
func (c *Codemasters) GetPort() int {
	return c.port
}

// IsRunning returns true if the game is currently sending data.
func (c *Codemasters) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// Start begins listening for telemetry data.
func (c *Codemasters) Start(ctx context.Context, dataChan chan<- core.TelemetryData) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("Codemasters client already running")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	addr := &net.UDPAddr{
		Port: c.port,
		IP:   net.ParseIP(c.address),
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to listen on UDP port %d: %w", c.port, err)
	}

	c.conn = conn
	c.running = true
	c.mu.Unlock()

	slog.Info("listening", "game", "Codemasters", "address", c.address, "port", c.port)

	go c.listen(dataChan)

	return nil
}

// Stop stops listening for telemetry data.
func (c *Codemasters) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	c.running = false

	if c.cancel != nil {
		c.cancel()
	}

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	return nil
}

func (c *Codemasters) listen(dataChan chan<- core.TelemetryData) {
	buffer := make([]byte, 2048)
	lastDataTime := time.Now()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()
			if conn == nil {
				return
			}

			conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					if time.Since(lastDataTime) > 5*time.Second {
						c.mu.Lock()
						c.running = false
						c.mu.Unlock()
					}
					continue
				}
				return
			}

			if n < 20 {
				continue
			}

			lastDataTime = time.Now()
			c.mu.Lock()
			wasRunning := c.running
			c.running = true
			c.mu.Unlock()

			if !wasRunning {
				slog.Info("connected and receiving data", "game", "Codemasters")
			}

			data := c.parseCodemastersPacket(buffer[:n])

			// Skip non-car-telemetry F1 packets (Source is empty for those)
			if data.Source == "" {
				continue
			}

			select {
			case dataChan <- data:
			default:
			}

			select {
			case <-c.ctx.Done():
				return
			default:
			}
		}
	}
}

// parseCodemastersPacket detects the protocol version and routes to the appropriate parser.
// F1 format: bytes 0-1 are the packetFormat year (uint16 LE, e.g. 2024).
// Legacy Dirt format: bytes 0-1 are the low bytes of a float32 time value — never in [2018,2030].
func (c *Codemasters) parseCodemastersPacket(packet []byte) core.TelemetryData {
	if len(packet) < 4 {
		return core.TelemetryData{Timestamp: time.Now()}
	}

	year := int(binary.LittleEndian.Uint16(packet[0:2]))
	if year >= 2018 && year <= 2030 {
		return c.parseF1Packet(packet, year)
	}
	return c.parseDirtPacket(packet)
}

// parseDirtPacket handles the legacy Dirt/Codemasters float-array packet format.
func (c *Codemasters) parseDirtPacket(packet []byte) core.TelemetryData {
	if len(packet) < 252 {
		return core.TelemetryData{Timestamp: time.Now()}
	}

	// RPM values are stored divided by 10 in the Dirt format
	engineRPM := readFloat32LE(packet, EngineRPMOffset) * 10.0
	maxRPM := readFloat32LE(packet, MaxRPMOffset) * 10.0

	c.mu.Lock()
	c.gameName = "Dirt/Codemasters"
	if maxRPM > 0 {
		c.maxRPM = maxRPM
	} else if engineRPM > c.maxRPM {
		// Packet didn't provide maxRPM — auto-detect from observed RPM
		c.maxRPM = float32(math.Ceil(float64(engineRPM)/100) * 100)
	}
	currentMax := c.maxRPM
	c.mu.Unlock()

	return core.TelemetryData{
		RPM:       engineRPM,
		MaxRPM:    currentMax,
		Source:    "Dirt/Codemasters",
		Timestamp: time.Now(),
	}
}

// parseF1Packet handles the F1 series multi-packet format.
// Only car telemetry packets (packetId=6) produce output; others return an empty
// TelemetryData (Source="") which the listen loop silently drops.
//
// Header layout by year:
//
//	2018:  [packetFormat(2)][packetVersion(1)][packetId(1)][sessionUID(8)][sessionTime(4)][frameId(4)][playerCarIdx(1)]  = 21 bytes
//	2019:  [packetFormat(2)][major(1)][minor(1)][packetVersion(1)][packetId(1)][sessionUID(8)][sessionTime(4)][frameId(4)][playerCarIdx(1)]  = 23 bytes
//	2020+: same as 2019 + [secondaryPlayerCarIdx(1)] = 24 bytes
//
// Car telemetry entry: engineRPM (uint16) is always at offset 16 within the entry.
// Entry size: 60 bytes for 2018-2020, 62 bytes for 2021+ (added revLightsBitValue uint16).
func (c *Codemasters) parseF1Packet(packet []byte, year int) core.TelemetryData {
	// Determine offsets based on format year
	var headerSize, packetIDOffset, playerCarIdxOffset int
	switch {
	case year == 2018:
		headerSize, packetIDOffset, playerCarIdxOffset = 21, 3, 20
	case year == 2019:
		headerSize, packetIDOffset, playerCarIdxOffset = 23, 5, 22
	default: // 2020+
		headerSize, packetIDOffset, playerCarIdxOffset = 24, 5, 22
	}

	if len(packet) < headerSize {
		return core.TelemetryData{Timestamp: time.Now()}
	}

	// Only process car telemetry packets
	if packet[packetIDOffset] != f1PacketIDCarTelemetry {
		return core.TelemetryData{Timestamp: time.Now()} // Source="" signals "skip"
	}

	var carEntrySize int
	if year >= 2021 {
		carEntrySize = 62
	} else {
		carEntrySize = 60
	}

	playerCar := int(packet[playerCarIdxOffset])
	rpmOffset := headerSize + playerCar*carEntrySize + f1CarEntryRPMOffset

	if len(packet) < rpmOffset+2 {
		return core.TelemetryData{Timestamp: time.Now()}
	}

	rpm := float32(binary.LittleEndian.Uint16(packet[rpmOffset : rpmOffset+2]))
	gameName := fmt.Sprintf("F1 %d", year)

	c.mu.Lock()
	c.gameName = gameName
	if rpm > c.maxRPM {
		c.maxRPM = float32(math.Ceil(float64(rpm)/100) * 100)
	}
	currentMax := c.maxRPM
	c.mu.Unlock()

	return core.TelemetryData{
		RPM:       rpm,
		MaxRPM:    currentMax,
		Source:    gameName,
		Timestamp: time.Now(),
	}
}

// SetMaxRPM allows manually setting the maximum RPM.
func (c *Codemasters) SetMaxRPM(maxRPM float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxRPM = maxRPM
	slog.Debug("max RPM set", "game", c.gameName, "rpm", maxRPM)
}

func readFloat32LE(data []byte, offset int) float32 {
	bits := binary.LittleEndian.Uint32(data[offset : offset+4])
	return *(*float32)(unsafe.Pointer(&bits))
}
