package cluster

import (
	"context"
	"goredis/config"
	databaseinstance "goredis/database"
	"goredis/interface/database"
	"goredis/interface/resp"
	consistenthash "goredis/lib/consistent_hash"
	"goredis/lib/logger"
	"goredis/resp/reply"
	"strings"

	pool "github.com/jolestar/go-commons-pool/v2"
)

var routerMap = makeRouter()

// 命令处理函数类型
type CmdFunc func(cluster *ClusterDatabase, conn resp.Connection, args [][]byte) resp.Reply

type ClusterDatabase struct {
	self       string                      // 节点自己的地址
	nodes      []string                    // 集群中所有节点的地址
	peerPicker *consistenthash.NodeMap     // 一致性哈希算法的节点映射
	peerConn   map[string]*pool.ObjectPool // 节点连接池
	db         database.Database           // 当前节点的数据库实例
}

func NewClusterDatabase() *ClusterDatabase {
	return &ClusterDatabase{}
}

func (c *ClusterDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("ClusterDatabase Exec panic:" + err.(error).Error())
			result = reply.MakeUnknownCmdErrReply()
		}
	}()

	cmdName := strings.ToLower(string(args[0]))

	if cmdFunc, ok := routerMap[cmdName]; ok {
		return cmdFunc(c, client, args)
	} else {
		result = reply.MakeStandardErrorReply("ERR unknown command '" + cmdName + "'")
	}

	return
}

func (c *ClusterDatabase) AfterClientClose(client resp.Connection) {
	c.db.AfterClientClose(client)
}

func (c *ClusterDatabase) Close() {
	c.db.Close()
}

// 创建一个新的ClusterDatabase实例，并初始化节点连接池和一致性哈希算法的节点映射。
func MakeClusterDatabase() *ClusterDatabase {
	// 创建一个新的ClusterDatabase实例
	cluster := &ClusterDatabase{
		self:       config.Properties.Self,
		db:         databaseinstance.NewStandaloneDatabase(),
		peerPicker: consistenthash.NewNodeMap(nil),
		peerConn:   make(map[string]*pool.ObjectPool),
	}

	// 初始化节点连接池
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	nodes = append(nodes, config.Properties.Peers...)
	nodes = append(nodes, config.Properties.Self)

	// 将所有节点添加到一致性哈希算法的节点映射中
	cluster.peerPicker.AddNode(nodes...)

	// 创建连接池对象
	ctx := context.Background()
	for _, peer := range config.Properties.Peers {
		cluster.peerConn[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}

	// 保存完整的节点列表
	cluster.nodes = nodes
	return cluster
}
