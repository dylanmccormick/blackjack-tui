package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	conn *websocket.Conn
	mu   sync.Mutex
	id   string
	hub  *Hub
	send chan []byte
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

func main() {
	hub := newHub()
	go hub.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received connection")
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// serve ws should take the client and register them with the hub. They should then go through the onboarding process... (login, authenticate, provide a username)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("An error occurred upgrading the http connection", "error", err)
		panic(err)
	}
	client := &Client{
		conn: conn,
		hub:  hub,
		send: make(chan []byte, 10),
	}
	slog.Info("Registering client to hub")
	client.hub.register <- client

	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
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
		c.hub.inbound <- message
	}
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
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			for range len(c.send) {
				w.Write(newline)
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
