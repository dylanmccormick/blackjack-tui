package main

import (
	"os"

	"github.com/dylanmccormick/blackjack-tui/client"
	"github.com/dylanmccormick/blackjack-tui/server"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		os.Exit(1)
	}
	switch args[1] {
	case "tui":
		client.RunTui()
	case "server":
		server.RunServer()
	}
}
