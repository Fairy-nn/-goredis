package main

import (
	"fmt"
	"goredis/config"
	"goredis/lib/logger"
	"goredis/resp/handler"
	"goredis/tcp/tcp"
)

const defaultConfigFileName string = "redis.conf"

func main() {
	config.SetupConfig(defaultConfigFileName)
	// 启动一个TCP服务器
	err := tcp.ListenServerWithSig(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d",
				config.Properties.Bind,
				config.Properties.Port),
		},

		handler.MakeHandler()) // 处理器
	if err != nil {
		logger.Error(err)
	}
}
