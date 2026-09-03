package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ManoloEsS/http_go/internal/request"
)

const (
	inputFilePath = "messages.txt"
	network       = "tcp"
	port          = ":42069"
)

func main() {
	listener, err := net.Listen(network, port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("could not start connection: %s\n", err.Error())
		}

		fmt.Printf("Connection has been accepted from %s\n\n", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error parsing request: %s\n", err.Error())
		}

		fmt.Printf(
			"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n\n",
			req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion,
		)
		fmt.Printf("Headers:\n")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}

		fmt.Println("Connection has been closed to", conn.RemoteAddr())
	}
}

// func getLinesChannel(f io.ReadCloser) <-chan string {
// 	lines := make(chan string)
// 	go func() {
// 		defer f.Close()
// 		defer close(lines)
// 		buf := make([]byte, 8, 8)
// 		var line strings.Builder
//
// 		for {
// 			n, err := f.Read(buf)
// 			if err != nil {
// 				if line.Len() > 0 {
// 					lines <- line.String()
// 					line.Reset()
// 				}
// 				if errors.Is(err, io.EOF) {
// 					break
// 				}
// 				fmt.Printf("error: %s\n", err.Error())
// 				return
// 			}
//
// 			str := string(buf[:n])
// 			parts := strings.Split(str, "\n")
//
// 			for i := 0; i < len(parts)-1; i++ {
// 				line.WriteString(parts[i])
// 				lines <- line.String()
// 				line.Reset()
// 			}
//
// 			line.WriteString(parts[len(parts)-1])
//
// 		}
// 	}()
// 	return lines
// }
