package main

import "log/slog"

type Table struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	inbound    chan []byte
	outbound   chan []byte
	stopChan   chan struct{}
	id         string
}

func newTable() *Table {
	return &Table{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		inbound:    make(chan []byte),
		outbound:   make(chan []byte),
		stopChan:   make(chan struct{}),
		id:         "placeholder",
	}
}

func (h *Table) run() {
	for {
		select {
		case <-h.stopChan:
			slog.Info("Killing Table")
			return
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			slog.Info("attempting to unregister client", "client", client.id)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.inbound:
			slog.Info("Received message", "message", string(message))
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
