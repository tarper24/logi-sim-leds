package beamng

import (
	"context"
	"encoding/binary"
	"math"
	"net"
	"testing"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

func TestNewBeamNG(t *testing.T) {
	b := NewBeamNG()
	if b.port != 5555 {
		t.Errorf("expected port 5555, got %d", b.port)
	}
	if b.address != "127.0.0.1" {
		t.Errorf("expected address 127.0.0.1, got %s", b.address)
	}
	if b.maxRPM != 1000 {
		t.Errorf("expected maxRPM 1000, got %f", b.maxRPM)
	}
}

func TestNewBeamNGWithPort(t *testing.T) {
	b := NewBeamNGWithPort(9999)
	if b.port != 9999 {
		t.Errorf("expected port 9999, got %d", b.port)
	}
}

func TestGetName(t *testing.T) {
	b := NewBeamNG()
	if name := b.GetName(); name != "BeamNG.drive" {
		t.Errorf("expected BeamNG.drive, got %s", name)
	}
}

func TestGetPort(t *testing.T) {
	b := NewBeamNGWithPort(7777)
	if p := b.GetPort(); p != 7777 {
		t.Errorf("expected 7777, got %d", p)
	}
}

func TestIsRunning_InitiallyFalse(t *testing.T) {
	b := NewBeamNG()
	if b.IsRunning() {
		t.Error("expected IsRunning to be false initially")
	}
}

func TestStartStop(t *testing.T) {
	b := NewBeamNGWithPort(15555)
	ctx := context.Background()
	dataChan := make(chan core.TelemetryData, 1)

	if err := b.Start(ctx, dataChan); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !b.IsRunning() {
		t.Error("expected IsRunning true after Start")
	}

	if err := b.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if b.IsRunning() {
		t.Error("expected IsRunning false after Stop")
	}
}

func TestStartAlreadyRunning(t *testing.T) {
	b := NewBeamNGWithPort(15556)
	ctx := context.Background()
	dataChan := make(chan core.TelemetryData, 1)

	if err := b.Start(ctx, dataChan); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer b.Stop()

	if err := b.Start(ctx, dataChan); err == nil {
		t.Error("expected error on second Start")
	}
}

func TestStopWhenNotRunning(t *testing.T) {
	b := NewBeamNG()
	if err := b.Stop(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func makeOutGaugePacket(rpm float32) []byte {
	packet := make([]byte, 64)
	binary.LittleEndian.PutUint32(packet[16:20], math.Float32bits(rpm))
	return packet
}

func TestParseOutGauge(t *testing.T) {
	b := NewBeamNG()
	data := b.parseOutGauge(makeOutGaugePacket(3500))
	if data.RPM != 3500 {
		t.Errorf("expected RPM 3500, got %f", data.RPM)
	}
	if data.Source != "BeamNG.drive" {
		t.Errorf("expected source BeamNG.drive, got %s", data.Source)
	}
}

func TestParseOutGauge_TooSmall(t *testing.T) {
	b := NewBeamNG()
	data := b.parseOutGauge(make([]byte, 10))
	if data.RPM != 0 {
		t.Errorf("expected RPM 0 for small packet, got %f", data.RPM)
	}
}

func TestSetMaxRPM(t *testing.T) {
	b := NewBeamNG()
	b.SetMaxRPM(6750)
	b.mu.RLock()
	got := b.maxRPM
	b.mu.RUnlock()
	if got != 6800 {
		t.Errorf("expected 6800, got %f", got)
	}
}

func TestRoundedMaxRPM(t *testing.T) {
	tests := []struct {
		input    float32
		expected float32
	}{
		{6750, 6800},
		{7000, 7000},
		{7001, 7100},
	}
	for _, tc := range tests {
		got := roundedMaxRPM(tc.input)
		if got != tc.expected {
			t.Errorf("roundedMaxRPM(%v) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestUDPTelemetry(t *testing.T) {
	port := 15557
	b := NewBeamNGWithPort(port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dataChan := make(chan core.TelemetryData, 1)

	if err := b.Start(ctx, dataChan); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer b.Stop()

	// Send a UDP packet
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	packet := makeOutGaugePacket(4200)
	conn.Write(packet)

	select {
	case data := <-dataChan:
		if data.RPM != 4200 {
			t.Errorf("expected RPM 4200, got %f", data.RPM)
		}
		if data.Source != "BeamNG.drive" {
			t.Errorf("expected source BeamNG.drive, got %s", data.Source)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for telemetry data")
	}
}
