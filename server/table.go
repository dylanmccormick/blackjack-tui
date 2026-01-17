package server

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/dylanmccormick/blackjack-tui/game"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

const (
	ACTION_TIMEOUT = 30 * time.Second
	TABLE_TIMEOUT  = 5 * time.Minute
)

type inboundMessage struct {
	data *protocol.TransportMessage
	// data   []byte // TODO: This should be transport message type
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
	tableTimer  *time.Timer
}

func newTable(name string) *Table {
	t := &Table{
		clients:        make(map[*Client]bool),
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		inbound:        make(chan inboundMessage),
		outbound:       make(chan []byte), // TODO: This should also be transport message type
		stopChan:       make(chan struct{}),
		id:             name,
		game:           game.NewGame(),
		betTimer:       time.NewTimer(ACTION_TIMEOUT),
		actionTimer:    time.NewTimer(ACTION_TIMEOUT),
		tableTimer:     time.NewTimer(TABLE_TIMEOUT),
	}
	t.betTimer.Stop()
	t.actionTimer.Stop()
	t.tableTimer.Stop()
	return t
}

func (t *Table) register(c *Client) {
	t.registerChan <- c
}

func (t *Table) unregister(c *Client) {
	slog.Info("Unregistering client")
	// this should always be "unintentional" because it comes from the readpump only
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
			playerReconnecting := t.game.GetPlayer(client.id) != nil
			if !playerReconnecting {
				p := game.NewPlayer(client.id)
				p.Name = "lugubrious_bagel"
				t.game.AddPlayer(p)
			}
			t.clients[client] = true
			if t.game.State == game.WAIT_FOR_START {
				t.game.State = game.WAITING_FOR_BETS
			}
		case client := <-t.unregisterChan:
			// Unintentionally Left Table (connection issue or pkill or whatever)
			slog.Info("attempting to unregister client", "client", client.id)
			t.DisconnectPlayer(client, false)
			if _, ok := t.clients[client]; ok {
				delete(t.clients, client)
				close(client.send)
			}
		case message := <-t.inbound:
			slog.Info("Received message", "message", message.data)
			t.handleCommand(message)
			t.autoProgress()
		case <-t.betTimer.C:
			slog.Info("BET TIMER EXPIRED")
			err := t.game.StartRound()
			if err != nil {
			}
			t.autoProgress()
		case <-t.actionTimer.C:
			slog.Info("ACTION TIMER EXPIRED")
			if t.game.State != game.PLAYER_TURN {
				// we don't need to reset anything if there are no actions to be waited for. i.e. the table is dead
				continue
			}
			t.game.Stay(t.game.CurrentPlayer())
			t.actionTimer.Reset(ACTION_TIMEOUT)
			if t.game.State == game.DEALER_TURN {
				t.autoProgress()
			}
		case <-t.tableTimer.C:
			slog.Info("KILLING TABLE")
			lobby.deleteTable(t.id)
			// TODO: figure out how to kill the table
		}
	}
}

// the read pump should be handling this

func (t *Table) handleCommand(msg inboundMessage) {
	switch msg.data.Type {
	case protocol.MsgStartGame:
		// TODO: add a check to make sure that the game hasn't already been started. You can spam this button to constantly reset the timer
		slog.Info("Starting game")
		err := t.game.StartGame()
		if err != nil {
			slog.Warn("Attempted to start the game after it has already been started")
			t.tableTimer.Reset(TABLE_TIMEOUT)
			return
		}
		t.betTimer.Reset(ACTION_TIMEOUT)
	case protocol.MsgPlaceBet:
		slog.Info("Betting")
		value := protocol.ValueMessage{}
		err := json.Unmarshal(msg.data.Data, &value)
		if err != nil {
			slog.Error("Got bad data from command", "command", msg.data)
		}
		bet, err := strconv.Atoi(value.Value)
		t.game.PlaceBet(t.game.GetPlayer(msg.client.id), bet)
	case protocol.MsgDealCards:
		t.game.DealCards()
	case protocol.MsgHit:
		slog.Info("Hitting")
		t.game.Hit(t.game.GetPlayer(msg.client.id))
		t.actionTimer.Reset(ACTION_TIMEOUT)
	case protocol.MsgStand:
		t.game.Stay(t.game.GetPlayer(msg.client.id))
		t.actionTimer.Reset(ACTION_TIMEOUT)
		slog.Info("Standing")
	case protocol.MsgLeaveTable:
		// intentionally left table
		// press ctrl+c or leave button
		t.cmdLeaveTable(msg.client)
	}
}

func (t *Table) DisconnectPlayer(c *Client, intentional bool) {
	player := t.game.GetPlayer(c.id)
	if player != nil {
		slog.Info("Disconnecting player", "id", player.ID, "intentional?", intentional)
		player.MarkDisconnected(intentional)
	}
}

func (t *Table) cmdLeaveTable(c *Client) {
	t.DisconnectPlayer(c, true)
	lobby.register(c)
	c.manager = lobby
	delete(t.clients, c)
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
			t.actionTimer.Reset(ACTION_TIMEOUT)
		case game.DEALER_TURN:
			slog.Info("PLAYING DEALER")
			t.game.PlayDealer()
		case game.RESOLVING_BETS:
			slog.Info("RESOLVING BETS")
			t.game.ResolveBets()
			t.betTimer.Reset(ACTION_TIMEOUT)
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
