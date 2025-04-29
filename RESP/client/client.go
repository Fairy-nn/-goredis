package client

import (
	"goredis/interface/resp"
	"goredis/lib/sync/wait"
	"goredis/resp/parser"
	"goredis/resp/reply"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"goredis/lib/logger"
)

type request struct {
	args      [][]byte
	reply     resp.Reply
	err       error
	heartbeat bool
	waiting   *wait.Wait
}
type Client struct {
	conn        net.Conn
	pendingReqs chan *request // wait to send
	waitGroup   chan *request // wait for reply
	ticker      *time.Ticker  // Heartbeat ticker
	addr        string
	working     *sync.WaitGroup
}

// creats a new client instance
func MakeClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:        conn,
		addr:        addr,
		pendingReqs: make(chan *request, 256),
		waitGroup:   make(chan *request, 256),
		working:     &sync.WaitGroup{},
	}, nil
}

func (c *Client) Start() {

	go c.handleWrite() // 1. 发送请求到服务端

	go func() {
		err := c.handleRead() // 2. 读取服务端的请求
		if err != nil {
			logger.Error(err)
		}
	}()

	c.ticker = time.NewTicker(10 * time.Second)
	go c.heartbeat() // 3. 周期性地触发心跳操作
}

// 1. 发送请求到服务端
func (c *Client) handleWrite() {
	// 从通道循环读取请求
	for req := range c.pendingReqs {
		c.doRequest(req)
	}
}

// 2. 读取服务端的请求
func (c *Client) handleRead() error {
	// 读取并解析数据
	ch := parser.ParseStream(c.conn)

	for payload := range ch {
		// 如果解析出错，创建一个标准错误回复，并跳过这个请求
		if payload.Err != nil {
			c.finishRequest(reply.MakeStandardErrorReply(payload.Err.Error()))
			continue
		}

		c.finishRequest(payload.Data)

	}
	return nil
}

// 3. 周期性地触发心跳操作
func (c *Client) heartbeat() {
	// 每隔10秒发送一次心跳请求
	for range c.ticker.C {
		c.doHeartbeat()
	}

}

// 向服务端发送请求
func (c *Client) doRequest(req *request) {
	if req == nil || len(req.args) == 0 {
		return
	}

	// 创建响应并转换为字节切片
	re := reply.MakeMultiBulkReply(req.args) 
	bytes := re.ToBytes()

	// 发送请求到服务端
	_, err := c.conn.Write(bytes)
	i := 0
	// 如果写入出错，重试三次
	for err != nil && i < 3 {
		err = c.handleConnectionError(err)
		if err == nil {
			_, err = c.conn.Write(bytes) // fix variable shadowing
		}
		i++
	}

	// 请求发送成功，将请求放入等待响应的队列
	if err == nil {
		c.waitGroup <- req
	} else {
		req.err = err
		req.waiting.Done()
	}
}

// 处理客户端连接错误的情况
func (c *Client) handleConnectionError(err error) error {
	// 尝试关闭连接
	err1 := c.conn.Close()
	if err1 != nil {
		if opErr, ok := err.(*net.OpError); ok {
			if opErr.Err.Error() != "use of closed network connection" {
				return err1
			}
		} else {
			return err1
		}
	}

	// 重新建立TCP连接
	conn, err1 := net.Dial("tcp", c.addr)
	if err1 != nil {
		logger.Error(err1)
		return err1
	}

	// 将新连接赋值给客户端实例
	c.conn = conn
	go func() {
		_ = c.handleRead() // 重新开始读取数据
	}()

	return nil
}

// 完成请求
func (c *Client) finishRequest(reply resp.Reply) {
	// 捕获panic异常
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			logger.Error(err)
		}
	}()

	// 从通道里获取请求
	request := <-c.waitGroup
	if request == nil {
		return
	}

	// 响应结果赋值给请求
	request.reply = reply
	if request.waiting != nil {
		request.waiting.Done()
	}
}

// 执行心跳请求
func (c *Client) doHeartbeat() {
	// 创建一个心跳请求
	request := &request{
		args:      [][]byte{[]byte("PING")},
		heartbeat: true,
		waiting:   &wait.Wait{},
	}

	// 等待计数器+1，有一个请求等待响应
	request.waiting.Add(1)
	// 工作计时器+1，表示有一个请求在工作
	c.working.Add(1)

	defer c.working.Done()

	// 将请求放入通道中，等待发送
	c.pendingReqs <- request
	// 等待响应
	request.waiting.WaitWithTimeout(3 * time.Second)
}

// 关闭客户端连接
func (client *Client) Close() {
	client.ticker.Stop() // 停止心跳定时器
	close(client.pendingReqs) // 关闭发送请求的通道
	client.working.Wait()
	_ = client.conn.Close()
	close(client.waitGroup) // 关闭等待响应的通道
}

// 发送请求到服务器
func (c *Client) Send(args [][]byte) resp.Reply {
	// 创建请求实例
	request := &request{
		args:      args,
		waiting:   &wait.Wait{},
		heartbeat: false,
	}

	// 等待计数器+1，有一个请求等待响应
	request.waiting.Add(1)
	// 工作计时器+1，表示有一个请求在工作
	c.working.Add(1)
	// 工作完成
	defer c.working.Done()

	// 将请求放入通道中，等待发送
	c.pendingReqs <- request
	// 等待响应
	timeout := request.waiting.WaitWithTimeout(3 * time.Second)	
	if timeout {
		return reply.MakeStandardErrorReply("timeout")
	}
	if request.err != nil {
		return reply.MakeStandardErrorReply(request.err.Error())
	}

	return request.reply
}
