package client

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HeaderModel struct {
	// Header at the top of the screen. Will display server info. Username Etc
	Username string
	State    string
	Width    int
	Height   int
}

const banner = `
 ____  _        _    ____ _  __   _   _    ____ _  __ 
| __ )| |      / \  / ___| |/ /  | | / \  / ___| |/ / 
|  _ \| |     / _ \| |   | ' /_  | |/ _ \| |   | ' /  
| |_) | |___ / ___ \ |___| . \ |_| / ___ \ |___| . \  
|____/|_____/_/   \_\____|_|\_\___/_/   \_\____|_|\_\ 
`

func NewHeader(height, width int) *HeaderModel {
	return &HeaderModel{
		Username: "not_logged_in",
		Width:    width,
		Height:   height,
	}
}

func (hm *HeaderModel) View() string {
	banner := hm.renderBanner()
	style := lipgloss.NewStyle().Width(hm.Width).Height(hm.Height)
	return style.Render(banner)
}

func (hm *HeaderModel) renderBanner() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(highlight)).
		Width(hm.Width / 2).
		Height(hm.Height).
		Align(lipgloss.Center)

	var sb strings.Builder

	sb.WriteString(banner)
	return style.Render(sb.String())
}

func (hm *HeaderModel) renderUserData() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(foreground)).
		Width(hm.Width / 4).
		Height(hm.Height).
		Align(lipgloss.Center)
	var sb strings.Builder
	sb.WriteString(hm.Username)
	return style.Render(sb.String())
}

func (hm *HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		hm.Width = (msg.Width - 6) / 2
	case AuthPollMsg:
		hm.Username = msg.UserName
	}
	return hm, cmd
}

func (hm *HeaderModel) Init() tea.Cmd {
	return nil
}
