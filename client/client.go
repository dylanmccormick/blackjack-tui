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

	mock bool

	transporter BackendClient
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

func (rm *RootModel) Login(url string) tea.Cmd {
	return func() tea.Msg {
		return rm.transporter.StartAuth(url)
	}
}

func (rm *RootModel) CheckLoginStatus() tea.Cmd {
	return func() tea.Msg {
		return rm.transporter.PollAuth()
	}
}

func (rm *RootModel) Send(data *protocol.TransportMessage) tea.Cmd {
	return func() tea.Msg {
		rm.transporter.QueueData(data)
		return nil
	}
}

func (rm *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case LoginRequested:
		cmds = append(cmds, rm.Login(msg.Url))
	case AuthLoginMsg:
		cmds = append(cmds, rm.CheckLoginStatus())
	case AuthPollMsg:
		cmds = append(cmds, StartWSCmd())
	case ChangeRootPageMsg:
		rm.page = msg.page
	case SendMsg:
		cmds = append(cmds, rm.Send(msg.data))
	case StartWSMsg:
		rm.transporter.Connect()
		cmd := SendData(protocol.PackageClientMessage(protocol.MsgTableList, ""))
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		// Top Level Keys. Kill the program type keys
		switch msg.Type {
		case tea.KeyCtrlC:
			return rm, tea.Quit
		case tea.KeyRunes:
			switch string(msg.Runes) {
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

func NewRootModel(tmio BackendClient) *RootModel {
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

func RunTui(mock *bool) {
	var rm *RootModel
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()
	if *mock {
		log.Println("running in mock mode")
		rm = NewRootModel(NewMockTransporter())
	} else {
		log.Println("running in LIVE mode")
		rm = NewRootModel(NewWsBackendClient())
	}
	p := tea.NewProgram(rm)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there has been an error: %v", err)
		os.Exit(1)
	}
}

type StartWSMsg struct{}

func StartWSCmd() tea.Cmd {
	return func() tea.Msg {
		return StartWSMsg{}
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

type AuthLoginMsg struct {
	UserCode  string
	Url       string
	SessionId string
}

type AuthPollMsg struct {
	Authenticated bool
}
