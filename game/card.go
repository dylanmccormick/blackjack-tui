package game

import "fmt"

type (
	suit     string
	cardRank int
	Card     struct {
		Suit suit     `json:"suit"`
		Rank cardRank `json:"rank"`
	}
)

const (
	JACK  cardRank = 11
	QUEEN cardRank = 12
	KING  cardRank = 13
	ACE   cardRank = 1
)

func ValToString(i cardRank) (string, error) {
	vals := map[cardRank]string{
		1:  "ace",
		2:  "two",
		3:  "three",
		4:  "four",
		5:  "five",
		6:  "six",
		7:  "seven",
		8:  "eight",
		9:  "nine",
		10: "ten",
		11: "jack",
		12: "queen",
		13: "king",
	}

	s, ok := vals[i]
	if !ok {
		return "", fmt.Errorf("no string value found for %d", i)
	}

	return s, nil
}

func calculateValue(c []Card) (int, bool) {
	value := 0
	highAces := 0
	for _, card := range c {
		val := card.Value()
		if card.Rank == ACE {
			highAces += 1
			value += 11
		} else if val > 10 {
			value += 10
		} else {
			value += card.Value()
		}
	}
	for value > 21 && highAces > 0 {
		value = value - 10
		highAces -= 1
	}
	return value, highAces > 0
}

func (c Card) Value() int {
	return int(c.Rank)
}
