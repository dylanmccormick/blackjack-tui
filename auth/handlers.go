package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func LoginHandler(w http.ResponseWriter, r *http.Request, sm *SessionManager) *Session {
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
		Authenticated: false,
	}
	// 2. send GH request
	err = sm.sendDeviceRequest(r.Context(), &session)
	if err != nil {
		slog.Error("error in handler", "error", err)
		return nil
	}

	data := map[string]string{"session_id": session.SessionId, "user_code": session.userCode}

	err = WriteHttpResponse(r.Context(), w, 200, data)
	if err != nil {
		slog.Error("Error in loginHandler", "error", err, "request_id", r.Context().Value("requestId"))
		WriteHttpResponse(r.Context(), w, 500, map[string]string{"message": "InternalServerError"})
	}
	return &session
}

func WriteHttpResponse(ctx context.Context, w http.ResponseWriter, statusCode int, body any) error {
	response, err := json.Marshal(body)
	if err != nil {
		slog.Error("Error writing json", "error", err, "request_id", ctx.Value("requestId"))
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
