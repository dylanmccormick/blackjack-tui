package server

import (
	"testing"

	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/google/uuid"
)

func TestCreateTable(t *testing.T) {
	table := newTable("test_table")

	if len(table.clients) != 0 {
		t.Fatalf("No clients map created")
	}
}

func TestTableClientInteractions(t *testing.T) {
	table := newTable("test_table")

	id, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID. err=%#v", err)
	}
	client := &Client{
		send: make(chan *protocol.TransportMessage, 10),
		id:   id,
	}

	table.RegisterClient(client)

	if table.clients[client] != true {
		t.Fatalf("Client not registered properly.")
	}

	table.UnregisterClient(client)
	if _, ok := table.clients[client]; ok {
		t.Fatalf("Client not unregistered properly.")
	}
}

