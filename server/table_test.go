package server

import (
	"encoding/json"
	"testing"

	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/dylanmccormick/blackjack-tui/store"
	"github.com/google/uuid"
)

func TestCreateTable(t *testing.T) {
	store, _ := store.NewStore(":memory:", "../sql/schema")
	lobby := newLobby(store)
	table := newTable("test_table", lobby, store)

	if len(table.clients) != 0 {
		t.Fatalf("No clients map created")
	}
}

func TestTableClientInteractions(t *testing.T) {
	store, _ := store.NewStore(":memory:", "../sql/schema")
	lobby := newLobby(store)
	table := newTable("test_table", lobby, store)

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

func TestAutoProgress(t *testing.T) {
	store, _ := store.NewStore(":memory:", "../sql/schema")
	lobby := newLobby(store)
	tab := newTable("test_table", lobby, store)
	client := clientHelper(1)[0]
	tab.RegisterClient(client)
	p := tab.game.GetPlayer(client.id)
	tab.game.StartGame()
	tab.game.PlaceBet(p, 5)
	tab.autoProgress()
	tab.game.Stay(p)
	tab.autoProgress()
	count := 0
	close(client.send)
	for msg := range client.send {
		var msgData protocol.GameDTO
		json.Unmarshal(msg.Data, &msgData)
		count += 1
		t.Logf("count: %d, msg: %#v\n", count, msgData)
	}
}
