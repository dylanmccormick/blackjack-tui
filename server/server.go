package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dylanmccormick/blackjack-tui/protocol"
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

func RunServer() {
	handlerOptions := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	slog.SetDefault(logger)
	serverLog = slog.With("component", "server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	lobby := newLobby()

	var wg sync.WaitGroup

	wg.Go(func() {
		lobby.run(ctx)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serverLog.Info("Received connection")
		serveWs(lobby, w, r)
	})

	server := &http.Server{Addr: ":8080"}
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

func serveWs(l *Lobby, w http.ResponseWriter, r *http.Request) {
	// serve ws should take the client and register them with the table. They should then go through the onboarding process... (login, authenticate, provide a username)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		serverLog.Error("An error occurred upgrading the http connection", "error", err)
		// TODO: this should send some sort of error back to the client
		return
	}
	client := &Client{
		conn:    conn,
		send:    make(chan *protocol.TransportMessage, 10),
		id:      generateId(conn),
		manager: l,
		log:     slog.With("component", "client"),
	}
	client.manager.register(client)

	go client.readPump()
	go client.writePump()
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
				c.log.Error("error: %v", err)
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
