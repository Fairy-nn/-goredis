package main

import (
	"fmt"
	"goredis/TCP/config"
	"goredis/TCP/lib/logger"
	"goredis/TCP/tcp"

	//"goredis/TCP/lib/sync/automic"

	//"goredis/TCP/lib/logger"
	"os"
)

func main() {
	// var b automic.Boolean
	// b.Set(true)
	// fmt.Println("Expected true, got:", b.Get()) // 应输出 true

	// b.Set(false)
	// fmt.Println("Expected false, got:", b.Get()) // 应输出 false

	// // 配置日志
	// settings := &logger.Settings{
	// 	Path:       "./logs",
	// 	Name:       "app",
	// 	Ext:        "log",
	// 	TimeFormat: "2006-01-02",
	// }
	// logger.Setup(settings)

	// // 测试日志输出
	// logger.Debug("This is a debug message")
	// logger.Info("This is an info message")
	// logger.Warn("This is a warning message")
	// logger.Error("This is an error message")
	// logger.Fatal("This is a fatal message")

	// Check if the file exists
	// filename := "test.conf"
	// if fileExists(filename) {
	// 	config.SetupConfig(filename) // Setup the configuration using the file
	// } else {
	// 	fmt.Println("File does not exist:", filename)
	// 	return
	// }

	const configFile string = "test.conf"

	var defaultProperties = &config.ServerProperties{
		Bind: "0.0.0.0",
		Port: 6379,
	}
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "godis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}

	err := tcp.ListenServerWithSig(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d",
				config.Properties.Bind,
				config.Properties.Port),
		},
		tcp.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil && !os.IsNotExist(err)
}
