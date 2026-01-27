package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"
)

var clientId = os.Getenv("GIT_CLIENT_ID")

type Session struct {
	SessionId     string
	deviceCode    string
	userCode      string
	githubToken   string
	githubUserId  string
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
	sb.WriteString("userId: " + s.githubUserId)
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

	data := map[string]string{"authenticated": fmt.Sprintf("%v", session.Authenticated)}
	err = WriteHttpResponse(w, 200, data)
	if err != nil {
		slog.Error("Error in HandleAuthCheck", "error", err)
		WriteHttpResponse(w, 500, map[string]string{"message": "InternalServerError"})
	}
	return session.Authenticated
}

func sendDeviceRequest(session *Session) error {
	client := &http.Client{Timeout: 20 * time.Second}

	data := map[string]string{"client_id": clientId}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	url := "https://github.com/login/device/code"
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

	slog.Info("URL", "uri", returnData.VerificationUri)

	session.userCode = returnData.UserCode
	session.deviceCode = returnData.DeviceCode
	session.createdAt = time.Now()
	return nil
}

func RunTest() {
	sm := NewSessionManager()
	go sm.Run(context.TODO())

	session := &Session{}
	session.SessionId = "1234"
	slog.Info("Sending request")
	sendDeviceRequest(session)
	slog.Info("sessionInfo", "session", session)
	sm.AddSession(session)

	ticker := time.NewTicker(20 * time.Second)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	for range ticker.C {
		if HandleAuthCheck(sm, session.SessionId, w, r) {
			ticker.Stop()
			fmt.Printf("%#v\n", session)
			fmt.Println("AUTHENTICATED BEOTCH")
			break
		}
	}
}
