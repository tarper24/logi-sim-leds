package logitech

const G29ProductID = 0xC24F

// G29 represents a Logitech G29 racing wheel
type G29 struct {
	*LogitechWheel
}

// NewG29 creates a new Logitech G29 device instance
func NewG29() *G29 {
	return &G29{
		LogitechWheel: NewLogitechWheel("Logitech G29", G29ProductID),
	}
}
