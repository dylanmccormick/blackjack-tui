package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/dylanmccormick/blackjack-tui/game"
	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/dylanmccormick/blackjack-tui/store"
	"github.com/google/uuid"
)

const (
	ACTION_TIMEOUT = 30 * time.Second
	TABLE_TIMEOUT  = 5 * time.Minute
)

type inboundMessage struct {
	data   *protocol.TransportMessage
	client *Client
}

type Table struct {
	clients        map[*Client]bool
	registerChan   chan *Client
	unregisterChan chan *Client
	inbound        chan inboundMessage
	outbound       chan []byte
	id             string
	cancel         context.CancelFunc
	lobby          *Lobby

	maxPlayers  int
	game        *game.Game
	betTimer    *time.Timer
	actionTimer *time.Timer
	tableTimer  *time.Timer

	log *slog.Logger
	db  *store.Store
}

func newTable(name string, lobby *Lobby, store *store.Store) *Table {
	t := &Table{
		clients:        make(map[*Client]bool),
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		inbound:        make(chan inboundMessage),
		outbound:       make(chan []byte), // TODO: This should also be transport message type
		id:             name,
		game:           game.NewGame(),
		betTimer:       time.NewTimer(ACTION_TIMEOUT),
		actionTimer:    time.NewTimer(ACTION_TIMEOUT),
		tableTimer:     time.NewTimer(TABLE_TIMEOUT),
		lobby:          lobby,
		log:            slog.With("component", "table"),
		db:             store,
	}
	t.log = t.log.With("table_id", t.id)
	t.betTimer.Stop()
	t.actionTimer.Stop()
	t.tableTimer.Stop()
	return t
}

func (t *Table) register(c *Client) {
	t.registerChan <- c
}

func (t *Table) unregister(c *Client) {
	// this should always be "unintentional" because it comes from the readpump only
	t.unregisterChan <- c
}

func (t *Table) sendMessage(msg inboundMessage) {
	t.inbound <- msg
}

func (t *Table) cleanUp() {
	t.log.Info("Cleaning up table", "table_name", t.id)
	close(t.registerChan)
	close(t.unregisterChan)
	close(t.inbound)
	close(t.outbound)
}

func (t *Table) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			t.log.Info("Killing Table")
			t.cleanUp()
			return
		case client := <-t.registerChan:
			t.RegisterClient(client)
		case client := <-t.unregisterChan:
			t.UnregisterClient(client)
		case message := <-t.inbound:
			t.log.Debug("Received message", "message", message.data)
			t.handleCommand(message)
			t.autoProgress()
		case <-t.betTimer.C:
			t.log.Info("BET TIMER EXPIRED")
			err := t.game.StartRound()
			if err != nil {
			}
			t.autoProgress()
		case <-t.actionTimer.C:
			t.log.Info("ACTION TIMER EXPIRED")
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
			t.log.Info("KILLING TABLE")
			t.sendDeleteMsg()
			t.cleanUp()
			return
		}
	}
}

func (t *Table) sendDeleteMsg() {
	msg := protocol.PackageClientMessage(protocol.MsgDeleteTable, t.id)
	t.lobby.inbound <- inboundMessage{msg, &Client{}}
}

func (t *Table) handleCommand(msg inboundMessage) {
	switch msg.data.Type {
	case protocol.MsgStartGame:
		// TODO: add a check to make sure that the game hasn't already been started. You can spam this button to constantly reset the timer
		t.log.Info("Starting game")
		err := t.game.StartGame()
		if err != nil {
			t.log.Warn("Attempted to start the game after it has already been started")
			t.tableTimer.Reset(TABLE_TIMEOUT)
			return
		}
		t.betTimer.Reset(ACTION_TIMEOUT)
	case protocol.MsgGetState:
		t.log.Debug("Client requested game state")
		t.broadcastGameState()
	case protocol.MsgPlaceBet:
		value := protocol.ValueMessage{}
		err := json.Unmarshal(msg.data.Data, &value)
		if err != nil {
			t.log.Error("Got bad data from command", "command", msg.data)
		}
		bet, err := strconv.Atoi(value.Value)
		t.game.PlaceBet(t.game.GetPlayer(msg.client.id), bet)
	case protocol.MsgDealCards:
		t.game.DealCards()
	case protocol.MsgHit:
		t.log.Debug("Hitting", "client", msg.client.id)
		t.game.Hit(t.game.GetPlayer(msg.client.id))
		t.actionTimer.Reset(ACTION_TIMEOUT)
	case protocol.MsgStand:
		t.game.Stay(t.game.GetPlayer(msg.client.id))
		t.actionTimer.Reset(ACTION_TIMEOUT)
		t.log.Debug("Standing", "client", msg.client.id)
	case protocol.MsgLeaveTable:
		// intentionally left table
		// press ctrl+c or leave button
		t.cmdLeaveTable(msg.client)
	}
}

func (t *Table) DisconnectPlayer(c *Client, intentional bool) {
	player := t.game.GetPlayer(c.id)
	if player != nil {
		t.log.Info("Disconnecting player", "id", player.ID, "intentional?", intentional)
		player.MarkDisconnected(intentional)
	}
}

func (t *Table) cmdLeaveTable(c *Client) {
	t.DisconnectPlayer(c, true)
	t.lobby.register(c)
	c.mu.Lock()
	c.manager = t.lobby
	c.mu.Unlock()
	delete(t.clients, c)
}

func (t *Table) autoProgress() {
OuterLoop:
	for {
		switch t.game.State {
		case game.WAITING_FOR_BETS:
			t.log.Debug("WAITING FOR MORE BETS")
			if t.game.AllPlayersBet() {
				t.game.StartRound()
			} else {
				// if you don't do this there will be an infinite loop of WAITING_FOR_MORE_BETS checks
				t.broadcastGameState()
				return
			}
		case game.DEALING:
			t.log.Debug("dealing cards")
			t.game.DealCards()
			t.actionTimer.Reset(ACTION_TIMEOUT)
		case game.DEALER_TURN:
			t.log.Debug("PLAYING DEALER")
			t.game.PlayDealer()
		case game.RESOLVING_BETS:
			t.log.Debug("RESOLVING BETS")
			pmap, err := t.game.ResolveBets()
			if err != nil {
				slog.Error("Error in autoprogress. Unable to resolve bets", "error", err)
			}
			t.StoreGameData(pmap)
			t.betTimer.Reset(ACTION_TIMEOUT)
		default:
			t.broadcastGameState()
			break OuterLoop
		}
		t.broadcastGameState()
	}
}

func (t *Table) StoreGameData(results map[uuid.UUID]store.RoundResult) {
	slog.Info("STORING GAME DATA")
	tempMap := map[uuid.UUID]*Client{}
	for client := range t.clients {
		tempMap[client.id] = client
	}

	for playerId, result := range results {
		client, ok := tempMap[playerId]
		if !ok {
			slog.Error("player id not found in table clients", "id", playerId)
		}
		githubId := client.username
		err := t.db.RecordResult(context.TODO(), githubId, result)
		if err != nil {
			slog.Error("Unable to record results to db", "username", githubId, "result", result)
		}
	}
}

func (t *Table) broadcastGameState() {
	gameData := protocol.GameToDTO(t.game)
	wrapped, err := protocol.PackageMessage(gameData)
	if err != nil {
		t.log.Error("unable to package message", "error", err)
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

func (t *Table) RegisterClient(client *Client) {
	t.log.Info("attempting to register client", "client", client.id)
	playerReconnecting := t.game.GetPlayer(client.id) != nil
	if !playerReconnecting {
		user, err := t.db.GetOrCreateUser(client.username)
		if err != nil {
			slog.Error("error getting user", "username", client.username)
			// probably should crash here?
		}
		p := game.NewPlayer(client.id, int(user.Wallet))
		p.Name = client.username
		t.game.AddPlayer(p)
	}
	t.clients[client] = true
	if t.game.State == game.WAIT_FOR_START {
		t.game.State = game.WAITING_FOR_BETS
	}
}

func (t *Table) UnregisterClient(client *Client) {
	t.log.Info("attempting to unregister client", "client", client.id)
	t.DisconnectPlayer(client, false)
	if _, ok := t.clients[client]; ok {
		delete(t.clients, client)
		close(client.send)
	}
}
