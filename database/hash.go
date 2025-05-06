package database

import (
	"fmt"
	"goredis/interface/resp"
	"goredis/lib/utils"
	"goredis/resp/reply"
)

// HSet 函数实现了 Redis 中 HSET 命令的功能，
// 将存储在键 key 的哈希表中的字段 field 设置为值 value
func execHSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	field := string(args[1])
	value := string(args[2])

	// 获取哈希表
	hash, _ := db.getOrCreateHash(key)
	result := hash.Set(field, value)

	db.addAof(utils.ToCmdLineWithName("HSET", args...))
	return reply.MakeIntegerReply(int64(result))
}

// HGet 函数实现了 Redis 中 HGET 命令的功能，
// 获取存储在键 key 的哈希表中字段 field 的值
func execHGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	field := string(args[1])

	// 获取哈希表
	hash, _ := db.getOrCreateHash(key)
	value, ok := hash.Get(field)

	if !ok {
		return reply.MakeNullReply()
	}
	return reply.MakeBulkReply([]byte(value))
}

// HExists 函数实现了 Redis 中 HEXISTS 命令的功能，
// 检查存储在键 key 的哈希表中字段 field 是否存在
func execHExists(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	field := string(args[1])

	// 获取哈希表
	hash, ok := db.getOrCreateHash(key)
	if !ok {
		return reply.MakeNullReply()
	}

	// 检查字段是否存在
	exists := hash.Exists(field)
	if !exists {
		return reply.MakeIntegerReply(0)
	}
	return reply.MakeIntegerReply(1)
}

// HDel 函数实现了 Redis 中 HDEL 命令的功能，
// 从存储在键 key 的哈希表中删除指定的字段
func execHDel(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// 获取哈希表
	hash, ok := db.getAsHash(key)
	if !ok {
		return reply.MakeIntegerReply(0)
	}

	// 删除字段
	deleted := 0
	for _, field := range args[1:] {
		deleted += hash.Delete(string(field))
	}
	// 如果删除字段后哈希表为空，从数据库中删除该键
	if hash.Len() == 0 {
		db.Remove(key)
	}

	// 如果有字段被删除，记录到 AOF 文件中
	if deleted > 0 {
		db.addAof(utils.ToCmdLineWithName("HDEL", args...))
	}

	return reply.MakeIntegerReply(int64(deleted))
}

// HLen 函数实现了 Redis 中 HLEN 命令的功能，
// 返回存储在键 key 的哈希表中的字段数量
func execHLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// 获取哈希表
	hash, ok := db.getAsHash(key)
	if !ok {
		return reply.MakeIntegerReply(0)
	}

	// 返回哈希表的字段数量
	return reply.MakeIntegerReply(int64(hash.Len()))
}

// HGetAll 函数实现了 Redis 中 HGETALL 命令的功能，
// 返回存储在键 key 的哈希表中的所有字段和值
func execHGetAll(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// 获取哈希表
	hash, ok := db.getAsHash(key)
	if !ok {
		return reply.MakeNullReply()
	}

	allMap := hash.GetAll()
	result := make([][]byte, 0, len(allMap)*2)
	for k, v := range allMap {
		result = append(result, []byte(k), []byte(v))
	}

	return reply.MakeMultiBulkReply(result)
}

// HKeys 函数实现了 Redis 中 HKEYS 命令的功能，
// 返回存储在键 key 的哈希表中的所有字段名
func execHKeys(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// 获取哈希表
	hash, ok := db.getAsHash(key)
	if !ok {
		return reply.MakeNullReply()
	}

	// 返回哈希表的所有字段名
	fields := hash.Fields()
	result := make([][]byte, len(fields))
	for i, field := range fields {
		result[i] = []byte(field)
	}
	return reply.MakeMultiBulkReply(result)
}

// HVals 函数实现了 Redis 中 HVALS 命令的功能，
// 返回存储在键 key 的哈希表中的所有值
func execHVals(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// 获取哈希表
	hash, ok := db.getAsHash(key)
	if !ok {
		return reply.MakeNullReply()
	}

	// 返回哈希表的所有值
	values := hash.Values()
	result := make([][]byte, 0, len(values))
	for _, value := range values {
		result = append(result, []byte(value))
	}
	return reply.MakeMultiBulkReply(result)
}

// HMGet 函数实现了 Redis 中 HMGET 命令的功能，
// 返回存储在键 key 的哈希表中多个字段的值
func execHMGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// 获取哈希表
	hash, ok := db.getAsHash(key)
	if !ok {
		results := make([][]byte, len(args)-1)
		for i := range results {
			results[i] = nil
		}
		return reply.MakeMultiBulkReply(results)
	}

	// 返回哈希表中多个字段的值
	values := make([][]byte, len(args)-1)
	for i, field := range args[1:] {
		value, exists := hash.Get(string(field))
		fmt.Printf("field: %s, value: %s\n", string(field), value)
		if exists {
			values[i] = []byte(value)
		} else {
			fmt.Println("Field not found:", string(field))
			values[i] = nil
		}
	}
	return reply.MakeMultiBulkReply(values)
}

// HMSet 函数实现了 Redis 中 HMSET 命令的功能，
// 在存储在键 key 的哈希表中设置多个字段的值。
// 命令格式：HMSET key field value [field value ...]
func execHMSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	if len(args)%2 == 0 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'hmset' command")
	}

	// 获取哈希表
	hash, _ := db.getOrCreateHash(key)

	// 设置多个字段的值
	for i := 1; i < len(args)-1; i += 2 {
		field := string(args[i])
		value := string(args[i+1])
		hash.Set(field, value)
	}

	// 将该命令添加到 AOF 文件中
	db.addAof(utils.ToCmdLineWithName("HMSET", args...))
	return reply.MakeOKReply()
}

// HEncoding 函数是一个自定义函数，用于获取存储在键 key 的哈希表的编码类型。
// 0 表示使用 listpack 编码，1 表示使用字典（哈希表）编码
func execHEncoding(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// 获取哈希表
	hash, ok := db.getAsHash(key)
	if !ok {
		return reply.MakeNullReply()
	}

	return reply.MakeIntegerReply(int64(hash.Encoding()))
}

// execHSetNX 函数实现了 Redis 中 HSETNX 命令的功能，
// 只有当存储在键 key 的哈希表中的字段 field 不存在时，才将其设置为值 value。
// 命令格式：HSETNX key field value
func execHSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	field := string(args[1])
	value := string(args[2])

	// 获取哈希表
	hash, _ := db.getOrCreateHash(key)

	_, ok := hash.Get(field)
	if ok {
		return reply.MakeIntegerReply(0)
	}

	// 设置字段的值
	result := hash.Set(field, value)

	// 将该命令添加到 AOF 文件中
	db.addAof(utils.ToCmdLineWithName("HSETNX", args...))
	return reply.MakeIntegerReply(int64(result))
}

func init() {
	RegisterCommand("HSET", execHSet, 4)
	RegisterCommand("HGET", execHGet, 3)
	RegisterCommand("HEXISTS", execHExists, 3)
	RegisterCommand("HDEL", execHDel, -3)
	RegisterCommand("HLEN", execHLen, 2)
	RegisterCommand("HGETALL", execHGetAll, 2)
	RegisterCommand("HKEYS", execHKeys, 2)
	RegisterCommand("HVALS", execHVals, 2)
	RegisterCommand("HMGET", execHMGet, -3)
	RegisterCommand("HMSET", execHMSet, -4)
	RegisterCommand("HENCODING", execHEncoding, 2)
	RegisterCommand("HSETNX", execHSetNX, 4)
}
