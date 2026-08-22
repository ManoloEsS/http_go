package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("could not open %s: %s\n", inputFilePath, err)
	}

	linesChan := getLinesChannel(file)
	for line := range linesChan {
		fmt.Println("read:", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
		defer close(lines)
		buf := make([]byte, 8, 8)
		var line strings.Builder

		for {
			n, err := f.Read(buf)
			if err != nil {
				if line.Len() > 0 {
					lines <- line.String()
					line.Reset()
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}

			str := string(buf[:n])
			parts := strings.Split(str, "\n")

			for i := 0; i < len(parts)-1; i++ {
				line.WriteString(parts[i])
				lines <- line.String()
				line.Reset()
			}

			line.WriteString(parts[len(parts)-1])

		}
	}()
	return lines
}
