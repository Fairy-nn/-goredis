package connection

import (
	"goredis/tcp/lib/sync/wait"
	"net"
	"sync"
	"time"
)

type Connection struct {
	conn         net.Conn   // network connection
	waitingReply wait.Wait  // wait for reply
	mutex        sync.Mutex // mutex for connection
	selectedDB   int        // selected database
}

// create a new connection instance
func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

// get the address of the connection
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// close the connection
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

// write data to the connection
func (c *Connection) Write(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	c.mutex.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mutex.Unlock()
	}()
	_, err := c.conn.Write(data)
	return err
}

// get the index of the selected database
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// select a database
func (c *Connection) SelectDB(db int) error {
	c.selectedDB = db
	return nil
}
