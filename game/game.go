package game

import (
	"fmt"
	"log/slog"
	"time"
)

type (
	GameState   int
	PlayerState int
	Player      struct {
		ID     int // TODO: Change this so it's unique later (for actually handling connections)
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
	WAITING_FOR_BETS GameState = iota
	DEALING
	PLAYER_TURN
	DEALER_TURN
	RESOLVING_BETS
)

const (
	PLAYER_LIMIT int = 5
	DECK_COUNT   int = 6
	CUT_LOCATION int = 150
)

type Game struct {
	State      GameState
	Deck       *Deck
	Players    []*Player
	DealerHand *Hand
	BetTime    int
}

func NewGame() *Game {
	return &Game{
		State:      WAITING_FOR_BETS,
		Deck:       CreateDeck(DECK_COUNT, CUT_LOCATION), // TODO: Get from config file
		Players:    make([]*Player, PLAYER_LIMIT),        // This can change. Maybe a table can have 6 players?
		DealerHand: &Hand{},
		BetTime:    30, // TODO: Get from Config file
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
	if g.State != RESOLVING_BETS {
		return fmt.Errorf("GameState WAITING_FOR_BETS is not reachable from state: %#v", g.State)
	}
	g.State = WAITING_FOR_BETS
	return nil
}

func (g *Game) StartDealing() error {
	if g.State != WAITING_FOR_BETS {
		return fmt.Errorf("GameState DEALING is not reachable from state: %#v", g.State)
	}
	g.State = DEALING
	return nil
}

func (g *Game) StartPlayerTurn() error {
	if g.State != DEALING {
		return fmt.Errorf("GameState PLAYER_TURN is not reachable from state: %#v", g.State)
	}
	g.State = PLAYER_TURN
	return nil
}

func (g *Game) StartDealerTurn() error {
	if g.State != PLAYER_TURN {
		return fmt.Errorf("GameState DEALER_TURN is not reachable from state: %#v", g.State)
	}
	g.State = DEALER_TURN
	return nil
}

func (g *Game) StartResolvingBets() error {
	if g.State != DEALER_TURN {
		return fmt.Errorf("GameState RESOLVING_BETS is not reachable from state: %#v", g.State)
	}
	g.State = RESOLVING_BETS
	return nil
}

func (g *Game) HandleBets() {
	tick := time.NewTimer(time.Duration(g.BetTime) * time.Second)

	wait := false
	timerExpired := false
	go func() {
		<-tick.C
		timerExpired = true
	}()
	for {
		for _, player := range g.Players {
			if player.Bet == 0 {
				wait = true
			}
		}
		if !wait || timerExpired {
			g.StartDealing()
			return
		}
	}
}

func (g *Game) HandleDealing() {
	g.DealerHand.Cards = make([]card, 2)
	for _, player := range g.Players {
		player.Hand.Cards = make([]card, 2)
	}
	// Deal Player Cards
	for i := range 2 {
		for _, player := range g.Players {
			card, err := g.Deck.DrawCard()
			if err != nil {
				slog.Error("Unable to deal card to player", "error", err)
			}
			player.Hand.Cards[i] = card
		}
	}
	for i := range 2 {
		card, err := g.Deck.DrawCard()
		if err != nil {
			slog.Error("Unable to deal card to dealer", "error", err)
		}
		g.DealerHand.Cards[i] = card
	}
}

func (g *Game) HandlePlayerTurns() {
	// TODO: Handle player turns. Game should wait 15 seconds for player to decide to hit, stay, stand, split, double, etc...
	for _, player := range g.Players {
		timerExpired := false
		go func() {
			tick := time.NewTimer(time.Duration(g.BetTime) * time.Second)
			<-tick.C
			timerExpired = true
		}()
		for {
			if player.State != WAITING_FOR_ACTION || timerExpired {
				break
			}
		}
	}
	g.StartDealerTurn()
}

func (g *Game) HandleDealerTurn() {
	// TODO: Handle the dealer logic here. Update the game state
	g.StartResolvingBets()
}

func (g *Game) ResolveBets() {
	// TODO: Resolve bets
	g.StartBetting()
}

func (g *Game) Run() error {
	// The orchestrator for the game. The game loop as they say
	for {
		switch g.State {
		case WAITING_FOR_BETS:
			g.HandleBets()
		case DEALING:
			g.HandleDealing()
		case PLAYER_TURN:
			g.HandlePlayerTurns()
		case DEALER_TURN:
			g.HandleDealerTurn()
		case RESOLVING_BETS:
			// TODO
		default:
			return fmt.Errorf("Unknown state reached for game %#v", g.State)
		}
	}
}
