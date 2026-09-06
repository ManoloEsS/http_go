package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/ManoloEsS/http_go/internal/request"
	"github.com/ManoloEsS/http_go/internal/response"
)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	statusCode response.StatusCode
	message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func writeError(w io.Writer, handlerError HandlerError) {
	err := response.WriteStatusLine(w, handlerError.statusCode)
	if err != nil {
		log.Printf("could not respond with error: %v", err)
	}

	messageBytes := []byte(handlerError.message)

	err = response.WriteHeaders(w, response.GetDefaultHeaders(len(messageBytes)))
	if err != nil {
		log.Printf("could not respond with error: %v\n", err)
	}

	n, err := response.WriteBody(w, messageBytes)
	if err != nil {
		log.Printf("could not respond with error: %v\n", err)
	}
	if n < len(messageBytes) {
		log.Print("could not write full error response body\n")
	}
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: ln,
		closed:   atomic.Bool{},
		handler:  handler,
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		writeError(conn, HandlerError{
			statusCode: 400,
			message:    "Bad Request",
		})
		return
	}

	var buff bytes.Buffer

	err = s.handler(conn, req)
	if err != nil {
		writeError(conn, HandlerError{
			statusCode: 500,
			message:    "Internal Server Error",
		})
		return
	}

	err = response.WriteStatusLine(buff, 200)
	if err != nil {
		return
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	if err != nil {
		return
	}
}
