package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/dylanmccormick/blackjack-tui/internal/database"
)

func TestIsYesterday(t *testing.T) {
	tim := time.Now()
	tim = tim.AddDate(0, 0, -1)

	if !isYesterday(tim) {
		t.Fatalf("expected isYesterday to be %v got=%v", true, isYesterday(tim))
	}
}

func TestIsToday(t *testing.T) {
	tim := time.Now()

	if !isToday(tim) {
		t.Fatalf("expected isToday to be %v got=%v", true, isToday(tim))
	}
}

func TestCalculateIncome(t *testing.T) {
	tests := []struct {
		name        string
		streak      int64
		expectedVal int64
	}{
		{"zero_streak", 0, 100},
		{"one_streak", 1, 150},
		{"two_streak", 2, 200},
		{"three_streak", 3, 300},
		{"greater_than_three_streak", 4, 300},
		{"week_streak", 7, 1100},
		{"two_week_streak", 14, 1100},
		{"month_streak", 30, 10100},
		{"year_streak", 365, 10000100},
		{"year_plus_one_streak", 366, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := calculateIncome(tt.streak)
			if got != tt.expectedVal {
				t.Errorf("Income calculation incorrect. expected=%d got=%d", tt.expectedVal, got)
			}
		})
	}
}

// Database Tests
func TestProcessLogin_NewUser(t *testing.T) {
	mockRepo := &MockUserRepo{}
	store, err := NewStoreWithRepo(mockRepo)
	if err != nil {
		t.Fatalf("Unalbe to initialize test. err:%v", err)
	}
	// Return error no rows for user
	mockRepo.GetUserReturn = database.User{}
	mockRepo.GetUserError = sql.ErrNoRows

	mockRepo.CreateUserReturn = database.User{LoginStreak: 0, CreatedAt: time.Now(), LastLogin: time.Time{}}
	mockRepo.CreateUserError = nil

	mockRepo.UpdateUserAddIncomeReturn = database.User{Wallet: 100}

	u, income, err := store.ProcessLogin(context.Background(), "TEST_GH_ID")
	if err != nil {
		t.Errorf("Got an unexpected error with process login. err=%v", err)
	}
	if income != 100 {
		t.Errorf("income calculated incorrectly. expected=%d got=%d", 100, income)
	}
	if u.Wallet == 0 {
		t.Errorf("Expected user to have money")
	}
	if mockRepo.GetUserCalls[0] != "TEST_GH_ID" {
		t.Errorf("get user called with wrong ID. expected=%s got=%s", "TEST_GH_ID", mockRepo.GetUserCalls[0])
	}
}

func TestProcessLogin_ExistingUser(t *testing.T) {
	mockRepo := &MockUserRepo{}
	store, err := NewStoreWithRepo(mockRepo)
	if err != nil {
		t.Fatalf("Unalbe to initialize test. err:%v", err)
	}
	// Return error no rows for user
	mockRepo.GetUserReturn = database.User{LoginStreak: 1, CreatedAt: time.Now(), LastLogin: time.Now().AddDate(0, 0, -1), GithubID: "TEST_GH_ID"}
	mockRepo.UpdateUserAddIncomeReturn = database.User{Wallet: 200}
	mockRepo.UpdateLoginStreakReturn = database.User{LoginStreak: 2, CreatedAt: time.Now(), LastLogin: time.Now().AddDate(0, 0, -1), GithubID: "TEST_GH_ID"}

	u, income, err := store.ProcessLogin(context.Background(), "TEST_GH_ID")
	if err != nil {
		t.Errorf("Got an unexpected error with process login. err=%v", err)
	}
	if income != 200 {
		t.Errorf("income calculated incorrectly. expected=%d got=%d", 200, income)
	}
	if u.Wallet == 0 {
		t.Errorf("Expected user to have money")
	}
	if mockRepo.GetUserCalls[0] != "TEST_GH_ID" {
		t.Errorf("get user called with wrong ID. expected=%s got=%s", "TEST_GH_ID", mockRepo.GetUserCalls[0])
	}
	if len(mockRepo.CreateUserCalls) > 0 {
		t.Errorf("Expected 0 calls to CreateUser. got=%d", len(mockRepo.CreateUserCalls))
	}
}
