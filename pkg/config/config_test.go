package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg == nil {
		t.Fatal("Default() returned nil")
	}
	if !cfg.AutoDetect {
		t.Error("expected AutoDetect true")
	}
	if !cfg.Devices.Logitech.Enabled {
		t.Error("expected Logitech enabled")
	}
	if len(cfg.Devices.Logitech.Models) != 3 {
		t.Errorf("expected 3 models, got %d", len(cfg.Devices.Logitech.Models))
	}
	// Games
	if cfg.Games.BeamNG.Port != 4444 {
		t.Errorf("BeamNG port = %d, want 4444", cfg.Games.BeamNG.Port)
	}
	if cfg.Games.AssettoCorsa.Port != 9996 {
		t.Errorf("AssettoCorsa port = %d, want 9996", cfg.Games.AssettoCorsa.Port)
	}
	if cfg.Games.Dirt.Port != 20777 {
		t.Errorf("Dirt port = %d, want 20777", cfg.Games.Dirt.Port)
	}
	for _, g := range []GameConfig{cfg.Games.BeamNG, cfg.Games.AssettoCorsa, cfg.Games.Dirt} {
		if !g.Enabled {
			t.Error("expected game enabled")
		}
		if g.Address != "127.0.0.1" {
			t.Errorf("address = %q, want 127.0.0.1", g.Address)
		}
	}
	// LEDs
	if cfg.LEDs.LED1Threshold != 45 {
		t.Errorf("LED1Threshold = %f, want 45", cfg.LEDs.LED1Threshold)
	}
	if cfg.LEDs.FlashThreshold != 93 {
		t.Errorf("FlashThreshold = %f, want 93", cfg.LEDs.FlashThreshold)
	}
	if cfg.LEDs.FlashInterval != 100 {
		t.Errorf("FlashInterval = %d, want 100", cfg.LEDs.FlashInterval)
	}
	// Logging
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want info", cfg.Logging.Level)
	}
	if cfg.Logging.File != "" {
		t.Errorf("Logging.File = %q, want empty", cfg.Logging.File)
	}
}

func TestLoad_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`
auto_detect: false
games:
  beamng:
    enabled: false
    port: 1234
    address: "0.0.0.0"
leds:
  led1_threshold: 50
  flash_interval: 200
logging:
  level: debug
  file: "app.log"
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AutoDetect {
		t.Error("expected AutoDetect false")
	}
	if cfg.Games.BeamNG.Enabled {
		t.Error("expected BeamNG disabled")
	}
	if cfg.Games.BeamNG.Port != 1234 {
		t.Errorf("BeamNG port = %d, want 1234", cfg.Games.BeamNG.Port)
	}
	if cfg.LEDs.LED1Threshold != 50 {
		t.Errorf("LED1Threshold = %f, want 50", cfg.LEDs.LED1Threshold)
	}
	if cfg.LEDs.FlashInterval != 200 {
		t.Errorf("FlashInterval = %d, want 200", cfg.LEDs.FlashInterval)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want debug", cfg.Logging.Level)
	}
}

func TestLoad_PartialOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`
games:
  beamng:
    port: 9999
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	// Overridden value
	if cfg.Games.BeamNG.Port != 9999 {
		t.Errorf("BeamNG port = %d, want 9999", cfg.Games.BeamNG.Port)
	}
	// Defaults preserved
	if !cfg.AutoDetect {
		t.Error("expected AutoDetect default true")
	}
	if cfg.Games.Dirt.Port != 20777 {
		t.Errorf("Dirt port = %d, want default 20777", cfg.Games.Dirt.Port)
	}
	if cfg.LEDs.FlashThreshold != 93 {
		t.Errorf("FlashThreshold = %f, want default 93", cfg.LEDs.FlashThreshold)
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(":\n\t:bad: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	def := Default()
	if cfg.AutoDetect != def.AutoDetect {
		t.Error("empty file should preserve defaults")
	}
	if cfg.Games.BeamNG.Port != def.Games.BeamNG.Port {
		t.Error("empty file should preserve default BeamNG port")
	}
}

func TestLoadOrDefault_FallsBackToDefaults(t *testing.T) {
	cfg := LoadOrDefault()
	if cfg == nil {
		t.Fatal("LoadOrDefault returned nil")
	}
	def := Default()
	if cfg.AutoDetect != def.AutoDetect {
		t.Error("expected default AutoDetect")
	}
	if cfg.Games.BeamNG.Port != def.Games.BeamNG.Port {
		t.Error("expected default BeamNG port")
	}
}
