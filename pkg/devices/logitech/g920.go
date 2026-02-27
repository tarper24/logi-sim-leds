package logitech

const G920ProductID = 0xC262

// G920 represents a Logitech G920 racing wheel
type G920 struct {
	*LogitechWheel
}

// NewG920 creates a new Logitech G920 device instance
func NewG920() *G920 {
	return &G920{
		LogitechWheel: NewLogitechWheel("Logitech G920", G920ProductID),
	}
}
