package game

import (
	"testing"
)

func TestDeckHasCorrectCardCount(t *testing.T) {
	deck := CreateDeck(1, 0)
	if len(deck.Cards) != 52 {
		t.Errorf("Deck does not contain the proper amount of cards. expected=%d, got=%d", 52, len(deck.Cards))
	}
}

func TestDeckHasProperSuits(t *testing.T) {
	deck := CreateDeck(1, 0)
	suitCounts := make(map[suit]int)
	for _, c := range deck.Cards {
		suitCounts[c.Suit]++
	}

	for suit, count := range suitCounts {
		if count != 13 {
			t.Errorf("suit %s does not have %d cards. got=%d", suit, 13, count)
		}
	}
}

func TestDeckCompositionShuffle(t *testing.T) {
	deck := CreateDeck(1, 0)
	beforeCounts := make(map[suit]int)
	for _, c := range deck.Cards {
		beforeCounts[c.Suit]++
	}
	deck.Shuffle()
	suitCounts := make(map[suit]int)
	for _, c := range deck.Cards {
		suitCounts[c.Suit]++
	}

	for suit, count := range suitCounts {
		if count != 13 {
			t.Errorf("suit %s does not have %d cards. got=%d", suit, 13, count)
		}
	}
}

func TestDeckShuffle(t *testing.T) {
	deck := CreateDeck(1, 0)
	previous := make([]card, 52)
	copy(previous, deck.Cards)

	deck.Shuffle()

	// Count how many cards are in different positions
	movedCards := 0
	for i := range 52 {
		if previous[i] != deck.Cards[i] {
			movedCards++
		}
	}

	// With a proper shuffle, it's EXTREMELY unlikely all 52 cards stay in place
	// Even 1 card moving proves shuffle happened
	if movedCards == 0 {
		t.Error("No cards moved position after shuffling")
	}
}

func TestNeedsReshuffle(t *testing.T) {
	threshold := 25
	decks := 3
	deck := CreateDeck(decks, threshold)
	// draw to threshold
	for len(deck.Cards) > 25 {
		deck.DrawCard()
	}
	if deck.NeedsReshuffle() {
		t.Errorf("deck.NeedsReshuffle should be False. got=%#v", deck.NeedsReshuffle())
	}
	deck.DrawCard()
	if !deck.NeedsReshuffle() {
		t.Errorf("deck.NeedsReshuffle should be True. got=%#v. threshold=%d count=%d", deck.NeedsReshuffle(), threshold, len(deck.Cards))
	}
}

func TestUsedCardsLimit(t *testing.T) {
	threshold := 25
	decks := 3
	deck := CreateDeck(decks, threshold)

	for range 500 {
		deck.DrawCard()
		if deck.NeedsReshuffle() {
			deck.Shuffle()
		}
		if len(deck.UsedCards) > decks*52 {
			t.Fatalf("Used cards is higher than starting deck size length: %d", len(deck.UsedCards))
		}
	}
}

func TestOverdrawnDeck(t *testing.T) {
	deck := CreateDeck(1, 0)

	for range 52 {
		deck.DrawCard()
	}

	_, err := deck.DrawCard()
	if err == nil {
		t.Fatalf("Expected an error drawing above threshold. got=%#v deckLen=%d", err, len(deck.Cards))
	}
}

func TestHandValue(t *testing.T) {
	suit := suit("spade")

	tests := []struct {
		hand          Hand
		expectedValue int
		expectedSoft  bool
		expectedState HandState
	}{
		{Hand{Cards: []card{{suit, ACE}, {suit, KING}}}, 21, true, BLACKJACK},
		{Hand{Cards: []card{{suit, ACE}, {suit, QUEEN}}}, 21, true, BLACKJACK},
		{Hand{Cards: []card{{suit, ACE}, {suit, JACK}}}, 21, true, BLACKJACK},
		{Hand{Cards: []card{{suit, ACE}, {suit, 10}}}, 21, true, BLACKJACK},
		{Hand{Cards: []card{{suit, ACE}, {suit, ACE}}}, 12, true, LIVE},
		{Hand{Cards: []card{{suit, ACE}, {suit, ACE}, {suit, ACE}}}, 13, true, LIVE},
		{Hand{Cards: []card{{suit, KING}, {suit, 10}, {suit, ACE}}}, 21, false, TWENTYONE},
		{Hand{Cards: []card{{suit, 7}, {suit, 10}, {suit, ACE}}}, 18, false, LIVE},
		{Hand{Cards: []card{{suit, 3}, {suit, 7}, {suit, ACE}}}, 21, true, TWENTYONE},
		{Hand{Cards: []card{{suit, 6}, {suit, ACE}}}, 17, true, LIVE}, // soft 17
		{Hand{Cards: []card{{suit, 10}, {suit, 10}, {suit, 10}}}, 30, false, BUST},
	}

	for i, tt := range tests {
		if tt.hand.GetValue() != tt.expectedValue {
			t.Errorf("hand value does not match expected. expected=%d got=%d testCase=%d", tt.expectedValue, tt.hand.GetValue(), i)
		}
		if tt.hand.IsSoft() != tt.expectedSoft {
			t.Errorf("hand softness does not match expected. expected=%#v got=%#v testCase=%d", tt.expectedSoft, tt.hand.IsSoft(), i)
		}
		if tt.hand.GetState() != tt.expectedState {
			t.Errorf("hand state does not match expected. expected=%#v got=%#v testCase=%d", tt.expectedState, tt.hand.GetState(), i)
		}
	}
}

func TestAddPlayer(t *testing.T) {
	game := NewGame()
	p1 := &Player{ID: 1}
	p2 := &Player{ID: 2}
	p3 := &Player{ID: 3}
	p4 := &Player{ID: 4}
	p5 := &Player{ID: 5}
	p6 := &Player{ID: 6}
	pList := []*Player{p1, p2, p3, p4, p5, p6}
	for i := range 5 {
		p := pList[i]
		err := game.AddPlayer(p)
		if err != nil {
			t.Fatalf("Expected to add player to game. Got error=%#v", err)
		}
		if game.Players[i] != p {
			t.Fatalf("Expected player %d to be in index %d. got=%#v", p.ID, i, game.Players[0])
		}
	}
	err := game.AddPlayer(pList[5])
	if err == nil {
		t.Fatalf("Expected to get an error adding player 6. got=%#v", err)
	}
}

func TestPlaceBet(t *testing.T) {
	game := NewGame()
	p1 := &Player{ID: 1, Wallet: 10}
	p2 := &Player{ID: 2, Wallet: 1}
	game.AddPlayer(p1)
	game.AddPlayer(p2)
	err := game.PlaceBet(p1, 5)
	if err != nil {
		t.Fatalf("No error expected for placing bet. got=%#v", err)
	}
	err = game.PlaceBet(p2, 5)
	if err == nil {
		t.Fatalf("Error expected for placing bet, but got nil")
	}
	err = game.PlaceBet(p2, -5)
	if err == nil {
		t.Fatalf("Error expected for placing negative bet, but got nil")
	}
}

func TestNextPlayer(t *testing.T) {
	game := NewGame()
	p1 := &Player{ID: 1, Wallet: 10, State: BETS_MADE}
	p2 := &Player{ID: 2, Wallet: 1, State: BETS_MADE}
	p3 := &Player{ID: 3, Wallet: 1, State: BETS_MADE}
	game.AddPlayer(p1)
	game.AddPlayer(p2)
	game.AddPlayer(p3)
	game.State = PLAYER_TURN
	res := game.NextPlayer()
	if !res {
		t.Fatalf("Incorrect result for next player. expected=%v got=%v", true, res)
	}
	res = game.NextPlayer()
	if !res {
		t.Fatalf("Incorrect result for next player. expected=%v got=%v", true, res)
	}
	res = game.NextPlayer()
	if res {
		t.Fatalf("Incorrect result for next player. expected=%v got=%v", false, res)
	}
}

func TestGameFlow(t *testing.T) {
	game := NewGame()
	p1 := &Player{ID: 1, Wallet: 10, State: BETTING}
	err := game.AddPlayer(p1)
	genericErrHelper(t, err)
	err = game.PlaceBet(p1, 5)
	genericErrHelper(t, err)
	game.StartDealing()
	genericErrHelper(t, err)
	game.DealCards()
	genericErrHelper(t, err)
	if game.State != PLAYER_TURN {
		t.Fatalf("Game in incorrect state. expected=%s got=%s", PLAYER_TURN, game.State)
	}
	if game.CurrentPlayer() != p1 {
		t.Fatalf("Got incorrect player as current player. expected=%#v got=%#v", p1, game.CurrentPlayer())
	}
	err = game.Stay(p1)
	genericErrHelper(t, err)
	err = game.PlayDealer()
	genericErrHelper(t, err)
	err = game.ResolveBets()
	genericErrHelper(t, err)
	if game.State != WAITING_FOR_BETS {
		t.Fatalf("Game in incorrect state. expected=%s got=%s", WAIT_FOR_START, game.State)
	}
}

func genericErrHelper(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestGameFlowErrors(t *testing.T) {
	game := NewGame()
	p1 := &Player{ID: 1, Wallet: 10, State: BETTING}
	err := game.AddPlayer(p1)
	genericErrHelper(t, err)
	err = game.PlaceBet(p1, 5)
	genericErrHelper(t, err)
	err = game.PlayDealer()
	if err == nil {
		t.Fatalf("Expected error from game.PlayDealer(). got nil")
	}
	err = game.Hit(p1)
	if err == nil {
		t.Fatalf("Expected error from game.Hit(). got nil")
	}
}
