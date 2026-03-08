# logi-sim-leds

A modern, modular Golang application for controlling Logitech racing wheel LEDs based on telemetry data from racing simulation games.

![License](https://img.shields.io/badge/license-GPL--3.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)

## Features

- 🖥️ **Desktop UI**: Visual LED display, telemetry monitoring, and device/game management
- 🎮 **Multi-Game Support**: BeamNG.drive, Assetto Corsa, Dirt/Codemasters games
- 🎯 **Multi-Device Support**: Logitech G29, G920, and G923 racing wheels
- 🔄 **Hot-Swappable**: Automatically detect and switch between devices and games
- ⚡ **Real-Time LED Control**: RPM-based LED visualization with configurable thresholds
- 🧩 **Modular Architecture**: Clean interface-based design for easy extensibility
- 🔌 **Automatic Detection**: No manual configuration needed

## Supported Hardware

| Device                      | Model        | Status      |
| --------------------------- | ------------ | ----------- |
| Logitech G29                | Racing Wheel | ⚠️ Untested |
| Logitech G920               | Racing Wheel | ⚠️ Untested |
| Logitech G923 (PlayStation) | Racing Wheel | ✅ Tested   |
| Logitech G923 (Xbox)        | Racing Wheel | ⚠️ Untested |

## Supported Games

| Game           | Protocol    | Port  | Status      |
| -------------- | ----------- | ----- | ----------- |
| BeamNG.drive   | OutGauge    | 4444  | ✅ Tested   |
| Assetto Corsa  | AC Protocol | 9996  | ✅ Tested   |
| DiRT Rally     | Codemasters | 20777 | ✅ Tested   |
| DiRT Rally 2.0 | Codemasters | 20777 | ✅ Tested   |
| DiRT 4         | Codemasters | 20777 | ✅ Tested   |
| F1 2018        | Codemasters | 20777 | ✅ Tested   |
| F1 2019        | Codemasters | 20777 | ✅ Tested   |
| F1 2020        | Codemasters | 20777 | ✅ Tested   |
| F1 2021–25     | Codemasters | 20777 | ⚠️ Untested |

## Architecture

```text
logi-sim-leds/
├── cmd/
│   └── logi-sim-leds/      # Main application entry point
├── pkg/
│   ├── config/              # Configuration loading
│   ├── core/                # Core interfaces and types
│   ├── devices/             # Device drivers (modular)
│   │   └── logitech/        # Logitech wheel implementations
│   ├── games/               # Game telemetry parsers (modular)
│   │   ├── beamng/          # BeamNG.drive driver
│   │   ├── assettocorsa/    # Assetto Corsa driver
│   │   └── codemasters/     # Dirt/Codemasters/F1 driver
│   ├── manager/             # Orchestration layer
│   └── ui/                  # Desktop UI (Fyne)
├── config.yaml              # Configuration file
└── Makefile                 # Build automation
```

### Design Principles

- **Modular**: Each game and device is a separate module
- **Pluggable**: Easy to add new games or devices
- **Concurrent**: Uses Go's goroutines for parallel processing
- **Graceful**: Handles disconnections and reconnections seamlessly

## Installation

### Prerequisites

- Go 1.21 or later
- C compiler (GCC/MinGW) for CGo (required by Fyne desktop UI)
  - **Windows**: Install TDM-GCC from [tdm-gcc.tdragon.net](https://jmeubank.github.io/tdm-gcc/) or MinGW-w64 via [MSYS2](https://www.msys2.org/)
  - **Linux**: `sudo apt-get install gcc` (Debian/Ubuntu) or `sudo dnf install gcc` (Fedora)
  - **macOS**: `xcode-select --install` (installs Command Line Tools)
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
   ./build/logi-sim-leds.exe  # Windows
   # or
   ./build/logi-sim-leds      # Linux/macOS
   ```

3. The desktop UI will open, showing:
   - **LED Display**: Visual representation of wheel LEDs
   - **Telemetry**: Current and Max RPM display
   - **Device Selector**: Choose between connected wheels
   - **Game Selector**: View detected games
   - **Max RPM Editor**: Manually set max RPM if needed

4. Launch your racing game and enable telemetry output
5. The LEDs (physical and on-screen) will automatically light up based on your engine RPM!

### Game-Specific Setup

#### BeamNG.drive

BeamNG uses the [OutGauge protocol](https://documentation.beamng.com/modding/protocols/) to broadcast telemetry over UDP.

1. Launch BeamNG.drive
2. Go to **Options → Other**
3. Scroll down to the **Protocols** section
4. Enable **OutGauge UDP protocol**
5. Set IP: `127.0.0.1`
6. Set Port: `4444` (this is BeamNG's default — no change needed)
7. If already in a vehicle, press **Ctrl+R** to reset — this may be needed for changes to take effect

**Alternative — config file method:**

Edit the settings file directly (while the game is **closed**):

```
%LOCALAPPDATA%\BeamNG.drive\<version>\settings\cloud\game-settings-cloud.ini
```

> The game must have been launched at least once for this file to exist. Do not edit while the game is running — it will be overwritten on exit.

Works with all vehicles. Uses the same protocol as Live For Speed, so LFS-compatible tools also work.

#### Assetto Corsa

**No configuration required.** Assetto Corsa has a built-in UDP telemetry server on port 9996 that activates automatically during driving sessions.

1. Start logi-sim-leds
2. Launch Assetto Corsa and enter any driving session (practice, race, etc.)
3. Telemetry connects automatically

**How it works:** logi-sim-leds sends a handshake request to `127.0.0.1:9996`, AC responds with car/track info, then logi-sim-leds subscribes to real-time physics updates. If AC isn't running, it retries every 2 seconds.

**Notes:**
- AC must be **in a driving session** — the main menu does not broadcast telemetry
- Works in both single player and multiplayer
- **Only one telemetry client at a time** — close other telemetry apps (SimHub, dashboard apps) first
- Windows Firewall may block UDP port 9996 — allow logi-sim-leds if telemetry isn't connecting
- **Assetto Corsa Competizione (ACC) is NOT supported** — ACC uses a completely different protocol

#### DiRT Rally / DiRT Rally 2.0 / DiRT 4

All DiRT/Codemasters games use the same config file method — no in-game menu. Edit while the game is **closed**.

Find the `<udp>` line in `hardware_settings_config.xml` and set:

```xml
<udp enabled="true" extradata="3" ip="127.0.0.1" port="20777" delay="1" />
```

The config file location depends on the game. Replace `<USERNAME>` with your Windows username:

| Game           | Windows Path                                                              | Linux Path                                                                                                                         |
| -------------- | ------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| DiRT Rally     | `C:\Users\<USERNAME>\Documents\My Games\DiRT Rally\hardwaresettings\`     | `~/.local/share/feral-interactive/DiRT Rally/VFS/User/AppData/Roaming/My Games/DiRT Rally/hardwaresettings/`                       |
| DiRT Rally 2.0 | `C:\Users\<USERNAME>\Documents\My Games\DiRT Rally 2.0\hardwaresettings\` | `~/.local/share/Steam/steamapps/compatdata/690790/pfx/drive_c/users/steamuser/Documents/My Games/DiRT Rally 2.0/hardwaresettings/` |
| DiRT 4         | `C:\Users\<USERNAME>\Documents\My Games\DiRT 4\hardwaresettings\`         | `~/.local/share/feral-interactive/DiRT 4/VFS/User/AppData/Roaming/My Games/DiRT 4/hardwaresettings/`                               |

**Notes:**
- The `extradata="3"` setting enables all telemetry fields including RPM
- The game must have been launched at least once for the config file to exist
- Steam updates may occasionally reset this file
- DiRT Rally and DiRT 4 have native Linux (Feral) ports; DiRT Rally 2.0 runs via Proton only

#### F1 Series (F1 2018 – F1 25)

F1 games have an **in-game settings menu** — no config file editing needed.

1. Go to **Home → Game Options → Settings → Telemetry Settings**
2. Configure:

   | Setting            | Value                                   |
   | ------------------ | --------------------------------------- |
   | UDP Telemetry      | **On**                                  |
   | UDP Broadcast Mode | **Off**                                 |
   | UDP IP Address     | **127.0.0.1**                           |
   | UDP Port           | **20777**                               |
   | UDP Send Rate      | **20Hz** (recommended)                  |
   | UDP Format         | Match game year (e.g. "2024" for F1 24) |

Settings persist across sessions — configure once and forget. The menu path may vary slightly between game years, but the settings are the same from F1 2019 through F1 25.

> **Note:** EA Sports WRC uses a completely different telemetry format and is not currently supported.

### Configuration

Edit `config.yaml` to customize settings. The config file allows overriding default values for ports, LED thresholds, and other parameters:

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
    port: 4444
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
- `pkg/config/`: Configuration loading and defaults
- `pkg/devices/`: Device driver implementations
- `pkg/games/`: Game telemetry parser implementations
- `pkg/manager/`: Orchestration logic for connecting games to devices
- `pkg/ui/`: Desktop UI built with Fyne
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

3. Register your game in `pkg/manager/manager.go`'s `NewManager()` function by adding it to the `games` slice:

   ```go
   games := []core.GameInterface{
       beamng.NewBeamNG(),
       assettocorsa.NewAssettoCorsa(),
       codemasters.NewCodemasters(),
       yourgame.NewYourGame(), // Add your game here
   }
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

3. Register your device in `pkg/devices/logitech/detector.go`'s `NewDetector()` function by adding it to the `supportedDevices` map:

   ```go
   supportedDevices: map[uint16]func() core.DeviceInterface{
       // ...existing devices...
       YourProductID: func() core.DeviceInterface { return NewYourDevice() },
   }
   ```

## Troubleshooting

### Build Errors (CGo/GCC)

If you encounter errors like `C compiler "gcc" not found` or `build constraints exclude all Go files`:

- **Cause**: The desktop UI (Fyne) requires CGo and a C compiler
- **Windows**: Install TDM-GCC or MinGW-w64 and ensure it's in your PATH
- **Linux**: Install gcc: `sudo apt-get install gcc`
- **macOS**: Install Xcode Command Line Tools: `xcode-select --install`

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

## Known Limitations

- Only one device can be active at a time (connects to the first detected)
- Assetto Corsa Competizione (ACC) and EA Sports WRC are not supported

## Roadmap / TODO

- [ ] **UI LED Threshold Editor** — LED thresholds are configurable via `config.yaml`, but an in-app editor would improve UX

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Credits

Inspired by:

- [beamng-shifting-leds](https://github.com/gamingdoom/beamng-shifting-leds) by gamingdoom
- [ac_shifting_leds](https://github.com/d4rk/ac_shifting_leds) by d4rk
- [out-gauge-cluster](https://github.com/fuelsoft/out-gauge-cluster) by fuelsoft

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
