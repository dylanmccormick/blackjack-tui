package main

import (
	"testing"
	"time"
)

func TestCreateTable(t *testing.T) {
	table := newTable()

	if len(table.clients) != 0 {
		t.Fatalf("No clients map created")
	}
}

func TestTableClientInteractions(t *testing.T) {
	table := newTable()

	go table.run()
	client := &Client{
		table: table,
		send:  make(chan []byte, 10),
		id:    1,
	}

	table.register <- client
	if table.clients[client] != true {
		t.Fatalf("Client not registered properly.")
	}

	table.unregister <- client
	time.Sleep(1 * time.Second)
	if _, ok := table.clients[client]; ok {
		t.Fatalf("Client not unregistered properly.")
	}

	close(table.stopChan)
}
