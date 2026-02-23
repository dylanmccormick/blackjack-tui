package client

import (
	"bytes"
	"encoding/json"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

func ReceiveMessage(sub <-chan *protocol.TransportMessage) tea.Cmd {
	return func() tea.Msg {
		msg := <-sub
		switch msg.Type {
		case protocol.MsgGameState:
			body := &protocol.GameDTO{}
			err := json.Unmarshal(msg.Data, body)
			if err != nil {
				slog.Error("error unmarshalling body", "error", err)
			}
			return body
		case protocol.MsgTableList:
			body := []*protocol.TableDTO{}
			err := json.Unmarshal(msg.Data, &body)
			if err != nil {
				slog.Error("error unmarshalling body", "error", err)
			}
			return body
		case protocol.MsgPopUp:
			body := protocol.PopUpDTO{}
			err := json.Unmarshal(msg.Data, &body)
			if err != nil {
				slog.Error("error unmarshalling body", "error", err)
			}
			return body
		case protocol.MsgUserStats:
			body := protocol.StatsDTO{}
			err := json.Unmarshal(msg.Data, &body)
			if err != nil {
				slog.Error("error unmarshalling body", "error", err)
			}
			slog.Info("Parsed user stats", "body", body)
			return body
		}

		return nil
	}
}

func ParseTransportMessage(msg []byte) []*protocol.TransportMessage {
	// Parses a transport message. Returns the type and the packaged message
	messages := []*protocol.TransportMessage{}
	decoder := json.NewDecoder(bytes.NewReader(msg))
	for decoder.More() {
		var tm protocol.TransportMessage
		err := decoder.Decode(&tm)
		if err != nil {
			slog.Error("Unable to decode JSON", "error", err)
			return []*protocol.TransportMessage{}
		}
		messages = append(messages, &tm)
	}
	return messages
}
