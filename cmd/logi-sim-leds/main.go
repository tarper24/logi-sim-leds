package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tarper24/logi-sim-leds/pkg/manager"
)

const (
	version = "1.0.0"
)

func main() {
	printBanner()

	// Create manager with auto-detect enabled
	mgr := manager.NewManager(true)

	// Start the manager
	if err := mgr.Start(); err != nil {
		fmt.Printf("Failed to start manager: %v\n", err)
		os.Exit(1)
	}

	// Print status periodically
	statusTicker := time.NewTicker(10 * time.Second)
	defer statusTicker.Stop()

	go func() {
		for range statusTicker.C {
			fmt.Println("\n" + mgr.GetStatus())
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("\nPress Ctrl+C to exit...")
	<-sigChan

	fmt.Println("\nShutting down...")
	if err := mgr.Stop(); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Goodbye!")
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                   LOGI-SIM-LEDS v%s                   ║
║          Logitech Racing Wheel LED Controller         ║
╚═══════════════════════════════════════════════════════════╝

Supported Devices:
  • Logitech G29
  • Logitech G920
  • Logitech G923

Supported Games:
  • BeamNG.drive
  • Assetto Corsa
  • Dirt/Codemasters (Dirt Rally, Dirt 4, F1 series)

Features:
  ✓ Automatic device detection
  ✓ Hot-swappable devices and games
  ✓ RPM-based LED control
  ✓ Multi-game support

`
	fmt.Printf(banner, version)
}
