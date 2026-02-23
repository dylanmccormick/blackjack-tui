package client

import (
	"log/slog"

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
		StatsMenu    tea.Model
		Height       int
		Width        int
		commands     map[string]string
	}
)

const (
	mainMenu mPage = iota
	serverMenu
	tableMenu
	settingsMenu
	statsMenu
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
		TableMenu:     NewTableMenu(0, 0), // TODO
		SettingsMenu:  NewSettingsMenu(),
		StatsMenu:     NewStatsMenu(),
	}
}

func (mm *MenuModel) Commands() map[string]string {
	return mm.commands
}

func (mm *MenuModel) Init() tea.Cmd {
	// commands := map[string]string{
	// 	"j":     "down",
	// 	"k":     "up",
	// 	"enter": "select",
	// 	"esc":   "back",
	// }
	return nil
}

func (mm *MenuModel) View() string {
	// I should probably use a bubbles list
	viewStyle := lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Height(mm.Height).Width(mm.Width)
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
	case statsMenu:
		view = mm.StatsMenu.View()
	}
	return viewStyle.Render(view)
}

func (mm *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case ChangeMenuPage:
		slog.Debug("changing menu page", "page", msg.page)
		mm.page = msg.page
	case tea.WindowSizeMsg:
		mm.Height = (msg.Height * 2 / 3)
		mm.Width = (msg.Width - 6) / 2
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
	case statsMenu:
		mm.StatsMenu, cmd = mm.StatsMenu.Update(msg)
		cmds = append(cmds, cmd)
	}
	return mm, tea.Batch(cmds...)
}
