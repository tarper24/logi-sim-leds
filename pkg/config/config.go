package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AutoDetect bool          `yaml:"auto_detect"`
	Devices    DevicesConfig `yaml:"devices"`
	Games      GamesConfig   `yaml:"games"`
	LEDs       LEDsConfig    `yaml:"leds"`
	Logging    LoggingConfig `yaml:"logging"`
}

type DevicesConfig struct {
	Logitech LogitechConfig `yaml:"logitech"`
}

type LogitechConfig struct {
	Enabled bool     `yaml:"enabled"`
	Models  []string `yaml:"models"`
}

type GamesConfig struct {
	BeamNG       GameConfig `yaml:"beamng"`
	AssettoCorsa GameConfig `yaml:"assetto_corsa"`
	Dirt         GameConfig `yaml:"dirt"`
}

type GameConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Address string `yaml:"address"`
}

type LEDsConfig struct {
	LED1Threshold  float64 `yaml:"led1_threshold"`
	LED2Threshold  float64 `yaml:"led2_threshold"`
	LED3Threshold  float64 `yaml:"led3_threshold"`
	LED4Threshold  float64 `yaml:"led4_threshold"`
	LED5Threshold  float64 `yaml:"led5_threshold"`
	FlashThreshold float64 `yaml:"flash_threshold"`
	FlashInterval  int     `yaml:"flash_interval"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// Default returns a Config with hardcoded default values.
func Default() *Config {
	return &Config{
		AutoDetect: true,
		Devices: DevicesConfig{
			Logitech: LogitechConfig{
				Enabled: true,
				Models:  []string{"g29", "g920", "g923"},
			},
		},
		Games: GamesConfig{
			BeamNG:       GameConfig{Enabled: true, Port: 4444, Address: "127.0.0.1"},
			AssettoCorsa: GameConfig{Enabled: true, Port: 9996, Address: "127.0.0.1"},
			Dirt:         GameConfig{Enabled: true, Port: 20777, Address: "127.0.0.1"},
		},
		LEDs: LEDsConfig{
			LED1Threshold:  45,
			LED2Threshold:  55,
			LED3Threshold:  62.5,
			LED4Threshold:  71,
			LED5Threshold:  85,
			FlashThreshold: 93,
			FlashInterval:  100,
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "",
		},
	}
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// LoadOrDefault tries to load config.yaml from the executable's directory,
// then the current working directory. Falls back to defaults if not found.
func LoadOrDefault() *Config {
	// Try executable directory first
	if exePath, err := os.Executable(); err == nil {
		path := filepath.Join(filepath.Dir(exePath), "config.yaml")
		if cfg, err := Load(path); err == nil {
			fmt.Printf("Config loaded from %s\n", path)
			return cfg
		}
	}

	// Try current working directory
	if cwd, err := os.Getwd(); err == nil {
		path := filepath.Join(cwd, "config.yaml")
		if cfg, err := Load(path); err == nil {
			fmt.Printf("Config loaded from %s\n", path)
			return cfg
		}
	}

	fmt.Println("No config.yaml found, using defaults")
	return Default()
}
