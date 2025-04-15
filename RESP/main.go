package main

import (
	"fmt"
	resp "goredis/RESP/interface"
)

func main() {
	// 测试 BulkReply
	bulkReply := resp.MakeBulkReply([]byte("hello"))
	fmt.Println("BulkReply ToBytes():", string(bulkReply.ToBytes()))

	// 测试空的 BulkReply
	emptyBulkReply := resp.MakeBulkReply([]byte(""))
	fmt.Println("Empty BulkReply ToBytes():", string(emptyBulkReply.ToBytes()))

	// 测试 MultiBulkReply
	multiBulkReply := resp.MakeMultiBulkReply([][]byte{
		[]byte("hello"),
		[]byte("world"),
		[]byte(""),
	})
	fmt.Println("MultiBulkReply ToBytes():", string(multiBulkReply.ToBytes()))

	// 测试 StandardErrorReply
	errorReply := resp.MakeStandardErrorReply("unknown command")
	fmt.Println("StandardErrorReply ToBytes():", string(errorReply.ToBytes()))

	// 测试 IntegerReply
	integerReply := resp.MakeIntegerReply(12345)
	fmt.Println("IntegerReply ToBytes():", string(integerReply.ToBytes()))

	// 测试 StatusReply
	statusReply := &resp.StatusReply{Code: "OK"}
	fmt.Println("StatusReply ToBytes():", string(statusReply.ToBytes()))

	// 测试 StatusReply 是否为错误
	errorStatusReply := &resp.StatusReply{Code: "-ERR something went wrong"}
	fmt.Println("IsErrorReply():", errorStatusReply.IsErrorReply())
	// // 测试 ArgNumErrReply
	// argNumErr := resp.MakeArgNumErrReply("set")
	// fmt.Println("ArgNumErrReply Error():", argNumErr.Error())
	// fmt.Println("ArgNumErrReply ToBytes():", string(argNumErr.ToBytes()))

	// // 创建 OKReply 实例
	// okReply := resp.MakeOKReply()
	// // 将响应转换为字节切片
	// bytes := okReply.ToBytes()
	// // 打印字节切片转换后的字符串
	// fmt.Println(string(bytes))
}
