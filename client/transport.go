package client

import (
	"bytes"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/url"
	"sync"
	"time"

	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/gorilla/websocket"
)

func NewMockTransporter() *MockTransportMessageIO {
	return &MockTransportMessageIO{
		out:        make(chan *protocol.TransportMessage),
		disconnect: make(chan struct{}),
		data:       make(chan *protocol.TransportMessage),
	}
}

func NewWsTransportMessageIO() *WsTransportMessageIO {
	return &WsTransportMessageIO{
		out:        make(chan *protocol.TransportMessage),
		disconnect: make(chan struct{}),
		data:       make(chan *protocol.TransportMessage),
	}
}

type TransportMessageIO interface {
	GetChan() chan *protocol.TransportMessage
	Connect(string) error // I think later we'll add an address you can connect to as a param
	Stop()                // Stops fetch data goroutine and disconnects from server.
	SendData()            // Reads from data chan sends JSON data across the wire to the server
	FetchData()           // Runs goroutine to pull data from the server connection
	QueueData(*protocol.TransportMessage)
}

type WsTransportMessageIO struct {
	mut        sync.Mutex
	out        chan *protocol.TransportMessage
	conn       *websocket.Conn
	data       chan *protocol.TransportMessage
	disconnect chan struct{}
}

// this seems like a bad thing to do, but I'm not sure how else to interact with an interface
func (ws *WsTransportMessageIO) QueueData(data *protocol.TransportMessage) {
	ws.data <- data
}

func (ws *WsTransportMessageIO) GetChan() chan *protocol.TransportMessage {
	return ws.out
}

func (ws *WsTransportMessageIO) Stop() {
	// This may need to do more later?
	ws.disconnect <- struct{}{}
}

func (ws *WsTransportMessageIO) Connect(sessionId string) error {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/"}
	q := u.Query()
	q.Set("session", sessionId)
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

func (ws *WsTransportMessageIO) SendData() {
	for msg := range ws.data {
		ws.mut.Lock()
		err := ws.conn.WriteJSON(msg)
		if err != nil {
			slog.Error("error writing to connection", "error", err)
		}
		ws.mut.Unlock()
	}
}

func (ws *WsTransportMessageIO) FetchData() {
	log.Println("starting fetch data")
	for {
		select {
		case <-ws.disconnect:
			close(ws.out)
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
				ws.out <- m
			}
		}
	}
}

type mockState int

const (
	menu mockState = iota
	table
)

func (m *MockTransportMessageIO) QueueData(data *protocol.TransportMessage) {
	m.data <- data
}

type MockTransportMessageIO struct {
	out        chan *protocol.TransportMessage
	conn       *websocket.Conn
	disconnect chan struct{}
	state      mockState
	data       chan *protocol.TransportMessage
}

func (m *MockTransportMessageIO) SendData() {
	for data := range m.data {
		switch data.Type {
		case protocol.MsgJoinTable:
			m.state = table
		case protocol.MsgLeaveTable:
			m.state = menu
		}
	}
}

func (m *MockTransportMessageIO) FetchData() {
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
			close(m.out)
		case <-tick.C:
			log.Println("Adding mock messages to output channel")
			for _, msg := range tm {
				m.out <- msg
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

func (m *MockTransportMessageIO) GetChan() chan *protocol.TransportMessage {
	return m.out
}

func (m *MockTransportMessageIO) Stop() {
	m.disconnect <- struct{}{}
}

func (m *MockTransportMessageIO) Connect(sessionId string) error {
	log.Println("got connect message")
	go m.FetchData()
	go m.SendData()
	return nil
}
