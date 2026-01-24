package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type SessionManager struct {
	sessions map[string]*Session
	log      *slog.Logger
	Commands chan *SessionCmd
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		log:      slog.With("component", "sessionManager"),
		Commands: make(chan *SessionCmd, 10),
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
			slog.Info("Removing session", "session_id", session.SessionId)
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
		return nil, fmt.Errorf("Session %s not found", id)
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

func (sm *SessionManager) pollGit(s *Session) error {
	client := &http.Client{Timeout: 20 * time.Second}

	grantType := fmt.Sprintf("urn:ietf:params:oauth:grant-type:%s", "device_code")

	data := map[string]string{"client_id": clientId, "device_code": s.deviceCode, "grant_type": grantType}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	url := "https://github.com/login/oauth/access_token"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return err
	}
	defer resp.Body.Close()

	// 5. Read and handle the response as needed (similar to the GET example).
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
		return err
	}
	var returnData GHDeviceResponse
	err = json.Unmarshal(body, &returnData)
	if err != nil {
		fmt.Printf("Error reading response, err:%#v\n", err)
	}

	fmt.Println(string(body))

	return nil
}

func (sm *SessionManager) PollSession(ctx context.Context, s *Session) {
	// polls the git endpoint at a 5 second interval until 15 minutes have passed or the session is authenticated
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// poll github
		err := sm.pollGit(s)
		if err != nil {
			// do some garbage
			sm.log.Error("Github error", "error", err)
		}

		// if authorized. update session and return
		sm.Commands <- &SessionCmd{
			Action:    "updateSession",
			SessionId: s.SessionId,

			Authenticated: true,
			GHUserId:      "#TODO",
			GHToken:       "#TODO",
		}
		// if past 15 minutes return
	}
}

func (sm *SessionManager) AddSession(s *Session) {
	cmd := &SessionCmd{Action: "add", Session: s}
	sm.Commands <- cmd
}

func (sm *SessionManager) UpdateSession(s *Session) {
	// TODO: refactor so we do our own commands. Maybe we don't need the UpdateSession method? not sure
	cmd := &SessionCmd{Action: "updateSession", Session: s}
	sm.Commands <- cmd
}

func (sm *SessionManager) updateSession(c *SessionCmd) {
	s, ok := sm.sessions[c.SessionId]
	if !ok {
		return
	}
	s.authenticated = c.Authenticated
	s.githubToken = c.GHToken
	s.githubUserId = c.GHUserId
}

func (sm *SessionManager) addSession(s *Session) {
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
