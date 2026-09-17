package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/ManoloEsS/http_go/internal/request"
	"github.com/ManoloEsS/http_go/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode  response.StatusCode
	Message     string
	ContentType string
}

func (hErr *HandlerError) WriteError(w io.Writer) {
	writer := response.Writer{
		Writer: w,
	}
	err := writer.WriteStatusLine(hErr.StatusCode)
	if err != nil {
		log.Printf("could not respond with error: %v", err)
	}

	messageBytes := []byte(hErr.Message)

	headers := response.GetDefaultHeaders(len(messageBytes), hErr.ContentType)

	err = writer.WriteHeaders(headers)
	if err != nil {
		log.Printf("could not respond with error: %v\n", err)
	}

	n, err := writer.WriteBody(messageBytes)
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
		handlerError := &HandlerError{
			StatusCode: 400,
			Message:    "Bad Request",
		}
		handlerError.WriteError(conn)
		return
	}

	writer := &response.Writer{
		Writer: conn,
	}

	s.handler(writer, req)
}
