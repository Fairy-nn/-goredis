package reply

// 处理 RESP 协议的回复类型
// reply PING --> PONG
type PongReply struct{}

func (p *PongReply) ToBytes() []byte {
	return []byte("+PONG\r\n")
}
func MakePongReply() *PongReply { //工厂函数
	return &PongReply{}
}

// reply set --> OK
type OKReply struct{}

func (o *OKReply) ToBytes() []byte {
	return []byte("+OK\r\n")
}
func MakeOKReply() *OKReply {
	return &OKReply{}
}

// reply nil --> Null
type NullReply struct{}

func (n *NullReply) ToBytes() []byte {
	return []byte("$-1\r\n")
}
func MakeNullReply() *NullReply {
	return &NullReply{}
}

// reply empty bulk --> $0\r\n\r\n
type EmptyBulkReply struct{}

func (e *EmptyBulkReply) ToBytes() []byte {
	return []byte("$0\r\n\r\n")
}
func MakeEmptyBulkReply() *EmptyBulkReply {
	return &EmptyBulkReply{}
}

// reply empty multi bulk --> *0\r\n
type EmptyMultiBulkReply struct{}

func (e *EmptyMultiBulkReply) ToBytes() []byte {
	return []byte("*0\r\n")
}
func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}

// no reply --> nil
type NoReply struct{}

func (n *NoReply) ToBytes() []byte {
	return nil
}
func MakeNoReply() *NoReply {
	return &NoReply{}
}
