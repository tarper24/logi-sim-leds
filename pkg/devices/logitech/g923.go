package logitech

const (
	G923XBoxProductID = 0xC267 // Xbox version
	G923PSProductID   = 0xC266 // PlayStation version
)

// G923 represents a Logitech G923 racing wheel
type G923 struct {
	*LogitechWheel
}

// NewG923XBox creates a new Logitech G923 (Xbox) device instance
func NewG923XBox() *G923 {
	return &G923{
		LogitechWheel: NewLogitechWheel("Logitech G923 (Xbox)", G923XBoxProductID),
	}
}

// NewG923PS creates a new Logitech G923 (PlayStation) device instance
func NewG923PS() *G923 {
	return &G923{
		LogitechWheel: NewLogitechWheel("Logitech G923 (PS)", G923PSProductID),
	}
}

// NewG923XBoxWithConfig creates a new Logitech G923 (Xbox) device instance with custom LED config
func NewG923XBoxWithConfig(ledCfg LEDConfig) *G923 {
	return &G923{
		LogitechWheel: NewLogitechWheelWithConfig("Logitech G923 (Xbox)", G923XBoxProductID, ledCfg),
	}
}

// NewG923PSWithConfig creates a new Logitech G923 (PlayStation) device instance with custom LED config
func NewG923PSWithConfig(ledCfg LEDConfig) *G923 {
	return &G923{
		LogitechWheel: NewLogitechWheelWithConfig("Logitech G923 (PS)", G923PSProductID, ledCfg),
	}
}
