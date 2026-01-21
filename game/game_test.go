package game

import (
	"testing"

	"github.com/google/uuid"
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
	previous := make([]Card, 52)
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
		{Hand{Cards: []Card{{suit, ACE}, {suit, KING}}}, 21, true, BLACKJACK},
		{Hand{Cards: []Card{{suit, ACE}, {suit, QUEEN}}}, 21, true, BLACKJACK},
		{Hand{Cards: []Card{{suit, ACE}, {suit, JACK}}}, 21, true, BLACKJACK},
		{Hand{Cards: []Card{{suit, ACE}, {suit, 10}}}, 21, true, BLACKJACK},
		{Hand{Cards: []Card{{suit, ACE}, {suit, ACE}}}, 12, true, LIVE},
		{Hand{Cards: []Card{{suit, ACE}, {suit, ACE}, {suit, ACE}}}, 13, true, LIVE},
		{Hand{Cards: []Card{{suit, KING}, {suit, 10}, {suit, ACE}}}, 21, false, TWENTYONE},
		{Hand{Cards: []Card{{suit, 7}, {suit, 10}, {suit, ACE}}}, 18, false, LIVE},
		{Hand{Cards: []Card{{suit, 3}, {suit, 7}, {suit, ACE}}}, 21, true, TWENTYONE},
		{Hand{Cards: []Card{{suit, 6}, {suit, ACE}}}, 17, true, LIVE}, // soft 17
		{Hand{Cards: []Card{{suit, 10}, {suit, 10}, {suit, 10}}}, 30, false, BUST},
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
	var err error
	game := NewGame()

	p1 := &Player{}
	p2 := &Player{}
	p3 := &Player{}
	p4 := &Player{}
	p5 := &Player{}
	p6 := &Player{}
	pList := []*Player{p1, p2, p3, p4, p5, p6}
	for i := range 5 {
		p := pList[i]
		p.ID, err = uuid.NewUUID()
		if err != nil {
			t.Fatalf("Unable to create UUID err:%#v", err)
		}
		err = game.AddPlayer(p)
		if err != nil {
			t.Fatalf("Expected to add player to game. Got error=%#v", err)
		}
		if game.Players[i] != p {
			t.Fatalf("Expected player %d to be in index %d. got=%#v", p.ID, i, game.Players[0])
		}
	}
	err = game.AddPlayer(pList[5])
	if err == nil {
		t.Fatalf("Expected to get an error adding player 6. got=%#v", err)
	}
}

func TestPlaceBet(t *testing.T) {
	u1, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}

	u2, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}

	game := NewGame()
	p1 := &Player{ID: u1, Wallet: 10}
	p2 := &Player{ID: u2, Wallet: 1}
	game.AddPlayer(p1)
	game.AddPlayer(p2)
	err = game.StartGame()
	genericErrHelper(t, err)
	err = game.PlaceBet(p1, 5)
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
	u1, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	u2, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	u3, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	game := NewGame()
	p1 := &Player{ID: u1, Wallet: 10, State: BETS_MADE}
	p2 := &Player{ID: u2, Wallet: 100, State: BETS_MADE}
	p3 := &Player{ID: u3, Wallet: 100, State: BETS_MADE}
	game.AddPlayer(p1)
	game.AddPlayer(p2)
	game.AddPlayer(p3)
	p1.State = BETS_MADE
	p2.State = BETS_MADE
	p3.State = BETS_MADE
	genericErrHelper(t, err)
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
	u1, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	game := NewGame()
	p1 := &Player{ID: u1, Wallet: 10, State: BETTING}
	err = game.AddPlayer(p1)
	genericErrHelper(t, err)
	err = game.StartGame()
	genericErrHelper(t, err)
	err = game.PlaceBet(p1, 5)
	genericErrHelper(t, err)
	if p1.Wallet != 5 {
		t.Errorf("player wallet not updated when betting. expected=%d got=%d", 5, p1.Wallet)
	}
	game.StartRound()
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
	u1, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	game := NewGame()
	p1 := &Player{ID: u1, Wallet: 10, State: BETTING}
	err = game.AddPlayer(p1)
	genericErrHelper(t, err)
	err = game.StartGame()
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

func TestCalculatePayout(t *testing.T) {
	suit := suit("spade")
	tests := []struct {
		PlayerHand     *Hand
		DealerHand     *Hand
		ExpectedPayout int
	}{
		{&Hand{Cards: []Card{{suit, 10}, {suit, 10}, {suit, 10}}}, &Hand{Cards: []Card{{suit, 10}, {suit, 10}, {suit, 10}}}, 0},
		{&Hand{Cards: []Card{{suit, ACE}, {suit, 10}}}, &Hand{Cards: []Card{{suit, 10}, {suit, 10}, {suit, 10}}}, 25},
		{&Hand{Cards: []Card{{suit, 10}, {suit, 10}}}, &Hand{Cards: []Card{{suit, 10}, {suit, 10}, {suit, 10}}}, 20},
		{&Hand{Cards: []Card{{suit, 10}, {suit, 5}, {suit, 5}, {suit, ACE}}}, &Hand{Cards: []Card{{suit, 10}, {suit, ACE}}}, 0},
		{&Hand{Cards: []Card{{suit, 10}, {suit, 10}}}, &Hand{Cards: []Card{{suit, 10}, {suit, 10}}}, 10},
	}
	g := NewGame()
	u1, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	p1 := &Player{ID: u1, Wallet: 100, State: BETTING, Bet: 10}
	err = g.AddPlayer(p1)
	genericErrHelper(t, err)
	for _, tt := range tests {
		g.DealerHand = tt.DealerHand
		p1.Hand = tt.PlayerHand
		payout := g.calculatePayout(p1)
		if payout != tt.ExpectedPayout {
			t.Errorf("payout calculation incorrect. expected=%d got=%d", tt.ExpectedPayout, payout)
		}
	}
}

func TestHitUntilBust(t *testing.T) {
	suit := suit("spade")
	g := NewGame()
	g.Deck.Cards = append(
		[]Card{
			{suit, 10}, // player cards
			{suit, 10},
			{suit, 10}, // dealer cards
			{suit, 10},
			{suit, 10}, // busting card (player)
		},
		g.Deck.Cards...,
	)
	u1, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	p1 := &Player{ID: u1, Wallet: 100, State: BETTING, Bet: 10}
	err = g.AddPlayer(p1)
	genericErrHelper(t, err)
	err = g.StartGame()
	genericErrHelper(t, err)
	err = g.PlaceBet(p1, 5)
	genericErrHelper(t, err)
	err = g.StartRound()
	genericErrHelper(t, err)
	err = g.DealCards() // bet is already set in player struct
	genericErrHelper(t, err)
	err = g.Hit(p1) // This should bust.
	genericErrHelper(t, err)
	if p1.State != DONE {
		t.Fatalf("player state not advanced after busting. expected=%s got=%s", DONE, p1.State)
	}
	if g.State != DEALER_TURN {
		t.Fatalf("incorrect state after player turn ended. expected=%s got=%s", DEALER_TURN, g.State)
	}
}

func TestHitUntilStay(t *testing.T) {
	suit := suit("spade")
	g := NewGame()
	g.Deck.Cards = append(
		[]Card{
			{suit, 2}, // player cards
			{suit, 2},
			{suit, 10}, // dealer cards
			{suit, 10},
			{suit, 2}, // busting card (player)
			{suit, 2},
			{suit, 2},
			{suit, 2},
			{suit, 2},
			{suit, 2},
			{suit, 2},
			{suit, 2},
		},
		g.Deck.Cards...,
	)
	u1, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	p1 := &Player{ID: u1, Wallet: 100, State: BETTING, Bet: 10}
	err = g.AddPlayer(p1)
	genericErrHelper(t, err)
	err = g.StartGame()
	genericErrHelper(t, err)
	err = g.PlaceBet(p1, 5)
	genericErrHelper(t, err)
	err = g.StartRound()
	genericErrHelper(t, err)
	err = g.DealCards() // bet is already set in player struct
	genericErrHelper(t, err)
	for range 8 {
		err = g.Hit(p1) // This should bust.
		genericErrHelper(t, err)
		if p1.State != WAITING_FOR_ACTION {
			t.Fatalf("player state incorrectly advanced after hitting. expected=%s got=%s", WAITING_FOR_ACTION, p1.State)
		}
		if g.State != PLAYER_TURN {
			t.Fatalf("incorrect state during player turn. expected=%s got=%s", PLAYER_TURN, g.State)
		}
	}
	err = g.Hit(p1) // This should bust.
	genericErrHelper(t, err)
	if p1.State != DONE {
		t.Fatalf("player state not advanced after busting. expected=%s got=%s", DONE, p1.State)
	}
	if g.State != DEALER_TURN {
		t.Fatalf("incorrect state after player turn ended. expected=%s got=%s", DEALER_TURN, g.State)
	}
}

func TestDealerLogicHitSoft17(t *testing.T) {
	StandOnSoft17 = false
	suit := suit("spade")
	g := NewGame()
	g.Deck.Cards = append(
		[]Card{
			{suit, 10}, // player cards
			{suit, 10},
			{suit, ACE}, // dealer cards
			{suit, 6},
			{suit, 10},
		},
		g.Deck.Cards...,
	)
	u1, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	p1 := &Player{ID: u1, Wallet: 100, State: BETTING, Bet: 10}
	err = g.AddPlayer(p1)
	genericErrHelper(t, err)
	err = g.StartGame()
	genericErrHelper(t, err)
	err = g.PlaceBet(p1, 5)
	genericErrHelper(t, err)
	err = g.StartRound()
	genericErrHelper(t, err)
	err = g.DealCards() // bet is already set in player struct
	genericErrHelper(t, err)
	err = g.Stay(p1)
	genericErrHelper(t, err)
	err = g.PlayDealer()
	genericErrHelper(t, err)
	if len(g.DealerHand.Cards) != 3 {
		t.Fatalf("dealer logic incorrect. dealer has incorrect amount of cards. expected=%d got=%d", 3, len(g.DealerHand.Cards))
	}
}

func TestDealerLogicStandSoft17(t *testing.T) {
	StandOnSoft17 = true
	suit := suit("spade")
	g := NewGame()
	g.Deck.Cards = append(
		[]Card{
			{suit, 10}, // player cards
			{suit, 10},
			{suit, ACE}, // dealer cards
			{suit, 6},
			{suit, 10},
		},
		g.Deck.Cards...,
	)
	u1, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("Unable to create UUID err:%#v", err)
	}
	p1 := &Player{ID: u1, Wallet: 100}
	err = g.AddPlayer(p1)
	genericErrHelper(t, err)
	err = g.StartGame()
	genericErrHelper(t, err)
	err = g.PlaceBet(p1, 5)
	genericErrHelper(t, err)
	err = g.StartRound()
	genericErrHelper(t, err)
	err = g.DealCards() // bet is already set in player struct
	genericErrHelper(t, err)
	err = g.Stay(p1)
	genericErrHelper(t, err)
	err = g.PlayDealer()
	genericErrHelper(t, err)
	if len(g.DealerHand.Cards) != 2 {
		t.Fatalf("dealer logic incorrect. dealer has incorrect amount of cards. expected=%d got=%d", 2, len(g.DealerHand.Cards))
	}
}
