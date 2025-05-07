package database

// hash 数据结构实现
import (
	"goredis/datastruct/dict"
	"goredis/datastruct/hash"
	"goredis/datastruct/set"
	"goredis/datastruct/zset"
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
	return hash, true // 修复：返回hash和true，表示成功找到哈希表
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

// getAsSet 函数用于从数据库中获取指定键对应的集合数据
func getAsSet(db *DB, key string) (set.Set, reply.ErrorReply) {
	entity, ok := db.GetEntity(key)
	if !ok {
		return nil, nil
	}
	// 使用类型断言，尝试将实体对象的数据转换为 set.Set 类型
	s, ok := entity.Data.(set.Set)
	if !ok {
		return nil, reply.MakeWrongTypeErrReply()
	}
	return s, nil
}

// getOrInitSet 函数用于获取指定键对应的集合数据，如果该键不存在，则初始化一个新的集合
func getOrInitSet(db *DB, key string) (set.Set, bool, reply.ErrorReply) {
	s, err := getAsSet(db, key) // 尝试从数据库中获取指定键对应的集合对象
	if err != nil {
		return nil, false, err
	}
	// 用于标记集合是否为新创建的
	isNew := false
	if s == nil {
		s = set.NewHashSet()
		isNew = true
	}
	return s, isNew, nil
}

// 
func getAsZSet(db *DB, key string) (zset.ZSet, bool) {
	// Get entity from database
	entity, exists := db.GetEntity(key)
	if !exists {
		return zset.NewZSet(), false
	}

	// Check if entity is a ZSet
	zsetObj, ok := entity.Data.(zset.ZSet)
	if !ok {
		return nil, true // Key exists but is not a ZSet
	}

	return zsetObj, true
}
