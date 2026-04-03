package main

import (
	"fmt"
	"net"

	"github.com/NeerajRijhwani/peer-cdn/internal/storage"
	"github.com/NeerajRijhwani/peer-cdn/internal/swarm"
)

func main() {
	db, err := storage.Connect()
	if err != nil {
		fmt.Printf("Db Connection Failed")
		return
	}
	swarmManager := swarm.Initalize_SwarmManager()

	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8000")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Handle client connection in a goroutine
		go swarm.Handleconection(conn, swarmManager, db, "", "")
	}
}
