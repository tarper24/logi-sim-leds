# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-26

### Added
- Initial release of logi-sim-leds
- Support for Logitech G29, G920, and G923 racing wheels
- Support for BeamNG.drive with OutGauge protocol
- Support for Assetto Corsa with native UDP telemetry
- Support for Dirt Rally, Dirt 4, and Codemasters F1 games
- Automatic device detection and hot-swapping
- Automatic game detection and switching
- RPM-based LED control with configurable thresholds
- LED flashing at high RPM
- Modular architecture for easy extensibility
- Cross-platform support (Windows, Linux)
- Configuration file support (YAML)
- Comprehensive documentation
- Build automation with Makefile

### Features
- Thread-safe concurrent processing
- Graceful handling of device disconnections
- Graceful handling of game switches
- Automatic max RPM detection
- Real-time telemetry processing
- Zero-configuration operation (works out of the box)

## [Unreleased]

### Planned
- Support for additional racing wheels (Thrustmaster, Fanatec)
- Support for additional games (iRacing, rFactor 2, Project CARS)
- Web-based configuration interface
- Enhanced logging and diagnostics
- Custom LED patterns and effects
- TCP telemetry support for remote setups
- Multiple device support (control multiple wheels simultaneously)
- Game-specific LED profiles
- Per-car RPM profiles
