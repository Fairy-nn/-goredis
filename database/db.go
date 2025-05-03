package database

import (
	"goredis/datastruct/dict"
	"goredis/interface/database"
	"goredis/interface/resp"
	"goredis/resp/reply"
	"strings"
)

type DB struct {
	index  int
	data   dict.Dict
	addAof func(line CmdLine) // addAof is a function to add commands to AOF.
}

func MakeDB() *DB {
	return &DB{
		index: 0,
		data:  dict.MakeSyncDict(),
		addAof: func(line CmdLine) {
			// do nothing
		},
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

// getenity returrns dataentity by key
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	enity, _ := raw.(*database.DataEntity)
	return enity, true
}

// put entity by key
func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

func (db *DB) Remove(key string) int {
	return db.data.Remove(key)
}

func (db *DB) Removes(keys ...string) int {
	deleted := 0
	for _, key := range keys {
		_, ok := db.data.Get(key)
		if ok {
			db.data.Remove(key)
			deleted++
		}
	}
	return deleted
}

func (dr *DB) Flush() {
	dr.data.Clear()
}
