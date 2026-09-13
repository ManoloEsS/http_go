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

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (hErr *HandlerError) WriteError(w io.Writer) {
	err := response.WriteStatusLine(w, hErr.StatusCode)
	if err != nil {
		log.Printf("could not respond with error: %v", err)
	}

	messageBytes := []byte(hErr.Message)

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
		handlerError := &HandlerError{
			StatusCode: 400,
			Message:    "Bad Request",
		}
		handlerError.WriteError(conn)
		return
	}

	var buff bytes.Buffer

	handlerErr := s.handler(&buff, req)
	if handlerErr != nil {
		handlerErr.WriteError(conn)
		return
	}

	err = response.WriteStatusLine(conn, 200)
	if err != nil {
		log.Println("error writing response status line")
		return
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(buff.Len()))
	if err != nil {
		log.Println("error response writing headers")
		return
	}

	_, err = buff.WriteTo(conn)
	if err != nil {
		log.Println("error writing response body")
		return
	}
}
