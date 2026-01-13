package game

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const DISCONNECT_TIMEOUT = 3 * time.Minute

type Player struct {
	Name   string
	ID     uuid.UUID
	State  PlayerState
	Bet    int // Used per round. How much the player is betting that round
	Wallet int // Used for a session. How much the player has at a session
	Hand   *Hand

	// Connection logic
	ConnectedAt           time.Time
	DisconnectedAt        time.Time // this will be good for time-in-game metrics or stats later
	IntentionalDisconnect bool
}

func NewPlayer(id uuid.UUID) *Player {
	slog.Debug("Creating new player")
	return &Player{
		ID:     id,
		Hand:   &Hand{},
		Bet:    0,
		Wallet: 100,
		Name:   "placeholder",
	}
}

const (
	BETTING PlayerState = iota
	BETS_MADE
	WAITING_FOR_ACTION
	DONE
	INACTIVE
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
	case INACTIVE:
		return "INACTIVE"
	}
	return "UNKNOWN"
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

func (p *Player) IsActive() bool {
	return p.State != INACTIVE
}

func (p *Player) ShouldRemove() bool {
	if p.IntentionalDisconnect {
		return true
	} else if time.Since(p.DisconnectedAt) > DISCONNECT_TIMEOUT {
		return true
	}
	return false
}

func (p *Player) MarkReconnected() {
	// I think this will reset to zero
	p.DisconnectedAt = time.Time{}
	p.IntentionalDisconnect = true
}

func (p *Player) MarkDisconnected(intentional bool) {
	p.DisconnectedAt = time.Now()
	p.IntentionalDisconnect = intentional
}
