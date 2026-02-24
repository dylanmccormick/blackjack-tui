package server

import (
	"net/http"

	"github.com/dylanmccormick/blackjack-tui/auth"
)

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	auth.WriteHttpResponse(w, http.StatusOK, `{"message": "healthy"}`)
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("Received login request")
	session := auth.LoginHandler(w, r, s.SessionManager)
	s.SessionManager.AddSession(session)
}

func (s *Server) authStatusHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("getting status for request")
	id := r.URL.Query().Get("id")
	auth.HandleAuthCheck(s.SessionManager, id, w, r)
}
