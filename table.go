package main

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"

	"github.com/dylanmccormick/blackjack-tui/game"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type Table struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	inbound    chan []byte
	outbound   chan []byte
	stopChan   chan struct{}
	id         string

	maxPlayers int
	game       *game.Game
}

func newTable() *Table {
	return &Table{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		inbound:    make(chan []byte),
		outbound:   make(chan []byte),
		stopChan:   make(chan struct{}),
		id:         "placeholder",
		game:       game.NewGame(),
	}
}

func (h *Table) run() {
	for {
		select {
		case <-h.stopChan:
			slog.Info("Killing Table")
			return
		case client := <-h.register:
			h.clients[client] = true
			p := game.NewPlayer(1)
			client.id = 1
			h.game.AddPlayer(p)
		case client := <-h.unregister:
			slog.Info("attempting to unregister client", "client", client.id)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.inbound:
			slog.Info("Received message", "message", string(message))
			h.processMessage(message)
			// for client := range h.clients {
			// 	select {
			// 	case client.send <- message:
			// 	default:
			// 		close(client.send)
			// 		delete(h.clients, client)
			// 	}
			// }
			h.autoProgress()
		}
	}
}

func (h *Table) processMessage(msg []byte) {
	// jsonMsg := &protocol.TransportMessage{}
	// json.Unmarshal(msg, jsonMsg)
	strMsg := string(msg)
	parts := strings.Split(strMsg, " ")
	for _, p := range parts {
		slog.Info(p)
	}
	h.handleCommand(parts)
}

func (t *Table) handleCommand(command []string) {
	switch command[0] {
	case "start":
		slog.Info("Starting game")
		t.game.StartDealing()
		t.game.Deck.Shuffle()
	case "bet":
		slog.Info("Betting")
		bet, err := strconv.Atoi(command[1])
		if err != nil {
			slog.Error("Got bad data from command", "command", command)
		}
		t.game.PlaceBet(t.game.GetPlayer(1), bet)
	case "deal":
		t.game.DealCards()
	case "hit":
		slog.Info("Hitting")
		t.game.Hit(t.game.GetPlayer(1))
	case "stay":
		t.game.Stay(t.game.GetPlayer(1))
		slog.Info("Standing")
	case "dealer":
		t.game.PlayDealer()
	}
}

func (t *Table) autoProgress() {
OuterLoop:
	for {
		switch t.game.State {
		case game.DEALER_TURN:
			t.game.PlayDealer()
		case game.RESOLVING_BETS:
			t.game.ResolveBets()
		default:
			t.broadcastGameState()
			break OuterLoop
		}
		t.broadcastGameState()
	}
}

func (t *Table) broadcastGameState() {
	gameData := protocol.GameToDTO(t.game)
	data, err := json.Marshal(gameData)
	if err != nil {
		slog.Error("Marshalling bad")
		return
	}
	for client := range t.clients {
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(t.clients, client)
		}
	}
}
