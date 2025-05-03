package database

import (
	"strings"
)

type command struct {
	exec  ExecFunc
	arity int
}

var cmdTable = make(map[string]*command)

// all redis like ping,set,commands are implemented in the form of a function
func RegisterCommand(name string, exec ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		exec:  exec,
		arity: arity,
	}
}
