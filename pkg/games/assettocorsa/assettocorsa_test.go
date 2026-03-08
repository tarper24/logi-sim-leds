package assettocorsa

import (
	"context"
	"encoding/binary"
	"math"
	"net"
	"testing"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

func TestNewAssettoCorsa(t *testing.T) {
	ac := NewAssettoCorsa()
	if ac.port != 9996 {
		t.Errorf("expected port 9996, got %d", ac.port)
	}
}

func TestNewAssettoCorsaWithPort(t *testing.T) {
	ac := NewAssettoCorsaWithPort(11111)
	if ac.port != 11111 {
		t.Errorf("expected port 11111, got %d", ac.port)
	}
}

func TestGetName(t *testing.T) {
	ac := NewAssettoCorsa()
	if name := ac.GetName(); name != "Assetto Corsa" {
		t.Errorf("expected Assetto Corsa, got %s", name)
	}
}

func makeRTCarInfoPacket(rpm float32) []byte {
	packet := make([]byte, 128)
	binary.LittleEndian.PutUint32(packet[RTCarInfoEngineRPMOffset:RTCarInfoEngineRPMOffset+4], math.Float32bits(rpm))
	return packet
}

func TestParseRTCarInfo(t *testing.T) {
	ac := NewAssettoCorsa()
	data := ac.parseRTCarInfo(makeRTCarInfoPacket(6750))

	if data.RPM != 6750 {
		t.Errorf("expected RPM 6750, got %f", data.RPM)
	}
	// maxRPM should be rounded up to 6800
	if data.MaxRPM != 6800 {
		t.Errorf("expected MaxRPM 6800, got %f", data.MaxRPM)
	}
	if data.Source != "Assetto Corsa" {
		t.Errorf("expected source Assetto Corsa, got %s", data.Source)
	}
}

func TestParseRTCarInfo_TooSmall(t *testing.T) {
	ac := NewAssettoCorsa()
	data := ac.parseRTCarInfo(make([]byte, 50))
	if data.RPM != 0 {
		t.Errorf("expected RPM 0, got %f", data.RPM)
	}
}

func TestSetMaxRPM(t *testing.T) {
	ac := NewAssettoCorsa()
	ac.SetMaxRPM(8500)
	ac.mu.RLock()
	got := ac.maxRPM
	ac.mu.RUnlock()
	if got != 8500 {
		t.Errorf("expected 8500, got %f", got)
	}
}

func TestParseUTF16String(t *testing.T) {
	// Encode "Hello" as UTF-16LE with null terminator
	input := "Hello"
	data := make([]byte, (len(input)+1)*2)
	for i, r := range input {
		binary.LittleEndian.PutUint16(data[i*2:(i+1)*2], uint16(r))
	}
	// null terminator already zero

	got := parseUTF16String(data)
	if got != "Hello" {
		t.Errorf("expected Hello, got %s", got)
	}
}

func TestParseUTF16String_Empty(t *testing.T) {
	got := parseUTF16String([]byte{})
	if got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
	got = parseUTF16String([]byte{0})
	if got != "" {
		t.Errorf("expected empty string for single byte, got %s", got)
	}
}

func TestStartStop(t *testing.T) {
	port := 19996
	// Start a local UDP listener to act as the AC server
	serverAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port}
	server, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer func() { _ = server.Close() }()

	ac := NewAssettoCorsaWithPort(port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dataChan := make(chan core.TelemetryData, 1)

	if err := ac.Start(ctx, dataChan); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// running should be true (connected may be false since no handshake)
	ac.mu.RLock()
	running := ac.running
	ac.mu.RUnlock()
	if !running {
		t.Error("expected running true after Start")
	}

	if err := ac.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Give goroutines time to exit
	time.Sleep(50 * time.Millisecond)

	ac.mu.RLock()
	running = ac.running
	ac.mu.RUnlock()
	if running {
		t.Error("expected running false after Stop")
	}
}
