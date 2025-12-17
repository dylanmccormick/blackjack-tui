package game

import (
	"fmt"
	"log/slog"
	"slices"
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

func (gs GameState) String() string {
	switch gs {
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

func (g *Game) DealCards() error {
	if g.State != DEALING {
		return fmt.Errorf("Cannot deal when not in dealing state. CurrentState=%s", g.State)
	}
	g.DealerHand = NewHand()
	g.activePlayers = g.ActivePlayers()
	for _, player := range g.activePlayers {
		player.Hand = NewHand()
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
	g.StartPlayerTurn()
	return nil
}

func (g *Game) Stay(p *Player) error {
	p.State = DONE
	return nil
}

func (g *Game) Hit(p *Player) error {
	// handle hit
	return nil
}

func (g *Game) ResolveBets() {
	// TODO: Resolve bets
	g.StartBetting()
}

func (g *Game) CurrentPlayer() *Player {
	return g.activePlayers[g.CurrentPlayerIndex]
}

// func (g *Game) Run() error {
// 	// The orchestrator for the game. The game loop as they say
// 	for {
// 		switch g.State {
// 		case WAITING_FOR_BETS:
// 			g.HandleBets()
// 		case DEALING:
// 			g.HandleDealing()
// 		case PLAYER_TURN:
// 			g.HandlePlayerTurns()
// 		case DEALER_TURN:
// 			g.HandleDealerTurn()
// 		case RESOLVING_BETS:
// 			// TODO
// 		default:
// 			return fmt.Errorf("Unknown state reached for game %#v", g.State)
// 		}
// 	}
// }

func (g *Game) PlaceBet(p *Player, bet int) error {
	if !slices.Contains(g.Players, p) {
		return fmt.Errorf("Player %d not in this game", p.ID)
	}
	if g.State != WAITING_FOR_BETS {
		return fmt.Errorf("game is not in proper state to accept bets. CurrentState=%#v", g.State)
	}
	i := slices.Index(g.Players, p)
	err := p.ValidateBet(bet)
	if err != nil {
		return err
	}
	g.Players[i].Bet = bet
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
		if p != nil && p.State != SITTING {
			active = append(active, p)
		}
	}
	return active
}
