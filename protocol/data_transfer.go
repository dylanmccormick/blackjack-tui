package protocol

import "github.com/dylanmccormick/blackjack-tui/game"

type HandDTO struct {
	// Does this need hand value?
	Cards []CardDTO `json:"cards"`
}

type CardDTO struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

type PlayerDTO struct {
	Bet    int     `json:"bet"`
	Wallet int     `json:"wallet"`
	Hand   HandDTO `json:"hand"`
}

type GameDTO struct {
	Players    []*PlayerDTO
	DealerHand HandDTO
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
	}
}

func PlayerToDTO(p *game.Player) PlayerDTO {
	return PlayerDTO{
		Bet:    p.Bet,
		Wallet: p.Wallet,
		Hand:   HandToDTO(p.Hand),
	}
}
