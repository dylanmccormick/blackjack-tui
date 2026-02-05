package client

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type RightBar struct {
	Height   int
	Width    int
	Messages []protocol.PopUpDTO
}

func NewRightBar(height, width int) *RightBar {
	return &RightBar{
		Height:   height,
		Width:    width,
		Messages: []protocol.PopUpDTO{},
	}
}

func (rb *RightBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		rb.Width = (msg.Width - 8) / 4
		rb.Height = msg.Height * 2 / 3
	case protocol.PopUpDTO:
		rb.Messages = append(rb.Messages, msg)
		cmd = PopUpTimer()
	case PopUpRemoveMsg:
		if len(rb.Messages) > 0 {
			rb.Messages = rb.Messages[1:]
		}
	}

	return rb, cmd
}

func (rb *RightBar) renderMultiplePopUps() string {
	maxPopUps := rb.Height / 8
	var color string
	views := []string{}
	for i, popUp := range rb.Messages {
		if i > maxPopUps {
			break
		}
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
		views = append(views, rb.renderPopupMessage(msg, color))
	}
	return lipgloss.JoinVertical(lipgloss.Top, views...)
}

func (rb *RightBar) renderPopupMessage(message, color string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(foreground)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Width((rb.Width) - 4).
		Height(6).
		Align(lipgloss.Center)

	var sb strings.Builder
	sb.WriteString(message)
	return style.Render(sb.String())
}

func (rb *RightBar) View() string {
	style := lipgloss.NewStyle().Width(rb.Width).Height(rb.Height).Align(lipgloss.Right)
	messages := rb.renderMultiplePopUps()
	return style.Render(messages)
}

func (lb *RightBar) Init() tea.Cmd {
	return nil
}

type LeftBar struct {
	Height int
	Width  int
}

func NewLeftBar(height, width int) *LeftBar {
	return &LeftBar{
		Height: height,
		Width:  width,
	}
}

func (lb *LeftBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		lb.Width = (msg.Width - 8) / 4
		lb.Height = msg.Height * 2 / 3
	}

	return lb, cmd
}

func (lb *LeftBar) View() string {
	style := lipgloss.NewStyle().Width(lb.Width).Height(lb.Height)
	return style.Render("THIS IS THE LEFT BAR. IN LEFT BAR WE TRUST")
}

func (lb *LeftBar) Init() tea.Cmd {
	return nil
}
