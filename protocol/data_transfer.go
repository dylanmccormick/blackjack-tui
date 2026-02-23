package protocol

import (
	"log/slog"

	"github.com/dylanmccormick/blackjack-tui/game"
	"github.com/dylanmccormick/blackjack-tui/internal/database"
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

type PopUpDTO struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type StatsDTO struct {
	LifetimeBet   int `json:"lifetime_bet"`
	LifetimeLoss  int `json:"lifetime_loss"`
	LifetimeWon   int `json:"lifetime_won"`
	Blackjacks    int `json:"lifetime_blackjacks"`
	Wallet        int `json:"wallet"`
	HandsPlayed   int `json:"hands_played"`
	HandsWon      int `json:"hands_won"`
	HandsLost     int `json:"hands_lost"`
	WinPercentage int `json:"win_percentage"`
}

func CardToDTO(c game.Card) CardDTO {
	return CardDTO{
		Suit: string(c.Suit),
		Rank: int(c.Rank),
	}
}

func UserToStatsDTO(u *database.User) StatsDTO {
	slog.Info("Translating stats", "user", u)
	var winPercentage int
	if u.HandsPlayed > 0 {
		winPercentage = int(100 * u.HandsWon / u.HandsPlayed)
	} else {
		winPercentage = 0
	}
	return StatsDTO{
		LifetimeBet:   int(u.AmountBetLifetime),
		LifetimeLoss:  int(u.AmountLostLifetime),
		LifetimeWon:   int(u.AmountWonLifetime),
		Blackjacks:    int(u.Blackjacks),
		Wallet:        int(u.Wallet),
		HandsPlayed:   int(u.HandsPlayed),
		HandsWon:      int(u.HandsWon),
		HandsLost:     int(u.HandsLost),
		WinPercentage: winPercentage,
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

func MessageToDTO(message string, lvl PopUpType) PopUpDTO {
	return PopUpDTO{
		Message: message,
		Type:    string(lvl),
	}
}
