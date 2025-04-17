package resp

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}
