package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ManoloEsS/http_go/internal/headers"
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
		h := response.GetDefaultHeaders(0, res.Header.Get("Content-Type"))
		h.Set("Transfer-Encoding", "chunked")
		h.Append("Trailer", "X-Content-SHA256")
		h.Append("Trailer", "X-Content-Length")
		h.Remove("content-length")

		_ = writer.WriteHeaders(h)

		buff := make([]byte, 1024)
		body := []byte{}

		for {
			n, err := res.Body.Read(buff)
			fmt.Printf("read: %d\n", n)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(buff[:n]) > 0 {
						body = append(body, buff[:n]...)
						chunkSize, _ := writer.WriteChunkedBody(buff[:n])
						fmt.Printf("wrote chunk of size: %d\n", chunkSize)
					}
				}
				done, err := writer.WriteChunkedBodyDone()
				if err != nil {
					log.Printf("could not write ending chunk")
				}
				fmt.Printf("done writing, end was %d bytes\n", done)
				break
			}

			body = append(body, buff[:n]...)
			written, err := writer.WriteChunkedBody(buff[:n])
			if err != nil {
				log.Printf("error writing chunk to client\n")
				break
			}
			fmt.Printf("wrote chunk of size: %d\n", written)

		}

		trailers := headers.NewHeaders()
		bodyHash := sha256.Sum256(body)
		trailers.Append("X-Content-Length", fmt.Sprintf("%d", len(body)))
		trailers.Append("X-Content-SHA256", hex.EncodeToString(bodyHash[:]))
		err = writer.WriteTrailers(trailers)
		if err != nil {
			fmt.Printf("error writing headers: %v\n", err)
		}

		return
	}

	_ = writer.WriteStatusLine(http.StatusNotFound)
	_ = writer.WriteHeaders(response.GetDefaultHeaders(len("Not Found"), "text/html"))
	_, _ = writer.WriteBody([]byte("Not Found"))

}

func video_handler(writer *response.Writer, req *request.Request) {
	currentWd, _ := os.Getwd()
	videoPath := filepath.Join(currentWd + "/assets/vim.mp4")

	videoFile, err := os.ReadFile(videoPath)
	_ = writer.WriteStatusLine(http.StatusOK)
	_ = writer.WriteHeaders(response.GetDefaultHeaders(len(videoFile), "video/mp4"))
	if err != nil {
		fmt.Printf("could not read video file: %v", err)
	}
	_, _ = writer.WriteBody([]byte(videoFile))
}

func main() {
	server, err := server.Serve(port, video_handler)
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
