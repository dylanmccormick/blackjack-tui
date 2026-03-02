package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/dylanmccormick/blackjack-tui/auth"
	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/dylanmccormick/blackjack-tui/server"
	"github.com/dylanmccormick/blackjack-tui/store"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

type ChaosClient struct {
	conn     *websocket.Conn
	send     chan *protocol.TransportMessage
	username string
	behavior string

	gameState *protocol.GameDTO
}

var yamlConfig = `
server:
  sqlite_schema_path: '../sql/schema'
  port: 42069
  git_client_id: 'TEST'
  sqlite_db_name: ':memory:'

log_level: INFO

table_action_timeout_seconds: 5
table_auto_delete_timeout_minutes: 5

# Game Config
stand_on_soft_17: true
bet_time_seconds: 5
deck_count: 6
cut_location: 150
`

var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

func setUpTestServer(t *testing.T) *server.Server {
	Config := server.Config{}
	err := yaml.Unmarshal([]byte(yamlConfig), &Config)
	if err != nil {
		panic(err)
	}
	Store, err := store.NewStore(":memory:", "../sql/schema")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	registry := prometheus.NewRegistry()
	metrics := server.NewMetrics(registry)
	SessionManager := auth.NewSessionManager(Config.Server.GitClientID)
	Lobby := server.NewLobby(Store, metrics)
	srv := &server.Server{
		Config:         &Config,
		SessionManager: SessionManager,
		Lobby:          Lobby,
		Store:          Store,
		Metrics:        metrics,
		Registry:       registry,
		Log:            logger,
	}

	go srv.Run()

	time.Sleep(100 * time.Millisecond)
	return srv
}

func createTestSession(username string) *auth.Session {
	return &auth.Session{
		SessionId:     uuid.New().String(),
		GithubUserId:  username,
		Authenticated: true,
		LastRequest:   time.Now(),
	}
}

func spawnChaosClients(t *testing.T, n int, serverUrl string, sm *auth.SessionManager) []*ChaosClient {
	u := url.URL{Scheme: "ws", Host: serverUrl, Path: "/"}
	var clients []*ChaosClient
	for i := range n {
		t.Logf("Creating client no. %d", i)
		username := fmt.Sprintf("chaos_user%d", i)
		session := createTestSession(username)
		sm.Sessions[session.SessionId] = session
		sm.Commands <- &auth.SessionCmd{
			Action:    "updateSession",
			SessionId: session.SessionId,

			Authenticated: true,
			GHToken:       "1234",
		}

		url := fmt.Sprintf("%s?session=%s", u.String(), session.SessionId)
		t.Log("Connecting to server", "server", url)
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			panic(err)
		}
		c := &ChaosClient{
			conn:     conn,
			username: username,
			behavior: "golden",
		}

		clients = append(clients, c)
	}
	return clients
}

func (c *ChaosClient) RandomAction() {
	actions := []string{
		"stand",
		"stand",
		"stand",
		"stand",
		"bet",
		"bet",
		"bet",
		"bet",
		// "join_table",
		// "leave_table",
	}
	// TODO: Use that string to generate an action
	action := actions[rand.IntN(len(actions))]
	slog.Info("Chaos agent running action", "agent", c.username, "action", action)
	switch action {
	case "start":
		c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgStartGame, ""))
	case "hit":
		c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgHit, ""))
	case "stand":
		c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgStand, ""))
	case "bet":
		c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgPlaceBet, fmt.Sprintf("%d", rand.IntN(5000))))
	case "join_table":
		c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgJoinTable, fmt.Sprintf("chaos%d", rand.IntN(5))))
	case "leave_table":
		c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgLeaveTable, ""))
	}
}

func (c *ChaosClient) readMessages(ctx context.Context) {
	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping message reading")
			return
		case <-ticker.C:
			var msg protocol.TransportMessage
			err := c.conn.ReadJSON(&msg)
			if err != nil {
				return
			}
			if msg.Type == "game_state" {
				var gameState protocol.GameDTO
				json.Unmarshal(msg.Data, &gameState)
				c.gameState = &gameState
			}
		}
	}
}

func (c *ChaosClient) Act() {
	switch c.behavior {
	case "golden":
		c.ActGolden()
	case "random":
		c.RandomAction()
	}
}

func (c *ChaosClient) isMyTurn() bool {
	slog.Info("Checking turn")
	for _, player := range c.gameState.Players {
		if player.Name == c.username && player.CurrentPlayer {
			return true
		}
	}
	return false
}

func (c *ChaosClient) ActGolden() {
	if c.gameState == nil {
		c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgPlaceBet, "5"))
		return
	}

	switch c.gameState.State {
	case "WAITING_FOR_BETS":
		c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgPlaceBet, "5"))
	case "PLAYER_TURN":
		if c.isMyTurn() {
			c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgStand, ""))
		}
	default:
		slog.Info("gamestate", "state", c.gameState.State)
	}
}

func runChaos(ctx context.Context, clients []*ChaosClient, duration time.Duration) {
	timer := time.NewTimer(duration)
	ticker := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			slog.Info("Chaos test interrupted by user")
			return
		case <-ticker.C:
			for _, c := range clients {
				c.Act()
				// c.RandomAction()
			}
		case <-timer.C:
			return
		}
	}
}

func createTables(client *ChaosClient, num int) {
	for i := range num {
		client.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgCreateTable, fmt.Sprintf("chaos%d", i)))
	}
	time.Sleep(1 * time.Second)
}

func TestChaosMonkey(t *testing.T) {
	slog.SetDefault(logger)
	t.Log("Logger set up")
	if testing.Short() {
		t.Skip()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		select {
		case <-sigChan:
			t.Log("Shutting down chaos tests")
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	t.Log("Setting up test server")
	server := setUpTestServer(t)

	t.Log("Creating clients")
	clients := spawnChaosClients(t, 4, "localhost:42069", server.SessionManager)
	createTables(clients[0], 6)
	for i, c := range clients {
		t.Logf("Client %d (%s) joining table chaos%d", i, c.username, 0)
		c.conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgJoinTable, fmt.Sprintf("chaos%d", 0)))
		t.Logf("Starting read message for chaos%d", i)
		go c.readMessages(ctx)
	}
	// Wait for all joins to complete
	time.Sleep(5 * time.Second)
	// Then, start the games (only one client per table needs to do this)
	clients[0].conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgStartGame, ""))
	// for i := 0; i < 6; i++ {
	// 	t.Logf("Starting game on table chaos%d", i)
	// 	clients[i].conn.WriteJSON(protocol.PackageClientMessage(protocol.MsgStartGame, ""))
	// }

	t.Log("Running chaos tests")
	runChaos(ctx, clients, 6000*time.Second)

	// assertNoPanics(t)
	// assertNoGoroutineLeaks(t)
}
