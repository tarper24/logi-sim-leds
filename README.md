# logi-sim-leds

A modern, modular Golang application for controlling Logitech racing wheel LEDs based on telemetry data from racing simulation games.

![License](https://img.shields.io/badge/license-GPL--3.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)

## Features

- 🎮 **Multi-Game Support**: BeamNG.drive, Assetto Corsa, Dirt/Codemasters games
- 🎯 **Multi-Device Support**: Logitech G29, G920, and G923 racing wheels
- 🔄 **Hot-Swappable**: Automatically detect and switch between devices and games
- ⚡ **Real-Time LED Control**: RPM-based LED visualization with configurable thresholds
- 🧩 **Modular Architecture**: Plugin-based design for easy extensibility
- 🔌 **Automatic Detection**: No manual configuration needed
- 💪 **Production Ready**: Thread-safe, error-resilient, optimized performance

## Supported Hardware

| Device        | Model        | Status       |
| ------------- | ------------ | ------------ |
| Logitech G29  | Racing Wheel | ✅ Supported |
| Logitech G920 | Racing Wheel | ✅ Supported |
| Logitech G923 | Racing Wheel | ✅ Supported |

## Supported Games

| Game                | Protocol    | Port  | Status       |
| ------------------- | ----------- | ----- | ------------ |
| BeamNG.drive        | OutGauge    | 5555  | ✅ Supported |
| Assetto Corsa       | AC Protocol | 9996  | ✅ Supported |
| Dirt Rally / Dirt 4 | Codemasters | 20777 | ✅ Supported |
| F1 Series           | Codemasters | 20777 | ✅ Supported |

## Architecture

```text
logi-sim-leds/
├── cmd/
│   └── logi-sim-leds/      # Main application entry point
├── pkg/
│   ├── core/               # Core interfaces and types
│   ├── devices/            # Device drivers (modular)
│   │   └── logitech/       # Logitech wheel implementations
│   ├── games/              # Game telemetry parsers (modular)
│   │   ├── beamng/         # BeamNG.drive driver
│   │   ├── assettocorsa/   # Assetto Corsa driver
│   │   └── dirt/           # Dirt/Codemasters driver
│   └── manager/            # Orchestration layer
├── config.yaml             # Configuration file
└── Makefile               # Build automation

```

### Design Principles

- **Modular**: Each game and device is a separate module
- **Pluggable**: Easy to add new games or devices
- **Concurrent**: Uses Go's goroutines for parallel processing
- **Graceful**: Handles disconnections and reconnections seamlessly

## Installation

### Prerequisites

- Go 1.21 or later
- HID library (libhidapi)
  - **Windows**: No additional dependencies (uses native HID)
  - **Linux**: `sudo apt-get install libhidapi-dev` (Debian/Ubuntu) or `sudo dnf install hidapi-devel` (Fedora)
  - **macOS**: `brew install hidapi`

### Linux Permissions

On Linux, you need to set up udev rules to access the HID device without root:

```bash
# Create udev rule
sudo tee /etc/udev/rules.d/99-logitech-wheel.rules << EOF
# Logitech G29
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="046d", ATTRS{idProduct}=="c24f", MODE="0666"
# Logitech G920
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="046d", ATTRS{idProduct}=="c262", MODE="0666"
# Logitech G923
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="046d", ATTRS{idProduct}=="c267", MODE="0666"
EOF

# Reload udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/tarper24/logi-sim-leds.git
cd logi-sim-leds

# Download dependencies
make deps

# Build the application
make build

# Run the application
make run
```

### Cross-Platform Build

```bash
# Build for Windows
make build-windows

# Build for Linux
make build-linux

# Build for all platforms
make build-all
```

## Usage

### Quick Start

1. Connect your Logitech racing wheel (G29, G920, or G923)
2. Run the application:

   ```bash
   ./build/logi-sim-leds
   ```

3. Launch your racing game and enable telemetry output
4. The LEDs will automatically light up based on your engine RPM!

### Game-Specific Setup

#### BeamNG.drive

1. Launch BeamNG.drive
2. Press ESC → Settings → Other
3. Enable "OutGauge Support"
4. Set IP: `127.0.0.1`
5. Set Port: `5555`

#### Assetto Corsa

1. The application automatically connects to Assetto Corsa
2. No additional configuration needed
3. Just start driving!

#### Dirt Rally / Dirt 4

1. Enable UDP telemetry in game settings:
   - Edit: `~/.local/share/feral-interactive/DiRT 4/VFS/User/AppData/Roaming/My Games/DiRT 4/hardwaresettings/hardware_settings_config.xml`
   - Change `<udp enabled="false"` to `<udp enabled="true"`
2. Ensure IP is `127.0.0.1` and port is `20777`

### Configuration

Edit `config.yaml` to customize settings:

```yaml
# LED thresholds (% of max RPM)
leds:
  led1_threshold: 45
  led2_threshold: 55
  led3_threshold: 62.5
  led4_threshold: 71
  led5_threshold: 85
  flash_threshold: 93
  flash_interval: 100

# Game ports
games:
  beamng:
    port: 5555
  assetto_corsa:
    port: 9996
  dirt:
    port: 20777
```

## How It Works

### LED Behavior

The wheel LEDs light up progressively as RPM increases:

- **LED 1**: 45% of max RPM
- **LED 2**: 55% of max RPM
- **LED 3**: 62.5% of max RPM
- **LED 4**: 71% of max RPM
- **LED 5**: 85% of max RPM
- **Flash**: 93%+ of max RPM (all LEDs flash rapidly)

### Hot-Swapping

The application automatically handles:

- **Device Changes**: Plug/unplug wheels while running
- **Game Switching**: Switch between games without restarting
- **Connection Loss**: Automatic reconnection when connection is restored

## Development

### Project Structure

- `pkg/core/`: Core interfaces defining contracts for devices and games
- `pkg/devices/`: Device driver implementations
- `pkg/games/`: Game telemetry parser implementations
- `pkg/manager/`: Orchestration logic for connecting games to devices
- `cmd/`: Application entry points

### Adding a New Game

1. Create a new package in `pkg/games/yourgame/`
2. Implement the `core.GameInterface`:

   ```go
   type YourGame struct {}
   
   func (g *YourGame) GetName() string { return "Your Game" }
   func (g *YourGame) Start(ctx context.Context, dataChan chan<- core.TelemetryData) error { ... }
   func (g *YourGame) Stop() error { ... }
   func (g *YourGame) IsRunning() bool { ... }
   func (g *YourGame) GetPort() int { ... }
   ```

### Adding a New Device

1. Create a new package in `pkg/devices/yourdevice/`
2. Implement the `core.DeviceInterface`:

   ```go
   type YourDevice struct {}
   
   func (d *YourDevice) GetName() string { ... }
   func (d *YourDevice) GetID() string { ... }
   func (d *YourDevice) Connect() error { ... }
   func (d *YourDevice) Disconnect() error { ... }
   func (d *YourDevice) IsConnected() bool { ... }
   func (d *YourDevice) UpdateLEDs(data core.TelemetryData) error { ... }
   func (d *YourDevice) SetLEDMask(mask uint8) error { ... }
   ```

## Troubleshooting

### Device Not Found

- **Windows**: Ensure the wheel is recognized in Device Manager
- **Linux**: Check udev rules (see Linux Permissions above)
- **All**: Try unplugging and replugging the wheel

### No Telemetry Data

- Verify the game's telemetry output is enabled
- Check that the correct port is configured
- Ensure no firewall is blocking UDP traffic

### LEDs Not Lighting Up

- Max RPM might be incorrectly detected
- Try manually setting max RPM in config for your car
- Check that device is properly connected

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Credits

Inspired by:

- [beamng-shifting-leds](https://github.com/gamingdoom/beamng-shifting-leds) by gamingdoom
- [ac_shifting_leds](https://github.com/d4rk/ac_shifting_leds) by d4rk
- [out-gauge-cluster](https://github.com/fuelsoft/out-gauge-cluster) by fuelsoft

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes in each version.
