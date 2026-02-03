package client

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type TuiTable struct {
	Players    []TuiPlayer
	Height     int
	Width      int
	Commands   map[string]string
	betInput   textinput.Model
	commandSet bool
	username   string
}

var GAME_COMMANDS = map[string]string{
	"n": "start game",
	"b": "place bet",
	"h": "hit",
	"s": "stand",
}

func NewTable(height, width int) *TuiTable {
	betText := textinput.New()
	betText.Placeholder = "5"
	betText.Width = 5
	return &TuiTable{
		Players: []TuiPlayer{{Name: "dealer", Cards: []*Card{}}, {}, {}, {}, {}, {}},
		Commands: map[string]string{
			"n": "start game",
			"b": "place bet",
			"h": "hit",
			"s": "stand",
			"L": "leave server",
		},
		betInput: betText,
		Height:   height,
		Width:    width,
	}
}

func (t *TuiTable) Init() tea.Cmd {
	return AddCommands(t.Commands)
}

func (t *TuiTable) GameMessageToState(msg *protocol.GameDTO) {
	for i := 1; i < 6; i++ {
		player := t.Players[i]
		if len(msg.Players) < i {
			break
		}
		receivedPlayer := msg.Players[i-1]
		if len(receivedPlayer.Hand.Cards) > 0 {
			player.Cards = []*Card{}
			player.Value = receivedPlayer.Hand.Value
		}
		for _, card := range receivedPlayer.Hand.Cards {
			player.Cards = append(player.Cards, CardToCard(card))
		}
		player.Bet = receivedPlayer.Bet
		player.Wallet = receivedPlayer.Wallet
		player.Name = receivedPlayer.Name
		player.Current = receivedPlayer.CurrentPlayer
		log.Printf("Adding player %s to board", player.Name)
		t.Players[i] = player
	}
	dealer := t.Players[0]
	if len(msg.DealerHand.Cards) > 0 {
		dealer.Cards = []*Card{}
	}
	for _, card := range msg.DealerHand.Cards {
		dealer.Cards = append(dealer.Cards, CardToCard(card))
		dealer.Value = msg.DealerHand.Value
	}
	t.Players[0] = dealer
}

func (t *TuiTable) Resize(height, width int) {
	t.Height = height
	t.Width = width
}

func (t *TuiTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	if !t.commandSet {
		cmds = append(cmds, AddCommands(t.Commands))
		t.commandSet = true
	}
	switch msg := msg.(type) {
	case AuthPollMsg:
		t.username = msg.UserName
	case TextFocusMsg:
		t.betInput.Focus()
	case *protocol.GameDTO:
		t.GameMessageToState(msg)
	case SaveBetMsg:
		cmds = append(cmds, SendData(protocol.PackageClientMessage(protocol.MsgPlaceBet, t.betInput.Value())))
	case tea.KeyMsg:
		// Top Level Keys. Kill the program type keys
		switch msg.Type {
		case tea.KeyEnter:
			if t.betInput.Focused() {
				cmds = append(cmds, SaveBetCmd())
				t.betInput.Blur()
			}
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "n":
				cmds = append(cmds, SendData(protocol.PackageClientMessage(protocol.MsgStartGame, "")))
			case "b":
				cmds = append(cmds, TextFocusCmd())
			case "h":
				cmds = append(cmds, SendData(protocol.PackageClientMessage(protocol.MsgHit, "")))
			case "s":
				cmds = append(cmds, SendData(protocol.PackageClientMessage(protocol.MsgStand, "")))
			case "u":
				cmd = SendData(protocol.PackageClientMessage(protocol.MsgGetState, ""))
				cmds = append(cmds, cmd)
			case "L":
				cmd = SendData(protocol.PackageClientMessage(protocol.MsgLeaveTable, ""))
				cmds = append(cmds, cmd)
				cmds = append(cmds, ChangeRootPage(menuPage))
				t.commandSet = false
			}
		}
	}
	if t.betInput.Focused() {
		t.betInput, cmd = t.betInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return t, tea.Batch(cmds...)
}

func (t *TuiTable) View() string {
	color := lipgloss.Color("#FFFFFF")
	style := lipgloss.NewStyle().Foreground(color).Border(lipgloss.DoubleBorder()).Height(t.Height-2).Width(t.Width-2).Align(lipgloss.Center, lipgloss.Center)
	return style.Render(t.renderMiddle())
}

func (t *TuiTable) renderMiddle() string {
	vzone1 := t.renderVerticalZone1()
	vzone2 := t.renderVerticalZone2()
	vzone3 := t.renderVerticalZone3()
	return lipgloss.JoinHorizontal(lipgloss.Left, vzone1, vzone2, vzone3)
}

func (t *TuiTable) renderVerticalZone1() string {
	p4Style := lipgloss.NewStyle().PaddingTop(2).PaddingLeft(3).PaddingRight(4).Foreground(lipgloss.Color(foreground))
	p5Style := lipgloss.NewStyle().PaddingTop(2).PaddingLeft(3).PaddingRight(4).PaddingBottom(5).Foreground(lipgloss.Color(foreground))
	playerFive := p4Style.Render(t.Players[5].renderPlayerZone(t.username))
	playerFour := p5Style.Render(t.Players[4].renderPlayerZone(t.username))
	return lipgloss.JoinVertical(lipgloss.Top, playerFive, playerFour)
}

func (t *TuiTable) renderBetDialogue() string {
	betPrompt := "Input Bet Amount:"
	if t.betInput.Focused() {
		return lipgloss.JoinVertical(lipgloss.Top, betPrompt, t.betInput.View())
	}
	return ""
}

func (t *TuiTable) renderVerticalZone2() string {
	dealerStyle := lipgloss.NewStyle().PaddingRight(4).PaddingTop(1).PaddingBottom(1).Foreground(lipgloss.Color(foreground))
	betStyle := lipgloss.NewStyle().Height(2).Foreground(lipgloss.Color(highlight)).Align(lipgloss.Center, lipgloss.Center)
	p3Style := lipgloss.NewStyle().PaddingTop(3).PaddingRight(4).PaddingBottom(2).Foreground(lipgloss.Color(foreground))
	dealer := dealerStyle.Render(t.Players[0].renderPlayerZone(t.username))
	betDialogue := betStyle.Render(t.renderBetDialogue())
	player3 := p3Style.Render(t.Players[3].renderPlayerZone(t.username))
	return lipgloss.JoinVertical(lipgloss.Top, dealer, betDialogue, player3)
}

func (t *TuiTable) renderVerticalZone3() string {
	p1Style := lipgloss.NewStyle().PaddingRight(4).PaddingTop(2).Foreground(lipgloss.Color(foreground))
	p2Style := lipgloss.NewStyle().PaddingRight(4).PaddingTop(2).PaddingBottom(5).Foreground(lipgloss.Color(foreground))
	playerTwo := p2Style.Render(t.Players[2].renderPlayerZone(t.username))
	playerOne := p1Style.Render(t.Players[1].renderPlayerZone(t.username))
	return lipgloss.JoinVertical(lipgloss.Top, playerOne, playerTwo)
}

func (p *TuiPlayer) renderPlayerZone(username string) string {
	currPlayer := lipgloss.NewStyle().Foreground(lipgloss.Color(highlight))
	myPlayer := lipgloss.NewStyle().Foreground(lipgloss.Color(highlight))
	if p.Name == "" { // we have an empty slot
		return renderEmptyPlayer()
	}
	nameTag := p.Name
	if p.Current {
		nameTag = currPlayer.Render(p.Name)
	}
	bet := p.Bet
	wallet := p.Wallet
	valueStr := fmt.Sprintf("%d", (p.Value))
	if p.Value == -1 {
		valueStr = "?"
	}
	status := fmt.Sprintf("V:%s B:%d W:%d", valueStr, bet, wallet)
	if p.Name == username {
		status = myPlayer.Render(status)
	}
	return lipgloss.Place(16, 5, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Top, nameTag, renderMultipleCards(p.Cards, 16, 6), status))
}

func renderEmptyPlayer() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(softForeground)).BorderStyle(lipgloss.RoundedBorder()).Width(16-2).Height(6-2).Align(lipgloss.Center, lipgloss.Center)
	return lipgloss.Place(16, 5, lipgloss.Center, lipgloss.Center, style.Render("empty"))
}
