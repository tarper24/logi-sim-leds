# Getting Started with logi-sim-leds

This guide will help you get up and running with logi-sim-leds quickly.

## Quick Start (5 minutes)

### Step 1: Install Go

If you don't have Go installed:

**Windows:**
- Download from https://go.dev/dl/
- Run the installer
- Verify: `go version` in PowerShell

**Linux:**
```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install golang-go libhidapi-dev

# Fedora
sudo dnf install golang hidapi-devel
```

**macOS:**
```bash
brew install go hidapi
```

### Step 2: Set Up Linux Permissions (Linux only)

```bash
# Create udev rule
sudo tee /etc/udev/rules.d/99-logitech-wheel.rules << EOF
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="046d", ATTRS{idProduct}=="c24f", MODE="0666"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="046d", ATTRS{idProduct}=="c262", MODE="0666"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="046d", ATTRS{idProduct}=="c267", MODE="0666"
EOF

# Reload
sudo udevadm control --reload-rules
sudo udevadm trigger
```

### Step 3: Build and Run

```bash
# Navigate to the project directory
cd logi-sim-leds

# Download dependencies
go mod download

# Build
make build

# Run
./build/logi-sim-leds
```

You should see:
```
╔═══════════════════════════════════════════════════════════╗
║                   LOGI-SIM-LEDS v1.0.0                   ║
║          Logitech Racing Wheel LED Controller         ║
╚═══════════════════════════════════════════════════════════╝

Starting logi-sim-leds manager...
Connected to device: Logitech G29
Manager started successfully

Press Ctrl+C to exit...
```

### Step 4: Configure Your Game

#### For BeamNG.drive:
1. Launch BeamNG.drive
2. Press ESC → Settings → Other
3. Enable "OutGauge Support"
4. IP: `127.0.0.1`, Port: `5555`
5. Start driving!

#### For Assetto Corsa:
1. Just launch and drive!
2. The app auto-connects (no setup needed)

#### For Dirt Rally / Dirt 4:
1. Enable UDP telemetry in game config
2. Ensure port is `20777`

### Step 5: Drive and Watch the LEDs!

The LEDs will progressively light up as your RPM increases and flash when you reach the redline.

## Common Issues

### "Device not found"
- Make sure your wheel is plugged in and powered on
- On Linux, check udev rules (see Step 2)
- Try unplugging and replugging

### "No telemetry data"
- Make sure the game's telemetry is enabled
- Check that no firewall is blocking UDP ports
- Verify the correct port in config

### "Permission denied" (Linux)
- Run the udev rule setup again (Step 2)
- Log out and back in
- Or run with `sudo` (not recommended)

## What's Next?

- Customize LED thresholds in `config.yaml`
- Try different games and cars
- Check [README.md](README.md) for advanced features
- See [CONTRIBUTING.md](CONTRIBUTING.md) to add new games/devices

## Need Help?

Open an issue on GitHub with:
- Your OS and Go version
- Device model (G29/G920/G923)
- Game and version
- Error messages or logs

Happy racing! 🏁
