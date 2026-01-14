package client

import (
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/gorilla/websocket"
)

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

type RootModel struct {
	page   page
	state  state
	width  int
	height int

	transporter TransportMessageIO
	wsMessages  <-chan *protocol.TransportMessage
	conn        *websocket.Conn

	table       tea.Model
	menuModel   tea.Model
	loginModel  tea.Model
	footerModel tea.Model
}

var ROOT_COMMANDS = map[string]string{"ctrl+c": "quit"}

func (rm *RootModel) Init() tea.Cmd {
	rm.page = loginPage
	commands := map[string]string{"ctrl+c": "quit"}
	return tea.Batch(
		rm.menuModel.Init(),
		tea.ClearScreen,
		ReceiveMessage(rm.wsMessages),
		AddCommands(commands),
	)
}

type errMsg struct {
	err error
}

func (rm *RootModel) Send(data *protocol.TransportMessage) tea.Cmd {
	return func() tea.Msg {
		err := rm.transporter.SendData(data)
		if err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (rm *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ChangeRootPageMsg:
		rm.page = msg.page
	case SendMsg:
		cmds = append(cmds, rm.Send(msg.data))
	case tea.KeyMsg:
		// Top Level Keys. Kill the program type keys
		switch msg.Type {
		case tea.KeyCtrlC:
			return rm, tea.Quit
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "c":
				// todo... turn this into a command!
				rm.transporter.Connect()
				cmd := SendData(protocol.PackageClientMessage(protocol.MsgTableList, ""))
				cmds = append(cmds, cmd)
			}
		}
	case tea.WindowSizeMsg:
		rm.width = msg.Width
		rm.height = msg.Height
	}
	switch rm.page {
	case menuPage:
		rm.menuModel, cmd = rm.menuModel.Update(msg)
		cmds = append(cmds, cmd)
	case loginPage:
		rm.loginModel, cmd = rm.loginModel.Update(msg)
		cmds = append(cmds, cmd)
	case gamePage:
		rm.table, cmd = rm.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	rm.footerModel, cmd = rm.footerModel.Update(msg)
	return rm, tea.Batch(append(cmds, ReceiveMessage(rm.wsMessages))...)
}

func NewRootModel(tmio TransportMessageIO) *RootModel {
	wsChan := tmio.GetChan()
	return &RootModel{
		transporter: tmio,
		wsMessages:  wsChan,
		table:       NewTable(),
		menuModel:   NewMenuModel(),
		loginModel:  &LoginModel{},
		footerModel: NewFooter(),
	}
}

func (rm *RootModel) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(80).
		Height(30).
		Margin(((rm.height - 25) / 2), ((rm.width - 80) / 2))
	var mainView string
	switch rm.page {
	case menuPage:
		mainView = rm.menuModel.View()
	case gamePage:
		mainView = rm.table.View()
	case loginPage:
		mainView = rm.loginModel.View()
		return style.Render(mainView)
	}
	bannerHeight := 5
	mainViewStyle := lipgloss.NewStyle().
		Width(80-2).
		Height(25-bannerHeight).
		Align(lipgloss.Center, lipgloss.Center)
	fullView := lipgloss.Place(80, 25, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Top, rm.RenderHeader(), mainViewStyle.Render(mainView), rm.footerModel.View()))
	return style.Render(fullView)
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

func RunTui(mock *bool) {
	var rm *RootModel
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}
	if *mock {
		log.Println("running in mock mode")
		rm = NewRootModel(NewMockTransporter())
	} else {
		log.Println("running in LIVE mode")
		rm = NewRootModel(NewWsTransportMessageIO())
	}
	p := tea.NewProgram(rm)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there has been an error: %v", err)
		os.Exit(1)
	}
}

func ChangeRootPage(p page) tea.Cmd {
	return func() tea.Msg {
		return ChangeRootPageMsg{p}
	}
}

type ChangeRootPageMsg struct {
	page page
}

func (rm *RootModel) RenderHeader() string {
	var sb strings.Builder
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(highlight)).
		Width(80).
		Align(lipgloss.Center)

	sb.WriteString(banner)

	return style.Render(sb.String())
}

const banner = `
 ____  _        _    ____ _  __   _   _    ____ _  __ 
| __ )| |      / \  / ___| |/ /  | | / \  / ___| |/ / 
|  _ \| |     / _ \| |   | ' /_  | |/ _ \| |   | ' /  
| |_) | |___ / ___ \ |___| . \ |_| / ___ \ |___| . \  
|____/|_____/_/   \_\____|_|\_\___/_/   \_\____|_|\_\ 
`
