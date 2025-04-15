package main

import (
	"bufio"
	"bytes"
	"fmt"
	"goredis/RESP/parser"
)

func main() {
	// 测试简单字符串请求
	//	testSimpleStringRequest()

	//	// 测试整数请求
	//	testIntegerRequest()
	//
	// 测试批量字符串请求
	//testBulkStringRequest()

	// 测试多条批量字符串请求
	testMultiBulkRequest()

	// 测试空多条批量字符串请求
	//testEmptyMultiBulkRequest()
}

func testSimpleStringRequest() {
	input := "+PING\r\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))
	ch := parser.ParseStream(reader)

	result := <-ch
	if result.Err != nil {
		fmt.Println("Simple String Request Test Failed:", result.Err)
	} else {
		fmt.Println("Simple String Request Test Passed:", string(result.Data.ToBytes()))
	}
}

func testIntegerRequest() {
	input := ":1000\r\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))
	ch := parser.ParseStream(reader)

	result := <-ch
	if result.Err != nil {
		fmt.Println("Integer Request Test Failed:", result.Err)
	} else {
		fmt.Println("Integer Request Test Passed:", string(result.Data.ToBytes()))
	}
}

func testBulkStringRequest() {
	input := "$4\r\nPING\r\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))
	ch := parser.ParseStream(reader)

	result := <-ch
	if result.Err != nil {
		fmt.Println("Bulk String Request Test Failed:", result.Err)
	} else {
		fmt.Println("Bulk String Request Test Passed:", string(result.Data.ToBytes()))
	}
}
func testMultiBulkRequest() {
	input := "*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))
	ch := parser.ParseStream(reader)

	result := <-ch
	if result.Err != nil {
		fmt.Println("Multi Bulk Request Test Failed:", result.Err)
	} else {
		fmt.Println("Multi Bulk Request Test Passed:", string(result.Data.ToBytes()))
	}
}

func testEmptyMultiBulkRequest() {
	input := "*0\r\n"
	reader := bufio.NewReader(bytes.NewReader([]byte(input)))
	ch := parser.ParseStream(reader)

	result := <-ch
	if result.Err != nil {
		fmt.Println("Empty Multi Bulk Request Test Failed:", result.Err)
	} else {
		fmt.Println("Empty Multi Bulk Request Test Passed:", string(result.Data.ToBytes()))
	}
}
