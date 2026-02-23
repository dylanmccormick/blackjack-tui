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
	sessions    map[string]*Session
	log         *slog.Logger
	Commands    chan *SessionCmd
	gitClientId string
}

func NewSessionManager(gitClientID string) *SessionManager {
	return &SessionManager{
		sessions:    make(map[string]*Session),
		log:         slog.With("component", "sessionManager"),
		Commands:    make(chan *SessionCmd, 10),
		gitClientId: gitClientID,
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

func (sm *SessionManager) pollGit(s *Session) (bool, error) {
	auth := false
	client := &http.Client{Timeout: 20 * time.Second}

	grantType := fmt.Sprintf("urn:ietf:params:oauth:grant-type:%s", "device_code")

	data := map[string]string{"client_id": sm.gitClientId, "device_code": s.deviceCode, "grant_type": grantType}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	url := "https://github.com/login/oauth/access_token"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		sm.log.Error("Error sending request", "error", err)
		return false, err
	}
	defer resp.Body.Close()

	// 5. Read and handle the response as needed (similar to the GET example).
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sm.log.Error("Error reading response body", "error", err)
		return false, err
	}

	sm.log.Info("Poll git body", "body", string(body), "resp", resp, "status_code", resp.StatusCode, "status", resp.Status)

	var returnData struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
	}
	err = json.Unmarshal(body, &returnData)
	if err != nil {
		sm.log.Error("Error reading response", "error", err)
	}

	sm.log.Info("Response", "response", returnData)

	if returnData.Error == "" {
		auth = true
		sm.Commands <- &SessionCmd{
			Action:    "updateSession",
			SessionId: s.SessionId,

			Authenticated: true,
			GHToken:       returnData.AccessToken,
		}
	}

	return auth, nil
}

func (sm *SessionManager) CheckStarredStatus(context context.Context, s *Session) (bool, error) {
	client := &http.Client{Timeout: 20 * time.Second}

	sm.log.Info("Bearer token", "token", s.githubToken)

	url := "https://api.github.com/user/starred/dylanmccormick/blackjack-tui"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/vnd.github+json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.githubToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := client.Do(req)
	if err != nil {
		sm.log.Error("Error sending request: %s\n", "error", err)
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sm.log.Error("Error reading response body: %s\n", "error", err)
		return false, err
	}
	sm.log.Info("Response", "body", string(body))

	if resp.StatusCode == 204 {
		return true, nil
	}
	return false, nil
}

func (sm *SessionManager) UpdateUsername(ctx context.Context, s *Session) error {
	client := &http.Client{Timeout: 20 * time.Second}

	sm.log.Info("Bearer token", "token", s.githubToken)

	url := "https://api.github.com/user"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/vnd.github+json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.githubToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		sm.log.Error("Error sending request", "error", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		sm.log.Error("Error reading response body", "error", err)
		return err
	}

	var returnData struct {
		UserName string `json:"login"`
	}
	err = json.Unmarshal(body, &returnData)
	if err != nil {
		sm.log.Error("Error reading response", "error", err)
	}
	if resp.StatusCode == 200 {
		sm.Commands <- &SessionCmd{
			Action:    "updateSession",
			GHUserId:  returnData.UserName,
			SessionId: s.SessionId,
		}
	}
	if resp.StatusCode != 200 {
		sm.log.Warn("Error getting username from github")
		sm.log.Info("Response", "status", resp.StatusCode, "response body", resp.Body, "responseStatus", resp.Status)
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
