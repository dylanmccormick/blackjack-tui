package protocol

import (
	"log/slog"

	"github.com/dylanmccormick/blackjack-tui/game"
)

type HandDTO struct {
	Cards []CardDTO `json:"cards"`
	Value int       `json:"value"`
	State string    `json:"state"`
}

type CardDTO struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

type PlayerDTO struct {
	Bet           int     `json:"bet"`
	Wallet        int     `json:"wallet"`
	Hand          HandDTO `json:"hand"`
	Name          string  `json:"name"`
	CurrentPlayer bool    `json:"current"`
}

type GameDTO struct {
	Players    []PlayerDTO
	DealerHand HandDTO
}

type TableDTO struct {
	Id             string
	Capacity       int
	CurrentPlayers int
}

func CardToDTO(c game.Card) CardDTO {
	return CardDTO{
		Suit: string(c.Suit),
		Rank: int(c.Rank),
	}
}

func HandToDTO(h *game.Hand) HandDTO {
	cards := []CardDTO{}
	for _, c := range h.Cards {
		cards = append(cards, CardToDTO(c))
	}
	return HandDTO{
		Cards: cards,
		Value: h.GetValue(),
		State: h.GetState().String(),
	}
}

func PlayerToDTO(p *game.Player) PlayerDTO {
	slog.Info("translating player", "player", p)
	slog.Info("Current player?", "val", p.State == game.PLAYING_TURN)
	return PlayerDTO{
		Bet:           p.Bet,
		Wallet:        p.Wallet,
		Hand:          HandToDTO(p.Hand),
		Name:          p.Name,
		CurrentPlayer: (p.State == game.PLAYING_TURN),
	}
}

func GameToDTO(g *game.Game) GameDTO {
	players := []PlayerDTO{}
	for _, p := range g.Players {
		if p != nil {
			players = append(players, PlayerToDTO(p))
		} else {
			// Send empty spaces for table
			players = append(players, PlayerDTO{})
		}
	}
	return GameDTO{
		DealerHand: DealerToDTO(g.State, g.DealerHand),
		Players:    players,
	}
}

func DealerToDTO(state game.GameState, h *game.Hand) HandDTO {
	hand := HandToDTO(h)
	if state != game.DEALER_TURN && state != game.RESOLVING_BETS && state != game.WAITING_FOR_BETS && len(hand.Cards) > 0 {
		hand.Cards = hand.Cards[:1]
		hand.Value = -1
	}
	return hand
}
