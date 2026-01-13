package client

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Footer struct {
	commands map[string]string
}

func NewFooter() *Footer {
	return &Footer{
		commands: make(map[string]string),
	}
}

type AddCommandsMsg struct {
	commands map[string]string
}

func (f *Footer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AddCommandsMsg:
		f.commands = msg.commands
	}
	return f, nil
}

func (f *Footer) View() string {
	views := []string{}
	for key, cmd := range f.commands {
		views = (append(views, fmt.Sprintf("%s - %s", key, cmd)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

func (f *Footer) Init() tea.Cmd {
	return nil
}

func AddCommands(cmds map[string]string) tea.Cmd {
	return func() tea.Msg {
		return AddCommandsMsg{cmds}
	}
}
