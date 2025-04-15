package resp

// argument number error reply
type ArgNumErrReply struct {
	Cmd string
}

func (e *ArgNumErrReply) Error() string {
	return "ERR wrong number of arguments for '" + e.Cmd + "' command"
}
func (e *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + e.Cmd + "' command\r\n")
}
func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{Cmd: cmd}
}

// unknown command error reply
type UnknownCmdErrReply struct{}

func (e *UnknownCmdErrReply) Error() string {
	return "ERR unknown command"
}
func (e *UnknownCmdErrReply) ToBytes() []byte {
	return []byte("-ERR unknown command\r\n")
}
func MakeUnknownCmdErrReply() *UnknownCmdErrReply {
	return &UnknownCmdErrReply{}
}

// wrong type error reply
type WrongTypeErrReply struct{}

func (e *WrongTypeErrReply) Error() string {
	return "ERR Operation against a key holding the wrong kind of value"
}
func (e *WrongTypeErrReply) ToBytes() []byte {
	return []byte("-ERR Operation against a key holding the wrong kind of value\r\n")
}
func MakeWrongTypeErrReply() *WrongTypeErrReply {
	return &WrongTypeErrReply{}
}

// syntax error reply
type SyntaxErrReply struct{}

func (e *SyntaxErrReply) Error() string {
	return "ERR syntax error"
}
func (e *SyntaxErrReply) ToBytes() []byte {
	return []byte("-ERR syntax error\r\n")
}
func MakeSyntaxErrReply() *SyntaxErrReply {
	return &SyntaxErrReply{}
}

// protocol error reply
type ProtocolErrReply struct{}

func (e *ProtocolErrReply) Error() string {
	return "ERR protocol error"
}
func (e *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR protocol error\r\n")
}
func MakeProtocolErrReply() *ProtocolErrReply {
	return &ProtocolErrReply{}
}
