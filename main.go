package main

import (
	"fmt"

	"github.com/dylanmccormick/blackjack-tui/game"
)



func main() {
	d := game.CreateDeck(1, 65)
	fmt.Printf("There are %d cards in the deck\n", len(d.Cards))
	fmt.Printf("Shuffling deck\n")
	d.Shuffle()
	for _, c := range d.Cards {
		v, err := game.ValToString(c.Rank)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Drew card %s of %ss\n", v, c.Suit)
	}
}
