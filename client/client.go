package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/gorilla/websocket"
)

// Pages... Menu Page and Game Page and Maybe a Settings Page?? Later. If I decide to add themes
const (
	tableHor = "═"
	tableVer = "║"
	tableTL  = "╔"
	tableTR  = "╗"
	tableBL  = "╚"
	tableBR  = "╝"
	cardHor  = "─"
	cardVer  = "│"
	cardTL   = "╭"
	cardTR   = "╮"
	cardBL   = "╰"
	cardBR   = "╯"
)

type page int

var (
	values = []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	suits  = []string{"♠", "♦", "♥", "♣"}
)

const (
	loginPage page = iota
	menuPage
	gamePage
	Width  = 6
	Height = 5
)

type state struct {
	todo int
}

type RootModel struct {
	page   page
	state  state
	width  int
	height int

	wsMessages chan *protocol.GameDTO
	conn       *websocket.Conn

	table *TuiTable
}

func (rm *RootModel) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		listenForMessages(rm.conn, rm.wsMessages),
		ReceiveMessage(rm.wsMessages),
	)
}

func (rm *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return rm, tea.Quit
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "b":
				err := rm.conn.WriteMessage(websocket.TextMessage, []byte("bet 5"))
				if err != nil {
					log.Print(err)
				}
				// bet 5
			case "h":
				err := rm.conn.WriteMessage(websocket.TextMessage, []byte("hit"))
				if err != nil {
					log.Print(err)
				}
				// hit
			case "s":
				// stay
				err := rm.conn.WriteMessage(websocket.TextMessage, []byte("stay"))
				if err != nil {
					log.Print(err)
				}
			}
		}
	case tea.WindowSizeMsg:
		rm.width = msg.Width
		rm.height = msg.Height
	case *protocol.GameDTO:
		rm.GameMessageToState(msg)
		return rm, ReceiveMessage(rm.wsMessages)
	}
	return rm, ReceiveMessage(rm.wsMessages)
}

func NewRootModel() *RootModel {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	return &RootModel{
		wsMessages: make(chan *protocol.GameDTO),
		conn:       c,
		table:      NewTable(),
	}
}

func (rm *RootModel) View() string {
	return lipgloss.Place(rm.width, rm.height, lipgloss.Center, lipgloss.Center, rm.table.View())
}

type Card struct {
	Value int
	Suit  int
}

type TuiPlayer struct {
	Name   string
	Cards  []*Card
	Value  int
	Wallet int
	Bet    int
}

type TuiTable struct {
	Players []TuiPlayer
}

var testPlayers = []TuiPlayer{
	{"dealer", []*Card{NewCard(0, 0), NewCard(0, 0)}, 0, 0, 0},
	{"player_1", []*Card{NewCard(0, 0), NewCard(0, 0)}, 0, 0, 0},
	{"player_2", []*Card{NewCard(0, 0), NewCard(0, 0)}, 0, 0, 0},
	{"player_3", []*Card{NewCard(0, 0), NewCard(0, 0)}, 0, 0, 0},
	{"player_4", []*Card{NewCard(0, 0), NewCard(0, 0)}, 0, 0, 0},
	{"player_5", []*Card{NewCard(0, 0), NewCard(0, 0)}, 0, 0, 0},
	{"player_6", []*Card{NewCard(0, 0), NewCard(0, 0)}, 0, 0, 0},
}

func NewTable() *TuiTable {
	return &TuiTable{
		Players: testPlayers,
	}
}

func NewCard(value, suit int) *Card {
	return &Card{
		Value: value,
		Suit:  suit,
	}
}

func (c *Card) String() string {
	return values[c.Value] + suits[c.Suit]
}

func (c *Card) ViewPartial() string {
	color := lipgloss.Color("#FFFFFF")
	style := lipgloss.NewStyle().Foreground(color)
	padding := strings.Repeat(cardHor, 2)
	view := style.Render(cardTL) + c.String() + "\n"
	for i := 1; i < Height-1; i++ {
		view += style.Render(cardVer+strings.Repeat(" ", 2)) + "\n"
	}
	view += style.Render("╰" + padding)
	return view
}

func (c *Card) View() string {
	color := lipgloss.Color("#FFFFFF")
	style := lipgloss.NewStyle().Foreground(color)
	padding := strings.Repeat(cardHor, Width-2-lipgloss.Width(c.String()))
	view := style.Render("╭") + c.String() + style.Render(padding+"╮") + "\n"
	for i := 1; i < Height-1; i++ {
		view += style.Render(cardVer+strings.Repeat(" ", Width-2)+cardVer) + "\n"
	}
	view += style.Render("╰"+padding) + c.String() + style.Render("╯")
	return view
}

func (t *TuiTable) View() string {
	color := lipgloss.Color("#FFFFFF")
	style := lipgloss.NewStyle().Foreground(color)
	// Render top line
	top := style.Render(tableTL + strings.Repeat(tableHor, 66-2) + tableTR)
	vertBorder := style.Render(strings.Repeat(tableVer+"\n", 24) + tableVer)
	// Render Middle
	middle := lipgloss.JoinHorizontal(lipgloss.Left, vertBorder, t.renderMiddle(), vertBorder)
	// Render Bottom Line
	bottom := style.Render(tableBL + strings.Repeat(tableHor, 66-2) + tableBR)
	return lipgloss.JoinVertical(lipgloss.Top, top, middle, bottom)
}

func (t *TuiTable) renderMiddle() string {
	vzone1 := t.renderVerticalZone1()
	vzone2 := t.renderVerticalZone2()
	vzone3 := t.renderVerticalZone3()
	return lipgloss.JoinHorizontal(lipgloss.Left, vzone1, vzone2, vzone3)
}

func (t *TuiTable) renderVerticalZone1() string {
	playerOne := t.Players[1].renderPlayerZone()
	playerTwo := t.Players[2].renderPlayerZone()
	middle1 := lipgloss.JoinHorizontal(lipgloss.Left, renderEmptyBlock(3, 6), playerOne, renderEmptyBlock(4, 6))
	middle2 := lipgloss.JoinHorizontal(lipgloss.Left, renderEmptyBlock(3, 6), playerTwo, renderEmptyBlock(4, 6))
	return lipgloss.JoinVertical(lipgloss.Top, renderEmptyBlock(22, 1), middle1, renderEmptyBlock(20, 2), middle2, renderEmptyBlock(20, 5))
}

func (t *TuiTable) renderVerticalZone2() string {
	dealer := t.Players[0].renderPlayerZone()
	player3 := t.Players[3].renderPlayerZone()
	middle1 := lipgloss.JoinHorizontal(lipgloss.Left, dealer, renderEmptyBlock(1, 6))
	middle2 := lipgloss.JoinHorizontal(lipgloss.Left, player3, renderEmptyBlock(1, 6))
	return lipgloss.JoinVertical(lipgloss.Top, middle1, renderEmptyBlock(20, 6), middle2, renderEmptyBlock(20, 2))
}

func (t *TuiTable) renderVerticalZone3() string {
	playerFour := t.Players[4].renderPlayerZone()
	playerFive := t.Players[4].renderPlayerZone()
	middle1 := lipgloss.JoinHorizontal(lipgloss.Left, playerFive, renderEmptyBlock(4, 6))
	middle2 := lipgloss.JoinHorizontal(lipgloss.Left, playerFour, renderEmptyBlock(4, 6))
	return lipgloss.JoinVertical(lipgloss.Top, renderEmptyBlock(16, 1), middle1, renderEmptyBlock(22, 2), middle2, renderEmptyBlock(22, 5))
}

func (p *TuiPlayer) renderPlayerZone() string {
	nameTag := p.Name
	bet := p.Bet
	wallet := p.Wallet
	value := p.Value
	status := fmt.Sprintf("V:%d B:%d W:%d", value, bet, wallet)
	return lipgloss.JoinVertical(lipgloss.Top, nameTag, renderMultipleCards(p.Cards), status)
}

func renderMultipleCards(cards []*Card) string {
	cardViews := []string{}
	for i, card := range cards {
		if i == len(cards)-1 {
			cardViews = append(cardViews, card.View())
		} else {
			cardViews = append(cardViews, card.ViewPartial())
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, cardViews...)
}

func renderEmptyBlock(width, height int) string {
	view := ""
	for range height {
		view += strings.Repeat(" ", width) + "\n"
	}
	return view
}

func RunTui() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}
	rm := NewRootModel()
	p := tea.NewProgram(rm)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there has been an error: %v", err)
		os.Exit(1)
	}
}

func listenForMessages(conn *websocket.Conn, c chan *protocol.GameDTO) tea.Cmd {
	return func() tea.Msg {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				panic(err)
			}
			data = bytes.TrimSpace(bytes.ReplaceAll(data, []byte("\n"), []byte(" ")))

			log.Printf("adding message to chan: %s", string(data))
			msg := ParseGameMessage(data)
			for _, m := range msg {
				c <- m
			}
		}
	}
}

func ReceiveMessage(sub chan *protocol.GameDTO) tea.Cmd {
	return func() tea.Msg {
		log.Println("Reading message from chan")
		msg := <-sub
		log.Println(msg)
		return msg
	}
}

func ParseGameMessage(msg []byte) []*protocol.GameDTO {
	decoder := json.NewDecoder(bytes.NewReader(msg))
	var messages []*protocol.GameDTO
	for decoder.More() {
		var data *protocol.GameDTO
		err := decoder.Decode(&data)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		log.Println("adding dto", data)
		messages = append(messages, data)
	}
	return messages
}

func (rm *RootModel) GameMessageToState(msg *protocol.GameDTO) {
	log.Println("Translating game to state")
	for i := 1; i < 6; i++ {
		player := rm.table.Players[i]
		if len(msg.Players) < i {
			break
		}
		receivedPlayer := msg.Players[i-1]
		if len(receivedPlayer.Hand.Cards) > 0 {
			player.Cards = []*Card{}
		}
		for _, card := range receivedPlayer.Hand.Cards {
			player.Cards = append(player.Cards, CardToCard(card))
		}
		player.Bet = receivedPlayer.Bet
		player.Value = receivedPlayer.Hand.Value
		player.Wallet = receivedPlayer.Wallet
		rm.table.Players[i] = player
	}
	dealer := rm.table.Players[0]
	if len(msg.DealerHand.Cards) > 0 {
		dealer.Cards = []*Card{}
	}
	for _, card := range msg.DealerHand.Cards {
		dealer.Cards = append(dealer.Cards, CardToCard(card))
	}
	dealer.Value = msg.DealerHand.Value
	rm.table.Players[0] = dealer
}

func CardToCard(pc protocol.CardDTO) *Card {
	card := &Card{}
	if strings.ToLower(pc.Suit) == "spade" {
		card.Suit = 0
	} else if strings.ToLower(pc.Suit) == "diamond" {
		card.Suit = 1
	} else if strings.ToLower(pc.Suit) == "heart" {
		card.Suit = 2
	} else if strings.ToLower(pc.Suit) == "club" {
		card.Suit = 3
	}
	card.Value = pc.Rank - 1

	return card
}
