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
	defer file.Close()

	buf := make([]byte, 8, 8)
	var line strings.Builder

	for {
		n, err := file.Read(buf)
		if err != nil {
			if line.Len() > 0 {
				fmt.Printf("read: %s\n", line.String())
				line.Reset()
			}
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Printf("error: %s\n", err.Error())
			break
		}

		str := string(buf[:n])
		parts := strings.Split(str, "\n")

		for i := 0; i < len(parts)-1; i++ {
			line.WriteString(parts[i])
			fmt.Printf("read: %s\n", line.String())
			line.Reset()
		}

		line.WriteString(parts[len(parts)-1])

	}
}
