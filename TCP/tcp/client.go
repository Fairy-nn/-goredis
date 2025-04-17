package tcp

import (
	"bufio"
	"context"
	"goredis/TCP/lib/logger"
	"goredis/TCP/lib/sync/atomic"
	"goredis/TCP/lib/sync/wait"
	"net"
	"sync"
	"time"
)

// EchoHandler is a TCP handler that echoes back the received data to the client.
type EchoHandler struct {
	activation sync.Map
	closing    atomic.Boolean
}

// EchoHandler implements the Handler interface for echoing data back to the client.
func MakeHandler() *EchoHandler {
	return &EchoHandler{
		activation: sync.Map{},
		closing:    atomic.Boolean(0),
	}
}

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

// EchoClient represents a client connection to the echo server.
func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	err := c.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// incoming connections and echoes back the received data.
func (e *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if e.closing.Get() { // Check if the handler is closing
		conn.Close()
		return
	}
	client := &EchoClient{
		Conn:    conn,
		Waiting: wait.Wait{},
	}
	e.activation.Store(client, struct{}{}) // Store the client connection in the activation map
	reader := bufio.NewReader(conn)        // Read data from the connection
	for {
		msg, err := reader.ReadString('\n') // Read a line from the connection
		if err != nil {
			if err.Error() == "EOF" {
				logger.Info("Client disconnected")
				e.activation.Delete(client) // Remove the client from the activation map
			} else {
				logger.Error("Error reading from connection:", err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		_, err = conn.Write(b) // Echo the message back to the client
		if err != nil {
			logger.Error("Error writing to connection:", err)
			client.Waiting.Done()
			return
		}
		client.Waiting.Done()
	}
}

// close the handler and all active connections
func (e *EchoHandler) Close() error {
	logger.Info("Closing EchoHandler")
	e.closing.Set(true) // Set the closing flag to true
	e.activation.Range(func(key, value interface{}) bool {
		client := key.(*EchoClient) // Get the client connection from the activation map
		err := client.Close()       // Close the client connection
		if err != nil {
			logger.Error("Error closing client connection:", err)
		}
		return true
	})
	return nil
}
