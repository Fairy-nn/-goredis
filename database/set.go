package database

import (
	"goredis/datastruct/set"
	"goredis/interface/database"
	"goredis/interface/resp"
	"goredis/lib/utils"
	"goredis/resp/reply"
	"strconv"
)

// SADD 命令用于将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略。
// SADD key member1 [member2 ...]
func execSADD(db *DB, args [][]byte) resp.Reply {
	// 获取集合的key
	key := string(args[0])
	// 获取集合的值
	members := args[1:]
	// 获取或初始化集合
	// 根据给定的键key从数据库db中获取集合对象
	// 若集合不存在，就创建一个新的集合对象
	setObj, isNew, errReply := getOrInitSet(db, key)
	if errReply != nil {
		return errReply
	}
	// 向集合中添加成员
	count := 0
	for _, member := range members {
		count += setObj.Add(string(member))
	}
	// 如果是新创建的集合，更新数据库中的集合对象
	if isNew || count > 0 {
		// 将集合对象存储到数据库中
		db.PutEntity(key, &database.DataEntity{
			Data: setObj,
		})
	}
	db.addAof(utils.ToCmdLineWithName("sadd", args...))
	return reply.MakeIntegerReply(int64(count))
}

// SCARD 命令用于返回集合中元素的数量
// SCARD key
func execSCARD(db *DB, args [][]byte) resp.Reply {
	// 获取集合的key
	key := string(args[0])
	// 从数据库中获取集合对象
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil {
		return reply.MakeIntegerReply(0) // 如果集合不存在，返回0
	}
	// 返回集合的元素数量
	return reply.MakeIntegerReply(int64(setObj.Len()))
}

// SISMEMBER 命令用于判断某个成员是否是集合的成员
// SISMEMBER key member
func execSISMEMBER(db *DB, args [][]byte) resp.Reply {
	// 获取集合的key
	key := string(args[0])
	// 获取成员
	member := string(args[1])

	// 从数据库中获取集合对象
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil {
		return reply.MakeIntegerReply(0) // 如果集合不存在，返回0
	}

	// 判断成员是否在集合中
	if setObj.Contains(member) {
		return reply.MakeIntegerReply(1) // 成员在集合中，返回1
	}
	return reply.MakeIntegerReply(0) // 成员不在集合中，返回0
}

// SMEMBERS 命令用于返回集合中的所有成员
func execSMEMBERS(db *DB, args [][]byte) resp.Reply {
	// 获取集合的key
	key := string(args[0])
	// 从数据库中获取集合对象
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil {
		return reply.MakeMultiBulkReply(nil) // 如果集合不存在，返回空集合
	}

	// 遍历集合中的每个成员，并将其转换为字节切片
	members := setObj.Members()
	result := make([][]byte, len(members))
	for i, member := range members {
		result[i] = []byte(member)
	}
	return reply.MakeMultiBulkReply(result) // 返回集合中的所有成员
}

// SREM 命令用于移除集合中的一个或多个成员
// SREM key member1 [member2 ...]
func execSREM(db *DB, args [][]byte) resp.Reply {
	// 获取集合的key
	key := string(args[0])
	// 获取要移除的成员
	members := args[1:]
	// 从数据库中获取集合对象
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil {
		return reply.MakeIntegerReply(0) // 如果集合不存在，返回0
	}

	// 移除成员并统计成功移除的数量
	count := 0
	for _, member := range members {
		count += setObj.Remove(string(member))
	}
	if count > 0 {
		// 检查集合是否为空，如果为空则删除key
		if setObj.Len() == 0 {
			db.Remove(key) // 从数据库中删除集合
		} else {
			// 更新数据库中的集合对象
			db.PutEntity(key, &database.DataEntity{
				Data: setObj,
			})
		}
		db.addAof(utils.ToCmdLineWithName("SERM", args...))
	}
	// 返回成功移除的成员数量
	return reply.MakeIntegerReply(int64(count))
}

// SPOP 命令用于随机移除并返回集合中的一个成员
// SPOP key
func execSPOP(db *DB, args [][]byte) resp.Reply {
	// 获取集合的key
	key := string(args[0])

	// 默认要移除的成员数量为 1
	count := 1
	if len(args) >= 2 {
		var err error
		count, err = strToInt(string(args[1]))
		if err != nil || count < 0 {
			return reply.MakeStandardErrorReply("转换失败或者数量为负数")
		}
	}

	// 从数据库中获取集合对象
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil || setObj.Len() == 0 {
		return reply.MakeEmptyBulkReply()
	}

	// 如果count为0，返回空数组
	if count == 0 {
		return reply.MakeMultiBulkReply([][]byte{})
	}

	// 限制count的最大值
	// 如果count大于集合的长度，则将count设置为集合的长度
	if count > setObj.Len() {
		count = setObj.Len()
	}

	// 随机获取count个成员
	members := setObj.RandomDistinctMembers(count)
	// 从集合中移除这些成员
	for _, member := range members {
		setObj.Remove(member)
	}

	// 如果集合为空，则从数据库中删除该key
	// 否则更新数据库中的集合对象
	if setObj.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, &database.DataEntity{
			Data: setObj,
		})
	}

	// 将SPOP命令添加到AOF日志中
	cmdArgs := make([][]byte, 2)
	cmdArgs[0] = []byte(key)
	cmdArgs[1] = []byte(intToStr(count))
	db.addAof(utils.ToCmdLineWithName("SPOP", cmdArgs...))

	// 如果count为1，则返回单个成员的回复
	if count == 1 {
		return reply.MakeBulkReply([]byte(members[0]))
	}

	// 否则返回多个成员的回复
	result := make([][]byte, len(members))
	for i, member := range members {
		result[i] = []byte(member)
	}
	return reply.MakeMultiBulkReply(result)

}

// 将字符串转换为整数
func strToInt(str string) (int, error) {
	value, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// 将整数转换为字符串
func intToStr(n int) string {
	return strconv.Itoa(n)
}

// SRANDMEMBER 命令用于随机返回集合中的一个或多个成员
func execSRANDMEMBER(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// 从数据库中获取集合对象
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil || setObj.Len() == 0 {
		return reply.MakeEmptyBulkReply()
	}

	// 默认返回一个成员
	count := 1
	withReplacement := false
	if len(args) >= 2 {
		var err error
		count, err = strToInt(string(args[1]))
		if err != nil {
			return reply.MakeStandardErrorReply("ERR value is not an integer")
		}
		// 如果count为负数，则表示要返回的成员可以重复
		if count < 0 {
			withReplacement = true
			count = -count
		}
	}

	// 如果count为0，返回空数组
	var members []string
	if withReplacement {
		members = setObj.RandomMembers(count)
	} else {
		members = setObj.RandomDistinctMembers(count)
	}

	// 如果count为1，则返回单个成员的回复
	if len(args) == 1 || (count == 1 && len(members) > 0) {
		return reply.MakeBulkReply([]byte(members[0]))
	}

	// 否则返回多个成员的回复
	result := make([][]byte, len(members))
	for i, member := range members {
		result[i] = []byte(member)
	}
	return reply.MakeMultiBulkReply(result)
}

func init() {
	RegisterCommand("SADD", execSADD, -3)
	RegisterCommand("SCARD", execSCARD, 2)
	RegisterCommand("SISMEMBER", execSISMEMBER, 3)
	RegisterCommand("SMEMBERS", execSMEMBERS, 2)
	RegisterCommand("SREM", execSREM, -3)
	RegisterCommand("SPOP", execSPOP, -2)
	RegisterCommand("SRANDMEMBER", execSRANDMEMBER, -2)

	RegisterCommand("SUNION", execSUnion, -2)
	RegisterCommand("SUNIONSTORE", execSUnionStore, -3)
	RegisterCommand("SINTER", execSInter, -2)
	RegisterCommand("SINTERSTORE", execSInterStore, -3)
	RegisterCommand("SDIFF", execSDiff, -2)
	RegisterCommand("SDIFFSTORE", execSDiffStore, -3)
}

func execSUnion(db *DB, args [][]byte) resp.Reply {
	// Create empty result set
	result := set.NewHashSet()

	// Process each set
	for _, arg := range args {
		key := string(arg)
		setObj, errReply := getAsSet(db, key)
		if errReply != nil {
			return errReply
		}
		if setObj == nil {
			continue
		}

		// Add all members to result
		setObj.ForEach(func(member string) bool {
			result.Add(member)
			return true
		})
	}

	// Convert set to reply
	members := result.Members()
	resultBytes := make([][]byte, len(members))
	for i, member := range members {
		resultBytes[i] = []byte(member)
	}

	return reply.MakeMultiBulkReply(resultBytes)
}

// execSUnionStore implements SUNIONSTORE destination key [key...]
// Store the union of multiple sets in a new set
func execSUnionStore(db *DB, args [][]byte) resp.Reply {
	destKey := string(args[0])
	keys := args[1:]

	// Execute union
	unionReply := execSUnion(db, keys)
	if _, ok := unionReply.(reply.ErrorReply); ok {
		return unionReply
	}

	// Create new set with union result
	unionResult := unionReply.(*reply.MultiBulkReply)
	newSet := set.NewHashSet()
	for _, member := range unionResult.Args {
		newSet.Add(string(member))
	}

	// Store set in database
	db.PutEntity(destKey, &database.DataEntity{
		Data: newSet,
	})

	// Add to AOF
	db.addAof(utils.ToCmdLineWithName("SUNIONSTORE", args...))

	return reply.MakeIntegerReply(int64(newSet.Len()))
}

// execSInter implements SINTER key [key...]
// Return the intersection of multiple sets
func execSInter(db *DB, args [][]byte) resp.Reply {
	if len(args) == 0 {
		return reply.MakeEmptyMultiBulkReply()
	}

	// Get first set as base
	key := string(args[0])
	firstSet, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if firstSet == nil {
		return reply.MakeEmptyMultiBulkReply()
	}

	// Create result set with members of first set
	result := set.NewHashSet()
	firstSet.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})

	// Intersect with each other set
	for i := 1; i < len(args); i++ {
		key := string(args[i])
		currentSet, errReply := getAsSet(db, key)
		if errReply != nil {
			return errReply
		}

		// Empty set or key doesn't exist means empty intersection
		if currentSet == nil {
			return reply.MakeEmptyMultiBulkReply()
		}

		// Keep only members that exist in current set
		toRemove := make([]string, 0)
		result.ForEach(func(member string) bool {
			if !currentSet.Contains(member) {
				toRemove = append(toRemove, member)
			}
			return true
		})

		// Remove non-intersecting members
		for _, member := range toRemove {
			result.Remove(member)
		}

		// Early termination if result is already empty
		if result.Len() == 0 {
			return reply.MakeEmptyMultiBulkReply()
		}
	}

	// Convert result to reply
	members := result.Members()
	resultBytes := make([][]byte, len(members))
	for i, member := range members {
		resultBytes[i] = []byte(member)
	}

	return reply.MakeMultiBulkReply(resultBytes)
}

// execSInterStore implements SINTERSTORE destination key [key...]
// Store the intersection of multiple sets in a new set
func execSInterStore(db *DB, args [][]byte) resp.Reply {
	destKey := string(args[0])
	keys := args[1:]

	// Execute intersection
	interReply := execSInter(db, keys)
	if _, ok := interReply.(reply.ErrorReply); ok {
		return interReply
	}

	// Create new set with intersection result
	interResult, ok := interReply.(*reply.MultiBulkReply)
	if !ok {
		return reply.MakeEmptyMultiBulkReply()
	}

	newSet := set.NewHashSet()
	for _, member := range interResult.Args {
		newSet.Add(string(member))
	}

	// Store set in database
	db.PutEntity(destKey, &database.DataEntity{
		Data: newSet,
	})

	// Add to AOF
	db.addAof(utils.ToCmdLineWithName("SINTERSTORE", args...))

	return reply.MakeIntegerReply(int64(newSet.Len()))
}

// execSDiff implements SDIFF key [key...]
// Return the difference between sets
func execSDiff(db *DB, args [][]byte) resp.Reply {
	// Get first set as base
	key := string(args[0])
	firstSet, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if firstSet == nil {
		return reply.MakeEmptyMultiBulkReply()
	}

	// Create result set with members of first set
	result := set.NewHashSet()
	firstSet.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})

	// Remove members that appear in subsequent sets
	for i := 1; i < len(args); i++ {
		key := string(args[i])
		currentSet, errReply := getAsSet(db, key)
		if errReply != nil {
			return errReply
		}
		if currentSet == nil {
			continue
		}

		// Remove members that exist in current set
		currentSet.ForEach(func(member string) bool {
			result.Remove(member)
			return true
		})

		// Early termination if result is already empty
		if result.Len() == 0 {
			return reply.MakeEmptyMultiBulkReply()
		}
	}

	// Convert result to reply
	members := result.Members()
	resultBytes := make([][]byte, len(members))
	for i, member := range members {
		resultBytes[i] = []byte(member)
	}

	return reply.MakeMultiBulkReply(resultBytes)
}

// execSDiffStore implements SDIFFSTORE destination key [key...]
// Store the difference between sets in a new set
func execSDiffStore(db *DB, args [][]byte) resp.Reply {
	destKey := string(args[0])
	keys := args[1:]

	// Execute difference
	diffReply := execSDiff(db, keys)
	if _, ok := diffReply.(reply.ErrorReply); ok {
		return diffReply
	}

	// Create new set with difference result
	diffResult, ok := diffReply.(*reply.MultiBulkReply)
	if !ok {
		return reply.MakeIntegerReply(0)
	}

	newSet := set.NewHashSet()
	for _, member := range diffResult.Args {
		newSet.Add(string(member))
	}

	// Store set in database
	db.PutEntity(destKey, &database.DataEntity{
		Data: newSet,
	})

	// Add to AOF
	db.addAof(utils.ToCmdLineWithName("SDIFFSTORE", args...))

	return reply.MakeIntegerReply(int64(newSet.Len()))
}

