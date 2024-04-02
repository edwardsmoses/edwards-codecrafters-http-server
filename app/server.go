package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const okResponseHead = "HTTP/1.1 200 OK"
const notFoundResponseHead = "HTTP/1.1 404 Not Found"
const crlf = "\r\n"

func main() {
	fmt.Println("Listening to connections on port 4221")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221:", err)
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request line:", err)
		return
	}

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == crlf {
			break
		}
		split := strings.SplitN(line, ": ", 2)
		if len(split) == 2 {
			headers[split[0]] = strings.Trim(split[1], crlf)
		}
	}

	method, path, _ := parseRequestLine(requestLine)
	processRequest(method, path, headers, conn)
}

func parseRequestLine(requestLine string) (method, path, version string) {
	parts := strings.Fields(requestLine)
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2]
	}
	return "", "", ""
}

func processRequest(method, path string, headers map[string]string, conn net.Conn) {
	if path == "/user-agent" {
		responseBody := headers["User-Agent"]
		sendResponse(conn, okResponseHead, map[string]string{"Content-Type": "text/plain"}, responseBody)
		return
	}

	if strings.HasPrefix(path, "/echo/") {
		responseBody := path[len("/echo/"):]
		sendResponse(conn, okResponseHead, map[string]string{"Content-Type": "text/plain"}, responseBody)
		return
	}

	if path == "/" {
		sendResponse(conn, okResponseHead, nil, "")
		return
	}

	sendResponse(conn, notFoundResponseHead, nil, "")
}

func sendResponse(conn net.Conn, statusLine string, headers map[string]string, body string) {
	headers["Content-Length"] = strconv.Itoa(len(body))
	response := statusLine + crlf
	for key, value := range headers {
		response += key + ": " + value + crlf
	}
	response += crlf + body
	conn.Write([]byte(response))
}
