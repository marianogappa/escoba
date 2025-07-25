//go:build !tinygo
// +build !tinygo

package main

import (
	"fmt"
	"os"

	"github.com/marianogappa/escoba/exampleclient"
	"github.com/marianogappa/escoba/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: escoba server")
		fmt.Println("usage: escoba player1|player2 [address]")
		fmt.Println("Define the PORT environment variable for escoba server to change the default port (8080).")
		os.Exit(0)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf("localhost:%v", port)
	if len(os.Args) >= 3 {
		address = os.Args[2]
	}

	arg := os.Args[1]
	switch arg {
	case "server":
		server.New(port).Start()
	case "player1":
		exampleclient.Player(0, address)
	case "player2":
		exampleclient.Player(1, address)
	default:
		fmt.Println("Invalid argument. Please provide either server or client.")
	}
}
