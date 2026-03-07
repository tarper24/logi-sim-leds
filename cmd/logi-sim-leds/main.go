package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"
	"github.com/tarper24/logi-sim-leds/pkg/manager"
	"github.com/tarper24/logi-sim-leds/pkg/ui"
)

const (
	version = "1.0.0"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug logging to debug.log")
	flag.Parse()

	// Resolve log path relative to the executable so it's always findable
	exePath, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(exePath), "debug.log")

	if *debug {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err == nil {
			os.Stdout = f
			os.Stderr = f
		}
	} else {
		devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		os.Stdout = devNull
		os.Stderr = devNull
	}

	// Create the Fyne application
	a := app.New()

	// Create manager with auto-detect enabled
	mgr := manager.NewManager(true)

	// Start the manager
	if err := mgr.Start(); err != nil {
		fmt.Printf("Failed to start manager: %v\n", err)
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
			fmt.Printf("Failed to switch device: %v\n", err)
		}
	})

	appUI.SetOnMaxRPMChange(func(rpm float32) {
		if err := mgr.SetMaxRPM(rpm); err != nil {
			fmt.Printf("Failed to set max RPM: %v\n", err)
		}
	})

	// Start listening to manager channels and update UI
	go func() {
		for {
			select {
			case data := <-mgr.GetUITelemetryChan():
				appUI.UpdateTelemetry(data)
			case deviceName := <-mgr.GetUIDeviceChan():
				if deviceName != "" {
					appUI.UpdateDevice(deviceName)
				}
				appUI.SetAvailableDevices(mgr.GetAvailableDevices())
			case gameName := <-mgr.GetUIGameChan():
				appUI.UpdateGame(gameName)
			}
		}
	}()

	// Show the window and run (blocks until window is closed)
	appUI.Show()
	a.Run()

	// Cleanup when application exits
	fmt.Println("Shutting down...")
	if err := mgr.Stop(); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
	}
}
