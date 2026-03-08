package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"
	"github.com/tarper24/logi-sim-leds/pkg/config"
	"github.com/tarper24/logi-sim-leds/pkg/core"
	"github.com/tarper24/logi-sim-leds/pkg/logging"
	"github.com/tarper24/logi-sim-leds/pkg/manager"
	"github.com/tarper24/logi-sim-leds/pkg/ui"
)

const (
	version = "1.0.0"
)

// resolveLogPath returns the log file path relative to the executable.
func resolveLogPath() string {
	exePath, _ := os.Executable()
	return filepath.Join(filepath.Dir(exePath), "logi-sim-leds.log")
}

// uiUpdater holds callbacks for UI updates, decoupled from any specific UI framework.
type uiUpdater struct {
	onTelemetry    func(core.TelemetryData)
	onDeviceChange func(deviceName string)
	onGameChange   func(gameName string)
	getDevices     func() []string
}

// startUILoop reads from manager channels and updates the UI.
// It exits when the done channel closes.
func startUILoop(done <-chan struct{}, telemetryChan <-chan core.TelemetryData, deviceChan <-chan string, gameChan <-chan string, updater uiUpdater) {
	for {
		select {
		case <-done:
			return
		case data := <-telemetryChan:
			updater.onTelemetry(data)
		case deviceName := <-deviceChan:
			if deviceName != "" {
				updater.onDeviceChange(deviceName)
			}
			updater.getDevices()
		case gameName := <-gameChan:
			updater.onGameChange(gameName)
		}
	}
}

func main() {
	debug := flag.Bool("debug", false, "enable debug logging to debug.log")
	flag.Parse()

	logPath := resolveLogPath()

	if err := logging.Setup(*debug, logPath); err != nil {
		slog.Error("failed to setup logging", "error", err)
	}

	// Load configuration
	cfg := config.LoadOrDefault()

	// Create the Fyne application
	a := app.New()

	// Create manager with config
	mgr := manager.NewManager(cfg)

	// Start the manager
	if err := mgr.Start(); err != nil {
		slog.Error("failed to start manager", "error", err)
		return
	}

	// Create the UI
	appUI := ui.NewAppUI(a)

	// Set initial available devices and games
	appUI.SetAvailableDevices(mgr.GetAvailableDevices())
	appUI.SetAvailableGames(mgr.GetAvailableGames())

	// Set initial active device and game
	if activeDevice := mgr.GetActiveDevice(); activeDevice != nil {
		appUI.UpdateDevice(activeDevice.GetName())
	}
	if activeGame := mgr.GetActiveGame(); activeGame != nil {
		appUI.UpdateGame(activeGame.GetName())
	}

	// Wire up UI callbacks
	appUI.SetOnDeviceChange(func(deviceName string) {
		if err := mgr.SetActiveDevice(deviceName); err != nil {
			slog.Error("failed to switch device", "device", deviceName, "error", err)
		}
	})

	appUI.SetOnMaxRPMChange(func(rpm float32) {
		if err := mgr.SetMaxRPM(rpm); err != nil {
			slog.Error("failed to set max RPM", "rpm", rpm, "error", err)
		}
	})

	// Start listening to manager channels and update UI
	go startUILoop(mgr.Done(), mgr.GetUITelemetryChan(), mgr.GetUIDeviceChan(), mgr.GetUIGameChan(), uiUpdater{
		onTelemetry:    appUI.UpdateTelemetry,
		onDeviceChange: appUI.UpdateDevice,
		onGameChange:   appUI.UpdateGame,
		getDevices: func() []string {
			return mgr.GetAvailableDevices()
		},
	})

	// Show the window and run (blocks until window is closed)
	appUI.Show()
	a.Run()

	// Cleanup when application exits
	slog.Info("shutting down")
	if err := mgr.Stop(); err != nil {
		slog.Error("error during shutdown", "error", err)
	}
}
