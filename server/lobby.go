package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/dylanmccormick/blackjack-tui/protocol"
)

// The lobby will be the landing zone for any new connections to the game.
// Players will be able to update their username, choose a table, and do whatever else they need to do

type Lobby struct {
	clients        map[*Client]bool
	registerChan   chan *Client
	unregisterChan chan *Client
	inbound        chan inboundMessage
	outbound       chan []byte
	tables         map[string]*Table
	tableWg        sync.WaitGroup
	log            *slog.Logger
}

func newLobby() *Lobby {
	return &Lobby{
		clients:        make(map[*Client]bool),
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		inbound:        make(chan inboundMessage),
		outbound:       make(chan []byte),
		tables:         make(map[string]*Table),
		log:            slog.With("component", "lobby"),
	}
}

func (l *Lobby) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			l.log.Info("Lobby Shutting Down")
			l.shutdownTables()
			return
		case client := <-l.registerChan:
			l.RegisterClient(client)
		case client := <-l.unregisterChan:
			l.UnregisterClient(client)
		case msg := <-l.inbound:
			l.handleCommand(ctx, msg)
		}
	}
}

func (l *Lobby) shutdownTables() {
	for _, table := range l.tables {
		table.cancel()
	}
	l.tableWg.Wait()
}

func (l *Lobby) UnregisterClient(client *Client) {
	l.log.Info("Unregistering client", "client", client.id)
	delete(l.clients, client)
	close(client.send)
}

func getValueFromRawValueMessage(raw json.RawMessage) (string, error) {
	value := protocol.ValueMessage{}
	err := json.Unmarshal(raw, &value)
	if err != nil {
		slog.Error("Got bad data from value in transport message", "raw", string(raw))
	}
	return value.Value, nil
}

func (l *Lobby) handleCommand(ctx context.Context, msg inboundMessage) {
	// join table, change username, get stats, etc
	l.log.Debug("lobby got command", "command", msg.data)
	switch msg.data.Type {
	case protocol.MsgCreateTable:
		val, err := getValueFromRawValueMessage(msg.data.Data)
		if err != nil {
			return
		}
		l.log.Info("Attempting to create table", "name", val)
		l.createTable(ctx, val)
	case protocol.MsgJoinTable:
		val, err := getValueFromRawValueMessage(msg.data.Data)
		if err != nil {
			return
		}
		l.log.Info("Attempting to join table", "name", val, "client", msg.client)
		l.joinTable(val, msg.client)
	case protocol.MsgTableList:
		l.log.Debug("Listing Tables")
		l.listTables(msg.client)

	case protocol.MsgDeleteTable:
		val, err := getValueFromRawValueMessage(msg.data.Data)
		if err != nil {
			return
		}
		l.log.Info("Attempting to delete table", "name", val)
		l.deleteTable(val)
	}
}

func (l *Lobby) RegisterClient(client *Client) {
	l.log.Info("registering client", "client", client.id)
	l.clients[client] = true
}

func (l *Lobby) register(c *Client) {
	l.log.Debug("Registering client to lobby", "client", c.id)
	l.registerChan <- c
}

func (l *Lobby) unregister(c *Client) {
	l.log.Debug("adding to lobby unregister chan", "client", c.id)
	l.unregisterChan <- c
}

func (l *Lobby) sendMessage(msg inboundMessage) {
	l.inbound <- msg
}

func (l *Lobby) createTable(ctx context.Context, name string) {
	if _, ok := l.tables[name]; ok {
		l.log.Warn("Table name already exists... not creating new table")
		return
	}
	t := newTable(name, l)
	tableCtx, tableCancel := context.WithCancel(ctx)
	t.cancel = tableCancel
	l.tables[name] = t
	l.tableWg.Go(func() {
		t.run(tableCtx)
	})
	for client := range l.clients {
		l.listTables(client)
	}
}

func (l *Lobby) deleteTable(name string) {
	if t, ok := l.tables[name]; ok {
		t.cancel()
		// this may have to do some cleanup. send everyone in the table back to the lobby
		delete(l.tables, name)
		return
	}
	l.log.Warn("Table name doesn't exist. cannot delete anything")
}

func (l *Lobby) joinTable(name string, c *Client) {
	if t, ok := l.tables[name]; ok {
		t.register(c)
		c.mu.Lock()
		c.manager = t
		c.mu.Unlock()
		delete(l.clients, c)
		return
	}
	l.log.Warn("The table does not exist", "name", name)
}

func (l *Lobby) Id() string {
	return "lobby"
}

func (l *Lobby) listTables(c *Client) {
	out := []protocol.TableDTO{}
	for _, t := range l.tables {
		out = append(out, t.CreateDTO())
	}
	data, err := protocol.PackageMessage(out)
	if err != nil {
		l.log.Error("Unable to send list tables in lobby")
	}
	c.send <- data
}
