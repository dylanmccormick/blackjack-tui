package main

import (
	"flag"

	"github.com/alecthomas/kong"
	"github.com/dylanmccormick/blackjack-tui/client"
	"github.com/dylanmccormick/blackjack-tui/server"
)

var CLI struct {
	Tui struct {
		Mock bool `help:"Run in mock mode"`
	} `cmd:"Run the blackjack TUI"`
	Server struct{} `cmd:"Run the blackjack Server"`
}

func main() {
	ctx := kong.Parse(&CLI)
	flag.Parse()
	switch ctx.Command() {
	case "tui":
		client.RunTui(CLI.Tui.Mock)
	case "server":
		s := server.InitializeServer()
		s.Run()
	default:
		panic(ctx.Command())
	}
}
