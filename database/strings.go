package database

import (
	"goredis/interface/database"
	"goredis/interface/resp"
	"goredis/lib/utils"
	"goredis/resp/reply"
)

func init() {
	RegisterCommand("GET", execGet, 2)
	RegisterCommand("SET", execSet, 3)
	RegisterCommand("SETNX", execSetNX, 3)
	RegisterCommand("GETSET", execGetSet, 3)
	RegisterCommand("STRLEN", execStrlen, 2)
}

// get:get key
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	if entity, ok := db.GetEntity(key); ok {
		return reply.MakeBulkReply(entity.Data.([]byte))
	}
	return reply.MakeNullReply()
}

// set: set key value
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	db.PutEntity(key, entity)

	// store to aof file
	db.addAof(utils.ToCmdLineWithName("SET", args...))
	return reply.MakeOKReply()
}

// setnx: set if not exists
func execSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	result := db.PutIfAbsent(key, entity)
	// write to aof file
	db.addAof(utils.ToCmdLineWithName("SETNX", args...))

	return reply.MakeIntegerReply(int64(result))
}

// getnx: get if not exists
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity, ok := db.GetEntity(key)
	db.PutEntity(key, &database.DataEntity{
		Data: value,
	})
	// write to aof file
	db.addAof(utils.ToCmdLineWithName("GETSET", args...))
	if ok {
		return reply.MakeBulkReply(entity.Data.([]byte))
	}
	return reply.MakeNullReply()
}

// strlen: get the length of the string
func execStrlen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	if entity, ok := db.GetEntity(key); ok {
		return reply.MakeIntegerReply(int64(len(entity.Data.([]byte))))
	}
	return reply.MakeNullReply()
}
