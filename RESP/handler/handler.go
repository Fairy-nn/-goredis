package handler

import (
	"context"
	"goredis/RESP/connection"
	"goredis/RESP/parser"
	"goredis/TCP/lib/logger"
	"goredis/TCP/lib/sync/atomic"
	data "goredis/database"
	"goredis/interface/database"
	"goredis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownCommandError = []byte("-ERR unknown command\r\n")
)

type RespHandler struct {
	activeConn sync.Map
	db         database.Database
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	db := data.NewEchoDatabase() // 创建一个新的数据库实例
	return &RespHandler{
		db: db,
	}
}

// recive and process the command
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() { // Check if the handler is closing
		_ = conn.Close()
	}

	client := connection.NewConnection(conn) // create a new connection instance
	h.activeConn.Store(client, 1)            // mark the connection as active

	ch := parser.ParseStream(conn)

	for payload := range ch { // read the payload from the channel
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				h.closeClient(client)
				logger.Info("Client disconnected")
				return
			}

			errReply := reply.MakeStandardErrorReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())

			if err != nil {
				h.closeClient(client)
				logger.Error("Error writing to client:", err)
			}
			continue
		}
		if payload.Data == nil {
			logger.Error("Payload data is nil")
			continue
		}

		r, ok := payload.Data.(*reply.MultiBulkReply) // extract the MultiBulkReply from the payload
		if !ok {
			logger.Error("Payload data is not a MultiBulkReply")
			continue
		}

		result := h.db.Exec(client, r.Args)

		if result == nil {
			_ = client.Write(unknownCommandError)
		} else {
			_ = client.Write(result.ToBytes())
		}

		logger.Info("Command executed successfully")
	}
}

func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

func (h *RespHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// TODO: concurrent wait
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
	return nil
}
