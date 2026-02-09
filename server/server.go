package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dylanmccormick/blackjack-tui/auth"
	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/dylanmccormick/blackjack-tui/store"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	newline   = []byte{'\n'}
	space     = []byte{' '}
	serverLog *slog.Logger
)

type manager interface {
	register(*Client)
	unregister(*Client)
	sendMessage(inboundMessage)
	Id() string
}

type Client struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	id       uuid.UUID
	manager  manager
	send     chan *protocol.TransportMessage
	username string
	log      *slog.Logger
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Config struct {
	SqliteDBPath       string `yaml:"sqlite_db_path"`
	port               string `yaml:"port"`
	TableActionTimeout int    `yaml:"table_action_timeout_seconds"`
	TableDeleteTimeout int    `yaml:"table_auto_delete_timout_minutes"`
	StandOnSoft17      bool   `yaml:"stand_on_soft_17"`
	BetTimeout         int    `yaml:"bet_time_seconds"`
	DeckCount          int    `yaml:"deck_count"`
	CutLocation        int    `yaml:"cut_location"`
}

func RunServer() {
	handlerOptions := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	slog.SetDefault(logger)
	serverLog = slog.With("component", "server")

	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", Config{})

	store, err := store.NewStore(os.Getenv("SQLITE_DB"), "./sql/schema")
	if err != nil {
		slog.Error("Unable to load or create datastore", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	lobby := newLobby(store)
	sessionManager := auth.NewSessionManager()

	var wg sync.WaitGroup

	wg.Go(func() {
		lobby.run(ctx)
	})

	wg.Go(func() {
		sessionManager.Run(ctx)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serverLog.Info("Received connection")
		// todo... validate logged in
		serveWs(sessionManager, lobby, w, r, store)
	})

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		serverLog.Info("Received login request")
		session := auth.LoginHandler(w, r)
		sessionManager.AddSession(session)
	})

	http.HandleFunc("/auth/status", func(w http.ResponseWriter, r *http.Request) {
		serverLog.Debug("getting status for request")
		id := r.URL.Query().Get("id")
		auth.HandleAuthCheck(sessionManager, id, w, r)
	})

	server := &http.Server{Addr: ":42069"}
	go server.ListenAndServe()

	<-sigChan
	serverLog.Info("Received shutdown signal")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)

	wg.Wait()
	serverLog.Info("Shutdown Complete")
}

func generateId(c *websocket.Conn) uuid.UUID {
	// TODO: Generate uuid from websocket so it's sticky... figure that out later
	uuid, err := uuid.NewUUID()
	if err != nil {
		serverLog.Error("ERROR GENERATING UUID", "error", err)
	}
	return uuid
}

func serveWs(sm *auth.SessionManager, l *Lobby, w http.ResponseWriter, r *http.Request, store *store.Store) {
	// serve ws should take the client and register them with the table. They should then go through the onboarding process... (login, authenticate, provide a username)
	// CHECK SESSION MANAGER FOR KEY
	sessionId := r.URL.Query().Get("session")
	session, err := sm.GetSession(sessionId)
	if err != nil {
		// TODO: return a rejection response to requester
		serverLog.Error("error with session", "error", err)
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	if !session.Authenticated {
		// TODO: return a rejection response to requester
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		serverLog.Error("An error occurred upgrading the http connection", "error", err)
		// TODO: this should send some sort of error back to the client
		return
	}
	client := &Client{
		conn:     conn,
		send:     make(chan *protocol.TransportMessage, 10),
		id:       generateId(conn),
		manager:  l,
		log:      slog.With("component", "client"),
		username: session.GithubUserId,
	}
	client.manager.register(client)

	go client.readPump()
	go client.writePump()

	u, income, err := store.ProcessLogin(context.TODO(), client.username)
	if err != nil {
		slog.Error("error processing login", "error", err)
	}
	slog.Info("User got that money!", "income", income)

	msg := fmt.Sprintf("Thank you for logging in for %d day(s) in a row! You have earned %d income", u.LoginStreak, income)

	pack, err := protocol.PackageMessage(protocol.MessageToDTO(msg, protocol.InfoMsg))
	if err != nil {
		slog.Error("Unable to process message", "error", err)
	}

	client.send <- pack

	isStarred, err := sm.CheckStarredStatus(context.TODO(), session)
	if err != nil {
		slog.Error("error processing repo stars", "error", err)
	}
	if isStarred {
		slog.Info("USER STARRED THE REPO!")
	}

	newStar, err := store.UpdateUserStarred(context.TODO(), client.username)
	if newStar {
		slog.Info("Thank you for the star kind user", "user", client.username)
		msg = fmt.Sprintf("Thank you for starring the repo! You have earned %d bonus", 5000)

		pack, err = protocol.PackageMessage(protocol.MessageToDTO(msg, protocol.InfoMsg))
		if err != nil {
			slog.Error("Unable to process message", "error", err)
		}

		client.send <- pack
	}
}

func (c *Client) readPump() {
	defer func() {
		c.mu.Lock()
		mgr := c.manager
		c.mu.Unlock()
		mgr.unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Error("error with websocket", "error", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		uMsg, err := unpackMessage(message)
		c.mu.Lock()
		c.manager.sendMessage(inboundMessage{uMsg, c})
		c.mu.Unlock()
	}
}

func unpackMessage(msg []byte) (*protocol.TransportMessage, error) {
	jsonMsg := &protocol.TransportMessage{}
	err := json.Unmarshal(msg, jsonMsg)
	if err != nil {
		return &protocol.TransportMessage{}, err
	}
	return jsonMsg, nil
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.log.Debug("Writing message", "message", message)
			c.conn.WriteJSON(message)
			for range len(c.send) {
				msg := <-c.send
				c.log.Debug("Writing message", "message", msg)
				c.conn.WriteJSON(msg)
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
