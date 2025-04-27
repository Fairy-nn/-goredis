package config

import (
	"bufio"
	"fmt"
	"goredis/lib/logger"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// 实现一个配置文件解析器，用于读取redis服务器的配置文件，并将其映射到ServerProperties结构体中。
// 支持：默认配置初始化功能、配置文件解析功能、配置文件加载入口。
// config file
type ServerProperties struct {
	Bind           string   `cfg:"bind"`            // bind address
	Port           int      `cfg:"port"`            // port number
	AppendOnly     bool     `cfg:"append_only"`     // whether to enable AOF persistence
	AppendFilename string   `cfg:"append_filename"` // AOF file name
	MaxClients     int      `cfg:"max_clients"`     // maximum number of clients
	Databases      int      `cfg:"databases"`       // database number
	Requirepass    string   `cfg:"requirepass"`     // password
	Peers          []string `cfg:"peers"`           // cluster nodes
	Self           string   `cfg:"self"`            // self node
}

var Properties *ServerProperties

// init: initialize configuration file
func init() {
	Properties = &ServerProperties{
		Bind:       "127.0.0.1",
		Port:       6379,
		AppendOnly: true,
	}
}

func parse(src io.Reader) *ServerProperties {
	config := &ServerProperties{}

	// read config file
	rawMap := make(map[string]string)
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		pivot := strings.IndexAny(line, " ")
		if pivot > 0 && pivot < len(line)-1 { // separator found
			key := line[0:pivot]
			value := strings.Trim(line[pivot+1:], " ")
			rawMap[strings.ToLower(key)] = value
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}

	// parse format
	t := reflect.TypeOf(config)
	v := reflect.ValueOf(config)
	n := t.Elem().NumField()
	for i := 0; i < n; i++ {
		field := t.Elem().Field(i)
		fieldVal := v.Elem().Field(i)
		key, ok := field.Tag.Lookup("cfg")
		if !ok {
			key = field.Name
		}
		value, ok := rawMap[strings.ToLower(key)]
		if ok {
			// fill config
			switch field.Type.Kind() {
			case reflect.String:
				fieldVal.SetString(value)
			case reflect.Int:
				intValue, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					fieldVal.SetInt(intValue)
				}
			case reflect.Bool:
				boolValue := "yes" == value
				fieldVal.SetBool(boolValue)
			case reflect.Slice:
				if field.Type.Elem().Kind() == reflect.String {
					slice := strings.Split(value, ",")
					fieldVal.Set(reflect.ValueOf(slice))
				}
			}
		}
	}
	return config
}

func SetupConfig(configFilename string) {
	file, err := os.Open(configFilename) // open the configuration file
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	Properties = parse(file)               // parse the configuration file
	fmt.Println("Properties:", Properties) // print the bind address
	if Properties == nil {
		panic("parse config error")
	}
}
