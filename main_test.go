package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
)

func startTestServer(t *testing.T) net.Listener {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to bind: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleConnection(conn)
		}
	}()

	return listener
}

func request(t *testing.T, addr string, req string) string {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "%s\r\n\r\n", req)

	resp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	return strings.TrimSpace(resp)
}

func TestConcurrentRequests(t *testing.T) {
	listener := startTestServer(t)
	defer listener.Close()
	addr := listener.Addr().String()

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := request(t, addr, "GET /health HTTP/1.1")
			if !strings.Contains(resp, "200 OK") {
				errs <- fmt.Errorf("health expected 200, got: %s", resp)
			}
		}()
	}

	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := request(t, addr, "GET /echo/hello HTTP/1.1")
			if !strings.Contains(resp, "200 OK") {
				errs <- fmt.Errorf("echo expected 200, got: %s", resp)
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestConcurrentMixedPaths(t *testing.T) {
	listener := startTestServer(t)
	defer listener.Close()
	addr := listener.Addr().String()

	var wg sync.WaitGroup
	errs := make(chan error, 200)

	paths := []string{
		"GET /health HTTP/1.1",
		"GET /echo/foo HTTP/1.1",
		"GET /echo/bar HTTP/1.1",
		"GET /unknown HTTP/1.1",
		"GET /echo/a HTTP/1.1",
	}

	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, p := range paths {
				resp := request(t, addr, p)
				if resp == "" {
					errs <- fmt.Errorf("empty response for %s", p)
				}
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
