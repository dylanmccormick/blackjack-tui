package client

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SplashModel struct {
	Height   int
	Width    int
	Commands map[string]string
}

func NewSplashModel() *SplashModel {
	return &SplashModel{
		0,
		0,
		map[string]string{"enter": "continue"},
	}
}

func (sm *SplashModel) Init() tea.Cmd {
	slog.Info("INITING YOUR MOTHER")
	return AddCommands(sm.Commands)
}

func (sm *SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	// var cmd tea.Cmd
	switch msg := msg.(type) {
	case ChangeRootPageMsg:
		cmds = append(cmds, AddCommands(sm.Commands))
	case tea.WindowSizeMsg:
		sm.Height = (msg.Height * 3 / 4) - 6
		sm.Width = (msg.Width-6)/2 - 4
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// this could be a command?
			cmds = append(cmds, ChangeRootPage(menuPage))
		}
	}
	return sm, tea.Batch(cmds...)
}

func (sm *SplashModel) View() string {
	return sm.renderSplash()
}

func (sm *SplashModel) renderSplash() string {
	box := lipgloss.NewStyle().
		Height(sm.Height).
		Width(sm.Width).
		Border(lipgloss.DoubleBorder()).
		Align(lipgloss.Center, lipgloss.Top)

	textStyle := lipgloss.NewStyle().
		Height(sm.Height-2*Height).
		Width(sm.Width/8*6).
		Align(lipgloss.Center, lipgloss.Center).
		MarginLeft(sm.Width / 8).
		MarginRight(sm.Width / 8)

	cardStyle := lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(sm.Width)

	var cards []string
	suits := []int{0, 1, 3, 2}
	for i := range sm.Width / Width {
		var val int
		var suit int
		if i%2 == 0 {
			val = 0
		} else {
			val = 12
		}
		suit = i % 4
		card := &Card{val, suits[suit]}
		cards = append(cards, card.View())

	}
	cardsView := lipgloss.JoinHorizontal(lipgloss.Left, cards...)

	welcome := "Welcome to Blackjack TUI. A project by @dylanmccormick. Have fun gambling with fake money in the terminal! Follow the instructions on the bottom of the screen to interact"

	thankYou := "Thank you for checking out my project. Hopefully this will help you feed your terminal-based gambling addiction. Come back for more coins tomorrow!"

	text := welcome + "\n\n" + thankYou

	totalView := lipgloss.JoinVertical(lipgloss.Top, cardStyle.Render(cardsView), textStyle.Render(text), cardsView)

	return box.Render(totalView)
}
