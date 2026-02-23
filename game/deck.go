package game

import (
	"fmt"
	"math/rand/v2"
	"slices"
)

type Deck struct {
	Cards     []Card
	UsedCards []Card
	Threshold int
}

func CreateDeck(numDecks, threshold int) *Deck {
	if numDecks == 0 {
		numDecks = DECK_COUNT
	}
	if threshold == 0 {
		threshold = CUT_LOCATION
	}
	d := &Deck{Cards: make([]Card, 52*numDecks), Threshold: threshold}
	i := 0
	for range numDecks {
		for _, s := range []suit{"club", "diamond", "heart", "spade"} {
			for _, v := range []cardRank{2, 3, 4, 5, 6, 7, 8, 9, 10, JACK, QUEEN, KING, ACE} {
				d.Cards[i] = Card{Suit: s, Rank: v}
				i++
			}
		}
	}

	d.Shuffle()

	return d
}

func (d *Deck) Shuffle() {
	d.Cards = append(d.Cards, d.UsedCards...)
	d.Cards = Shuffle(d.Cards)
	d.UsedCards = []Card{}
}

func (d *Deck) DrawCard() (Card, error) {
	if len(d.Cards) == 0 {
		return Card{}, fmt.Errorf("Deck is empty")
	}
	c := d.Cards[0]
	d.Cards = d.Cards[1:]
	d.UsedCards = append(d.UsedCards, c)
	return c, nil
}

func (d *Deck) NeedsReshuffle() bool {
	return len(d.Cards) < d.Threshold
}

func Shuffle[T any](items []T) []T {
	// Copy so we don't mutate the list. Good practice from assembly I think
	ret := slices.Clone(items)
	for i := len(ret) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		ret[i], ret[j] = ret[j], ret[i]
	}
	return ret
}
