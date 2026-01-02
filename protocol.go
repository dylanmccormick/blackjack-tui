package main

import "encoding/json"

// This is a wrapper for all message types between the server and the client
type TransportMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type GameStateMessage struct {
	// TODO
}

type GameActionMessage struct {
	// TODO
}

type ServerMessage struct{
	// TODO
}
