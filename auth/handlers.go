package auth

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) *Session {
	// 1. create session id
	sessionUUID, err := uuid.NewUUID()
	if err != nil {
		slog.Error("Unable to create UUID", "error", err)
	}
	SessionId := sessionUUID.String()
	session := Session{
		SessionId: SessionId,
	}
	// 2. send GH request
	err = sendDeviceRequest(&session)
	if err != nil {
		slog.Error("error in handler", "error", err)
		return nil
	}
	// 3. Write session info to response
	// 4. return something to server so it can save sessions
	return &session
}
