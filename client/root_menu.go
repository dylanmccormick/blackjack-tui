package client

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type (
	mPage     int
	MenuModel struct {
		currMenuIndex   int
		availableTables []*protocol.TableDTO
		page            mPage

		// Menu Pages
		MainMenu     tea.Model
		ServerMenu   tea.Model
		TableMenu    tea.Model
		SettingsMenu tea.Model
		CommandsSet  bool
	}
)

const (
	mainMenu mPage = iota
	serverMenu
	tableMenu
	settingsMenu
)

var dialogueCommands = map[string]string{
	"enter": "enter",
	"esc":   "cancel",
}

func NewMenuModel() *MenuModel {
	return &MenuModel{
		currMenuIndex: 0,
		page:          mainMenu,
		MainMenu:      NewMainMenu(),
		ServerMenu:    NewServerMenu(),
		TableMenu:     NewTableMenu(),
		SettingsMenu:  NewSettingsMenu(),
		CommandsSet:   false,
	}
}

func (mm *MenuModel) Init() tea.Cmd {
	commands := map[string]string{
		"j":     "down",
		"k":     "up",
		"enter": "select",
		"esc":   "back",
	}
	return AddCommands(commands)
}

func (mm *MenuModel) View() string {
	// I should probably use a bubbles list
	viewStyle := lipgloss.NewStyle().Width(40).Align(lipgloss.Center, lipgloss.Center)
	var view string
	switch mm.page {
	case mainMenu:
		view = mm.MainMenu.View()
	case tableMenu:
		view = mm.TableMenu.View()
	case serverMenu:
		view = mm.ServerMenu.View()
	case settingsMenu:
		view = mm.SettingsMenu.View()
	}
	return viewStyle.Render(view)
}

func (mm *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ChangeMenuPage:
		log.Printf("Changing menu page: %#v", msg.page)
		mm.page = msg.page
		mm.CommandsSet = false
	}
	switch mm.page {
	case mainMenu:
		mm.MainMenu, cmd = mm.MainMenu.Update(msg)
		cmds = append(cmds, cmd)
	case tableMenu:
		mm.TableMenu, cmd = mm.TableMenu.Update(msg)
		cmds = append(cmds, cmd)
	case serverMenu:
		mm.ServerMenu, cmd = mm.ServerMenu.Update(msg)
		cmds = append(cmds, cmd)
	case settingsMenu:
		mm.SettingsMenu, cmd = mm.SettingsMenu.Update(msg)
		cmds = append(cmds, cmd)
	}
	return mm, tea.Batch(cmds...)
}

func SendData(msg *protocol.TransportMessage) tea.Cmd {
	return func() tea.Msg {
		return SendMsg{msg}
	}
}

type SendMsg struct {
	data *protocol.TransportMessage
}

type ChangeMenuPage struct {
	page mPage
}

func ChangeMenuPageCmd(p mPage) tea.Cmd {
	return func() tea.Msg {
		return ChangeMenuPage{p}
	}
}

type TextFocusMsg struct{}

func TextFocusCmd() tea.Cmd {
	return func() tea.Msg {
		return TextFocusMsg{}
	}
}
