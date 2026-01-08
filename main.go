package main

import (
	"flag"
	"os"

	"github.com/dylanmccormick/blackjack-tui/client"
	"github.com/dylanmccormick/blackjack-tui/server"
)

func main() {
	function := flag.String("f", "tui", "use this flag to determine what function the app will serve. Available options: 'tui', 'server'")
	mockFlg := flag.Bool("mock", false, "add this flag to run in MOCK mode. You will not be able to connect to a websocket")
	flag.Parse()
	args := os.Args
	if len(args) < 2 {
		os.Exit(1)
	}
	switch *function {
	case "tui":
		client.RunTui(mockFlg)
	case "server":
		server.RunServer()
	}
}
