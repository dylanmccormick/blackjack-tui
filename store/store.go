package store

import "context"

type Store struct{}

type WonState int

const (
	Won WonState = iota
	Lost
	Tied
)

type RoundResult struct {
	Outcome     WonState
	Blackjack   bool
	Bet         int
	WalletDelta int
}

func (s *Store) RecordResult(ctx context.Context, githubID string, rr RoundResult) error {
	return nil
}
