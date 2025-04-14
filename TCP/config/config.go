package config

import (
	"bufio"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

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
		AppendOnly: false,
	}
}

// Parse : parse configuration file
func parse(src io.Reader) (*ServerProperties) {
	config := &ServerProperties{}
	rawMap := make(map[string]string)
	// read configuration file
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue // skip empty lines and comments
		}
		// split line into key and value
		pivot := strings.Index(line, ":")
		if pivot > 0 && pivot < len(line)-1 { // check if the line contains a key-value pair
			key := strings.TrimSpace(line[:pivot]) //  trailing white spaces
			value := strings.TrimSpace(line[pivot+1:])
			// add key-value pair to the map
			rawMap[key] = value
		} else {
			return nil
		}
	}
	// parse key-value pairs into struct fields
	t := reflect.TypeOf(config)  // get the type of the struct
	v := reflect.ValueOf(config) // get the value of the struct
	for i := 0; i < t.Elem().NumField(); i++ {
		filed := t.Elem().Field(i)
		filedVal := v.Elem().Field(i)
		key, ok := filed.Tag.Lookup("cfg")
		if !ok {
			key = filed.Name
		}
		value, ok := rawMap[strings.ToLower(key)]
		if ok {
			switch filed.Type.Kind() {
			case reflect.String:
				filedVal.SetString(value)
			case reflect.Int:
				if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
					filedVal.SetInt(intVal)
				} else {
					return nil
				}
			case reflect.Bool:
				boolValue := value == "yes"
				filedVal.SetBool(boolValue)
			case reflect.Slice:
				if filed.Type.Elem().Kind() == reflect.String {
					slice := strings.Split(value, ",")
					filedVal.Set(reflect.ValueOf(slice))
				}
			}
		}
	}
	return config
}
func SetupConfig(configFilename string) {
	file, err := os.Open(configFilename)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	Properties = parse(file)
}