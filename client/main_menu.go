package client

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	currMenuIndex int
	menuItems     []string
	Commands      map[string]string
	commandSet    bool
}

func NewMainMenu() *MainMenuModel {
	return &MainMenuModel{
		menuItems: []string{
			"Servers",
			"Tables",
			"Settings",
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
	if !mm.commandSet {
		cmds = append(cmds, AddCommands(mm.Commands))
		mm.commandSet = true
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			log.Printf("Got keypress enter: chainging to page %#v ", mm.menuItems[mm.currMenuIndex])
			var newPage mPage
			switch mm.menuItems[mm.currMenuIndex] {
			case "Servers":
				newPage = serverMenu
			case "Tables":
				newPage = tableMenu
			case "Settings":
				newPage = settingsMenu
			}
			log.Printf("New page: %#v", newPage)
			cmds = append(cmds, ChangeMenuPageCmd(newPage))
			mm.commandSet = false
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
