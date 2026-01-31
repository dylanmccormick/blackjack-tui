package client

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type StartWSMsg struct{}

func StartWSCmd() tea.Cmd {
	return func() tea.Msg {
		return StartWSMsg{}
	}
}

func ChangeRootPage(p page) tea.Cmd {
	return func() tea.Msg {
		return ChangeRootPageMsg{p}
	}
}

func AddCommands(cmds map[string]string) tea.Cmd {
	return func() tea.Msg {
		return AddCommandsMsg{cmds}
	}
}

type LoginRequested struct {
	Url string
}

func RequestLogin(url string) tea.Cmd {
	return func() tea.Msg {
		return LoginRequested{url}
	}
}

func SendData(msg *protocol.TransportMessage) tea.Cmd {
	return func() tea.Msg {
		return SendMsg{msg}
	}
}

type (
	AddCommandsMsg struct {
		commands map[string]string
	}
	SaveBetMsg        struct{}
	ChangeRootPageMsg struct {
		page page
	}
	SendMsg struct {
		data *protocol.TransportMessage
	}
	ChangeMenuPage struct {
		page mPage
	}
	TextFocusMsg struct{}
)

func ChangeMenuPageCmd(p mPage) tea.Cmd {
	return func() tea.Msg {
		return ChangeMenuPage{p}
	}
}

func TextFocusCmd() tea.Cmd {
	return func() tea.Msg {
		return TextFocusMsg{}
	}
}

func SaveBetCmd() tea.Cmd {
	return func() tea.Msg {
		return SaveBetMsg{}
	}
}
