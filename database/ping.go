package database

import (
	"goredis/interface/resp"
	"goredis/resp/reply"
)

func init() {
	RegisterCommand("ping", Ping, 1)
}

// Register the ping command with arity 0
func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}
