package handler

import (
	"context"
	"fmt"
	"goredis/cluster"
	"goredis/config"
	"goredis/database"
	databaseinterface "goredis/interface/database"
	"goredis/lib/logger"
	"goredis/lib/sync/atomic"
	"goredis/resp/connection"
	"goredis/resp/parser"
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
	db         databaseinterface.Database // database interface
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	// db := database.NewStandaloneDatabase() // create a new database instance
	// return &RespHandler{
	// 	db: db,
	// }
	var db databaseinterface.Database
	if config.Properties.Self != "" && len(config.Properties.Peers) > 0 {
		db = cluster.MakeClusterDatabase()
	} else {
		db = database.NewStandaloneDatabase()
	}
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
	fmt.Println("ch:", ch)
	for payload := range ch { // read the payload from the channel
		fmt.Println("在通道中读取到数据:", payload)
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
