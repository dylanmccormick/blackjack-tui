package client

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ServerMenuModel struct {
	textInput       textinput.Model
	savedServers    []string
	currServerIndex int
}

func NewServerMenu() *ServerMenuModel {
	serverText := textinput.New()
	serverText.Placeholder = "http://localhost:3030"
	serverText.Width = 40
	return &ServerMenuModel{
		textInput: serverText,
	}
}

func (sm *ServerMenuModel) Init() tea.Cmd {
	return nil
}

func (sm *ServerMenuModel) View() string {
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
	switch msg := msg.(type) {
	case TextFocusMsg:
		sm.textInput.Focus()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if sm.textInput.Focused() {
				sm.textInput.Blur()
			} else {
				cmds = append(cmds, ChangeMenuPageCmd(mainMenu))
			}
		case tea.KeyEnter:
			// This will be to join a server
			cmd = nil
			cmds = append(cmds, cmd)
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "n":
				cmd = TextFocusCmd()
				cmds = append(cmds, cmd)
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

	if sm.textInput.Focused() {
		sm.textInput, cmd = sm.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return sm, tea.Batch(cmds...)
}
