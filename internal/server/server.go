package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/ManoloEsS/http_go/internal/response"
)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
}

var resp = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!\n")

func Serve(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: ln,
		closed:   atomic.Bool{},
	}

	fmt.Println("starting listen from serve")

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		return
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	if err != nil {
		return
	}
}
