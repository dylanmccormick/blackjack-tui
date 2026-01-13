package client

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
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
		serverTextInput textinput.Model
		tableTextInput  textinput.Model
		savedServers    []string
	}
)

const (
	mainMenu mPage = iota
	serverMenu
	tableMenu
	settingsMenu
)

func NewMenuModel() *MenuModel {
	serverText := textinput.New()
	serverText.Placeholder = "http://localhost:3030"
	serverText.Width = 40

	tableText := textinput.New()
	tableText.Placeholder = "my_cool_table"
	tableText.Width = 40

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
		serverTextInput: serverText,
		tableTextInput:  tableText,
		savedServers: []string{
			"localhost:8080",
			"localhost:3030",
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
	case serverMenu:
		return mm.ServerView()
	case settingsMenu:
		return mm.ViewSettings()
	}
	return ""
}

func (mm *MenuModel) MainMenuView() string {
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

func (mm *MenuModel) TableView() string {
	items := []string{}
	selectedTableStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(highlight))
	for i, table := range mm.availableTables {
		if i == mm.currTableIndex {
			items = append(items, selectedTableStyle.Render(fmt.Sprintf("%d %s %d/%d\n", i, table.Id, table.CurrentPlayers, table.Capacity)))
		} else {
			items = append(items, fmt.Sprintf("%d %s %d/%d\n", i, table.Id, table.CurrentPlayers, table.Capacity))
		}
	}
	items = append(items, mm.tableTextInput.View())
	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (mm *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch mm.page {
	case mainMenu:
		return mm.UpdateMain(msg)
	case tableMenu:
		return mm.TableUpdate(msg)
	case serverMenu:
		return mm.UpdateServer(msg)
	case settingsMenu:
		return mm.UpdateSettings(msg)
	}
	return mm, nil
}

func (mm *MenuModel) TableUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	// for when the root model is on page lobby
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case TextFocusMsg:
		mm.tableTextInput.Focus()
	case ChangeMenuPage:
		mm.page = msg.page
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if mm.tableTextInput.Focused() {
				mm.tableTextInput.Blur()
			} else {
				cmds = append(cmds, ChangeMenuPageCmd(mainMenu))
			}
		case tea.KeyEnter:
			// this could be a command?
			var tableName string
			if mm.tableTextInput.Focused() {
				tableName = mm.tableTextInput.Value()
				cmd = SendData(protocol.PackageClientMessage(protocol.MsgCreateTable, tableName))
				cmds = append(cmds, cmd)
			} else {
				tableName = mm.availableTables[mm.currTableIndex].Id
				cmd = SendData(protocol.PackageClientMessage(protocol.MsgJoinTable, tableName))
				cmds = append(cmds, cmd)
				log.Printf("Attempting to join table: %s", tableName)
				cmd = ChangeRootPage(gamePage)
				cmds = append(cmds, cmd)
			}
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "n":
				cmd = TextFocusCmd()
				cmds = append(cmds, cmd)
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

	if mm.tableTextInput.Focused() {
		mm.tableTextInput, cmd = mm.tableTextInput.Update(msg)
		cmds = append(cmds, cmd)
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

type TextFocusMsg struct{}

func TextFocusCmd() tea.Cmd {
	return func() tea.Msg {
		return TextFocusMsg{}
	}
}

func (mm *MenuModel) ServerView() string {
	selectedServerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(highlight))
	view := []string{}
	for i, menuItem := range mm.savedServers {
		if i == mm.currServerIndex {
			view = append(view, selectedServerStyle.Render(fmt.Sprintf("%s\n", menuItem)))
		} else {
			view = append(view, fmt.Sprintf("%s\n", menuItem))
		}
	}
	view = append(view, mm.serverTextInput.View())
	return lipgloss.JoinVertical(lipgloss.Left, view...)
}

func (mm *MenuModel) UpdateServer(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ChangeMenuPage:
		mm.page = msg.page
	case TextFocusMsg:
		mm.serverTextInput.Focus()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if mm.serverTextInput.Focused() {
				mm.serverTextInput.Blur()
			} else {
				cmds = append(cmds, ChangeMenuPageCmd(mainMenu))
			}
		case tea.KeyEnter:
			// This will be to join a server
			cmd = nil
			cmds = append(cmds, cmd)
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "n":
				cmd = TextFocusCmd()
				cmds = append(cmds, cmd)
			case "j":
				if mm.currServerIndex+1 < len(mm.savedServers) {
					mm.currServerIndex += 1
				}
			case "k":
				if mm.currServerIndex-1 >= 0 {
					mm.currServerIndex -= 1
				}
			}
		}
	}

	if mm.serverTextInput.Focused() {
		mm.serverTextInput, cmd = mm.serverTextInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return mm, tea.Batch(cmds...)
}

func (mm *MenuModel) UpdateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ChangeMenuPage:
		mm.page = msg.page
	case TextFocusMsg:
		mm.serverTextInput.Focus()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			cmds = append(cmds, ChangeMenuPageCmd(mainMenu))
		case tea.KeyEnter:
			// This will be to join a server
			cmd = nil
			cmds = append(cmds, cmd)
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "j":
				if mm.currServerIndex+1 < len(mm.savedServers) {
					mm.currServerIndex += 1
				}
			case "k":
				if mm.currServerIndex-1 >= 0 {
					mm.currServerIndex -= 1
				}
			}
		}
	}
	return mm, tea.Batch(cmds...)
}

func (mm *MenuModel) ViewSettings() string {
	return ""
}
