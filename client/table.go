package client

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type TuiTable struct {
	Players  []TuiPlayer
	Height   int
	Width    int
	Commands map[string]string
}

var GAME_COMMANDS = map[string]string{
	"n": "start game",
	"b": "place bet",
	"h": "hit",
	"s": "stand",
}

func NewTable() *TuiTable {
	return &TuiTable{
		Players: testPlayers,
		Commands: map[string]string{
			"n": "start game",
			"b": "place bet",
			"h": "hit",
			"s": "stand",
		},
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

func (t *TuiTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case *protocol.GameDTO:
		t.GameMessageToState(msg)
	case tea.KeyMsg:
		// Top Level Keys. Kill the program type keys
		switch msg.Type {
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "n":
				cmds = append(cmds, SendData(protocol.PackageClientMessage(protocol.MsgStartGame, "")))
			case "b":
				cmds = append(cmds, SendData(protocol.PackageClientMessage(protocol.MsgPlaceBet, "5")))
			case "h":
				cmds = append(cmds, SendData(protocol.PackageClientMessage(protocol.MsgHit, "")))
			case "s":
				cmds = append(cmds, SendData(protocol.PackageClientMessage(protocol.MsgStand, "")))
			}
			cmds = append(cmds, cmd)
		}
	}
	return t, tea.Batch(cmds...)
}

func (t *TuiTable) View() string {
	color := lipgloss.Color("#FFFFFF")
	style := lipgloss.NewStyle().Foreground(color).Border(lipgloss.DoubleBorder())
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
	playerFive := p4Style.Render(t.Players[5].renderPlayerZone())
	playerFour := p5Style.Render(t.Players[4].renderPlayerZone())
	return lipgloss.JoinVertical(lipgloss.Top, playerFive, playerFour)
}

func (t *TuiTable) renderVerticalZone2() string {
	dealerStyle := lipgloss.NewStyle().PaddingRight(4).PaddingTop(1).Foreground(lipgloss.Color(foreground))
	p3Style := lipgloss.NewStyle().PaddingTop(6).PaddingRight(4).PaddingBottom(2).Foreground(lipgloss.Color(foreground))
	dealer := dealerStyle.Render(t.Players[0].renderPlayerZone())
	player3 := p3Style.Render(t.Players[3].renderPlayerZone())
	return lipgloss.JoinVertical(lipgloss.Top, dealer, player3)
}

func (t *TuiTable) renderVerticalZone3() string {
	p1Style := lipgloss.NewStyle().PaddingRight(4).PaddingTop(2).Foreground(lipgloss.Color(foreground))
	p2Style := lipgloss.NewStyle().PaddingRight(4).PaddingTop(2).PaddingBottom(5).Foreground(lipgloss.Color(foreground))
	playerTwo := p2Style.Render(t.Players[2].renderPlayerZone())
	playerOne := p1Style.Render(t.Players[1].renderPlayerZone())
	return lipgloss.JoinVertical(lipgloss.Top, playerOne, playerTwo)
}

func (p *TuiPlayer) renderPlayerZone() string {
	nameTag := p.Name
	bet := p.Bet
	wallet := p.Wallet
	valueStr := fmt.Sprintf("%d", (p.Value))
	if p.Value == -1 {
		valueStr = "?"
	}
	status := fmt.Sprintf("V:%s B:%d W:%d", valueStr, bet, wallet)
	return lipgloss.Place(16, 5, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Top, nameTag, renderMultipleCards(p.Cards, 16, 6), status))
}
