package game

import (
	"fmt"
	"math/rand/v2"
	"slices"
)

type Deck struct {
	Cards     []card
	UsedCards []card
	Threshold int
}

func CreateDeck(numDecks, threshold int) *Deck {
	d := &Deck{Cards: make([]card, 52*numDecks), Threshold: threshold}
	i := 0
	for range numDecks {
		for _, s := range []suit{"club", "diamond", "heart", "spade"} {
			for _, v := range []cardRank{2, 3, 4, 5, 6, 7, 8, 9, 10, JACK, QUEEN, KING, ACE} {
				d.Cards[i] = card{Suit: s, Rank: v}
				i++
			}
		}
	}

	return d
}

func (d *Deck) Shuffle() {
	d.Cards = append(d.Cards, d.UsedCards...)
	d.Cards = Shuffle(d.Cards)
	d.UsedCards = []card{}
}

func (d *Deck) DrawCard() (card, error) {
	if len(d.Cards) == 0 {
		return card{}, fmt.Errorf("Deck is empty")
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
