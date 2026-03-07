package logitech

const G920ProductID = 0xC262 // Xbox only

// G920 represents a Logitech G920 racing wheel
type G920 struct {
	*LogitechWheel
}

// NewG920 creates a new Logitech G920 device instance
func NewG920() *G920 {
	return &G920{
		LogitechWheel: NewLogitechWheel("Logitech G920 (Xbox)", G920ProductID),
	}
}

// NewG920WithConfig creates a new Logitech G920 device instance with custom LED config
func NewG920WithConfig(ledCfg LEDConfig) *G920 {
	return &G920{
		LogitechWheel: NewLogitechWheelWithConfig("Logitech G920 (Xbox)", G920ProductID, ledCfg),
	}
}
