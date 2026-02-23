package client

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ServerMenuModel struct {
	textInput       textinput.Model
	savedServers    []string
	currServerIndex int
	Commands        map[string]string
	commandSet      bool
	loginMenu       *LoginMenu
	state           string
}

func NewServerMenu() *ServerMenuModel {
	serverText := textinput.New()
	serverText.Placeholder = "http://localhost:3030"
	serverText.Width = 40
	return &ServerMenuModel{
		textInput: serverText,
		Commands: map[string]string{
			"j":     "down",
			"k":     "up",
			"enter": "select",
			"esc":   "back",
			"n":     "new server",
		},
		loginMenu:    &LoginMenu{userCodePage, "", "", make(map[string]string)},
		savedServers: []string{"http://blackjack.dylanjmccormick.com", "http://localhost:42069"},
	}
}

func (sm *ServerMenuModel) Init() tea.Cmd {
	return nil
}

func (sm *ServerMenuModel) View() string {
	if sm.state == "login" {
		return sm.loginMenu.View()
	}
	selectedServerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(highlight))
	view := []string{}
	for i, menuItem := range sm.savedServers {
		if i == sm.currServerIndex {
			view = append(view, selectedServerStyle.Render(fmt.Sprintf("%s\n", menuItem)))
		} else {
			view = append(view, fmt.Sprintf("%s\n", menuItem))
		}
	}
	view = append(view, sm.textInput.View())
	return lipgloss.JoinVertical(lipgloss.Left, view...)
}

func (sm *ServerMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	if !sm.commandSet {
		cmds = append(cmds, AddCommands(sm.Commands))
		sm.commandSet = true
	}
	switch msg := msg.(type) {
	case LoginRequested:
		sm.state = "login"
		sm.loginMenu.currentUrl = sm.savedServers[sm.currServerIndex]
	case TextFocusMsg:
		sm.textInput.Focus()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if sm.textInput.Focused() {
				sm.textInput.Blur()
			} else {
				cmds = append(cmds, ChangeMenuPageCmd(mainMenu))
				sm.commandSet = false
			}
		case tea.KeyEnter:
			var server string
			if sm.textInput.Focused() {
				server := sm.textInput.Value()
				sm.savedServers = append(sm.savedServers, server)
				cmds = append(cmds, AddCommands(sm.Commands))
			} else if sm.state != "login" {
				if len(sm.savedServers) > 0 {
					server = sm.savedServers[sm.currServerIndex]
					// TODO: join a saved server
					cmds = append(cmds, RequestLogin(server))
					slog.Info("Attempting to join server", "server", server)
				}
			}
			// This will be to join a server
			cmds = append(cmds, cmd)
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "n":
				cmd = TextFocusCmd()
				cmds = append(cmds, cmd)
				cmds = append(cmds, AddCommands(map[string]string{"enter": "create server address", "esc": "cancel"}))
			case "j":
				if sm.currServerIndex+1 < len(sm.savedServers) {
					sm.currServerIndex += 1
				}
			case "k":
				if sm.currServerIndex-1 >= 0 {
					sm.currServerIndex -= 1
				}
			}
		}
	}
	if sm.state == "login" {
		sm.loginMenu, cmd = sm.loginMenu.Update(msg)
		cmds = append(cmds, cmd)
	}

	if sm.textInput.Focused() {
		sm.textInput, cmd = sm.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return sm, tea.Batch(cmds...)
}
