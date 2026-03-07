package logitech

const G29ProductID = 0xC24F // PlayStation only

// G29 represents a Logitech G29 racing wheel
type G29 struct {
	*LogitechWheel
}

// NewG29 creates a new Logitech G29 device instance
func NewG29() *G29 {
	return &G29{
		LogitechWheel: NewLogitechWheel("Logitech G29 (PS)", G29ProductID),
	}
}

// NewG29WithConfig creates a new Logitech G29 device instance with custom LED config
func NewG29WithConfig(ledCfg LEDConfig) *G29 {
	return &G29{
		LogitechWheel: NewLogitechWheelWithConfig("Logitech G29 (PS)", G29ProductID, ledCfg),
	}
}
