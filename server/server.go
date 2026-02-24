package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Middleware func(http.Handler) http.Handler

type manager interface {
	register(*Client)
	unregister(*Client)
	sendMessage(inboundMessage)
	Id() string
}

type Client struct {
	conn        *websocket.Conn
	mu          sync.Mutex
	id          uuid.UUID
	manager     manager
	send        chan *protocol.TransportMessage
	username    string
	log         *slog.Logger
	connectedAt time.Time
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
		SqliteSchemaPath string `yaml:"sqlite_schema_path"`
		port             string `yaml:"port"`
		GitClientID      string `yaml:"git_client_id"`
		SqliteDBName     string `yaml:"sqlite_db_name"`
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
	Metrics        *Metrics
	Registry       *prometheus.Registry
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
	Store, err := store.NewStore(Config.Server.SqliteDBName, Config.Server.SqliteSchemaPath)
	if err != nil {
		slog.Error("Unable to load or create datastore", "error", err)
		os.Exit(1)
	}
	registry := prometheus.NewRegistry()
	metrics := NewMetrics(registry)
	SessionManager := auth.NewSessionManager(Config.Server.GitClientID)
	Lobby := newLobby(Store, metrics)
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
		Metrics:        metrics,
		Registry:       registry,
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

	mux := http.NewServeMux()

	// PUBLIC ROUTES (no auth)
	mux.HandleFunc("/healthz", s.healthCheckHandler)
	mux.Handle("/metrics", promhttp.HandlerFor(s.Registry, promhttp.HandlerOpts{Registry: s.Registry}))
	mux.HandleFunc("/auth", s.authHandler)
	mux.HandleFunc("/auth/status", s.authStatusHandler)

	// Auth required
	protectedWs := chainMiddleware(
		http.HandlerFunc(s.serveWs),
		s.authMiddleware,
	)
	mux.Handle("/", protectedWs)

	handler := chainMiddleware(
		mux,
		s.panicRecoveryMiddleware,
		s.loggingMiddleware,
		s.metricsMiddleware,
	)

	server := &http.Server{Addr: ":42069", Handler: handler}
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

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) panicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.log.Error("PANIC recovered",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.log.Info("[Request Start]", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
		duration := time.Since(start).Seconds()
		s.log.Info("[Request End]", "method", r.Method, "path", r.URL.Path, "duration", duration)
	})
}

type responseWriterWithStatus struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWithStatus) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("responsewriter does not support hijacking")
	}
	return hijacker.Hijack()
}

func (rw *responseWriterWithStatus) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		s.Metrics.HTTPRequestsInFlight.Inc()
		defer s.Metrics.HTTPRequestsInFlight.Dec()

		wrappedWriter := &responseWriterWithStatus{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start).Seconds()
		method := r.Method
		path := r.URL.Path
		status := fmt.Sprintf("%d", wrappedWriter.statusCode)
		s.Metrics.HTTPRequestsDuration.WithLabelValues(method, path).Observe(duration)
		s.Metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	})
}

func chainMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Apply middleware in REVERSE order so they execute in the order listed
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
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
	ctx := r.Context()
	session := ctx.Value("session").(*auth.Session)
	ctx = context.WithValue(ctx, "sessionId", session.SessionId)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error("An error occurred upgrading the http connection", "error", err)
		http.Error(w, "Error upgrading connection", http.StatusInternalServerError)
		return
	}
	client := &Client{
		conn:        conn,
		send:        make(chan *protocol.TransportMessage, 10),
		id:          generateId(),
		manager:     s.Lobby,
		log:         slog.With("component", "client"),
		username:    session.GithubUserId,
		connectedAt: time.Now(),
	}
	s.Metrics.ConnectedClients.Inc()
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
