package automic

type Boolean int32
// Boolean is a custom type that represents a boolean value as an integer (0 or 1).
// It provides methods to set and get the boolean value in a thread-safe manner.
func (b *Boolean) Set(value bool) {
	if value {
		*b = 1
	} else {
		*b = 0
	}
}

func (b *Boolean) Get() bool {
	return *b == 1
}