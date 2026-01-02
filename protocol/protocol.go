package protocol

import (
	"encoding/json"
)

// This is a wrapper for all message types between the server and the client
type TransportMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type CommandMessage struct {
	Command string `json:"type"`
	Target  string `json:"target"`
}
