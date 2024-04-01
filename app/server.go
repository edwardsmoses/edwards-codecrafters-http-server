package main

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	fmt.Println("Listening to connections")

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	fmt.Println("accepting connection", conn)

	// create a new buffer to store the incoming data
	data := make([]byte, 1024)

	// read the incoming connection into the buffer
	_, err = conn.Read(data)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
		os.Exit(1)
	}

	fmt.Println("Received data: ", string(data))
	dataString := strings.Split(string(data), " ")

	fmt.Println("Split data: ", dataString)
	fmt.Println("Method: ", dataString[0])
	fmt.Println("Path: ", dataString[1])

	requestPath := strings.Split(dataString[1], "/")

	if dataString[0] == "GET" && dataString[1] == "/" {
		fmt.Println("Responding with 200 OK")

		httpResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\n\r\n")
		_, err := conn.Write([]byte(httpResponse))

		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}

	} else if dataString[0] == "GET" && slices.Contains(requestPath, "echo") {
		fmt.Println("Secret: ", requestPath)

		fmt.Println("Responding with 200 OK")
		content := requestPath[2]

		httpResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
		_, err := conn.Write([]byte(httpResponse))

		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Println("Responding with 404 Not Found")
		_, err = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}

	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}
