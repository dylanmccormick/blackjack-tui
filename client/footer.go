package client

import (
	"fmt"
	"log"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Footer struct {
	commands map[string]string
	Width    int
	Height   int
}

func NewFooter(height, width int) *Footer {
	return &Footer{
		commands: make(map[string]string),
		Width:    height,
		Height:   width,
	}
}

func (f *Footer) Resize(height, width int) {
	f.Height = height
	f.Width = width
}

func (f *Footer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AddCommandsMsg:
		f.commands = msg.commands
	case tea.WindowSizeMsg:
		f.Height = 3
		f.Width = msg.Width / 2
		log.Printf("Terminal: %dx%d, My layout uses: %dx%d\n", msg.Width, msg.Height, f.Width, f.Height)
	}
	return f, nil
}

func (f *Footer) View() string {
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(highlight))
	views := []string{}
	for key, cmd := range f.commands {
		views = (append(views, fmt.Sprintf(" %s-%s ", key, cmd)))
	}
	slices.Sort(views)
	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, footerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, views...)))
}

func (f *Footer) Init() tea.Cmd {
	return nil
}
