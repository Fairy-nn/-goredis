package database

import (
	"goredis/interface/database"
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
