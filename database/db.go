package database

import (
	"goredis/datastruct/dict"
	"goredis/interface/resp"
	"goredis/resp/reply"
	"strings"
)

type DB struct {
	index int
	data  dict.Dict
}

func MakeDB() *DB {
	return &DB{
		index: 0,
		data:  dict.MakeSyncDict(),
	}
}

// all redis like ping,set,commands are implemented in the form of a function
type ExecFunc func(db *DB, args [][]byte) resp.Reply
type CmdLine = [][]byte

// parse and execute the command
func (db *DB) Exec(c resp.Connection, cmdLine CmdLine) resp.Reply {
	
	cmdName := strings.ToLower(string(cmdLine[0]))
	// find the command in the command table
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeStandardErrorReply("ERR unknown command '" + cmdName + "'")
	}
	// check the arity of the command
	if !ValidateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	// check the db index
	return cmd.exec(db, cmdLine[1:])
}

func ValidateArity(arity int, args [][]byte) bool {
	if arity >= 0 {
		return len(args) == arity
	} else {
		return len(args) >= -arity
	}
}
