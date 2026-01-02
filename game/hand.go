package game

type HandState int

const (
	LIVE HandState = iota
	BUST
	BLACKJACK
	TWENTYONE
)

func (hs HandState) String() string {
	switch hs {
	case LIVE:
		return "Live"
	case BUST:
		return "Bust"
	case BLACKJACK:
		return "BlackJack"
	case TWENTYONE:
		return "TwentyOne"
	}
	return ""
}

type Hand struct {
	Cards []Card
}

func NewHand() *Hand {
	return &Hand{Cards: []Card{}}
}

func (h *Hand) AddCard(c Card) {
	h.Cards = append(h.Cards, c)
}

func (h *Hand) IsSoft() bool {
	_, soft := calculateValue(h.Cards)
	return soft
}

func (h *Hand) GetValue() int {
	val, _ := calculateValue(h.Cards)
	return val
}

func (h *Hand) GetState() HandState {
	if h.GetValue() == 21 {
		if len(h.Cards) == 2 {
			return BLACKJACK
		}
		return TWENTYONE
	} else if h.GetValue() > 21 {
		return BUST
	}
	return LIVE
}
