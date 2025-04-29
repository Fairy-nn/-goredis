package cluster

import (
	"context"
	"errors"
	"goredis/resp/client"

	pool "github.com/jolestar/go-commons-pool/v2"
)

type connectionFactory struct {
	Peer string // 节点的地址
}

// 创建连接池对象
func (f *connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
    c, err := client.MakeClient(f.Peer)
    if err != nil {
        return nil, err
    }
	c.Start()
    return pool.NewPooledObject(c), nil
}

// 销毁一个连接池对象
func (f *connectionFactory) DestroyObject(ctx context.Context, pooledObject *pool.PooledObject) error {
	// 从连接池中获取连接对象
	c, ok := pooledObject.Object.(*client.Client)
	if !ok {
		return errors.New("invalid connection type")
	}

	// 关闭连接对象
	c.Close()
	return nil
}

// BorrowObject从连接池中借用一个连接对象
func (f *connectionFactory) ValidateObject(ctx context.Context, pooledObject *pool.PooledObject) bool {
	return true
}

// PassivateObject 从连接池中释放一个连接对象
func (f *connectionFactory) PassivateObject(ctx context.Context, pooledObject *pool.PooledObject) error {
	return nil
}

// ActivateObject 激活一个连接对象
func (f *connectionFactory) ActivateObject(ctx context.Context, pooledObject *pool.PooledObject) error {
	return nil
}
