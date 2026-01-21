package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var lobby = newLobby()

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
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	slog.SetDefault(logger)
	go lobby.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received connection")
		serveWs(w, r)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("Error serving http", "error", err)
		os.Exit(1)
	}
}

func generateId(c *websocket.Conn) uuid.UUID {
	// TODO: Generate uuid from websocket so it's sticky... figure that out later
	uuid, err := uuid.NewUUID()
	if err != nil {
		slog.Error("ERROR GENERATING UUID", "error", err)
	}
	return uuid
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	// serve ws should take the client and register them with the table. They should then go through the onboarding process... (login, authenticate, provide a username)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("An error occurred upgrading the http connection", "error", err)
		// TODO: this should send some sort of error back to the client
		return
	}
	client := &Client{
		conn:    conn,
		send:    make(chan *protocol.TransportMessage, 10),
		id:      generateId(conn),
		manager: lobby,
	}
	client.manager.register(client)

	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.mu.Lock()
		c.manager.unregister(c)
		c.mu.Unlock()
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			slog.Error("retrieved an error", "error", err)
			break
		}
		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		uMsg, err := unpackMessage(message)
		c.mu.Lock()
		slog.Info("writing message to manager", "manager", c.manager.Id())
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
			slog.Info("Writing message", "message", message)
			c.conn.WriteJSON(message)
			for range len(c.send) {
				msg := <-c.send
				slog.Info("Writing message", "message", msg)
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
