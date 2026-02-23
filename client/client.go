package client

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
	"github.com/gorilla/websocket"
)

type (
	page int
)

type PageModel interface {
	tea.Model
	Commands() map[string]string
}

const (
	splashPage page = iota
	menuPage
	gamePage
)

type RootModel struct {
	page   page
	state  string
	width  int
	height int

	currPage PageModel

	mock bool

	transporter BackendClient
	wsMessages  <-chan *protocol.TransportMessage
	conn        *websocket.Conn

	table       tea.Model
	menuModel   tea.Model
	splashModel tea.Model
	footerModel tea.Model
	headerModel tea.Model
	leftBar     tea.Model
	rightBar    tea.Model
	pageMap     map[page]PageModel
}

var ROOT_COMMANDS = map[string]string{"ctrl+c": "quit"}

func (rm *RootModel) Init() tea.Cmd {
	rm.page = splashPage
	// commands := map[string]string{"ctrl+c": "quit"}
	return tea.Batch(
		rm.menuModel.Init(),
		tea.ClearScreen,
		ReceiveMessage(rm.wsMessages),
		ChangeRootPage(splashPage),
	)
}

type errMsg struct {
	err error
}

func (rm *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	// Root-Only Commands
	case ChangeRootPageMsg:
		rm.page = msg.page
	case SendMsg:
		cmds = append(cmds, rm.Send(msg.data))
	case StartWSMsg:
		rm.transporter.Connect()
		cmd := SendData(protocol.PackageClientMessage(protocol.MsgTableList, ""))
		cmds = append(cmds, cmd)
	case ReloadStatsMsg:
		cmd := SendData(protocol.PackageClientMessage(protocol.MsgGetStats, ""))
		cmds = append(cmds, cmd)

	// Current Page Commands
	case LoginRequested:
		cmds = append(cmds, rm.Login(msg.Url))
	case AuthLoginMsg:
		cmds = append(cmds, rm.CheckLoginStatus())

	// Universal commands (updating state and whatever)
	case AuthPollMsg:
		cmds = append(cmds, StartWSCmd())
		rm.menuModel, cmd = rm.menuModel.Update(msg)
		cmds = append(cmds, cmd)
		rm.table, cmd = rm.table.Update(msg)
		cmds = append(cmds, cmd)
	case tea.WindowSizeMsg:
		rm.width = msg.Width - 1
		rm.height = msg.Height - 1
		slog.Info("Sizes:", "t_width", msg.Width, "t_height", msg.Height, "rootwidth", rm.width, "rootheight", rm.height)
	case tea.KeyMsg:
		// Top Level Keys. Kill the program type keys
		switch msg.Type {
		case tea.KeyCtrlC:
			return rm, tea.Quit
		case tea.KeyRunes:
			switch string(msg.Runes) {
			}
		}
	}
	switch rm.page {
	case menuPage:
		rm.menuModel, cmd = rm.menuModel.Update(msg)
		cmds = append(cmds, cmd)
	case splashPage:
		rm.splashModel, cmd = rm.splashModel.Update(msg)
		cmds = append(cmds, cmd)
	case gamePage:
		rm.table, cmd = rm.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	// header and footer always get updated since they don't handle keypresses
	rm.footerModel, cmd = rm.footerModel.Update(msg)
	cmds = append(cmds, cmd)
	rm.headerModel, cmd = rm.headerModel.Update(msg)
	cmds = append(cmds, cmd)
	rm.leftBar, cmd = rm.leftBar.Update(msg)
	cmds = append(cmds, cmd)
	rm.rightBar, cmd = rm.rightBar.Update(msg)
	cmds = append(cmds, cmd)
	return rm, tea.Batch(append(cmds, ReceiveMessage(rm.wsMessages))...)
}

func NewRootModel(tmio BackendClient) *RootModel {
	wsChan := tmio.GetChan()
	rm := &RootModel{
		height:      60,
		width:       200,
		transporter: tmio,
		wsMessages:  wsChan,
		table:       NewTable(20, 80),
		menuModel:   NewMenuModel(),
		splashModel: NewSplashModel(),
		footerModel: NewFooter(3, 78),
		headerModel: NewHeader(6, 200),
		state:       "menu",
		leftBar:     NewLeftBar(0, 0),
		rightBar:    NewRightBar(0, 0),
	}

	return rm
}

func (rm *RootModel) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(rm.width-6).
		Height(rm.height-6).
		Margin(3, 3)
	var mainView string
	switch rm.page {
	case menuPage:
		mainView = rm.menuModel.View()
	case gamePage:
		mainView = rm.table.View()
	case splashPage:
		mainView = rm.splashModel.View()
		// return style.Render(mainView)
	}
	mainViewStyle := lipgloss.NewStyle().
		Width(((rm.width-6)/2)-2).
		Height((rm.height*2/3)-2).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder())

	view := lipgloss.JoinHorizontal(lipgloss.Left, rm.leftBar.View(), lipgloss.JoinVertical(lipgloss.Top, rm.headerModel.View(), mainViewStyle.Render(mainView), rm.footerModel.View()), rm.rightBar.View())
	fullView := lipgloss.Place(rm.width/2, rm.height*3/4, lipgloss.Center, lipgloss.Center, view)
	return style.Render(fullView)
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

type TuiPlayer struct {
	Name    string
	Cards   []*Card
	Value   int
	Wallet  int
	Bet     int
	Current bool
}

func RunTui(mock bool) {
	var rm *RootModel
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
	defer f.Close()
	if mock {
		slog.Debug("running in mock mode")
		rm = NewRootModel(NewMockTransporter())
	} else {
		slog.Debug("running in LIVE mode")
		rm = NewRootModel(NewWsBackendClient())
	}
	p := tea.NewProgram(rm)
	if _, err := p.Run(); err != nil {
		slog.Error("Error running BubbleTea application", "error", err)
		os.Exit(1)
	}
}
