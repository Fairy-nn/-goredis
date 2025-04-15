package resp

// an interface for connection
type Connection interface {
	Write(data []byte) error
	GetDBIndex() int
	SelectDB(int) error
}

// an interface for reply
type Reply interface {
	ToBytes() []byte
}
