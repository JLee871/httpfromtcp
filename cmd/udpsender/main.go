package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const port = ":42069"

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		log.Fatalf("error resolving UDP address: %s\n", err.Error())
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("error making UDP connection: %s\n", err.Error())
	}
	defer udpConn.Close()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("> ")
		line, err := stdReader.ReadString('\n')

		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			break
		}

		_, err = udpConn.Write([]byte(line))

		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			break
		}
	}
}
