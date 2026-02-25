package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dylanmccormick/blackjack-tui/internal/errors"
)

type SessionManager struct {
	sessions map[string]*Session
	log      *slog.Logger
	Commands chan *SessionCmd
	ghClient *GithubClient
}

func NewSessionManager(gitClientID string) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		log:      slog.With("component", "sessionManager"),
		Commands: make(chan *SessionCmd, 10),
		ghClient: NewGithubClient(gitClientID),
	}
}

type SessionCmd struct {
	Action       string
	SessionId    string
	ResponseChan chan *Session

	// Needed for different commands. Optional.
	LastUpdated   time.Time
	Authenticated bool
	GHUserId      string
	GHToken       string
	Session       *Session
}

func (sm *SessionManager) cleanup() {
	for _, session := range sm.sessions {
		if int(time.Since(session.lastRequest).Seconds()) > 900 {
			sm.log.Info("Removing session", "session_id", session.SessionId)
			delete(sm.sessions, session.SessionId)
		}
	}
	// this will delete sessions which haven't had a request in 15 minutes
}

func (sm *SessionManager) GetSession(id string) (*Session, error) {
	resp := make(chan *Session)
	cmd := &SessionCmd{Action: "get", ResponseChan: resp, SessionId: id}
	sm.Commands <- cmd
	session := <-resp
	if session == nil {
		return nil, &errors.NotFoundError{Resource: "session", ID: "id"}
	}
	return session, nil
}

func (sm *SessionManager) getSession(resp chan<- *Session, id string) {
	s, ok := sm.sessions[id]
	if !ok {
		sm.log.Warn("Session doesn't exist", "sessionId", id)
		resp <- nil
		return
	}
	resp <- s
}

type GHPollData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

func (sm *SessionManager) pollGit(s *Session) (bool, error) {
	resp, err := sm.ghClient.PollAccessToken(context.Background(), s.deviceCode)
	if err != nil {
		return false, err
	}

	sm.Commands <- &SessionCmd{
		Action:    "updateSession",
		SessionId: s.SessionId,

		Authenticated: true,
		GHToken:       resp,
	}

	return true, nil
}

func (sm *SessionManager) CheckStarredStatus(ctx context.Context, s *Session) (bool, error) {
	starred, err := sm.ghClient.CheckStarred(ctx, s.githubToken, "dylanmccormick/blackjack-tui")
	if err != nil {
		sm.log.Error("Error checking starred status", "error", err)
		return false, err
	}
	return starred, nil
}

func (sm *SessionManager) UpdateUsername(ctx context.Context, s *Session) error {
	username, err := sm.ghClient.GetUsername(ctx, s.githubToken)
	if err != nil {
		return err
	}
	if username == "" {
		sm.log.Warn("Error getting username from github")
		return nil
	}
	sm.Commands <- &SessionCmd{
		Action:    "updateSession",
		GHUserId:  username,
		SessionId: s.SessionId,
	}

	return nil
}

func (sm *SessionManager) PollSession(ctx context.Context, s *Session) {
	// polls the git endpoint at a 5 second interval until 15 minutes have passed or the session is authenticated
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// poll github
		authenticated, err := sm.pollGit(s)
		sm.log.Debug("response from pollGit", "response", authenticated)
		if err != nil {
			// do some garbage
			sm.log.Error("Github error", "error", err)
		}
		if authenticated {
			sm.log.Info("Stopping poll. got positive auth check")
			ticker.Stop()
			sm.Commands <- &SessionCmd{Action: "getUsername", Session: s}
			return
		}
	}
}

func (sm *SessionManager) AddSession(s *Session) {
	cmd := &SessionCmd{Action: "add", Session: s}
	sm.Commands <- cmd
}

func (sm *SessionManager) updateSession(c *SessionCmd) {
	s, ok := sm.sessions[c.SessionId]
	if !ok {
		return
	}
	sm.log.Info("Updating session before", "session", fmt.Sprint(s.String()))
	if c.Authenticated != false {
		s.Authenticated = c.Authenticated
	}
	if c.GHToken != "" {
		s.githubToken = c.GHToken
	}
	if c.GHUserId != "" {
		s.GithubUserId = c.GHUserId
	}
	sm.log.Info("Updating session after", "session", s.String())
}

func (sm *SessionManager) addSession(s *Session) {
	sm.log.Info("Adding session to session manager", "sessionID", s.SessionId)
	sm.sessions[s.SessionId] = s
}

func (sm *SessionManager) ProcessCommand(ctx context.Context, cmd *SessionCmd) {
	switch cmd.Action {
	case "get":
		sm.getSession(cmd.ResponseChan, cmd.SessionId)
	case "add":
		sm.addSession(cmd.Session)
		go sm.PollSession(ctx, cmd.Session)
	case "updateSession":
		sm.updateSession(cmd)
	case "getUsername":
		sm.UpdateUsername(ctx, cmd.Session)
	}
}

func (sm *SessionManager) Run(ctx context.Context) {
	// all of these methods run synchronously
	// This means there should not be any concurrent access to the session map
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case cmd := <-sm.Commands:
			sm.ProcessCommand(ctx, cmd)
		case <-ctx.Done():
			sm.log.Info("Cleaning up session manager")
			return
		case <-ticker.C:
			// cleanup if 15 minutes of inactivity
			sm.cleanup()
		}
	}
}
