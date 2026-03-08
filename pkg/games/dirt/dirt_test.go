package dirt

import (
	"context"
	"encoding/binary"
	"math"
	"net"
	"testing"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

func TestNewDirt(t *testing.T) {
	d := NewDirt()
	if d.port != 20777 {
		t.Errorf("expected port 20777, got %d", d.port)
	}
}

func TestNewDirtWithPort(t *testing.T) {
	d := NewDirtWithPort(12345)
	if d.port != 12345 {
		t.Errorf("expected port 12345, got %d", d.port)
	}
}

func TestGetName(t *testing.T) {
	d := NewDirt()
	if name := d.GetName(); name != "Dirt/Codemasters" {
		t.Errorf("expected Dirt/Codemasters, got %s", name)
	}
}

func TestStartStop(t *testing.T) {
	d := NewDirtWithPort(30777)
	ctx := context.Background()
	dataChan := make(chan core.TelemetryData, 1)

	if err := d.Start(ctx, dataChan); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !d.IsRunning() {
		t.Error("expected IsRunning true after Start")
	}

	if err := d.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if d.IsRunning() {
		t.Error("expected IsRunning false after Stop")
	}
}

func makeCodemastersPacket(rpm, maxRPM float32) []byte {
	packet := make([]byte, 256)
	// RPM at offset 148, stored as value/10
	binary.LittleEndian.PutUint32(packet[148:152], math.Float32bits(rpm/10.0))
	// MaxRPM at offset 248, stored as value/10
	binary.LittleEndian.PutUint32(packet[248:252], math.Float32bits(maxRPM/10.0))
	return packet
}

func TestParseCodemastersPacket(t *testing.T) {
	d := NewDirt()
	packet := makeCodemastersPacket(5000, 8000)
	data := d.parseCodemastersPacket(packet)

	if data.RPM != 5000 {
		t.Errorf("expected RPM 5000, got %f", data.RPM)
	}
	if data.MaxRPM != 8000 {
		t.Errorf("expected MaxRPM 8000, got %f", data.MaxRPM)
	}
	if data.Source != "Dirt/Codemasters" {
		t.Errorf("expected source Dirt/Codemasters, got %s", data.Source)
	}
}

func TestParseCodemastersPacket_TooSmall(t *testing.T) {
	d := NewDirt()
	data := d.parseCodemastersPacket(make([]byte, 100))
	if data.RPM != 0 {
		t.Errorf("expected RPM 0, got %f", data.RPM)
	}
}

func TestSetMaxRPM(t *testing.T) {
	d := NewDirt()
	d.SetMaxRPM(7500)
	d.mu.RLock()
	got := d.maxRPM
	d.mu.RUnlock()
	if got != 7500 {
		t.Errorf("expected 7500, got %f", got)
	}
}

func TestUDPTelemetry(t *testing.T) {
	port := 30778
	d := NewDirtWithPort(port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dataChan := make(chan core.TelemetryData, 1)

	if err := d.Start(ctx, dataChan); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer d.Stop()

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	packet := makeCodemastersPacket(6000, 9000)
	conn.Write(packet)

	select {
	case data := <-dataChan:
		if data.RPM != 6000 {
			t.Errorf("expected RPM 6000, got %f", data.RPM)
		}
		if data.MaxRPM != 9000 {
			t.Errorf("expected MaxRPM 9000, got %f", data.MaxRPM)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for telemetry data")
	}
}
