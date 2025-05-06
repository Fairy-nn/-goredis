package database

import (
	"container/list" // 使用go标准库中的list包来实现双向链表
	"fmt"
	"goredis/interface/database"
	"goredis/interface/resp"
	"goredis/lib/utils"
	"goredis/resp/reply"
	"strconv"
)

// getAsList 函数用于获取指定键对应的列表，如果键不存在则创建一个新列表。
// 返回列表指针和一个布尔值，布尔值表示键是否原本就存在。
func getList(db *DB, key string) (*list.List, bool) {
	// 从数据库中获取指定键的实体
	entity, ok := db.GetEntity(key)
	if !ok {
		return list.New(), false
	}

	// 判断实体的类型是否为list
	lst, ok := entity.Data.(*list.List)
	if !ok {
		return nil, true
	}
	// 返回实体的值
	return lst, true
}

// execLPush 函数实现了LPUSH命令，用于将一个或多个值插入到列表的头部。
// 命令格式：LPUSH key value [value ...]
func execLPush(db *DB, args [][]byte) resp.Reply {
	// 获取key和value
	key := string(args[0])
	value := args[1:]
	// 列表存在但不是列表类型，返回错误
	lst, ok := getList(db, key)
	if lst == nil && ok {
	
		return reply.MakeWrongTypeErrReply()
	}

	// 将value中的每个值插入到列表的头部
	for _, v := range value {
		lst.PushFront(v)
	}

	// 将列表存入数据库
	db.PutEntity(key, &database.DataEntity{
		Data: lst,
	})
	// 判断是否存到数据库
	db.GetEntity(key)
	// 将该命令添加到AOF日志中
	db.addAof(utils.ToCmdLineWithName("LPUSH", args...))
	//返回新的列表长度
	return reply.MakeIntegerReply(int64(lst.Len()))
}

// execRPush 函数实现了RPUSH命令，用于将一个或多个值插入到列表的尾部。
// 命令格式：RPUSH key value [value ...]
func execRPush(db *DB, args [][]byte) resp.Reply {
	// 获取key和value
	key := string(args[0])
	value := args[1:]

	// 列表存在但不是列表类型，返回错误
	lst, ok := getList(db, key)
	if lst == nil && ok {
		return reply.MakeWrongTypeErrReply()
	}

	// 将value中的每个值插入到列表的头部
	for _, v := range value {
		lst.PushBack(v)
	}

	// 将列表存入数据库
	db.PutEntity(key, &database.DataEntity{
		Data: lst,
	})
	// 将该命令添加到AOF日志中
	db.addAof(utils.ToCmdLineWithName("RPUSH", args...))

	//返回新的列表长度
	return reply.MakeIntegerReply(int64(lst.Len()))
}

// execLPop 函数实现了LPOP命令，用于移除并返回列表的第一个元素。
// 命令格式：LPOP key
func execLPop(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	lst, ok := getList(db, key)
	
	if !ok {
		return reply.MakeNullReply()
	}
	if lst == nil {
		return reply.MakeWrongTypeErrReply()
	}
	if lst.Len() == 0 {
		fmt.Println("LPOP: lst is empty")
		return reply.MakeNullReply()
	}

	// 从列表中移除第一个元素
	element := lst.Front()
	lst.Remove(element)
	value := element.Value.([]byte)
	
	if lst.Len() == 0 {
		// 如果列表为空，则从数据库中删除该键
		db.Remove(key)
	} else {
		// 否则更新列表
		db.PutEntity(key, &database.DataEntity{
			Data: lst,
		})
	}

	// 将该命令添加到AOF日志中
	db.addAof(utils.ToCmdLineWithName("LPOP", args...))
	// 返回移除的元素
	return reply.MakeBulkReply(value)
}

// execRPop 函数实现了RPOP命令，用于移除并返回列表的最后一个元素。
// 命令格式：RPOP key
func execRPop(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	lst, ok := getList(db, key)
	if !ok {
		return reply.MakeNullReply()
	}
	if lst == nil {
		return reply.MakeWrongTypeErrReply()
	}
	if lst.Len() == 0 {
		return reply.MakeNullReply()
	}

	// 从列表中移除最后一个元素
	element := lst.Back()
	lst.Remove(element)
	value := element.Value.([]byte)
	if lst.Len() == 0 {
		// 如果列表为空，则从数据库中删除该键
		db.Remove(key)
	} else {
		// 否则更新列表
		db.PutEntity(key, &database.DataEntity{
			Data: lst,
		})
	}

	// 将该命令添加到AOF日志中
	db.addAof(utils.ToCmdLineWithName("RPOP", args...))
	// 返回移除的元素
	return reply.MakeBulkReply(value)
}

// execLRange 函数实现了LRANGE命令，用于返回列表中指定范围的元素。
// 命令格式：LRANGE key start stop
// start和stop可以是负数，表示从列表的尾部开始计数。
func execLRange(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	start, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid start index")
	}

	stop, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid stop index")
	}

	lst, ok := getList(db, key)
	if !ok {
		return reply.MakeNullReply()
	}
	if lst == nil {
		return reply.MakeWrongTypeErrReply()
	}

	size := int64(lst.Len())
	if start < 0 {
		start = size + start
	}
	if stop < 0 {
		stop = size + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= size {
		stop = size - 1
	}
	if start > stop {
		return reply.MakeEmptyMultiBulkReply()
	}

	// 创建一个新的列表来存储结果
	elements := make([][]byte, 0, stop-start+1)
	index := int64(0)
	for e := lst.Front(); e != nil; e = e.Next() {
		if index >= start && index <= stop {
			// 将元素添加到结果列表中
			elements = append(elements, e.Value.([]byte))
		} else if index > stop {
			break
		}
		index++
	}
	return reply.MakeMultiBulkReply(elements)
}

//	execLLen 函数实现了LLEN命令，用于返回列表的长度。
//
// 命令格式：LLEN key
func execLLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	lst, ok := getList(db, key)
	if lst == nil {
		return reply.MakeWrongTypeErrReply()
	}
	if !ok {
		return reply.MakeNullReply()
	}
	size := int64(lst.Len())
	return reply.MakeIntegerReply(size)
}

// execLIndex 函数实现了LINDEX命令，用于返回列表中指定索引位置的元素。
// 命令格式：LINDEX key index
func execLIndex(db *DB, args [][]byte) resp.Reply {
	// 获取键名
	key := string(args[0])
	index, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		// 若索引参数不是有效的整数，返回错误回复
		return reply.MakeStandardErrorReply("value is not an integer or out of range")
	}

	// 获取列表
	lst, ok := getList(db, key)
	if !ok {
		return reply.MakeStandardErrorReply("ERR no such key")
	}
	if lst == nil {
		return reply.MakeWrongTypeErrReply()
	}

	size := int64(lst.Len())
	if index < 0 {
		index = size + index
	}
	if index < 0 || index >= size {
		// 若索引越界，返回空回复
		return reply.MakeNullReply()
	}

	// 查找指定索引位置的元素
	var element *list.Element
	if index < size/2 {
		// 若索引在前半部分，从列表头部开始遍历
		element = lst.Front()
		for i := int64(0); i < index; i++ {
			element = element.Next()
		}
	} else {
		// 若索引在后半部分，从列表尾部开始遍历
		element = lst.Back()
		for i := size - 1; i > index; i-- {
			element = element.Prev()
		}
	}

	return reply.MakeBulkReply(element.Value.([]byte))
}

// execLSet 函数实现了LSET命令，用于将列表中指定索引位置的元素设置为指定值。
// 命令格式：LSET key index value
func execLSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	index, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid index")
	}
	value := args[2]

	lst, ok := getList(db, key)
	if !ok {
		return reply.MakeStandardErrorReply("ERR no such key")
	}
	if lst == nil {
		return reply.MakeWrongTypeErrReply()
	}

	size := int64(lst.Len())
	if index < 0 {
		index = size + index
	}
	if index < 0 || index >= size {
		return reply.MakeStandardErrorReply("ERR index out of range")
	}

	// 获取指定索引位置的元素
	var e *list.Element
	if index < size/2 {
		e = lst.Front()
		for i := int64(0); i < index; i++ {
			e = e.Next()
		}
	} else {
		e = lst.Back()
		for i := size - 1; i > index; i-- {
			e = e.Prev()
		}
	}
	// 设置元素的值
	e.Value = value
	// 将列表存入数据库
	db.PutEntity(key, &database.DataEntity{
		Data: lst,
	})
	// 将该命令添加到AOF日志中
	db.addAof(utils.ToCmdLineWithName("LSET", args...))
	return reply.MakeStatusReply("OK")
}

func init() {
	// 注册命令
	RegisterCommand("LPUSH", execLPush, -3)  // 命令格式：key value [value ...]，至少3个参数
	RegisterCommand("RPUSH", execRPush, -3)  // 命令格式：key value [value ...]，至少3个参数
	RegisterCommand("LPOP", execLPop, 2)     // 命令格式：key，2个参数
	RegisterCommand("RPOP", execRPop, 2)     // 命令格式：key，2个参数
	RegisterCommand("LRANGE", execLRange, 4) // 命令格式：key start stop，4个参数
	RegisterCommand("LLEN", execLLen, 2)     // 命令格式：LLEN key，2个参数
	RegisterCommand("LINDEX", execLIndex, 3) // 命令格式：LINDEX key index，3个参数
	RegisterCommand("LSET", execLSet, 4)     // 命令格式：LSET key index value，4个参数
}
