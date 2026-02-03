package store

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/dylanmccormick/blackjack-tui/internal/database"
	"github.com/pressly/goose"
	_ "modernc.org/sqlite"
)

func NewStore(dbPath, schemaLocation string,) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return &Store{}, err
	}

	goose.SetDialect("sqlite3")
	err = goose.Up(db, schemaLocation)
	if err != nil {
		slog.Error("Error running goose", "error", err)
		return &Store{}, err
	}
	return &Store{database.New(db)}, nil
}

type Store struct {
	DB *database.Queries
}

func calculateIncome(streak int64) int64 {
	baseAmt := int64(100)
	switch {
	// 1 year bonus
	case streak == 0:
		return baseAmt
	case streak%365 == 0:
		return 10_000_000 + baseAmt
	// 30 day bonus
	case streak%30 == 0:
		return 10000 + baseAmt
	// 1 week bonus
	case streak%7 == 0:
		return 1000 + baseAmt
	// 3 day bonus
	case streak == 3:
		return 200 + baseAmt
	// 2 day bonus
	case streak == 2:
		return 100 + baseAmt
	// 1 day bonus
	case streak == 1:
		return 50 + baseAmt
	default:
		return baseAmt
	}
}

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
	Wallet      int
	WalletDelta int
}

func (s *Store) RecordResult(ctx context.Context, githubID string, rr RoundResult) error {
	var addWinAmount int64
	var addLossAmount int64
	var addHandWin int64
	var addHandLoss int64
	var addBlackjacks int64
	if rr.Blackjack {
		addBlackjacks = 1
	} else {
		addBlackjacks = 0
	}
	switch rr.Outcome {
	case Won:
		addWinAmount = int64(rr.WalletDelta)
		addLossAmount = 0
		addHandWin = 1
		addHandLoss = 0
	case Lost:
		addWinAmount = 0
		addLossAmount = int64(-1 * rr.WalletDelta)
		addHandWin = 0
		addHandLoss = 1
	default:
		addWinAmount = 0
		addLossAmount = 0
		addHandWin = 0
		addHandLoss = 0
	}
	params := database.UpdateUserStatsParams{
		Wallet:             int64(rr.Wallet),
		AmountBetLifetime:  int64(rr.Bet),
		AmountWonLifetime:  addWinAmount,
		AmountLostLifetime: addLossAmount,
		HandsWon:           addHandWin,
		HandsLost:          addHandLoss,
		GithubID:           githubID,
		Blackjacks:         addBlackjacks,
	}
	_, err := s.DB.UpdateUserStats(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func isYesterday(t time.Time) bool {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	tStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return tStart.Equal(yesterdayStart)
}

func isToday(t time.Time) bool {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return tStart.Equal(todayStart)
}

func (s *Store) UpdateUserStarred(ctx context.Context, githubID string) (bool, error) {
	user, err := s.DB.GetUserByUsername(ctx, githubID)
	if err != nil {
		return false, err
	}
	if user.GithubStarred {
		return false, nil
	}

	_, err = s.DB.UpdateGithubStarred(
		ctx,
		database.UpdateGithubStarredParams{
			GithubStarred: true,
			GithubID:      githubID,
		},
	)
	if err != nil {
		return false, err
	}

	user, err = s.DB.UpdateUserAddIncome(ctx, database.UpdateUserAddIncomeParams{Wallet: 5000, GithubID: githubID})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) ProcessLogin(ctx context.Context, githubID string) (database.User, int, error) {
	user, err := s.GetOrCreateUser(githubID)
	if err != nil {
		return user, -1, err
	}
	if isYesterday(user.LastLogin) {
		user, err = s.DB.UpdateLoginStreak(
			ctx, database.UpdateLoginStreakParams{
				LoginStreak: user.LoginStreak + 1,
				GithubID:    githubID,
			},
		)
	} else if isToday(user.LastLogin) {
		// do nothing!
		return user, 0, nil
	} else {
		user, err = s.DB.UpdateLoginStreak(
			ctx,
			database.UpdateLoginStreakParams{
				LoginStreak: 0,
				GithubID:    githubID,
			},
		)
	}

	income := calculateIncome(user.LoginStreak)
	if user.GithubStarred {
		income = income * 2
	}
	user, err = s.DB.UpdateUserAddIncome(ctx, database.UpdateUserAddIncomeParams{Wallet: income, GithubID: githubID})
	if err != nil {
		return user, -1, err
	}
	return user, int(income), nil
}

func (s *Store) GetOrCreateUser(githubID string) (database.User, error) {
	user, err := s.DB.GetUserByUsername(context.TODO(), githubID)
	if err == sql.ErrNoRows {
		slog.Info("User not found in database", "githubID", githubID)
		params := database.CreateUserParams{
			GithubID:  githubID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			LastLogin: time.Now(),
		}
		user, err = s.DB.CreateUser(context.TODO(), params)
		if err != nil {
			slog.Error("Unable to create user in DB", "githubID", githubID)
			return user, err
		}
	} else if err != nil {
		slog.Error("Database issues detected", "error", err)
		return user, err
	}

	return user, nil
}
