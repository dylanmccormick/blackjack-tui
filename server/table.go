package server

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/dylanmccormick/blackjack-tui/game"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type inboundMessage struct {
	data   []byte
	client *Client
}

type Table struct {
	clients        map[*Client]bool
	registerChan   chan *Client
	unregisterChan chan *Client
	inbound        chan inboundMessage
	outbound       chan []byte
	stopChan       chan struct{}
	id             string

	maxPlayers  int
	game        *game.Game
	betTimer    *time.Timer
	actionTimer *time.Timer
}

func newTable(name string) *Table {
	return &Table{
		clients:        make(map[*Client]bool),
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		inbound:        make(chan inboundMessage),
		outbound:       make(chan []byte),
		stopChan:       make(chan struct{}),
		id:             name,
		game:           game.NewGame(),
		betTimer:       time.NewTimer(30 * time.Second),
		actionTimer:    time.NewTimer(30 * time.Second),
	}
}

func (t *Table) register(c *Client) {
	t.registerChan <- c
}

func (t *Table) unregister(c *Client) {
	slog.Info("Unregistering client")
	t.unregisterChan <- c
}

func (t *Table) sendMessage(msg inboundMessage) {
	t.inbound <- msg
}

func (t *Table) run() {
	for {
		select {
		case <-t.stopChan:
			slog.Info("Killing Table")
			return
		case client := <-t.registerChan:
			slog.Info("attempting to register client", "client", client.id)
			t.clients[client] = true
			p := game.NewPlayer(client.id)
			p.Name = "lugubrious_bagel"
			t.game.AddPlayer(p)
		case client := <-t.unregisterChan:
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
			slog.Info("BET TIMER EXPIRED")
			t.game.StartRound()
			t.autoProgress()
		case <-t.actionTimer.C:
			slog.Info("ACTION TIMER EXPIRED")
			t.game.Stay(t.game.CurrentPlayer())
			if t.game.State == game.DEALER_TURN {
				t.autoProgress()
			}
		}
	}
}

func unpackMessage(msg inboundMessage) (*protocol.TransportMessage, error) {
	jsonMsg := &protocol.TransportMessage{}
	err := json.Unmarshal(msg.data, jsonMsg)
	if err != nil {
		return &protocol.TransportMessage{}, err
	}
	return jsonMsg, nil
}

func (t *Table) processMessage(msg inboundMessage) {
	tranMsg, err := unpackMessage(msg)
	if err != nil {
		slog.Error("unable to unpack transport message", err)
		return
	}
	t.handleCommand(tranMsg, msg.client)
}

func (t *Table) handleCommand(command *protocol.TransportMessage, c *Client) {
	switch command.Type {
	case protocol.MsgStartGame:
		slog.Info("Starting game")
		t.betTimer.Reset(30 * time.Second)
	case protocol.MsgPlaceBet:
		slog.Info("Betting")
		value := protocol.ValueMessage{}
		err := json.Unmarshal(command.Data, &value)
		if err != nil {
			slog.Error("Got bad data from command", "command", command)
		}
		bet, err := strconv.Atoi(value.Value)
		t.game.PlaceBet(t.game.GetPlayer(c.id), bet)
	case protocol.MsgDealCards:
		t.game.DealCards()
	case protocol.MsgHit:
		slog.Info("Hitting")
		t.game.Hit(t.game.GetPlayer(c.id))
		t.actionTimer.Reset(30 * time.Second)
	case protocol.MsgStand:
		t.game.Stay(t.game.GetPlayer(c.id))
		slog.Info("Standing")
	case protocol.MsgLeaveTable:
		lobby.register(c)
		c.manager = lobby
		delete(t.clients, c)
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
				// if you don't do this there will be an infinite loop of WAITING_FOR_MORE_BETS checks
				t.broadcastGameState()
				return
			}
		case game.DEALING:
			slog.Info("DEALING CARDS")
			t.game.DealCards()
			t.actionTimer.Reset(30 * time.Second)
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
	wrapped, err := protocol.PackageMessage(gameData)
	if err != nil {
		slog.Error("unable to package message", "error", err)
		return
	}
	for client := range t.clients {
		select {
		case client.send <- wrapped:
		default:
			close(client.send)
			delete(t.clients, client)
		}
	}
}

func (t *Table) Id() string {
	return t.id
}

func (t *Table) CreateDTO() protocol.TableDTO {
	return protocol.TableDTO{
		Id:             t.id,
		Capacity:       t.maxPlayers,
		CurrentPlayers: len(t.clients),
	}
}
