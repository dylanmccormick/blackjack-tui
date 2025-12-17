package game

type HandState int

const (
	LIVE HandState = iota
	BUST
	BLACKJACK
	TWENTYONE
)

type Hand struct {
	Cards []card
}

func NewHand() *Hand {
	return &Hand{Cards: []card{}}
}

func (h *Hand) AddCard(c card) {
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
