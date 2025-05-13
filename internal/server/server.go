package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	netListener net.Listener

	closed atomic.Bool //track whether server is closed
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("error creating net listener")
	}

	server := &Server{
		netListener: listener,
	}

	go func() {
		server.listen()
	}()

	return server, nil
}

func (s *Server) Close() error {
	err := s.netListener.Close()
	s.closed.Store(true)
	return err
}

func (s *Server) listen() {
	for {
		if s.closed.Load() {
			break
		}

		conn, err := s.netListener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %s", err)
		}
		go func() {
			s.handle(conn)
		}()
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	response := "HTTP/1.1 200 OK\r\n" + // Status line
		"Content-Type: text/plain\r\n" + // Example header
		"\r\n" + // Blank line to separate headers from the body
		"Hello World!\n" // Body
	conn.Write([]byte(response))
}
