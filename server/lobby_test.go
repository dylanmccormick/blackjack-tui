package server

import (
	"testing"

	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/google/uuid"
)

func TestCreateLobby(t *testing.T) {
	lobby := newLobby()

	if len(lobby.clients) != 0 {
		t.Fatalf("No clients map created")
	}
}

func TestRegisterClient(t *testing.T) {
	lobby := newLobby()
	id, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID. err=%#v", err)
	}
	client := &Client{
		send: make(chan *protocol.TransportMessage, 10),
		id:   id,
	}

	lobby.RegisterClient(client)
	if lobby.clients[client] != true {
		t.Fatalf("Client not registered properly")
	}

	lobby.UnregisterClient(client)
	if lobby.clients[client] != false {
		t.Fatalf("Client not unregistered properly")
	}
}

func TestAddTable(t *testing.T) {
	lobby := newLobby()
	lobby.createTable("test_table")
	if len(lobby.tables) != 1 {
		t.Fatalf("expected lobby to have 1 table. got=%d", len(lobby.tables))
	}
	lobby.createTable("test_table_2")
	if len(lobby.tables) != 2 {
		t.Fatalf("expected lobby to have 2 tables. got=%d", len(lobby.tables))
	}
}

func TestListTable(t *testing.T) {
	clients := clientHelper(2)
	c0 := clients[0]
	c1 := clients[1]
	lobby := newLobby()
	lobby.createTable("test")
	lobby.RegisterClient(c0)
	lobby.RegisterClient(c1)
	lobby.listTables(c0)
	if len(c0.send) != 1 {
		t.Errorf("Expected a message in send channel for client got=%d", len(c0.send))
	}
	if len(c1.send) > 0 {
		t.Errorf("Expected no message in send channel for client")
	}
}

func TestDeleteTable(t *testing.T) {
	lobby := newLobby()
	lobby.createTable("test_table")
	lobby.createTable("test_table_2")
	test_table := lobby.tables["test_table"]
	lobby.deleteTable("test_table")
	_, ok := <-test_table.stopChan
	if ok {
		t.Fatalf("Stopchan should be closed")
	}
}

func clientHelper(n int) []*Client {
	var clients []*Client
	for range n {
		id, err := uuid.NewUUID()
		if err != nil {
			panic(err)
		}
		client := &Client{
			send: make(chan *protocol.TransportMessage, 10),
			id:   id,
		}
		clients = append(clients, client)
	}
	return clients
}
