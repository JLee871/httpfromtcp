package main

import (
	"fmt"
	"log"
	"net"

	request "github.com/JLee871/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		netConn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}

		fmt.Println("Connection has been accepted from", netConn.RemoteAddr())
		fmt.Println("=====================================")

		/*
			lineChannel := getLinesChannel(netConn)
			for line := range lineChannel {
				fmt.Println(line)
			}
		*/

		req, err := request.RequestFromReader(netConn)
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}

		fmt.Println("Request line:")
		fmt.Println("- Method:", req.RequestLine.Method)
		fmt.Println("- Target:", req.RequestLine.RequestTarget)
		fmt.Println("- Version:", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.RequestHeaders {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.RequestBody))

		fmt.Println("=====================================")
		fmt.Println("Connection to ", netConn.RemoteAddr(), "has been closed")
	}
}

/*
func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		defer f.Close()
		currentLine := ""
		for {
			byteArray := make([]byte, 8)
			_, err := f.Read(byteArray)
			if err != nil {
				if err == io.EOF {
					if currentLine != "" {
						lines <- currentLine
					}
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				break
			}
			stringSlice := strings.Split(string(byteArray), "\n")

			n := len(stringSlice)

			for i := 0; i < n-1; i++ {
				lines <- currentLine + stringSlice[i]
				currentLine = ""
			}

			currentLine += stringSlice[n-1]
		}
	}()
	return lines
}
*/
