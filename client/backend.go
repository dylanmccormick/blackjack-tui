package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/gorilla/websocket"
)

func NewMockTransporter() *MockBackendClient {
	return &MockBackendClient{
		wsOut:      make(chan *protocol.TransportMessage),
		disconnect: make(chan struct{}),
		data:       make(chan *protocol.TransportMessage),
	}
}

func NewWsBackendClient() *WsBackendClient {
	return &WsBackendClient{
		serverUrl:  &url.URL{},
		wsOut:      make(chan *protocol.TransportMessage),
		disconnect: make(chan struct{}),
		data:       make(chan *protocol.TransportMessage),
		mut:        sync.Mutex{},
	}
}

type BackendClient interface {
	GetChan() chan *protocol.TransportMessage
	Connect() error // I think later we'll add an address you can connect to as a param
	Stop()          // Stops fetch data goroutine and disconnects from server.
	SendData()      // Reads from data chan sends JSON data across the wire to the server
	FetchData()     // Runs goroutine to pull data from the server connection
	QueueData(*protocol.TransportMessage)

	// HTTP Methods
	StartAuth(url string) tea.Msg
	PollAuth() tea.Msg
}

type WsBackendClient struct {
	mut        sync.Mutex
	serverUrl  *url.URL
	wsOut      chan *protocol.TransportMessage
	conn       *websocket.Conn
	data       chan *protocol.TransportMessage
	disconnect chan struct{}
	sessionId  string
}

func (ws *WsBackendClient) StartAuth(u string) tea.Msg {
	// http://localhost:8080
	sUrl, err := url.Parse(u)
	if err != nil {
		slog.Error("Url parsing failed", "url", u, "error", err)
	}
	ws.serverUrl = sUrl
	ws.serverUrl.Path = "/"
	ws.serverUrl.Scheme = "http"

	endpoint := "auth"
	fullURL := ws.serverUrl.JoinPath(endpoint)
	if err != nil {
		slog.Debug("error creating url path", "error", err)
		return fmt.Errorf("")
	}
	slog.Info("Attmepting to connect to fullUrl", "url", fullURL.String(), "serverUrl", ws.serverUrl, "endpoint", endpoint)

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("GET", fullURL.String(), nil)
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

	ws.sessionId = data.SessionId
	return AuthLoginMsg{
		UserCode:  data.UserCode,
		Url:       "https://github.com/device/login",
		SessionId: data.SessionId,
	}
}

func (ws *WsBackendClient) PollAuth() tea.Msg {
	fullURL, err := url.JoinPath(ws.serverUrl.String(), "auth", "status")
	if err != nil {
		slog.Debug("error creating url path", "error", err)
		return fmt.Errorf("")
	}
	ticker := time.NewTicker(1 * time.Second)
	client := &http.Client{Timeout: 20 * time.Second}
	for range ticker.C {

		slog.Info("Full url", "url", fullURL)
		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			slog.Debug("error sending request", "error", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Session-Id", ws.sessionId)
		q := req.URL.Query()
		q.Add("id", ws.sessionId)
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
		slog.Info("Reading poll auth:", "body", body)

		var data struct {
			Authenticated string `json:"authenticated"`
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			slog.Error("error loading json", "error", err)
		}

		if data.Authenticated == "true" {
			ticker.Stop()
			return AuthPollMsg{true}
		}
	}
	return AuthPollMsg{false}
}

// this seems like a bad thing to do, but I'm not sure how else to interact with an interface
func (ws *WsBackendClient) QueueData(data *protocol.TransportMessage) {
	ws.data <- data
}

func (ws *WsBackendClient) GetChan() chan *protocol.TransportMessage {
	return ws.wsOut
}

func (ws *WsBackendClient) Stop() {
	// This may need to do more later?
	ws.disconnect <- struct{}{}
}

func (ws *WsBackendClient) Connect() error {
	u := url.URL{Scheme: "ws", Host: ws.serverUrl.Host, Path: "/"}
	q := u.Query()
	q.Set("session", ws.sessionId)
	u.RawQuery = q.Encode()
	slog.Info("URL STRING", "query", u.RawQuery, "host", u.Host, "path", u.RawPath)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	ws.conn = c
	go ws.FetchData() // Will this die when I exit the function? I don't think so?
	go ws.SendData()  // Will this die when I exit the function? I don't think so?
	return nil
}

func (ws *WsBackendClient) SendData() {
	slog.Info("WRITING DATA TO BACKEND")
	for msg := range ws.data {
		ws.mut.Lock()
		err := ws.conn.WriteJSON(msg)
		if err != nil {
			slog.Error("error writing to connection", "error", err)
		}
		ws.mut.Unlock()
	}
}

func (ws *WsBackendClient) FetchData() {
	log.Println("starting fetch data")
	for {
		select {
		case <-ws.disconnect:
			close(ws.wsOut)
			ws.conn.Close()
		default:
			// Not using ReadJson because there are potentially multiple transport messages
			_, data, err := ws.conn.ReadMessage()
			if err != nil {
				slog.Error("Unable to turn data into json", "error", err, "data", string(data))
			}
			data = bytes.TrimSpace(bytes.ReplaceAll(data, []byte("\n"), []byte(" ")))

			msg := ParseTransportMessage(data)
			for _, m := range msg {
				log.Printf("adding message to chan, %#v", m)
				ws.wsOut <- m
			}
		}
	}
}

type mockState int

const (
	menu mockState = iota
	table
)

func (m *MockBackendClient) QueueData(data *protocol.TransportMessage) {
	m.data <- data
}

func (m *MockBackendClient) PollAuth() tea.Msg {
	return AuthPollMsg{true}
}

func (m *MockBackendClient) StartAuth(_ string) tea.Msg {
	return AuthLoginMsg{
		UserCode:  "TEST-TEST",
		Url:       "https://github.com/device/login",
		SessionId: "1234",
	}
}

type MockBackendClient struct {
	wsOut      chan *protocol.TransportMessage
	conn       *websocket.Conn
	disconnect chan struct{}
	state      mockState
	data       chan *protocol.TransportMessage
}

func (m *MockBackendClient) SendData() {
	for data := range m.data {
		switch data.Type {
		case protocol.MsgJoinTable:
			m.state = table
		case protocol.MsgLeaveTable:
			m.state = menu
		}
	}
}

func (m *MockBackendClient) FetchData() {
	log.Println("Starting fetchdata")
	var tm []*protocol.TransportMessage
	// eventually this will like read a file or generate some random messages that go through the flow of a blackjack game
	// since I'm only currently working on the menu page / table selection... I don't need to send table data
	tick := time.NewTicker(1 * time.Second)

	for {
		switch m.state {
		case menu:
			log.Println("generating table data")
			tm = generateTableData()
		case table:
			log.Println("generating game data")
			tm = generateGameData()
		}
		select {
		case <-m.disconnect:
			close(m.wsOut)
		case <-tick.C:
			log.Println("Adding mock messages to output channel")
			for _, msg := range tm {
				m.wsOut <- msg
			}
		}
	}
}

func generateMockCards() []protocol.CardDTO {
	suitStrings := []string{"spade", "heart", "diamond", "club"}
	return []protocol.CardDTO{
		{Suit: suitStrings[rand.IntN(4)], Rank: rand.IntN(13) + 1},
		{Suit: suitStrings[rand.IntN(4)], Rank: rand.IntN(13) + 1},
		{Suit: suitStrings[rand.IntN(4)], Rank: rand.IntN(13) + 1},
		{Suit: suitStrings[rand.IntN(4)], Rank: rand.IntN(13) + 1},
	}
}

func generateGameData() []*protocol.TransportMessage {
	randomNames := []string{
		"l_bagel",
		"curious_shark",
		"uninhabited_warzone",
		"kevin",
	}

	// todo
	players := []protocol.PlayerDTO{
		{
			Bet:    5,
			Wallet: 300,
			Hand:   protocol.HandDTO{Cards: generateMockCards(), Value: 18, State: "LIVE"},
			Name:   randomNames[rand.IntN(len(randomNames))],
		},
		{
			Bet:    5,
			Wallet: 300,
			Hand:   protocol.HandDTO{Cards: generateMockCards(), Value: 18, State: "LIVE"},
			Name:   randomNames[rand.IntN(len(randomNames))],
		},
		{},
		{},
		{},
	}
	gameState := protocol.GameDTO{Players: players, DealerHand: protocol.HandDTO{Cards: generateMockCards(), Value: 18, State: "LIVE"}}
	dat, err := protocol.PackageMessage(gameState)
	if err != nil {
		slog.Error("Unable to generate game data. gameState encoding error:", "error", err)
		return nil
	}
	return []*protocol.TransportMessage{dat}
}

func generateTableData() []*protocol.TransportMessage {
	tblList := []protocol.TableDTO{{Id: "test1", Capacity: 5, CurrentPlayers: 1}, {Id: "test3", Capacity: 5, CurrentPlayers: 1}, {Id: "test2", Capacity: 5, CurrentPlayers: 1}}
	dat, err := protocol.PackageMessage(tblList) // This will need to be changed. PackageMessage is hankering for a refactor
	if err != nil {
		slog.Error("Unable to generate table data. tblList encoding error:", "error", err)
		return nil
	}
	return []*protocol.TransportMessage{dat}
}

func (m *MockBackendClient) GetChan() chan *protocol.TransportMessage {
	return m.wsOut
}

func (m *MockBackendClient) Stop() {
	m.disconnect <- struct{}{}
}

func (m *MockBackendClient) Connect() error {
	log.Println("got connect message")
	go m.FetchData()
	go m.SendData()
	return nil
}
