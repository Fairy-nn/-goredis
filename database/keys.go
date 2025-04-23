package database

import (
	"goredis/interface/resp"
	"goredis/resp/reply"
)

func init() {
	RegisterCommand("PING", Ping, 1)
	RegisterCommand("DEL", execDel, -2)
	RegisterCommand("EXISTS", execExists, -2)
	RegisterCommand("FLUSHDB", execFlushDB, -1)
	RegisterCommand("TYPE", execType, 2)
	RegisterCommand("RENAME", execRename, 3)
	RegisterCommand("RENAMENX", execRenameNX, 3)
}

// Register the ping command with arity 0
func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}

// Register the del command with arity -2 (variable number of arguments)
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i := 0; i < len(args); i++ {
		keys[i] = string(args[i])
	}
	deleted := db.Removes(keys...)
	return reply.MakeIntegerReply(int64(deleted))
}

// Register the exists command with arity -2 (variable number of arguments)
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		if _, ok := db.GetEntity(key); ok {
			result++
		}
	}
	return reply.MakeIntegerReply(result)
}

// Register the flushdb command with arity -1 (variable number of arguments)
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	return reply.MakeOKReply()
}

// Register the type command with arity 2 (1 argument)
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	if entity, ok := db.GetEntity(key); ok {
		switch entity.Data.(type) {
		case []byte:
			return reply.MakeBulkReply([]byte("string"))
		}
	} else {
		return reply.MakeStatusReply("none")
	}
	return reply.MakeUnknownCmdErrReply()
}

// Register the rename command with arity 3 (2 arguments)
func execRename(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dst := string(args[1])
	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeStandardErrorReply("ERR no such key")
	}
	db.PutEntity(dst, entity)
	db.Remove(src)
	return reply.MakeOKReply()
}

func execRenameNX(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dst := string(args[1])
	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeStandardErrorReply("ERR no such key")
	}
	if _, ok := db.GetEntity(dst); ok {
		return reply.MakeIntegerReply(0)
	}
	db.PutEntity(dst, entity)
	db.Remove(src)
	return reply.MakeIntegerReply(1)
}
