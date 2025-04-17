package database

import (
	"goredis/interface/resp"
)

type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	AfterClientClose(c resp.Connection)
	Close()
}

// used to store the data entity
type DataEntity struct {
	Data interface{}
}
