package client

import (
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	currMenuIndex int
	menuItems     []string
	Commands      map[string]string
}

func NewMainMenu() *MainMenuModel {
	return &MainMenuModel{
		menuItems: []string{
			"Servers",
			"Tables",
			// "Settings",
			"Stats",
		},
		Commands: map[string]string{
			"j":     "down",
			"k":     "up",
			"enter": "select",
			"esc":   "back",
		},
	}
}

func (mm *MainMenuModel) Init() tea.Cmd {
	return nil
}

func (mm *MainMenuModel) View() string {
	selectedTableStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(highlight))
	view := []string{}
	for i, menuItem := range mm.menuItems {
		if i == mm.currMenuIndex {
			view = append(view, selectedTableStyle.Render(fmt.Sprintf("%s\n", menuItem)))
		} else {
			view = append(view, fmt.Sprintf("%s\n", menuItem))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, view...)
}

func (mm *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case ChangeMenuPage:
		cmds = append(cmds, AddCommands(mm.Commands))
	case ChangeRootPageMsg:
		cmds = append(cmds, AddCommands(mm.Commands))
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			slog.Debug("Got keypress enter: changing page", "target", mm.menuItems[mm.currMenuIndex])
			var newPage mPage
			switch mm.menuItems[mm.currMenuIndex] {
			case "Servers":
				newPage = serverMenu
			case "Tables":
				newPage = tableMenu
			case "Settings":
				newPage = settingsMenu
			case "Stats":
				newPage = statsMenu
			}
			slog.Debug("New page:",  "page", newPage)
			cmds = append(cmds, ChangeMenuPageCmd(newPage))
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "j":
				if mm.currMenuIndex+1 < len(mm.menuItems) {
					mm.currMenuIndex += 1
				}
			case "k":
				if mm.currMenuIndex-1 >= 0 {
					mm.currMenuIndex -= 1
				}
			}
		}
	}

	return mm, tea.Batch(cmds...)
}
