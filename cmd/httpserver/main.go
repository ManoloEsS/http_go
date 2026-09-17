package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

func proxy_handler(writer *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

		res, err := http.Get(fmt.Sprintf("%s/%s", "https://httpbingo.org", path))
		if err != nil {
			_ = writer.WriteStatusLine(http.StatusNotFound)
			_ = writer.WriteHeaders(response.GetDefaultHeaders(len("Not Found"), "text/html"))
			_, _ = writer.WriteBody([]byte("Not Found"))
			return
		}
		defer res.Body.Close()

		_ = writer.WriteStatusLine(response.StatusCode(200))
		headers := response.GetDefaultHeaders(0, res.Header.Get("Content-Type"))
		headers.Append("Transfer-Encoding", "chunked")
		headers.Remove("content-length")

		_ = writer.WriteHeaders(headers)

		buff := make([]byte, 1024)

		for {
			n, err := res.Body.Read(buff)
			fmt.Printf("read: %d\n", n)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(buff[:n]) > 0 {
						chunkSize, _ := writer.WriteChunkedBody(buff[:n])
						fmt.Printf("wrote chunk of size: %d\n", chunkSize)
					}
				}
				done, err := writer.WriteChunkedBodyDone()
				if err != nil {
					log.Printf("could not write ending chunk")
				}
				fmt.Printf("done writing, end was %d bytes\n", done)
				return
			}

			written, err := writer.WriteChunkedBody(buff[:n])
			if err != nil {
				log.Printf("error writing chunk to client\n")
				return
			}
			fmt.Printf("wrote chunk of size: %d\n", written)

		}
	}

	_ = writer.WriteStatusLine(http.StatusNotFound)
	_ = writer.WriteHeaders(response.GetDefaultHeaders(len("Not Found"), "text/html"))
	_, _ = writer.WriteBody([]byte("Not Found"))

}

func main() {
	server, err := server.Serve(port, proxy_handler)
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
