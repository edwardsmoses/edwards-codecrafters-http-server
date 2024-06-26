package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const okResponseHead = "HTTP/1.1 200 OK"
const writeFilePOSTResponseHead = "HTTP/1.1 201 OK"
const notFoundResponseHead = "HTTP/1.1 404 Not Found"
const crlf = "\r\n"

var serverDir string

func main() {
	fmt.Println("Listening to connections on port 4221")

	dir := flag.String("directory", "", "The name of the directory")
	flag.Parse()

	serverDir = *dir
	fmt.Println("Server directory:", serverDir)

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

	//get the content length
	contentLength, err := strconv.Atoi(headers["Content-Length"])
	if err != nil {
		fmt.Println("Error reading content length:", err)
	}
	var requestBody string
	if err == nil {
		requestBody = parseRequestBody(reader, contentLength)
	} else {
		requestBody = ""
	}
	processRequest(method, path, headers, conn, requestBody)
}

func parseRequestLine(requestLine string) (method, path, version string) {
	parts := strings.Fields(requestLine)
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2]
	}
	return "", "", ""
}

func parseRequestBody(reader *bufio.Reader, contentLength int) string {
	fmt.Println("Content length:", contentLength)
	body := make([]byte, contentLength)
	_, err := reader.Read(body)
	if err != nil {
		fmt.Println("Error reading body:", err)
		return ""
	}
	return string(body)
}

func processRequest(method, path string, headers map[string]string, conn net.Conn, requestBody string) {

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

	if method == "GET" && strings.HasPrefix(path, "/files/") {
		filePath := path[len("/files/"):]
		fmt.Println("File path:", filePath)

		if serverDir != "" {
			filePath = serverDir + filePath
			fmt.Println("File path with directory:", filePath)

			fileData, readErr := os.ReadFile(filePath)
			if readErr != nil {
				sendResponse(conn, notFoundResponseHead, nil, "")
				return
			} else {
				content := string(fileData[:])
				fmt.Println("File content:", content)
				sendResponse(conn, okResponseHead, map[string]string{"Content-Type": "application/octet-stream"}, content)
				return
			}

		}
	}

	if method == "POST" && strings.HasPrefix(path, "/files/") {
		filePath := path[len("/files/"):]

		if serverDir != "" {
			filePath = serverDir + filePath

			fmt.Println("saving this content in the file", requestBody, filePath)
			err := os.WriteFile(filePath, []byte(requestBody), 0644) // 0644 is the permission - read and write
			if err != nil {
				fmt.Println("Error writing file:", err)
				sendResponse(conn, notFoundResponseHead, nil, "")
				return
			} else {
				sendResponse(conn, writeFilePOSTResponseHead, nil, "")
				return
			}
		}
	}

	if path == "/" {
		sendResponse(conn, okResponseHead, nil, "")
		return
	}

	sendResponse(conn, notFoundResponseHead, nil, "")
}

func sendResponse(conn net.Conn, statusLine string, headers map[string]string, body string) {
	// Initialize headers map if it is nil
	if headers == nil {
		headers = make(map[string]string)
	}

	headers["Content-Length"] = strconv.Itoa(len(body))
	response := statusLine + crlf
	for key, value := range headers {
		response += key + ": " + value + crlf
	}
	response += crlf + body
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error sending response:", err)
	}
}
