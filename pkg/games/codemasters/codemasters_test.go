package codemasters

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"testing"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

func TestNewCodemasters(t *testing.T) {
	c := NewCodemasters()
	if c.port != 20777 {
		t.Errorf("expected port 20777, got %d", c.port)
	}
}

func TestNewCodemastersWithPort(t *testing.T) {
	c := NewCodemastersWithPort(12345)
	if c.port != 12345 {
		t.Errorf("expected port 12345, got %d", c.port)
	}
}

func TestGetName(t *testing.T) {
	c := NewCodemasters()
	if name := c.GetName(); name != "Dirt/Codemasters" {
		t.Errorf("expected Dirt/Codemasters, got %s", name)
	}
}

func TestStartStop(t *testing.T) {
	c := NewCodemastersWithPort(30777)
	ctx := context.Background()
	dataChan := make(chan core.TelemetryData, 1)

	if err := c.Start(ctx, dataChan); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !c.IsRunning() {
		t.Error("expected IsRunning true after Start")
	}

	if err := c.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if c.IsRunning() {
		t.Error("expected IsRunning false after Stop")
	}
}

// makeDirtPacket builds a legacy Dirt/Codemasters float-array packet.
// Bytes 0-1 will be 0x00 0x00 (low bytes of float32 time=0.0), which is
// outside the F1 year range [2018,2030] and correctly routes to parseDirtPacket.
func makeDirtPacket(rpm, maxRPM float32) []byte {
	packet := make([]byte, 256)
	binary.LittleEndian.PutUint32(packet[148:152], math.Float32bits(rpm/10.0))
	binary.LittleEndian.PutUint32(packet[248:252], math.Float32bits(maxRPM/10.0))
	return packet
}

// makeF1CarTelemetryPacket builds a minimal F1 2024 car telemetry packet (packetId=6).
func makeF1CarTelemetryPacket(year int, rpm uint16) []byte {
	// Header: 24 bytes for 2020+, 23 for 2019, 21 for 2018
	var headerSize, packetIDOffset, playerCarIdxOffset int
	switch year {
	case 2018:
		headerSize, packetIDOffset, playerCarIdxOffset = 21, 3, 20
	case 2019:
		headerSize, packetIDOffset, playerCarIdxOffset = 23, 5, 22
	default:
		headerSize, packetIDOffset, playerCarIdxOffset = 24, 5, 22
	}

	var carEntrySize int
	switch year {
	case 2018:
		carEntrySize = f1CarEntrySize2018 // 53
	case 2019:
		carEntrySize = f1CarEntrySize2019 // 66
	case 2020:
		carEntrySize = f1CarEntrySize2020 // 58
	default: // 2021+
		carEntrySize = f1CarEntrySize2021 // 60
	}

	size := headerSize + 20*carEntrySize + 4 // 20 cars + buttonStatus
	packet := make([]byte, size)

	// Set packetFormat (year) at bytes 0-1
	binary.LittleEndian.PutUint16(packet[0:2], uint16(year))
	// Set packetId = 6 (car telemetry)
	packet[packetIDOffset] = f1PacketIDCarTelemetry
	// playerCarIndex = 5 (non-zero to exercise entry-size calculation)
	const playerIdx = 5
	packet[playerCarIdxOffset] = playerIdx
	// engineRPM offset within car entry differs by year (2018 uses integers, 2019+ uses floats)
	entryRPMOffset := 16
	if year == 2018 {
		entryRPMOffset = 7
	}
	rpmPos := headerSize + playerIdx*carEntrySize + entryRPMOffset
	binary.LittleEndian.PutUint16(packet[rpmPos:rpmPos+2], rpm)

	return packet
}

func TestParseCodemastersPacket(t *testing.T) {
	c := NewCodemasters()
	data := c.parseCodemastersPacket(makeDirtPacket(5000, 8000))

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

func TestParseDirtPacket_ZeroMaxRPM(t *testing.T) {
	// Dirt Rally sends maxRPM=0 — should auto-detect from observed RPM
	c := NewCodemasters()
	data := c.parseCodemastersPacket(makeDirtPacket(6750, 0))

	if data.RPM != 6750 {
		t.Errorf("expected RPM 6750, got %f", data.RPM)
	}
	// auto-detect: ceil(6750/100)*100 = 6800
	if data.MaxRPM != 6800 {
		t.Errorf("expected MaxRPM 6800 (auto-detected), got %f", data.MaxRPM)
	}
}

func TestParseCodemastersPacket_TooSmall(t *testing.T) {
	c := NewCodemasters()
	data := c.parseCodemastersPacket(make([]byte, 100))
	if data.RPM != 0 {
		t.Errorf("expected RPM 0, got %f", data.RPM)
	}
}

func TestParseF1CarTelemetry(t *testing.T) {
	for _, year := range []int{2018, 2019, 2020, 2021, 2024} {
		t.Run(fmt.Sprintf("F1 %d", year), func(t *testing.T) {
			c := NewCodemasters()
			packet := makeF1CarTelemetryPacket(year, 12000)
			data := c.parseCodemastersPacket(packet)

			if data.RPM != 12000 {
				t.Errorf("expected RPM 12000, got %f", data.RPM)
			}
			expected := fmt.Sprintf("F1 %d", year)
			if data.Source != expected {
				t.Errorf("expected source %s, got %s", expected, data.Source)
			}
			if c.GetName() != expected {
				t.Errorf("expected GetName %s, got %s", expected, c.GetName())
			}
		})
	}
}

func TestParseF1NonCarTelemetry(t *testing.T) {
	// Non-car-telemetry F1 packets should return Source="" (dropped by listen loop)
	c := NewCodemasters()
	packet := make([]byte, 100)
	binary.LittleEndian.PutUint16(packet[0:2], 2024) // F1 year
	packet[5] = 1                                     // packetId=1 (Session), not 6

	data := c.parseCodemastersPacket(packet)
	if data.Source != "" {
		t.Errorf("expected empty Source for non-CT packet, got %s", data.Source)
	}
}

func TestSetMaxRPM(t *testing.T) {
	c := NewCodemasters()
	c.SetMaxRPM(7500)
	c.mu.RLock()
	got := c.maxRPM
	c.mu.RUnlock()
	if got != 7500 {
		t.Errorf("expected 7500, got %f", got)
	}
}

func TestUDPTelemetry(t *testing.T) {
	port := 30778
	c := NewCodemastersWithPort(port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dataChan := make(chan core.TelemetryData, 1)

	if err := c.Start(ctx, dataChan); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = c.Stop() }()

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	_, _ = conn.Write(makeDirtPacket(6000, 9000))

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
