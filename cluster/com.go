package cluster

import (
	"context"
	"errors"
	"goredis/interface/resp"
	"goredis/lib/utils"
	"goredis/resp/client"
	"goredis/resp/reply"
	"strconv"
)

// 从连接池中获取一个连接对象
func (c *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	// 从连接池中查找指定对等节点的连接池
	pool, ok := c.peerConn[peer]
	if !ok {
		return nil, errors.New("peer not found")
	}

	// 从连接池中借用一个连接对象
	conn, err := pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}

	// 将连接对象转换为*client.Client类型
	clientConn, ok := conn.(*client.Client)
	if !ok {
		return nil, errors.New("invalid connection type")
	}

	return clientConn, nil
}

// 归还连接对象到连接池
func (c *ClusterDatabase) returnPeerClient(peer string, client *client.Client) {
	// 从连接池中查找指定对等节点的连接池
	pool, ok := c.peerConn[peer]
	if !ok {
		return
	}

	// 归还连接对象到连接池
	pool.ReturnObject(context.Background(), client)
}

// 转发请求到指定的对等节点
func (c *ClusterDatabase) relayExec(peer string, conn resp.Connection,
	args [][]byte) resp.Reply {
	// 检查指定的对等节点是否是当前节点本身
	if peer == c.self {
		// 如果是当前节点，直接在本地数据库执行命令
		return c.db.Exec(conn, args)
	}

	// 从连接池中获取指定对等节点的客户端连接
	client, err := c.getPeerClient(peer)
	if err != nil {
		return reply.MakeStandardErrorReply(err.Error())
	}
	defer func() {
		c.returnPeerClient(peer, client)
	}()

	// 设置连接的数据库索引
	client.Send(utils.ToCmdLine("SELECT", strconv.Itoa(conn.GetDBIndex())))

	// 发送实际的命令并返回执行结果
	return client.Send(args)
}

// 广播命令到所有对等节点
func (c *ClusterDatabase) broadcastExec(conn resp.Connection, args [][]byte) map[string]resp.Reply {
    results := make(map[string]resp.Reply)

    // 遍历集群中的所有节点
    for _, peer := range c.nodes {
        // 调用 relayExec 方法将命令转发到当前节点并执行，获取执行结果
        result := c.relayExec(peer, conn, args)
        results[peer] = result
    }

    return results
}
