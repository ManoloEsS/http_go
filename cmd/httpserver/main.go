package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ManoloEsS/http_go/internal/request"
	"github.com/ManoloEsS/http_go/internal/response"
	"github.com/ManoloEsS/http_go/internal/server"
)

const port = 42069

const badRequest = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const internalServerError = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const successResponse = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

func test_handler(writer *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		_ = writer.WriteStatusLine(http.StatusBadRequest)
		_ = writer.WriteHeaders(response.GetDefaultHeaders(len(badRequest), "text/html"))
		_, _ = writer.WriteBody([]byte(badRequest))

	case "/myproblem":
		_ = writer.WriteStatusLine(http.StatusInternalServerError)
		_ = writer.WriteHeaders(response.GetDefaultHeaders(len(internalServerError), "text/html"))
		_, _ = writer.WriteBody([]byte(internalServerError))
	default:
		_ = writer.WriteStatusLine(http.StatusOK)
		_ = writer.WriteHeaders(response.GetDefaultHeaders(len(successResponse), "text/html"))
		_, _ = writer.WriteBody([]byte(successResponse))
	}
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
