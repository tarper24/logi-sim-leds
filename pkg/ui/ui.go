package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/tarper24/logi-sim-leds/pkg/core"
)

// AppUI represents the main application UI
type AppUI struct {
	window fyne.Window

	// Bindings (thread-safe for goroutine updates)
	rpmDisplay binding.String

	// Selectors
	deviceSelect *widget.Select
	gameSelect   *widget.Select

	// Max RPM Entry
	maxRPMEntry *widget.Entry

	// Callbacks
	onDeviceChange  func(string)
	onMaxRPMChange  func(float32)

	// Guard flag to suppress OnChanged callbacks during programmatic updates
	updatingDevice bool
}

// NewAppUI creates a new application UI
func NewAppUI(app fyne.App) *AppUI {
	window := app.NewWindow("Logi-Sim-LEDs - Racing Wheel LED Controller")
	window.Resize(fyne.NewSize(420, 220))

	ui := &AppUI{
		window:     window,
		rpmDisplay: binding.NewString(),
	}

	_ = ui.rpmDisplay.Set("0 / 0")
	ui.setupUI()
	return ui
}

// setupUI creates the UI layout
func (ui *AppUI) setupUI() {
	ui.deviceSelect = widget.NewSelect([]string{}, func(value string) {
		if ui.updatingDevice {
			return
		}
		if ui.onDeviceChange != nil && value != "" {
			ui.onDeviceChange(value)
		}
	})
	ui.deviceSelect.PlaceHolder = "Select device..."

	ui.gameSelect = widget.NewSelect([]string{}, func(_ string) {})
	ui.gameSelect.PlaceHolder = "Auto-detect"

	ui.maxRPMEntry = widget.NewEntry()
	ui.maxRPMEntry.SetPlaceHolder("e.g. 7000")
	setBtn := widget.NewButton("Set", func() {
		if rpm, err := strconv.ParseFloat(ui.maxRPMEntry.Text, 32); err == nil {
			if ui.onMaxRPMChange != nil {
				ui.onMaxRPMChange(float32(rpm))
			}
		}
	})
	ui.maxRPMEntry.OnSubmitted = func(value string) {
		if rpm, err := strconv.ParseFloat(value, 32); err == nil {
			if ui.onMaxRPMChange != nil {
				ui.onMaxRPMChange(float32(rpm))
			}
		}
	}

	rpmLabel := widget.NewLabelWithData(ui.rpmDisplay)

	form := container.New(layout.NewFormLayout(),
		widget.NewLabel("Device:"), ui.deviceSelect,
		widget.NewLabel("Game:"), ui.gameSelect,
		widget.NewLabel("Max RPM:"), container.NewBorder(nil, nil, nil, setBtn, ui.maxRPMEntry),
		widget.NewLabel("RPM:"), rpmLabel,
	)

	ui.window.SetContent(container.NewPadded(form))
}

// UpdateTelemetry updates the RPM display — safe to call from goroutines via binding
func (ui *AppUI) UpdateTelemetry(data core.TelemetryData) {
	_ = ui.rpmDisplay.Set(fmt.Sprintf("%.0f / %.0f", data.RPM, data.MaxRPM))
}

// UpdateDevice reflects the currently active device in the dropdown
func (ui *AppUI) UpdateDevice(deviceName string) {
	if deviceName != "" {
		ui.updatingDevice = true
		ui.deviceSelect.SetSelected(deviceName)
		ui.updatingDevice = false
	}
}

// UpdateGame reflects the currently active game in the dropdown
func (ui *AppUI) UpdateGame(gameName string) {
	if gameName != "" {
		ui.gameSelect.SetSelected(gameName)
	}
}

// SetAvailableDevices populates the device dropdown
func (ui *AppUI) SetAvailableDevices(devices []string) {
	ui.updatingDevice = true
	ui.deviceSelect.Options = devices
	ui.deviceSelect.Refresh()
	if len(devices) > 0 && ui.deviceSelect.Selected == "" {
		ui.deviceSelect.SetSelected(devices[0])
	}
	ui.updatingDevice = false
}

// SetAvailableGames populates the game dropdown
func (ui *AppUI) SetAvailableGames(games []string) {
	ui.gameSelect.Options = games
	ui.gameSelect.Refresh()
}

// SetOnDeviceChange sets the callback for device selection changes
func (ui *AppUI) SetOnDeviceChange(callback func(string)) {
	ui.onDeviceChange = callback
}

// SetOnMaxRPMChange sets the callback for max RPM changes (receives parsed float32)
func (ui *AppUI) SetOnMaxRPMChange(callback func(float32)) {
	ui.onMaxRPMChange = callback
}

// Show shows the main window (non-blocking; caller must invoke app.Run)
func (ui *AppUI) Show() {
	ui.window.Show()
}

// StartUpdateLoop is a no-op; main.go owns the update goroutine
func (ui *AppUI) StartUpdateLoop() {}
