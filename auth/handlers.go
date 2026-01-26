package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

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
		SessionId:     SessionId,
		createdAt:     time.Now(),
		lastRequest:   time.Now(),
		authenticated: false,
	}
	// 2. send GH request
	err = sendDeviceRequest(&session)
	if err != nil {
		slog.Error("error in handler", "error", err)
		return nil
	}

	data := map[string]string{"session_id": session.SessionId, "user_code": session.userCode}

	err = WriteHttpResponse(w, 200, data)
	if err != nil {
		slog.Error("Error in loginHandler", "error", err)
		WriteHttpResponse(w, 500, map[string]string{"message": "InternalServerError"})
	}
	return &session
}

func WriteHttpResponse(w http.ResponseWriter, statusCode int, body any) error {
	response, err := json.Marshal(body)
	if err != nil {
		slog.Error("Error writing json", "error", err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(response)
	if err != nil {
		slog.Error("error writing to http", "error", err)
		return err
	}
	return nil
}
