package client

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type (
	mPage     int
	MenuModel struct {
		currMenuIndex   int
		currServerIndex int
		currTableIndex  int
		availableTables []*protocol.TableDTO
		page            mPage
		menuItems       []string
	}
)

const (
	mainMenu mPage = iota
	serverMenu
	tableMenu
	settingsMenu
)

func NewMenuModel() *MenuModel {
	return &MenuModel{
		availableTables: []*protocol.TableDTO{
			{Id: "placeholder", Capacity: 5, CurrentPlayers: 0},
			{Id: "placeholder", Capacity: 5, CurrentPlayers: 0},
			{Id: "placeholder", Capacity: 5, CurrentPlayers: 0},
		},
		currMenuIndex:   0,
		currServerIndex: 0,
		currTableIndex:  0,
		page:            mainMenu,
		menuItems: []string{
			"Servers",
			"Tables",
			"Settings",
		},
	}
}

func (mm *MenuModel) Init() tea.Cmd {
	return nil
}

func (mm *MenuModel) View() string {
	// I should probably use a bubbles list
	switch mm.page {
	case mainMenu:
		return mm.MainMenuView()
	case tableMenu:
		return mm.TableView()
	}
	return ""
}

func (mm *MenuModel) MainMenuView() string {
	selectedTableStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
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

func (mm *MenuModel) TableView() string {
	items := []string{}
	selectedTableStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	for i, table := range mm.availableTables {
		if i == mm.currTableIndex {
			items = append(items, selectedTableStyle.Render(fmt.Sprintf("%d %s %d/%d\n", i, table.Id, table.CurrentPlayers, table.Capacity)))
		} else {
			items = append(items, fmt.Sprintf("%d %s %d/%d\n", i, table.Id, table.CurrentPlayers, table.Capacity))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (mm *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch mm.page {
	case mainMenu:
		return mm.UpdateMain(msg)
	case tableMenu:
		return mm.TableUpdate(msg)
	}
	return mm, nil
}

func (mm *MenuModel) TableUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	// for when the root model is on page lobby
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// this could be a command?
			log.Printf("Attempting to join table: %s", mm.availableTables[mm.currTableIndex].Id)
			cmds = append(cmds, SendData(protocol.PackageClientMessage(protocol.MsgJoinTable, mm.availableTables[mm.currTableIndex].Id)))
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "j":
				if mm.currTableIndex+1 < len(mm.availableTables) {
					mm.currTableIndex += 1
				}
				// lower the index on the room
			case "k":
				if mm.currTableIndex-1 >= 0 {
					mm.currTableIndex -= 1
				}
				// raise the index on the room
			}
		}
	case []*protocol.TableDTO:
		mm.TablesToState(msg)
	}

	return mm, tea.Batch(cmds...)
}

func SendData(msg *protocol.TransportMessage) tea.Cmd {
	return func() tea.Msg {
		return SendMsg{msg}
	}
}

func (mm *MenuModel) UpdateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case ChangeMenuPage:
		mm.page = msg.page
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// this could be a command?
			var newPage mPage
			switch mm.menuItems[mm.currMenuIndex] {
			case "Servers":
				newPage = serverMenu
			case "Tables":
				newPage = tableMenu
			case "Settings":
				newPage = settingsMenu
			}
			cmds = append(cmds, ChangeMenuPageCmd(newPage))
			// this will be a command to change the page
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "j":
				if mm.currMenuIndex+1 < len(mm.menuItems) {
					mm.currMenuIndex += 1
				}
				// lower the index on the room
			case "k":
				if mm.currMenuIndex-1 >= 0 {
					mm.currMenuIndex -= 1
				}
				// raise the index on the room
			}
		}
	}

	return mm, tea.Batch(cmds...)
}

type SendMsg struct {
	data *protocol.TransportMessage
}

func (mm *MenuModel) TablesToState(msg []*protocol.TableDTO) {
	log.Println("Translating tables to table list")
	mm.availableTables = msg
}

type ChangeMenuPage struct {
	page mPage
}

func ChangeMenuPageCmd(p mPage) tea.Cmd {
	return func() tea.Msg {
		return ChangeMenuPage{p}
	}
}
