package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Session struct {
	SessionId     string
	deviceCode    string
	userCode      string
	githubToken   string
	GithubUserId  string
	lastRequest   time.Time
	Authenticated bool
	createdAt     time.Time
}

func (s *Session) String() string {
	sb := strings.Builder{}
	sb.WriteString("SessionId: " + s.SessionId)
	fmt.Fprint(&sb, "\n\t")
	sb.WriteString("auth: " + strconv.FormatBool(s.Authenticated))
	fmt.Fprint(&sb, "\n\t")
	sb.WriteString("userId: " + s.GithubUserId)
	return sb.String()
}

type GHDeviceResponse struct {
	UserCode        string `json:"user_code"`
	DeviceCode      string `json:"device_code"`
	VerificationUri string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func HandleAuthCheck(sm *SessionManager, id string, w http.ResponseWriter, r *http.Request) bool {
	session, err := sm.GetSession(id)
	if err != nil {
		data := map[string]string{"message": "session not found"}
		err = WriteHttpResponse(w, 404, data)
		if err != nil {
			slog.Error("Error in HandleAuthCheck", "error", err)
			WriteHttpResponse(w, 500, map[string]string{"message": "InternalServerError"})
		}
		return false
	}

	data := map[string]string{"authenticated": fmt.Sprintf("%v", session.Authenticated), "username": session.GithubUserId}
	err = WriteHttpResponse(w, 200, data)
	if err != nil {
		slog.Error("Error in HandleAuthCheck", "error", err)
		WriteHttpResponse(w, 500, map[string]string{"message": "InternalServerError"})
	}
	return session.Authenticated
}

func (sm *SessionManager) sendDeviceRequest(session *Session) error {
	resp, err := sm.ghClient.RequestDeviceCode(context.Background())
	if err != nil {
		return err
	}

	session.userCode = resp.UserCode
	session.deviceCode = resp.DeviceCode
	session.createdAt = time.Now()
	return nil
}
