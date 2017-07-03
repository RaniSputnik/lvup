package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
)

type Command string

const (
	CmdSetGlobal = Command("setglobal\n")
	CmdRestart   = Command("restart\n")
	CmdStop      = Command("stop\n")
)

// Server describes a server that will listen for connections
// from Love clients and is bound by the given context.
type Server interface {
	Listen(ctx context.Context, addr string) error
	Command(cmd Command) error
}

// ErrAlreadyListening is returned when a duplicate call to Server.Listen is made
var ErrAlreadyListening = errors.New("TCP server is already listening")

func NewServer(logger *log.Logger) Server {
	return &TCPServer{
		Log: logger,
	}
}

type Client struct {
	io.WriteCloser
}

type TCPServer struct {
	Log *log.Logger

	listener net.Listener

	connected    chan Client
	disconnected chan Client
	broadcast    chan Command
}

func (s *TCPServer) Listen(ctx context.Context, addr string) error {
	if s.listener != nil {
		return ErrAlreadyListening
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = l

	go func(ctx context.Context) {
		defer l.Close()

		clients := map[Client]struct{}{}
		s.connected = make(chan Client)
		s.disconnected = make(chan Client)
		s.broadcast = make(chan Command)

		go func(ctx context.Context, l net.Listener, connected chan<- Client, logger *log.Logger) {
			for ctx.Err() == nil {
				// Listen for an incoming connection.
				conn, err := l.Accept()
				if err != nil {
					logger.Printf("Failed to accept connection: %s\n", err)
					continue
				}

				connected <- Client{conn}
			}
		}(ctx, l, s.connected, s.Log)

		for {
			select {
			case client := <-s.connected:
				s.Log.Println("New connection established")
				clients[client] = struct{}{}

			case broadcast := <-s.broadcast:
				s.Log.Printf("Broadcasting message: %s\n", broadcast)
				// TODO this is pretty rubbish, we should probably
				// have the broadcasts running in another goroutine?
				// But how to maintain a list of clients without a mutex?
				for client := range clients {
					_, err := client.Write([]byte(broadcast))
					if err == io.EOF {
						// Disconnect - Can't send to disconnected
						// channel because it isn't buffered
						delete(clients, client)
					}
				}
			case <-ctx.Done():
				s.Log.Printf("Finishing listen: %s\n", ctx.Err())
				break
			}
		}
	}(ctx)

	return nil
}

func (s *TCPServer) Command(cmd Command) error {
	s.Log.Printf("Broadcasting command: %s", cmd)
	// TODO catch 'Server not listening'
	s.broadcast <- cmd
	return nil
}
