package logitech

const G923ProductID = 0xC267

// G923 represents a Logitech G923 racing wheel
type G923 struct {
	*LogitechWheel
}

// NewG923 creates a new Logitech G923 device instance
func NewG923() *G923 {
	return &G923{
		LogitechWheel: NewLogitechWheel("Logitech G923", G923ProductID),
	}
}
