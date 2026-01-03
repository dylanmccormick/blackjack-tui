package main

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/dylanmccormick/blackjack-tui/game"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type inboundMessage struct {
	data   []byte
	client *Client
}

type Table struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	inbound    chan inboundMessage
	outbound   chan []byte
	stopChan   chan struct{}
	id         string

	maxPlayers int
	game       *game.Game
	betTimer   *time.Timer
}

func newTable() *Table {
	return &Table{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		inbound:    make(chan inboundMessage),
		outbound:   make(chan []byte),
		stopChan:   make(chan struct{}),
		id:         "placeholder",
		game:       game.NewGame(),
		betTimer:   time.NewTimer(30 * time.Second),
	}
}

func (t *Table) run() {
	for {
		select {
		case <-t.stopChan:
			slog.Info("Killing Table")
			return
		case client := <-t.register:
			t.clients[client] = true
			p := game.NewPlayer(client.id)
			t.game.AddPlayer(p)
		case client := <-t.unregister:
			slog.Info("attempting to unregister client", "client", client.id)
			if _, ok := t.clients[client]; ok {
				delete(t.clients, client)
				close(client.send)
			}
		case message := <-t.inbound:
			slog.Info("Received message", "message", string(message.data))
			t.processMessage(message)
			t.autoProgress()
		case <-t.betTimer.C:
			slog.Info("TIMER EXPIRED")
			t.game.StartRound()
			t.autoProgress()
		}
	}
}

func (h *Table) processMessage(msg inboundMessage) {
	// jsonMsg := &protocol.TransportMessage{}
	// json.Unmarshal(msg, jsonMsg)
	data := msg.data
	strMsg := string(data)
	parts := strings.Split(strMsg, " ")
	for _, p := range parts {
		slog.Info(p)
	}
	h.handleCommand(parts, msg.client)
}

func (t *Table) handleCommand(command []string, c *Client) {
	switch command[0] {
	case "start":
		slog.Info("Starting game")
		t.betTimer.Reset(30 * time.Second)
	case "bet":
		slog.Info("Betting")
		bet, err := strconv.Atoi(command[1])
		if err != nil {
			slog.Error("Got bad data from command", "command", command)
		}
		t.game.PlaceBet(t.game.GetPlayer(c.id), bet)
	case "deal":
		t.game.DealCards()
	case "hit":
		slog.Info("Hitting")
		t.game.Hit(t.game.GetPlayer(c.id))
	case "stay":
		t.game.Stay(t.game.GetPlayer(c.id))
		slog.Info("Standing")
	case "dealer":
		t.game.PlayDealer()
	}
}

func (t *Table) autoProgress() {
OuterLoop:
	for {
		switch t.game.State {
		case game.WAITING_FOR_BETS:
			slog.Info("WAITING FOR MORE BETS")
			if t.game.AllPlayersBet() {
				t.game.StartRound()
			} else {
				t.broadcastGameState()
				return
			}
		case game.DEALING:
			slog.Info("DEALING CARDS")
			t.game.DealCards()
		case game.DEALER_TURN:
			slog.Info("PLAYING DEALER")
			t.game.PlayDealer()
		case game.RESOLVING_BETS:
			slog.Info("RESOLVING BETS")
			t.game.ResolveBets()
			t.betTimer.Reset(30 * time.Second)
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
