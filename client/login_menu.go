package client

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type loginMenuPage int

const (
	userCodePage     loginMenuPage = iota
	confirmationPage loginMenuPage = iota
)

type LoginMenu struct {
	page       loginMenuPage
	userCode   string
	currentUrl string
	Commands   map[string]string
}

func (lm *LoginMenu) View() string {
	switch lm.page {
	case userCodePage:
		return lm.viewUserCodePage()
	case confirmationPage:
		return lm.viewConfirmationPage()
	}
	return ""
}

func (lm *LoginMenu) viewConfirmationPage() string {
	return "YOU'RE LOGGED IN"
}

func (lm *LoginMenu) viewUserCodePage() string {
	sb := strings.Builder{}
	sb.WriteString("UserCode: ")
	sb.WriteString(lm.userCode)
	sb.WriteString("\n")

	sb.WriteString("Go to https://github.com/device/login and enter the above code\n")
	return sb.String()
}

func (lm *LoginMenu) Update(msg tea.Msg) (*LoginMenu, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeMenuPage:
		cmds = append(cmds, AddCommands(lm.Commands))
	case ChangeRootPageMsg:
		cmds = append(cmds, AddCommands(lm.Commands))
	case AuthLoginMsg:
		lm.userCode = msg.UserCode
		lm.page = userCodePage
	case AuthPollMsg:
		if msg.Authenticated {
			lm.page = confirmationPage
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			cmd = nil
			cmds = append(cmds, cmd)
		case tea.KeyRunes:
		}
	}

	return lm, tea.Batch(cmds...)
}

func (lm *LoginMenu) Init() tea.Cmd {
	return nil
}
