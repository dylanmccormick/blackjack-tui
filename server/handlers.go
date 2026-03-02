package server

import (
	"net/http"

	"github.com/dylanmccormick/blackjack-tui/auth"
)

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	s.Log.Debug("getting health status for request", "request_id", r.Context().Value("requestId"))
	auth.WriteHttpResponse(r.Context(), w, http.StatusOK, `{"message": "healthy"}`)
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	s.Log.Info("Received login request", "request_id", r.Context().Value("requestId"))
	session := auth.LoginHandler(w, r, s.SessionManager)
	s.SessionManager.AddSession(session)
}

func (s *Server) authStatusHandler(w http.ResponseWriter, r *http.Request) {
	s.Log.Debug("getting status for request", "request_id", r.Context().Value("requestId"))
	id := r.URL.Query().Get("id")
	auth.HandleAuthCheck(s.SessionManager, id, w, r)
}
