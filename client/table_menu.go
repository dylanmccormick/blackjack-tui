package client

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type TableMenuModel struct {
	textInput       textinput.Model
	currTableIndex  int
	availableTables []*protocol.TableDTO
	commandSet      bool
	Commands        map[string]string
}

func NewTableMenu() *TableMenuModel {
	ti := textinput.New()
	ti.Placeholder = "my_cool_table"
	ti.Width = 40
	return &TableMenuModel{
		textInput:  ti,
		commandSet: false,
		Commands: map[string]string{
			"j":     "down",
			"k":     "up",
			"enter": "select",
			"esc":   "back",
			"n":     "new table",
		},
	}
}

func (tm *TableMenuModel) Init() tea.Cmd {
	return nil
}

func (tm *TableMenuModel) View() string {
	items := []string{}
	selectedTableStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(highlight))
	for i, table := range tm.availableTables {
		if i == tm.currTableIndex {
			items = append(items, selectedTableStyle.Render(fmt.Sprintf("%d %s %d/%d\n", i, table.Id, table.CurrentPlayers, table.Capacity)))
		} else {
			items = append(items, fmt.Sprintf("%d %s %d/%d\n", i, table.Id, table.CurrentPlayers, table.Capacity))
		}
	}
	items = append(items, tm.textInput.View())
	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (tm *TableMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// for when the root model is on page lobby
	var cmds []tea.Cmd
	var cmd tea.Cmd
	if !tm.commandSet {
		cmds = append(cmds, AddCommands(tm.Commands))
		tm.commandSet = true
	}
	switch msg := msg.(type) {
	case TextFocusMsg:
		tm.textInput.Focus()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if tm.textInput.Focused() {
				tm.textInput.Blur()
				cmds = append(cmds, AddCommands(tm.Commands))
			} else {
				cmds = append(cmds, ChangeMenuPageCmd(mainMenu))
			}
		case tea.KeyEnter:
			// this could be a cotmand?
			var tableName string
			if tm.textInput.Focused() {
				tableName = tm.textInput.Value()
				cmd = SendData(protocol.PackageClientMessage(protocol.MsgCreateTable, tableName))
				cmds = append(cmds, cmd)
				cmds = append(cmds, AddCommands(tm.Commands))
			} else {
				if len(tm.availableTables) > 0 {
					tableName = tm.availableTables[tm.currTableIndex].Id
					cmd = SendData(protocol.PackageClientMessage(protocol.MsgJoinTable, tableName))
					cmds = append(cmds, cmd)
					log.Printf("Attempting to join table: %s", tableName)
					cmd = ChangeRootPage(gamePage)
					cmds = append(cmds, cmd)
					tm.commandSet = false
				}
			}
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "n":
				cmd = TextFocusCmd()
				cmds = append(cmds, cmd)
				cmds = append(cmds, AddCommands(map[string]string{"enter": "create table", "esc": "cancel"}))
			case "j":
				if tm.currTableIndex+1 < len(tm.availableTables) {
					tm.currTableIndex += 1
				}
				// lower the index on the room
			case "k":
				if tm.currTableIndex-1 >= 0 {
					tm.currTableIndex -= 1
				}
			case "u":
				cmd = SendData(protocol.PackageClientMessage(protocol.MsgTableList, ""))
				cmds = append(cmds, cmd)
			}
		}
	case []*protocol.TableDTO:
		tm.TablesToState(msg)
	}

	if tm.textInput.Focused() {
		tm.textInput, cmd = tm.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tm, tea.Batch(cmds...)
}

func (tm *TableMenuModel) TablesToState(msg []*protocol.TableDTO) {
	log.Println("Translating tables to table list")
	tm.availableTables = msg
}
