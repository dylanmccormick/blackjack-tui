package client

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
