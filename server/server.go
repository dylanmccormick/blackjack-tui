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
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
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
	Server struct {
		SqliteDBPath string `yaml:"sqlite_db_path"`
		port         string `yaml:"port"`
		GitClientID  string `yaml:"git_client_id"`
		SqliteDBName string `yaml:"sqlite_db_name"`
	} `yaml:"server"`

	TableActionTimeout int  `yaml:"table_action_timeout_seconds"`
	TableDeleteTimeout int  `yaml:"table_auto_delete_timeout_minutes"`
	StandOnSoft17      bool `yaml:"stand_on_soft_17"`
	BetTimeout         int  `yaml:"bet_time_seconds"`
	DeckCount          int  `yaml:"deck_count"`
	CutLocation        int  `yaml:"cut_location"`

	// Programming Config Items
	LogLevel string `yaml:"log_level"`
}

type Server struct {
	SessionManager *auth.SessionManager
	Lobby          *Lobby
	Store          *store.Store
	Config         *Config
	log            *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		LogLevel: "INFO",
	}
}

func LoadConfig() Config {
	err := godotenv.Load(".env")
	if err != nil {
		slog.Warn("No env file found", "error", err)
	}
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		slog.Error("Error reading yaml file", "error", err)
		os.Exit(1)
	}
	// Loading all the environment variables into config locations with ${VAR}
	expandedContent := []byte(os.ExpandEnv(string(yamlFile)))
	// Set Defaults for Config
	config := DefaultConfig()

	// Overwrite defaults
	err = yaml.Unmarshal(expandedContent, &config)
	if err != nil {
		slog.Error("Unable to unmarshal yaml file", "error", err)
		os.Exit(1)
	}
	return config
}

func InitializeServer() *Server {
	Config := LoadConfig()
	Store, err := store.NewStore(os.Getenv("SQLITE_DB"), "./sql/schema")
	if err != nil {
		slog.Error("Unable to load or create datastore", "error", err)
		os.Exit(1)
	}
	SessionManager := auth.NewSessionManager(Config.Server.GitClientID)
	Lobby := newLobby(Store)
	var level slog.Level
	err = level.UnmarshalText([]byte(Config.LogLevel))
	if err != nil {
		slog.Error("error parsing log level", "error", err)
		os.Exit(1)
	}
	handlerOptions := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	slog.SetDefault(logger)
	serverLog := slog.With("component", "server")
	return &Server{
		Config:         &Config,
		SessionManager: SessionManager,
		Lobby:          Lobby,
		Store:          Store,
		log:            serverLog,
	}
}

func (s *Server) Run() {
	slog.Debug("Server loaded with config", "config", s.Config)
	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", s.Config)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	wg.Go(func() {
		s.Lobby.run(ctx)
	})

	wg.Go(func() {
		s.SessionManager.Run(ctx)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		auth.WriteHttpResponse(w, http.StatusOK, `{"message": "healthy"}`)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.log.Info("Received connection")
		s.serveWs(w, r)
	})

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		s.log.Info("Received login request")
		session := auth.LoginHandler(w, r, s.SessionManager)
		s.SessionManager.AddSession(session)
	})

	http.HandleFunc("/auth/status", func(w http.ResponseWriter, r *http.Request) {
		s.log.Debug("getting status for request")
		id := r.URL.Query().Get("id")
		auth.HandleAuthCheck(s.SessionManager, id, w, r)
	})

	server := &http.Server{Addr: ":42069"}
	go server.ListenAndServe()

	<-sigChan
	s.log.Info("Received shutdown signal")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)

	wg.Wait()
	s.log.Info("Shutdown Complete")
}

func generateId() uuid.UUID {
	uuid, err := uuid.NewUUID()
	if err != nil {
		slog.Error("ERROR GENERATING UUID", "error", err)
	}
	return uuid
}

func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	// serve ws should take the client and register them with the table. They should then go through the onboarding process... (login, authenticate, provide a username)
	// CHECK SESSION MANAGER FOR KEY
	ctx := context.Background()
	sessionId := r.URL.Query().Get("session")
	session, err := s.SessionManager.GetSession(sessionId)
	if err != nil {
		s.log.Error("error with session", "error", err)
		http.Error(w, "Authentication Required", http.StatusUnauthorized)
		return
	}
	if !session.Authenticated {
		http.Error(w, "Authentication Required", http.StatusUnauthorized)
		return
	}

	ctx = context.WithValue(ctx, "sessionId", sessionId)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error("An error occurred upgrading the http connection", "error", err)
		http.Error(w, "Error upgrading connection", http.StatusInternalServerError)
		return
	}
	client := &Client{
		conn:     conn,
		send:     make(chan *protocol.TransportMessage, 10),
		id:       generateId(),
		manager:  s.Lobby,
		log:      slog.With("component", "client"),
		username: session.GithubUserId,
	}
	client.manager.register(client)
	ctx = context.WithValue(ctx, "ghUsername", client.username)

	go client.readPump(ctx)
	go client.writePump(ctx)

	u, income, err := s.Store.ProcessLogin(ctx, client.username)
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

	isStarred, err := s.SessionManager.CheckStarredStatus(ctx, session)
	if err != nil {
		slog.Error("error processing repo stars", "error", err)
	}
	if isStarred {
		slog.Info("USER STARRED THE REPO!")
	}

	newStar, err := s.Store.UpdateUserStarred(ctx, client.username)
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

func (c *Client) readPump(ctx context.Context) {
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
				c.log.Error("error with websocket", "error", err, "sessionId", ctx.Value("sessionId"), "client", ctx.Value("ghUsername"))
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

func (c *Client) writePump(ctx context.Context) {
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
			c.log.Debug("Writing message", "message", message, "sessionId", ctx.Value("sessionId"), "clientId", ctx.Value("ghUsername"))
			c.conn.WriteJSON(message)
			for range len(c.send) {
				msg := <-c.send
				c.log.Debug("Writing message", "message", msg, "sessionId", ctx.Value("sessionId"), "clientId", ctx.Value("ghUsername"))
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
