package main

import (
	"bufio"
	"bytes"
	"fmt"
	"goredis/RESP/parser"
)

func main() {

	// 测试正常的单行输入
	testReadLineSingleLine()

	// 测试无效的行终止符
	testReadLineInvalidTerminator()

	// 测试空输入
	testReadLineEmptyInput()
	// // 测试 BulkReply
	// bulkReply := resp.MakeBulkReply([]byte("hello"))
	// fmt.Println("BulkReply ToBytes():", string(bulkReply.ToBytes()))

	// // 测试空的 BulkReply
	// emptyBulkReply := resp.MakeBulkReply([]byte(""))
	// fmt.Println("Empty BulkReply ToBytes():", string(emptyBulkReply.ToBytes()))

	// // 测试 MultiBulkReply
	// multiBulkReply := resp.MakeMultiBulkReply([][]byte{
	// 	[]byte("hello"),
	// 	[]byte("world"),
	// 	[]byte(""),
	// })
	// fmt.Println("MultiBulkReply ToBytes():", string(multiBulkReply.ToBytes()))

	// // 测试 StandardErrorReply
	// errorReply := resp.MakeStandardErrorReply("unknown command")
	// fmt.Println("StandardErrorReply ToBytes():", string(errorReply.ToBytes()))

	// // 测试 IntegerReply
	// integerReply := resp.MakeIntegerReply(12345)
	// fmt.Println("IntegerReply ToBytes():", string(integerReply.ToBytes()))

	// // 测试 StatusReply
	// statusReply := &resp.StatusReply{Code: "OK"}
	// fmt.Println("StatusReply ToBytes():", string(statusReply.ToBytes()))

	// // 测试 StatusReply 是否为错误
	// errorStatusReply := &resp.StatusReply{Code: "-ERR something went wrong"}
	// fmt.Println("IsErrorReply():", errorStatusReply.IsErrorReply())
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
func testReadLineSingleLine() {
	input := "+OK\r\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))
	state := &parser.ParserResult{}

	line, ioErr, err := parser.ReadLine(reader, state)
	if err != nil {
		fmt.Println("Single Line Test Failed:", err)
	} else if ioErr {
		fmt.Println("Single Line Test Failed: Unexpected IO error")
	} else {
		fmt.Println("Single Line Test Passed:", string(line))
	}
}

func testReadLineInvalidTerminator() {
	input := "+OK\n" // 缺少 \r
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))
	state := &parser.ParserResult{}

	line, _, err := parser.ReadLine(reader, state)
	if err != nil {
		fmt.Println("Invalid Terminator Test Passed:", err)
	} else {
		fmt.Println("Invalid Terminator Test Failed: Expected error but got:", string(line))
	}
}

func testReadLineEmptyInput() {
	input := ""
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))
	state := &parser.ParserResult{}

	line, ioErr, err := parser.ReadLine(reader, state)
	if err != nil && ioErr {
		fmt.Println("Empty Input Test Passed: IO error as expected")
	} else {
		fmt.Println("Empty Input Test Failed: Expected IO error but got:", string(line))
	}
}
