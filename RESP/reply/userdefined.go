package reply

import (
	"bytes"
	"goredis/interface/resp"
	"strconv"
)

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

// bulk reply
type BulkReply struct {
	Arg []byte
}

func (b *BulkReply) ToBytes() []byte {
	if len(b.Arg) == 0 {
		return []byte("$0\r\n\r\n")

	}
	return []byte("$" + strconv.Itoa(len(b.Arg)) + "\r\n" + string(b.Arg) + "\r\n")
}
func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

// multi bulk reply
type MultiBulkReply struct {
	Args [][]byte
}

func (r *MultiBulkReply) ToBytes() []byte {
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(len(r.Args)) + "\r\n")
	for _, arg := range r.Args {
		if arg == nil {
			// nil 值使用 $-1\r\n 表示
			buf.WriteString("$-1\r\n")
		} else if len(arg) == 0 {
			buf.WriteString("$0\r\n\r\n")
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + "\r\n" + string(arg) + "\r\n")
		}
	}
	return buf.Bytes()
}
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: args}
}

// standard error reply
type StandardErrorReply struct {
	Err string
}

func (s *StandardErrorReply) ToBytes() []byte {
	return []byte("-ERR " + s.Err + "\r\n")
}
func MakeStandardErrorReply(err string) *StandardErrorReply {
	return &StandardErrorReply{Err: err}
}
func (r *StandardErrorReply) Error() string {
	return r.Err
}

// Integer reply
type IntegerReply struct {
	Code int64
}

func (i *IntegerReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(i.Code, 10) + "\r\n")
}
func MakeIntegerReply(code int64) *IntegerReply {
	return &IntegerReply{Code: code}
}

// status reply
type StatusReply struct {
	Code string
}

func (s *StatusReply) ToBytes() []byte {
	return []byte("+" + s.Code + "\r\n")
}
func MakeStatusReply(code string) *StatusReply {
	return &StatusReply{Code: code}
}

// what's more ,emplement a function to decide whether the reply is a error reply or not
func IsErrReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
