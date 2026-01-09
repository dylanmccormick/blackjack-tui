package client

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

const (
	tableHor = "═"
	tableVer = "║"
	tableTL  = "╔"
	tableTR  = "╗"
	tableBL  = "╚"
	tableBR  = "╝"
)

type TuiTable struct {
	Players []TuiPlayer
}

func NewTable() *TuiTable {
	return &TuiTable{
		Players: testPlayers,
	}
}

func (t *TuiTable) Init() tea.Cmd {
	return nil
}

func (t *TuiTable) GameMessageToState(msg *protocol.GameDTO) {
	log.Println("Translating game to state")
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
	case tea.KeyMsg:
		// Top Level Keys. Kill the program type keys
		switch msg.Type {
		case tea.KeyRunes:
			switch string(msg.Runes) {
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
	style := lipgloss.NewStyle().Foreground(color)
	// Render top line
	top := style.Render(tableTL + strings.Repeat(tableHor, 66-2) + tableTR)
	vertBorder := style.Render(strings.Repeat(tableVer+"\n", 24) + tableVer)
	// Render Middle
	middle := lipgloss.JoinHorizontal(lipgloss.Left, vertBorder, t.renderMiddle(), vertBorder)
	// Render Bottom Line
	bottom := style.Render(tableBL + strings.Repeat(tableHor, 66-2) + tableBR)
	return lipgloss.JoinVertical(lipgloss.Top, top, middle, bottom)
}

func (t *TuiTable) renderMiddle() string {
	vzone1 := t.renderVerticalZone1()
	vzone2 := t.renderVerticalZone2()
	vzone3 := t.renderVerticalZone3()
	return lipgloss.JoinHorizontal(lipgloss.Left, vzone1, vzone2, vzone3)
}

func (t *TuiTable) renderVerticalZone1() string {
	playerOne := t.Players[1].renderPlayerZone()
	playerTwo := t.Players[2].renderPlayerZone()
	middle1 := lipgloss.JoinHorizontal(lipgloss.Left, renderEmptyBlock(3, 6), playerOne, renderEmptyBlock(4, 6))
	middle2 := lipgloss.JoinHorizontal(lipgloss.Left, renderEmptyBlock(3, 6), playerTwo, renderEmptyBlock(4, 6))
	return lipgloss.JoinVertical(lipgloss.Top, renderEmptyBlock(22, 1), middle1, renderEmptyBlock(20, 2), middle2, renderEmptyBlock(20, 5))
}

func (t *TuiTable) renderVerticalZone2() string {
	dealer := t.Players[0].renderPlayerZone()
	player3 := t.Players[3].renderPlayerZone()
	middle1 := lipgloss.JoinHorizontal(lipgloss.Left, dealer, renderEmptyBlock(1, 6))
	middle2 := lipgloss.JoinHorizontal(lipgloss.Left, player3, renderEmptyBlock(1, 6))
	return lipgloss.JoinVertical(lipgloss.Top, middle1, renderEmptyBlock(20, 6), middle2, renderEmptyBlock(20, 2))
}

func (t *TuiTable) renderVerticalZone3() string {
	playerFour := t.Players[4].renderPlayerZone()
	playerFive := t.Players[4].renderPlayerZone()
	middle1 := lipgloss.JoinHorizontal(lipgloss.Left, playerFive, renderEmptyBlock(4, 6))
	middle2 := lipgloss.JoinHorizontal(lipgloss.Left, playerFour, renderEmptyBlock(4, 6))
	return lipgloss.JoinVertical(lipgloss.Top, renderEmptyBlock(16, 1), middle1, renderEmptyBlock(22, 2), middle2, renderEmptyBlock(22, 5))
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
	return lipgloss.JoinVertical(lipgloss.Top, nameTag, renderMultipleCards(p.Cards), status)
}
