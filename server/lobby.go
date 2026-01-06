package server

import (
	"log/slog"
	"strings"

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
	stopChan       chan struct{}

	tables map[string]*Table
}

func newLobby() *Lobby {
	return &Lobby{
		clients:        make(map[*Client]bool),
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		inbound:        make(chan inboundMessage),
		outbound:       make(chan []byte),
		stopChan:       make(chan struct{}),
		tables:         make(map[string]*Table),
	}
}

func (l *Lobby) run() {
	for {
		select {
		case <-l.stopChan:
			slog.Info("Closing lobby")
			return
		case client := <-l.registerChan:
			l.clients[client] = true
		case client := <-l.unregisterChan:
			if _, ok := l.clients[client]; ok {
				delete(l.clients, client)
				close(client.send)
			}
		case msg := <-l.inbound:
			l.processMessage(msg)
		}
	}
}

func (l *Lobby) processMessage(msg inboundMessage) {
	// jsonMsg := &protocol.TransportMessage{}
	// json.Unmarshal(msg, jsonMsg)
	data := msg.data
	strMsg := string(data)
	parts := strings.Split(strMsg, " ")
	for _, p := range parts {
		slog.Info(p)
	}
	l.handleCommand(parts, msg.client)

	slog.Info("returning from processMessage")
}

func (l *Lobby) handleCommand(cmd []string, c *Client) {
	// join table, change username, get stats, etc
	slog.Info("lobby got command", "command", cmd)
	switch cmd[0] {
	case "create":
		if len(cmd) < 2 {
			slog.Error("Not enough arguments")
			return
		}
		slog.Info("Attempting to create table", "name", cmd[1])
		l.createTable(cmd[1])
	case "join":
		if len(cmd) < 2 {
			slog.Error("Not enough arguments")
			return
		}
		slog.Info("Attempting to join table", "name", cmd[1])
		l.joinTable(cmd[1], c)
	case "list":
		slog.Info("Listing Tables")
		l.listTables(c)
	}
}

func (l *Lobby) register(c *Client) {
	slog.Info("Registering client to lobby")
	l.registerChan <- c
}

func (l *Lobby) unregister(c *Client) {
	slog.Info("adding to lobby unregister chan")
	l.unregisterChan <- c
}

func (l *Lobby) sendMessage(msg inboundMessage) {
	l.inbound <- msg
}

func (l *Lobby) createTable(name string) {
	if _, ok := l.tables[name]; ok {
		slog.Warn("Table name already exists... not creating new table")
		return
	}
	t := newTable(name)
	l.tables[name] = t
	go t.run()
	for client := range l.clients {
		l.listTables(client)
	}
}

func (l *Lobby) deleteTable(name string) {
	if t, ok := l.tables[name]; ok {
		t.stopChan <- struct{}{}
		// this may have to do some cleanup. send everyone in the table back to the lobby
		delete(l.tables, name)
		return
	}
	slog.Warn("Table name doesn't exist. cannot delete anything")
}

func (l *Lobby) joinTable(name string, c *Client) {
	if t, ok := l.tables[name]; ok {
		t.register(c)
		c.manager = t
		delete(l.clients, c)
		return
	}
	slog.Warn("The table does not exist", "name", name)
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
		slog.Error("Unable to send list tables in lobby")
	}
	c.send <- data
}
