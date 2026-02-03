package client

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type HeaderModel struct {
	// Header at the top of the screen. Will display server info. Username Etc
	Username string
	State    string
	Width    int
	Height   int
	Messages []protocol.PopUpDTO
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
	messages := hm.renderMultiplePopUps()
	return lipgloss.JoinHorizontal(lipgloss.Left, userData, banner, messages)
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

func (hm *HeaderModel) renderMultiplePopUps() string {
	var color string
	views := []string{}
	for _, popUp := range hm.Messages {
		msg := popUp.Message
		lvl := popUp.Type
		switch lvl {
		case "error":
			color = popUpErr
		case "warn":
			color = popUpWarn
		default:
			color = popUpInfo
		}
		views = append(views, hm.renderPopupMessage(msg, color))
	}
	return lipgloss.JoinVertical(lipgloss.Top, views...)
}

func (hm *HeaderModel) renderPopupMessage(message, color string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(foreground)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Width((hm.Width / 4) - 2).
		Height(hm.Height - 2).
		Align(lipgloss.Center)

	var sb strings.Builder
	sb.WriteString(message)
	return style.Render(sb.String())
}

func (hm *HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case AuthPollMsg:
		hm.Username = msg.UserName
	case protocol.PopUpDTO:
		hm.Messages = append(hm.Messages, msg)
		cmd = PopUpTimer()
	case PopUpRemoveMsg:
		if len(hm.Messages) > 0 {
			hm.Messages = hm.Messages[1:]
		}
	}
	return hm, cmd
}

func (hm *HeaderModel) Init() tea.Cmd {
	return nil
}
