package main

import (
	"fmt"
	resp "goredis/RESP/interface"
)

func main() {
	// 创建 OKReply 实例
	okReply := resp.MakeOKReply()
	// 将响应转换为字节切片
	bytes := okReply.ToBytes()
	// 打印字节切片转换后的字符串
	fmt.Println(string(bytes))
}
