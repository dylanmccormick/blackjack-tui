package server

import (
	"log/slog"

	"github.com/dylanmccormick/blackjack-tui/protocol"
)

func CreatePopUp(message, level string) *protocol.TransportMessage {
	msg := protocol.PopUpDTO{Message: message, Type: level}
	data, err := protocol.PackageMessage(msg)
	if err != nil {
		slog.Error("Unable to package popup message", "error", err)
		return nil
	}
	return data
}
