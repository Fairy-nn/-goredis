package cluster

import (
	"goredis/interface/resp"
	"goredis/resp/reply"
)

func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	routerMap["exists"] = defaultFunc
	routerMap["type"] = defaultFunc
	routerMap["set"] = defaultFunc
	routerMap["get"] = defaultFunc
	routerMap["setnx"] = defaultFunc
	routerMap["getset"] = defaultFunc

	routerMap["ping"] = pingFunc
	routerMap["rename"] = renameFunc
	routerMap["renamex"] = renameFunc
	routerMap["flushdb"] = flushDBFunc
	routerMap["del"] = delFunc
	routerMap["select"] = selectFunc

	routerMap["lpush"] = defaultFunc
	routerMap["rpush"] = defaultFunc
	routerMap["lpop"] = defaultFunc
	routerMap["rpop"] = defaultFunc
	routerMap["lrange"] = defaultFunc
	routerMap["llen"] = defaultFunc
	routerMap["lindex"] = defaultFunc
	routerMap["lset"] = defaultFunc

	return routerMap
}

// 根据给定的键选择一个节点，并将请求转发到该节点执行
func defaultFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	key := string(args[1])
	peer := cluster.peerPicker.GetNode(key)
	return cluster.relayExec(peer, conn, args)
}

// PING 命令的处理函数
func pingFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	return cluster.db.Exec(conn, args)
}

// rename 重命名键的处理函数
func renameFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeStandardErrorReply("ERR wrong number of arguments for 'rename' command")
	}
	// 源键
	src := string(args[1])
	// 目标键
	dest := string(args[2])

	// 使用一致性哈希算法，从集群中挑选一个对等节点
	srcPeer := cluster.peerPicker.GetNode(src)
	destPeer := cluster.peerPicker.GetNode(dest)

	// 判断所在节点是否相同
	if srcPeer != destPeer {
		return reply.MakeStandardErrorReply("ERR source and destination keys are on different nodes")
	}

	return cluster.relayExec(srcPeer, conn, args)
}

// flushallFunc 清空数据库的处理函数
func flushDBFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	// 广播执行命令到所有节点
	replies := cluster.broadcastExec(conn, args)

	var errReply reply.ErrorReply
	for _, r := range replies {
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}
	}
	if errReply == nil {
		return reply.MakeOKReply()
	}
	return reply.MakeStandardErrorReply("error: " + errReply.Error())
}

// delFunc 删除键的处理函数
func delFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return reply.MakeArgNumErrReply("del")
	}

	// 处理单个键的情况
	if len(args) == 2 {
		key := string(args[1])
		peer := cluster.peerPicker.GetNode(key)
		// Note: The full command, including "DEL", needs to be passed
		fullArgs := make([][]byte, 2)
		fullArgs[0] = []byte("DEL")
		fullArgs[1] = args[1]
		return cluster.relayExec(peer, conn, fullArgs)
	}

	// 处理多个键的情况
	groupedKeys := make(map[string][][]byte)
	for i := 1; i < len(args); i++ {
		key := string(args[i])
		peer := cluster.peerPicker.GetNode(key)
		if _, ok := groupedKeys[peer]; !ok {
			groupedKeys[peer] = make([][]byte, 0)
		}
		groupedKeys[peer] = append(groupedKeys[peer], args[i])
	}

	// 执行删除操作
	var deleted int64 = 0
	var firstErrReply reply.ErrorReply
	// 遍历每个节点，执行删除操作
	for peer, keys := range groupedKeys {
		nodeArgs := make([][]byte, len(keys)+1)
		nodeArgs[0] = []byte("DEL")
		copy(nodeArgs[1:], keys)
		nodeReply := cluster.relayExec(peer, conn, nodeArgs)

		// 检查返回的回复类型
		if reply.IsErrReply(nodeReply) {
			if firstErrReply == nil {
				if errReply, ok := nodeReply.(reply.ErrorReply); ok {
					firstErrReply = errReply
				} else {
					firstErrReply = reply.MakeStandardErrorReply("unknown error from peer")
				}
			}
			continue
		}
		if intReply, ok := nodeReply.(*reply.IntegerReply); ok {
			deleted += intReply.Code
		} else {
			if firstErrReply == nil {
				firstErrReply = reply.MakeStandardErrorReply("unexpected reply type from peer")
			}
			continue
		}
	}

	// 如果有错误发生，返回第一个错误的回复
	if firstErrReply != nil {
		return reply.MakeStandardErrorReply("error occurs during multi-key delete: " + firstErrReply.Error())
	}
	return reply.MakeIntegerReply(deleted)
}

// selectFunc 选择数据库的处理函数
func selectFunc(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply {
	return cluster.db.Exec(conn, args)
}
