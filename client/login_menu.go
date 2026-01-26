package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type loginMenuPage int

const (
	userCodePage     loginMenuPage = iota
	confirmationPage loginMenuPage = iota
)

type LoginMenu struct {
	page      loginMenuPage
	userCode  string
	sessionId string
}

func (lm *LoginMenu) View() string {
	switch lm.page {
	case userCodePage:
		return lm.viewUserCodePage()
	case confirmationPage:
		return lm.viewConfirmationPage()
	}
	return ""
}

func (lm *LoginMenu) viewConfirmationPage() string {
	return "YOU'RE LOGGED IN"
}

func (lm *LoginMenu) viewUserCodePage() string {
	sb := strings.Builder{}
	sb.WriteString("UserCode: ")
	sb.WriteString(lm.userCode)
	sb.WriteString("\n")

	sb.WriteString("Go to https://github.com/device/login and enter the above code\n")
	return sb.String()
}

func (lm *LoginMenu) Update(msg tea.Msg) (*LoginMenu, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case AuthLoginMsg:
		lm.userCode = msg.UserCode
		lm.sessionId = msg.SessionId
		lm.page = userCodePage
	case AuthPollMsg:
		if msg.Authenticated {
			lm.page = confirmationPage
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			cmd = AuthLoginCmd()
			cmds = append(cmds, cmd)
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "o":
				cmd = AuthPollAccess(lm.sessionId)
				cmds = append(cmds, cmd)
			}
		}
	}

	return lm, tea.Batch(cmds...)
}

func (lm *LoginMenu) Init() tea.Cmd {
	return AuthLoginCmd()
}

type AuthLoginMsg struct {
	UserCode  string
	Url       string
	SessionId string
}

type AuthPollMsg struct {
	Authenticated bool
}

func AuthLoginCmd() tea.Cmd {
	return func() tea.Msg {
		return SendLoginRequest()
	}
}

func AuthPollAccess(sessionId string) tea.Cmd {
	return func() tea.Msg {
		return CheckAccess(sessionId)
	}
}

func CheckAccess(sessionId string) AuthPollMsg {
	client := &http.Client{Timeout: 20 * time.Second}

	url := "http://localhost:8080/auth/status"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Debug("error sending request", "error", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	q := req.URL.Query()
	q.Add("id", sessionId)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
	}

	slog.Info("reading body", "body", body)
	var data struct {
		Authenticated string `json:"authenticated"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		slog.Error("error loading json", "error", err)
	}

	if data.Authenticated == "true" {
		return AuthPollMsg{true}
	}
	return AuthPollMsg{false}
}

func SendLoginRequest() AuthLoginMsg {
	client := &http.Client{Timeout: 20 * time.Second}

	url := "http://localhost:8080/auth"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Debug("error sending request", "error", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
	}

	slog.Info("reading body", "body", body)
	var data struct {
		SessionId string `json:"session_id"`
		UserCode  string `json:"user_code"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		slog.Error("error loading json", "error", err)
	}

	return AuthLoginMsg{
		UserCode:  data.UserCode,
		Url:       "https://github.com/device/login",
		SessionId: data.SessionId,
	}
}
