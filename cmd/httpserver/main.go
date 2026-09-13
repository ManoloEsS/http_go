package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ManoloEsS/http_go/internal/request"
	"github.com/ManoloEsS/http_go/internal/server"
)

const port = 42069

func test_handler(w io.Writer, req *request.Request) *server.HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return &server.HandlerError{
			StatusCode: http.StatusBadRequest,
			Message:    "Your problem is not my problem\n",
		}
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		return &server.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Woopsie, my bad\n",
		}
	}

	_, err := fmt.Fprintf(w, "All good, frfr\n")
	if err != nil {
		log.Println("could not write body to response")
	}

	return nil
}

func main() {
	server, err := server.Serve(port, test_handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
