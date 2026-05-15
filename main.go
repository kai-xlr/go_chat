package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	healthResponse   = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 2\r\n\r\nOK"
	notFoundResponse = "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain\r\nContent-Length: 9\r\n\r\nNot Found"
	badRequestResp   = "HTTP/1.1 400 Bad Request\r\nContent-Type: text/plain\r\nContent-Length: 11\r\n\r\nBad Request"
)

func writeResponse(conn net.Conn, response string) {
	if _, err := conn.Write([]byte(response)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("Accepted connection from %s", conn.RemoteAddr())

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("Failed to read from connection: %v", err)
		return
	}

	requestStr := string(buf[:n])
	log.Printf("Received request:\n%s", requestStr)

	lines := strings.Split(requestStr, "\r\n")
	if len(lines) == 0 {
		return
	}

	parts := strings.Split(lines[0], " ")
	if len(parts) < 2 {
		writeResponse(conn, badRequestResp)
		return
	}

	path := parts[1]

	switch {
	case path == "/health":
		writeResponse(conn, healthResponse)
	case strings.HasPrefix(path, "/echo/"):
		message := strings.TrimPrefix(path, "/echo/")
		writeResponse(conn, fmt.Sprintf(
			"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
			len(message),
			message,
		))
	default:
		writeResponse(conn, notFoundResponse)
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Failed to bind to port 8080: %v", err)
	}
	defer listener.Close()

	log.Println("Server listening on localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}
