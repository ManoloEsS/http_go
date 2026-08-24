package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	network = "udp"
	port    = "localhost:42069"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr(network, port)
	if err != nil {
		log.Fatalf("error starting udp address: %s\n", err.Error())
	}

	udpConn, err := net.DialUDP(network, nil, udpAddr)
	if err != nil {
		log.Fatalf("error creating udp connection at address %s: %s\n", udpAddr.String(), err.Error())
	}
	defer udpConn.Close()

	buf := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		in, err := buf.ReadString('\n')
		if err != nil {
			log.Printf("encountered error reading input: %w", err)
		}

		_, err = udpConn.Write([]byte(in))
		if err != nil {
			log.Printf("error sending input '%s' to %s\n", in, udpConn.LocalAddr().String())
		}
	}
}
