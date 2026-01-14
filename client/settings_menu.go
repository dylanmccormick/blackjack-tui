package client

import tea "github.com/charmbracelet/bubbletea"

type SettingsMenuModel struct{}

func NewSettingsMenu() *SettingsMenuModel {
	return &SettingsMenuModel{}
}

func (sm *SettingsMenuModel) Init() tea.Cmd{
	return nil
}

func (sm *SettingsMenuModel) View() string {
	return ""
}

func (sm *SettingsMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			cmds = append(cmds, ChangeMenuPageCmd(mainMenu))
		case tea.KeyEnter:
			cmd = nil
			cmds = append(cmds, cmd)
		case tea.KeyRunes:
			switch string(msg.Runes) {
			}
		}
	}
	return sm, tea.Batch(cmds...)
}
