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

func NewHeader() *HeaderModel {
	return &HeaderModel{
		Username: "not_logged_in",
		Width:    120,
		Height:   6,
	}
}

func (hm *HeaderModel) View() string {
	banner := hm.renderBanner()
	userData := hm.renderUserData()
	message := hm.renderPopupMessage()
	return lipgloss.JoinHorizontal(lipgloss.Left, userData, banner, message)
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

func (hm *HeaderModel) renderPopupMessage() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(foreground)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E69875")).
		Width((hm.Width / 4) - 2).
		Height(hm.Height - 2).
		Align(lipgloss.Center)

	var sb strings.Builder
	sb.WriteString("This is a popup message placeholder")
	return style.Render(sb.String())
}

func (hm *HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AuthPollMsg:
		hm.Username = msg.UserName
	}
	return hm, nil
}

func (hm *HeaderModel) Init() tea.Cmd {
	return nil
}

