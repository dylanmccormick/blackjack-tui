package client

import (
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

type page int

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

type MenuModel struct {
	currTableIndex  int
	availableTables []*protocol.TableDTO
}

type RootModel struct {
	page   page
	state  state
	width  int
	height int

	wsMessages chan *protocol.TransportMessage
	conn       *websocket.Conn

	table     *TuiTable
	menuModel *MenuModel
}

func (mm *MenuModel) Init() tea.Cmd {
	return nil
}

func (rm *RootModel) Init() tea.Cmd {
	data, err := protocol.PackageClientMessage(protocol.MsgTableList, "")
	if err != nil {
		log.Print(err)
	}
	err = rm.conn.WriteJSON(data)
	if err != nil {
		log.Print(err)
	}
	return tea.Batch(
		rm.menuModel.Init(),
		tea.ClearScreen,
		listenForMessages(rm.conn, rm.wsMessages),
		ReceiveMessage(rm.wsMessages),
	)
}

func (rm *RootModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	// for when the root model is on page lobby
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "j":
				// lower the index on the room
			case "k":
				// raise the index on the room
			case "enter":
				// join the selected room
			}
		}
	}

	return rm, ReceiveMessage(rm.wsMessages)
}

func (rm *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("Running update on message: %#v", msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return rm, tea.Quit
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "b":
				data, err := protocol.PackageClientMessage(protocol.MsgPlaceBet, "5")
				if err != nil {
					log.Print(err)
				}
				err = rm.conn.WriteJSON(data)
				if err != nil {
					log.Print(err)
				}
				// bet 5
			case "h":
				data, err := protocol.PackageClientMessage(protocol.MsgHit, "")
				if err != nil {
					log.Print(err)
				}
				err = rm.conn.WriteJSON(data)
				if err != nil {
					log.Print(err)
				}
				// hit
			case "s":
				// stand
				data, err := protocol.PackageClientMessage(protocol.MsgStand, "")
				if err != nil {
					log.Print(err)
				}
				err = rm.conn.WriteJSON(data)
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
	case []*protocol.TableDTO:
		rm.TablesToState(msg)
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
		wsMessages: make(chan *protocol.TransportMessage),
		conn:       c,
		table:      NewTable(),
		menuModel:  &MenuModel{availableTables: []*protocol.TableDTO{{Id: "placeholder", Capacity: 5, CurrentPlayers: 0}}},
	}
}

func (rm *RootModel) View() string {
	tables := rm.ViewTables()
	view := lipgloss.JoinHorizontal(lipgloss.Left, tables, rm.table.View())
	return lipgloss.Place(rm.width, rm.height, lipgloss.Center, lipgloss.Center, view)
}

func (rm *RootModel) ViewTables() string {
	view := ""
	for i, table := range rm.menuModel.availableTables {
		view += fmt.Sprintf("%d %s %d/%d\n", i, table.Id, table.CurrentPlayers, table.Capacity)
	}
	return view
}

type TuiPlayer struct {
	Name   string
	Cards  []*Card
	Value  int
	Wallet int
	Bet    int
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
			player.Value = receivedPlayer.Hand.Value
		}
		for _, card := range receivedPlayer.Hand.Cards {
			player.Cards = append(player.Cards, CardToCard(card))
		}
		player.Bet = receivedPlayer.Bet
		player.Wallet = receivedPlayer.Wallet
		player.Name = receivedPlayer.Name
		rm.table.Players[i] = player
	}
	dealer := rm.table.Players[0]
	if len(msg.DealerHand.Cards) > 0 {
		dealer.Cards = []*Card{}
	}
	for _, card := range msg.DealerHand.Cards {
		dealer.Cards = append(dealer.Cards, CardToCard(card))
		dealer.Value = msg.DealerHand.Value
	}
	rm.table.Players[0] = dealer
}

func (rm *RootModel) TablesToState(msg []*protocol.TableDTO) {
	log.Println("Translating tables to table list")
	rm.menuModel.availableTables = msg
}
