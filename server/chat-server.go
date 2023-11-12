package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

var (
	clients   = make(map[net.Conn]bool)
	clientsMu sync.Mutex
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	fmt.Println("New client connected:", conn.RemoteAddr())

	// Read messages from the client
	go func(conn net.Conn) {
		defer func() {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()

			conn.Close()
			fmt.Println("Client disconnected:", conn.RemoteAddr())
		}()

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			message := scanner.Text()

			// Broadcast the message to all clients
			broadcastMessage(conn, message)
		}
	}(conn)
}

func broadcastMessage(sender net.Conn, message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		if client != sender {
			fmt.Fprintf(client, "%s\n", message)
		}
	}
}
