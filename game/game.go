package game

import (
	"fmt"
	"log/slog"
	"math"
	"slices"

	"github.com/google/uuid"
)

type (
	GameState   int
	PlayerState int
	Player      struct {
		ID     uuid.UUID // TODO: Change this so it's unique later (for actually handling connections)
		State  PlayerState
		Bet    int // Used per round. How much the player is betting that round
		Wallet int // Used for a session. How much the player has at a session
		Hand   *Hand
	}
)

const (
	BETTING PlayerState = iota
	BETS_MADE
	WAITING_FOR_ACTION
	DONE
	SITTING
)

const (
	WAIT_FOR_START GameState = iota
	WAITING_FOR_BETS
	DEALING
	PLAYER_TURN
	DEALER_TURN
	RESOLVING_BETS
)

func (ps PlayerState) String() string {
	switch ps {
	case BETTING:
		return "BETTING"
	case BETS_MADE:
		return "BETS_MADE"
	case WAITING_FOR_ACTION:
		return "WAITING_FOR_ACTION"
	case DONE:
		return "DONE"
	case SITTING:
		return "SITTING"
	}
	return "UNKNOWN"
}

func (gs GameState) String() string {
	switch gs {
	case WAIT_FOR_START:
		return "WAIT_FOR_START"
	case WAITING_FOR_BETS:
		return "WAITING_FOR_BETS"
	case DEALING:
		return "DEALING"
	case PLAYER_TURN:
		return "PLAYER_TURN"
	case DEALER_TURN:
		return "DEALER_TURN"
	case RESOLVING_BETS:
		return "RESOLVING_BETS"
	}
	return "UNKNOWN"
}

const (
	PLAYER_LIMIT int = 5
	DECK_COUNT   int = 6
	CUT_LOCATION int = 150
)

var StandOnSoft17 bool = true

type Game struct {
	State              GameState
	Deck               *Deck
	Players            []*Player
	DealerHand         *Hand
	BetTime            int
	CurrentPlayerIndex int
	activePlayers      []*Player
}

func NewGame() *Game {
	slog.Debug("Creating game")
	return &Game{
		State:              WAITING_FOR_BETS,
		Deck:               CreateDeck(DECK_COUNT, CUT_LOCATION), // TODO: Get from config file
		Players:            make([]*Player, PLAYER_LIMIT),        // This can change. Maybe a table can have 6 players?
		DealerHand:         &Hand{},
		BetTime:            30, // TODO: Get from Config file
		CurrentPlayerIndex: 0,
	}
}

func NewPlayer(id uuid.UUID) *Player {
	slog.Debug("Creating new player")
	return &Player{
		ID:     id,
		Hand:   &Hand{},
		Bet:    0,
		Wallet: 100,
	}
}

func (g *Game) AddPlayer(p *Player) error {
	for i, player := range g.Players {
		if player == nil {
			// Fill in empty slot
			g.Players[i] = p
			return nil
		}
	}
	return fmt.Errorf("Table is full")
}

func (g *Game) StartBetting() error {
	err := g.checkState(RESOLVING_BETS, "StartBetting")
	if err != nil {
		return err
	}
	g.State = WAITING_FOR_BETS
	return nil
}

func (g *Game) StartRound() error {
	err := g.checkState(WAITING_FOR_BETS, "StartRound")
	if err != nil {
		return err
	}
	g.State = DEALING
	return nil
}

func (g *Game) StartPlayerTurn() error {
	err := g.checkState(DEALING, "StartPlayerTurn")
	if err != nil {
		return err
	}
	g.State = PLAYER_TURN
	return nil
}

func (g *Game) StartDealerTurn() error {
	err := g.checkState(PLAYER_TURN, "StartDealerTurn")
	if err != nil {
		return err
	}
	g.State = DEALER_TURN
	return nil
}

func (g *Game) StartResolvingBets() error {
	err := g.checkState(DEALER_TURN, "StartResolvingBets")
	if err != nil {
		return err
	}
	g.State = RESOLVING_BETS
	return nil
}

func (g *Game) DealCards() error {
	err := g.checkState(DEALING, "DealCards")
	if err != nil {
		return err
	}

	g.DealerHand = NewHand()
	g.activePlayers = g.ActivePlayers()
	for _, player := range g.activePlayers {
		player.Hand = NewHand()
		player.State = WAITING_FOR_ACTION
	}
	// Deal Player Cards
	for range 2 {
		for _, player := range g.activePlayers {
			card, err := g.Deck.DrawCard()
			if err != nil {
				slog.Error("Unable to deal card to player", "error", err)
			}
			player.Hand.AddCard(card)
		}
	}
	for range 2 {
		card, err := g.Deck.DrawCard()
		if err != nil {
			slog.Error("Unable to deal card to dealer", "error", err)
		}
		g.DealerHand.AddCard(card)
	}
	return g.StartPlayerTurn()
}

func (g *Game) Stay(p *Player) error {
	err := g.checkState(PLAYER_TURN, "Stay")
	if err != nil {
		return err
	}
	if p != g.CurrentPlayer() {
		return fmt.Errorf("It is not Player %d's turn", p.ID)
	}
	g.endPlayerTurn(p)
	return nil
}

func (g *Game) Hit(p *Player) error {
	err := g.checkState(PLAYER_TURN, "Hit")
	if err != nil {
		return err
	}
	if p != g.CurrentPlayer() {
		return fmt.Errorf("It is not Player %d's turn", p.ID)
	}

	// add card to hand
	c, err := g.Deck.DrawCard()
	if err != nil {
		return err
	}
	p.Hand.AddCard(c)

	// update player state
	if p.Hand.GetState() == BUST {
		g.endPlayerTurn(p)
	}

	return nil
}

func (g *Game) endPlayerTurn(p *Player) error {
	if p != g.CurrentPlayer() {
		return fmt.Errorf("It is not Player %d's turn", p.ID)
	}
	p.State = DONE
	if !g.NextPlayer() {
		slog.Info("Starting dealer turn")
		g.StartDealerTurn()
	}
	return nil
}

func (g *Game) ResolveBets() error {
	err := g.checkState(RESOLVING_BETS, "ResolveBets")
	if err != nil {
		return err
	}
	for _, player := range g.activePlayers {
		winAmt := g.calculatePayout(player)
		player.Wallet += winAmt
	}
	g.reset()
	return g.StartBetting()
}

func (g *Game) reset() {
	for _, p := range g.ActivePlayers() {
		p.Bet = 0
		p.Hand = &Hand{Cards: []Card{}}
	}
	g.DealerHand = &Hand{Cards: []Card{}}
	g.CurrentPlayerIndex = 0
}

func (g *Game) calculatePayout(p *Player) int {
	pVal := p.Hand.GetValue()
	pState := p.Hand.GetState()
	dVal := g.DealerHand.GetValue()
	dState := g.DealerHand.GetState()
	if pState == BUST {
		return 0
	}
	if pState == BLACKJACK && dState == BLACKJACK {
		return p.Bet
	}
	if pVal == dVal && dState != BLACKJACK {
		if pState == BLACKJACK && dState != BLACKJACK {
			return int(math.Floor(float64(p.Bet)*float64(1.5))) + p.Bet
		}
		return p.Bet // push
	}

	// Win Conditions. Player > dealer. Player is live when dealer busts. BlackJack, but not if dealer also gets blackjack
	if pVal > dVal || g.DealerHand.GetState() == BUST {
		if pState == BLACKJACK {
			// player won with blackjack
			return int(math.Floor(float64(p.Bet)*float64(1.5))) + p.Bet
		}
		// player won regular style
		return p.Bet * 2
	}
	// player did not win
	return 0
}

func (g *Game) CurrentPlayer() *Player {
	if len(g.activePlayers) == 0 {
		g.activePlayers = g.ActivePlayers()
	}
	return g.activePlayers[g.CurrentPlayerIndex]
}

func (g *Game) PlaceBet(p *Player, bet int) error {
	err := g.checkState(WAITING_FOR_BETS, "PlaceBet")
	if err != nil {
		return err
	}
	if !slices.Contains(g.Players, p) {
		return fmt.Errorf("Player %d not in this game", p.ID)
	}
	i := slices.Index(g.Players, p)
	err = p.ValidateBet(bet)
	if err != nil {
		return err
	}
	g.Players[i].Bet = bet
	g.Players[i].Wallet -= bet
	g.Players[i].State = BETS_MADE
	return nil
}

func (p *Player) ValidateBet(bet int) error {
	if bet < 1 {
		return fmt.Errorf("Bets cannot be negative")
	}
	if bet > p.Wallet {
		return fmt.Errorf("Bet cannot be higher than current wallet amount")
	}
	return nil
}

func (g *Game) NextPlayer() bool {
	g.CurrentPlayerIndex++
	return g.CurrentPlayerIndex < len(g.ActivePlayers())
}

func (g *Game) ActivePlayers() []*Player {
	active := []*Player{}
	for _, p := range g.Players {
		if p == nil {
			continue
		}
		if p.State != SITTING {
			active = append(active, p)
		} else {
			p.State = SITTING
		}
	}
	if len(active) < 1 {
		slog.Error("No active players")
	}
	return active
}

func (g *Game) PlayDealer() error {
	err := g.checkState(DEALER_TURN, "PlayDealer")
	if err != nil {
		return err
	}
hitPhase:
	for g.DealerHand.GetValue() < 17 {
		c, err := g.Deck.DrawCard()
		if err != nil {
			return err
		}
		g.DealerHand.AddCard(c)
	}

	if g.DealerHand.GetValue() == 17 && g.DealerHand.IsSoft() && !StandOnSoft17 {
		c, err := g.Deck.DrawCard()
		if err != nil {
			return err
		}
		g.DealerHand.AddCard(c)
		goto hitPhase
	}
	g.StartResolvingBets()
	return nil
}

func (g *Game) checkState(expected GameState, method string) error {
	if g.State != expected {
		return fmt.Errorf("Method %s() cannot be run from state %s", method, g.State)
	}
	return nil
}

func (g *Game) GetPlayer(playerId uuid.UUID) *Player {
	for _, p := range g.Players {
		if p.ID == playerId {
			return p
		}
	}
	return nil
}

func (g *Game) AllPlayersBet() bool {
	for _, p := range g.Players {
		if p != nil && p.Bet == 0 {
			return false
		}
	}
	return true
}
