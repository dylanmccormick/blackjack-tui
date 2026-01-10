package client

import (
	"bytes"
	"log"
	"math/rand/v2"
	"net/url"
	"time"

	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/gorilla/websocket"
)

func NewMockTransporter() *MockTransportMessageIO {
	return &MockTransportMessageIO{
		out:        make(chan *protocol.TransportMessage),
		disconnect: make(chan struct{}),
	}
}

func NewWsTransportMessageIO() *WsTransportMessageIO {
	return &WsTransportMessageIO{
		out:        make(chan *protocol.TransportMessage),
		disconnect: make(chan struct{}),
	}
}

type TransportMessageIO interface {
	GetChan() chan *protocol.TransportMessage
	Connect() error                            // I think later we'll add an address you can connect to as a param
	Stop()                                     // Stops fetch data goroutine and disconnects from server.
	SendData(*protocol.TransportMessage) error // Sends JSON data across the wire to the server
	FetchData()                                // Runs goroutine to pull data from the server connection
}

type WsTransportMessageIO struct {
	out        chan *protocol.TransportMessage
	conn       *websocket.Conn
	disconnect chan struct{}
}

func (ws *WsTransportMessageIO) GetChan() chan *protocol.TransportMessage {
	return ws.out
}

func (ws *WsTransportMessageIO) Stop() {
	// This may need to do more later?
	ws.disconnect <- struct{}{}
}

func (ws *WsTransportMessageIO) Connect() error {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	ws.conn = c
	go ws.FetchData() // Will this die when I exit the function? I don't think so?
	return nil
}

func (ws *WsTransportMessageIO) SendData(tm *protocol.TransportMessage) error {
	return nil
}

func (ws *WsTransportMessageIO) FetchData() {
	for {
		select {
		case <-ws.disconnect:
			close(ws.out)
			ws.conn.Close()
		default:
			_, data, err := ws.conn.ReadMessage()
			if err != nil {
				panic(err)
			}
			data = bytes.TrimSpace(bytes.ReplaceAll(data, []byte("\n"), []byte(" ")))

			log.Printf("adding message to chan: %s", string(data))
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

type MockTransportMessageIO struct {
	out        chan *protocol.TransportMessage
	conn       *websocket.Conn
	disconnect chan struct{}
	state      mockState
}

func (m *MockTransportMessageIO) SendData(tm *protocol.TransportMessage) error {
	log.Printf("Current state: %#v", m.state)
	switch tm.Type {
	case protocol.MsgJoinTable:
		log.Println("Changgin state")
		m.state = table
	case protocol.MsgLeaveTable:
		m.state = menu
	}
	return nil
}

func (m *MockTransportMessageIO) FetchData() {
	log.Println("Starting fetchdata")
	var tm []*protocol.TransportMessage
	// eventually this will like read a file or generate some random messages that go through the flow of a blackjack game
	// since I'm only currently working on the menu page / table selection... I don't need to send table data
	tick := time.NewTicker(5 * time.Second)

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
	return []protocol.CardDTO{
		{Suit: suits[rand.IntN(4)], Rank: rand.IntN(13) + 1},
		{Suit: suits[rand.IntN(4)], Rank: rand.IntN(13) + 1},
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
	}
	gameState := protocol.GameDTO{Players: players, DealerHand: protocol.HandDTO{Cards: generateMockCards(), Value: 18, State: "LIVE"}}
	dat, err := protocol.PackageMessage(gameState) // This will need to be changed. PackageMessage is hankering for a refactor
	if err != nil {
		panic(err)
	}
	return ParseTransportMessage(dat)
}

func generateTableData() []*protocol.TransportMessage {
	tblList := []protocol.TableDTO{{Id: "test1", Capacity: 5, CurrentPlayers: 1}, {Id: "test3", Capacity: 5, CurrentPlayers: 1}, {Id: "test2", Capacity: 5, CurrentPlayers: 1}}
	dat, err := protocol.PackageMessage(tblList) // This will need to be changed. PackageMessage is hankering for a refactor
	if err != nil {
		panic(err)
	}
	return ParseTransportMessage(dat)
}

func (m *MockTransportMessageIO) GetChan() chan *protocol.TransportMessage {
	return m.out
}

func (m *MockTransportMessageIO) Stop() {
	m.disconnect <- struct{}{}
}

func (m *MockTransportMessageIO) Connect() error {
	log.Println("got connect message")
	go m.FetchData()
	return nil
}
