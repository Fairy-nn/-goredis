package tcp

import (
	"context"
	"net"
)

// Handler interface defines the methods that a TCP handler must implement.
// It includes a method to handle incoming connections and a method to close the handler.
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
