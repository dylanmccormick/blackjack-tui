package client

import (
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Footer struct {
	commands map[string]string
	width    int
	height   int
}

func NewFooter() *Footer {
	return &Footer{
		commands: make(map[string]string),
		width:    78,
		height:   3,
	}
}

func (f *Footer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AddCommandsMsg:
		f.commands = msg.commands
	}
	return f, nil
}

func (f *Footer) View() string {
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(softForeground))
	views := []string{}
	for key, cmd := range f.commands {
		views = (append(views, fmt.Sprintf(" %s-%s ", key, cmd)))
	}
	slices.Sort(views)
	return lipgloss.Place(f.width, f.height, lipgloss.Center, lipgloss.Center, footerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, views...)))
}

func (f *Footer) Init() tea.Cmd {
	return nil
}

