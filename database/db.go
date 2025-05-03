package database

// hash 数据结构实现
import (
	"goredis/datastruct/dict"
	"goredis/datastruct/hash"
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

// getAsHash 函数从数据库中获取存储在指定键的哈希值，如果键不存在则返回 nil 和 false。
// 如果键存在但对应的数据不是哈希类型，则返回 nil 和 true。
func (db *DB) getAsHash(key string) (*hash.Hash, bool) {
	entity, ok := db.GetEntity(key)
	if !ok {
		return nil, false
	}
	hash, ok := entity.Data.(*hash.Hash)
	if !ok {
		return nil, true
	}
	return hash, false
}

// getOrCreateHash 函数用于获取或创建一个哈希对象。
// 首先尝试获取指定键对应的哈希对象，如果不存在则创建一个新的哈希对象并存储到数据库中。
func (db *DB) getOrCreateHash(key string) (*hash.Hash, bool) {
	// 获取指定键对应的哈希对象
	hashObj, exists := db.getAsHash(key)
	if exists {
		return hashObj, true
	}

	// 创建一个新的哈希对象
	hashObj = hash.NewHash()
	// 存储到数据库中
	db.PutEntity(key, &database.DataEntity{Data: hashObj})

	return hashObj, false
}
