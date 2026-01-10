package client

import (
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

type Card struct {
	Value int
	Suit  int
}

var (
	values = []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	suits  = []string{"♠", "♦", "♥", "♣"}
)

const (
	cardHor = "─"
	cardVer = "│"
	cardTL  = "╭"
	cardTR  = "╮"
	cardBL  = "╰"
	cardBR  = "╯"
)

func NewCard(value, suit int) *Card {
	return &Card{
		Value: value,
		Suit:  suit,
	}
}

func CardToCard(pc protocol.CardDTO) *Card {
	log.Printf("translating card %#v", pc)
	card := &Card{}
	if strings.ToLower(pc.Suit) == "spade" {
		card.Suit = 2
	} else if strings.ToLower(pc.Suit) == "diamond" {
		card.Suit = 1
	} else if strings.ToLower(pc.Suit) == "heart" {
		card.Suit = 2
	} else if strings.ToLower(pc.Suit) == "club" {
		card.Suit = 3
	}

	card.Value = pc.Rank - 1

	return card
}

func (c *Card) String() string {
	color := lipgloss.Color(foreground)
	switch c.Suit {
	case 0, 3: // black cards
		color = lipgloss.Color(blackCard)
	case 1, 2: // red cards
		color = lipgloss.Color(redCard)
	}
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(values[c.Value] + suits[c.Suit])
}

func (c *Card) ViewPartial() string {
	color := lipgloss.Color("#FFFFFF")
	style := lipgloss.NewStyle().Foreground(color)
	padding := strings.Repeat(cardHor, 2)
	view := style.Render(cardTL) + c.String() + "\n"
	for i := 1; i < Height-1; i++ {
		view += style.Render(cardVer+strings.Repeat(" ", 2)) + "\n"
	}
	view += style.Render("╰" + padding)
	return view
}

func (c *Card) View() string {
	color := lipgloss.Color("#FFFFFF")
	style := lipgloss.NewStyle().Foreground(color)
	padding := strings.Repeat(cardHor, Width-2-lipgloss.Width(c.String()))
	view := style.Render(cardTL) + c.String() + style.Render(padding+cardTR) + "\n"
	for i := 1; i < Height-1; i++ {
		view += style.Render(cardVer+strings.Repeat(" ", Width-2)+cardVer) + "\n"
	}
	view += style.Render(cardBL+padding) + c.String() + style.Render(cardBR)
	return view
}

func hiddenCardView() string {
	color := lipgloss.Color("#FFFFFF")
	style := lipgloss.NewStyle().Foreground(color)
	padding := strings.Repeat(cardHor, Width-2)
	view := style.Render("╭") + style.Render(padding+"╮") + "\n"
	for i := 1; i < Height-1; i++ {
		view += style.Render(cardVer+strings.Repeat(" ", Width-2)+cardVer) + "\n"
	}
	view += style.Render("╰"+padding) + style.Render("╯")
	return view
}

func renderMultipleCards(cards []*Card, w, h int) string {
	cardViews := []string{}
	if len(cards) == 1 {
		cardViews = append(cardViews, cards[0].ViewPartial())
		cardViews = append(cardViews, hiddenCardView())
	} else {
		for i, card := range cards {
			if i == len(cards)-1 {
				cardViews = append(cardViews, card.View())
			} else {
				cardViews = append(cardViews, card.ViewPartial())
			}
		}
	}
	return lipgloss.Place(w, h, lipgloss.Left, lipgloss.Center, lipgloss.JoinHorizontal(lipgloss.Left, cardViews...))
}
