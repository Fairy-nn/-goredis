package tcp

import (
	"context"
	"fmt"

	"goredis/interface/tcp"
	"goredis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Address string
}

// incoming TCP connections on the specified address and port, and handles them using the provided handler.
// also listens for termination signals (SIGINT, SIGTERM) to gracefully shut down the server.
func ListenServerWithSig(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT) // listen for termination signals
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address) // listen on the specified address and port
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("bind: %s, start listening", cfg.Address))
	ListenAndServe(listener, handler, closeChan) // start accepting connections
	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	go func() {
		<-closeChan // wait for termination signal
		logger.Info("shutting down")
		_ = listener.Close() // close the listener
		_ = handler.Close() // close the handler
	}()

	defer func() {
		// close during unexpected error
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background() // create a new context

	var waitDone sync.WaitGroup
	for { // accept incoming connections in a loop 
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accept link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn) // handle the connection
		}()
	}
	waitDone.Wait()
}
