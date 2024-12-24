package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	// Parse timeout flag
	timeoutPtr := flag.String("timeout", "10s", "connection timeout")
	flag.Parse()

	// Get host and port from positional arguments
	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: go-telnet [--timeout=<duration>] <host> <port>\n")
		os.Exit(1)
	}
	host := args[0]
	port := args[1]

	// Parse timeout duration
	var timeout time.Duration
	var err error
	if timeout, err = time.ParseDuration(*timeoutPtr); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid timeout value: %v\n", err)
		os.Exit(1)
	}

	// Create a TCP connection with timeout
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Set read and write deadlines based on timeout
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set deadline: %v\n", err)
		os.Exit(1)
	}

	// Start reading from STDIN and writing to connection
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			data, err := reader.ReadBytes('\n')
			if err != nil {
				// Handle EOF or error
				if err == io.EOF {
					// Close the connection when stdin is closed
					conn.Close()
				}
				return
			}
			_, writeErr := conn.Write(data)
			if writeErr != nil {
				fmt.Fprintf(os.Stderr, "Write error: %v\n", writeErr)
				return
			}
		}
	}()

	// Start reading from connection and writing to STDOUT
	go func() {
		reader := bufio.NewReader(conn)
		for {
			data, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					// Server closed connection
				} else {
					fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
				}
				return
			}
			_, writeErr := os.Stdout.Write(data)
			if writeErr != nil {
				fmt.Fprintf(os.Stderr, "Write to stdout error: %v\n", writeErr)
				return
			}
		}
	}()

	// Wait for both goroutines to finish
	// Since we can't wait on them directly, we'll just loop until connection is closed
	for {
		if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
			break
		}
		select {
		case <-time.After(1 * time.Second):
			continue
		}
	}
}
