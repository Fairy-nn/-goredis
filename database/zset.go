package database

import (
	"goredis/interface/database"
	"goredis/interface/resp"
	"goredis/lib/utils"
	"goredis/resp/reply"
	"strconv"
)

func init() {
	RegisterCommand("ZADD", execZAdd, -4)     // key score member [score member ...]
	RegisterCommand("ZSCORE", execZScore, 3)  // key member
	RegisterCommand("ZCARD", execZCard, 2)    // key
	RegisterCommand("ZRANGE", execZRANGE, -4) // key start stop [WITHSCORES]
	RegisterCommand("ZREM", execZREM, -3)     // key member [member ...]
	RegisterCommand("ZCOUNT", execZCOUNT, 4)  // key min max
	RegisterCommand("ZRANK", execZRank, 3)    // key member
	RegisterCommand("ZTYPE", execZType, 2)    // key
}

// ZADD 添加元素到有序集合中
func execZAdd(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 || len(args)%2 == 0 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zadd' command")
	}

	key := string(args[0])

	// 获取或创建 ZSet
	zsetObj, exists := getAsZSet(db, key)
	if exists && zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}

	// 遍历参数并添加到 ZSet
	added := 0
	for i := 1; i < len(args); i += 2 {
		scoreStr := string(args[i])
		member := string(args[i+1])

		// 解析分数
		score, err := parseFloat(scoreStr)
		if err != nil {
			return err
		}

		// 添加到 ZSet
		if zsetObj.Add(member, score) {
			added++
		}
	}

	// 更新数据库中的 ZSet
	db.PutEntity(key, &database.DataEntity{Data: zsetObj})
	// 添加AOF日志
	db.addAof(utils.ToCmdLineWithName("ZADD", args...))

	return reply.MakeIntegerReply(int64(added))
}

// parseFloat 解析字符串为浮点数
func parseFloat(scoreStr string) (float64, resp.ErrorReply) {
	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return 0, reply.MakeStandardErrorReply("ERR value is not a valid float")
	}
	return score, nil
}

// ZSCORE 获取有序集合中成员的分数
// ZSCORE key member
func execZScore(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zscore' command")
	}
	key := string(args[0])
	member := string(args[1])

	// 获取 ZSet
	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeEmptyBulkReply()
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}

	// 获取成员的分数
	// Get score
	score, exists := zsetObj.Score(member)
	if !exists {
		return reply.MakeEmptyBulkReply()
	}
	// 返回分数
	return reply.MakeBulkReply([]byte(strconv.FormatFloat(score, 'f', -1, 64)))
}

// ZCARD 用于获取有序集合的成员数量
// ZCARD key
func execZCard(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zcard' command")
	}
	key := string(args[0])

	// 获取 ZSet
	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeIntegerReply(0)
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}
	// 返回成员数量
	return reply.MakeIntegerReply(int64(zsetObj.Len()))
}

// ZRANK 用于获取有序集合中指定成员的排名
// ZRANK key member
func execZRank(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zrank' command")
	}

	key := string(args[0])
	member := string(args[1])

	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeEmptyBulkReply()
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}

	score, exists := zsetObj.Score(member)
	if !exists {
		return reply.MakeEmptyBulkReply()
	}

	rank := -1
	if zsetObj.Encoding() == 1 {
		skiplist := zsetObj.GetSkiplist()
		rank = skiplist.GetRank(member, score)
	} else {
		// For listpack encoding, we need to compute rank by sorting
		members := zsetObj.RangeByRank(0, -1)
		for i, m := range members {
			if m == member {
				rank = i
				break
			}
		}
	}

	if rank == -1 {
		return reply.MakeEmptyBulkReply()
	}

	return reply.MakeIntegerReply(int64(rank))

}

func execZType(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'ztype' command")
	}

	key := string(args[0])

	// Get ZSet
	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeEmptyBulkReply()
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}

	return reply.MakeIntegerReply(int64(zsetObj.Encoding()))
}

// ZRANGE 用于获取有序集合中指定范围内的成员
// ZRANGE key start stop [WITHSCORES]
func execZRANGE(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zrange' command")
	}

	withScores := false
	if len(args) > 3 && string(args[3]) == "WITHSCORES" {
		withScores = true
	}

	key := string(args[0])
	// 获取范围的起始和结束索引
	start, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR value is not an integer")
	}
	stop, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR value is not an integer")
	}

	// 获取有序集合对象
	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeEmptyMultiBulkReply()
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}

	// 获取成员
	members := zsetObj.RangeByRank(start, stop)

	if !withScores {
		// 如果不需要分数，直接返回成员
		result := make([][]byte, len(members))
		for i, member := range members {
			result[i] = []byte(member)
		}
		return reply.MakeMultiBulkReply(result)
	} else {
		// 如果需要分数，返回成员和分数
		result := make([][]byte, len(members)*2)
		for i, member := range members {
			result[i*2] = []byte(member)
			score, _ := zsetObj.Score(member)
			result[i*2+1] = []byte(strconv.FormatFloat(score, 'f', -1, 64))
		}
		return reply.MakeMultiBulkReply(result)
	}
}

// ZREM 用于删除有序集合中的成员
// ZREM key member1 [member2 ...]
func execZREM(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zrem' command")
	}

	key := string(args[0])

	// 获取有序集合对象
	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeIntegerReply(0)
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}

	// 删除成员并统计成功删除的数量
	removed := 0
	for i := 1; i < len(args); i++ {
		member := string(args[i])
		if zsetObj.Remove(member) {
			removed++
		}
	}

	if removed > 0 {
		if zsetObj.Len() == 0 {
			db.Remove(key) // 从数据库中删除有序集合
		} else {
			db.PutEntity(key, &database.DataEntity{
				Data: zsetObj,
			})
		}
		db.addAof(utils.ToCmdLineWithName("ZREM", args...))
	}
	return reply.MakeIntegerReply(int64(removed))
}

// ZCOUNT 用于获取有序集合中指定分数范围内的成员数量
// ZCOUNT key min max
func execZCOUNT(db *DB, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zcount' command")
	}

	key := string(args[0])
	min, err := parseFloat(string(args[1]))
	if err != nil {
		return err
	}

	max, err := parseFloat(string(args[2]))
	if err != nil {
		return err
	}

	// 获取有序集合对象
	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeIntegerReply(0) // 如果有序集合不存在，返回0
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply() // 如果有序集合不存在，返回0
	}

	count := zsetObj.Count(min, max)
	return reply.MakeIntegerReply(int64(count))
}
