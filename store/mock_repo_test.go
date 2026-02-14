package store

import (
	"context"

	"github.com/dylanmccormick/blackjack-tui/internal/database"
)

type MockUserRepo struct {
	GetUserReturn database.User
	GetUserError  error

	CreateUserReturn database.User
	CreateUserError  error

	UpdateGithubStarredReturn database.User
	UpdateGithubStarredError  error

	UpdateLoginStreakReturn database.User
	UpdateLoginStreakError  error

	UpdateUserAddIncomeReturn database.User
	UpdateUserAddIncomeError  error

	UpdateUserStatsReturn database.User
	UpdateUserStatsError  error

	GetUserCalls             []string // githubIDs passed
	UpdateStreakCalls        []database.UpdateLoginStreakParams
	CreateUserCalls          []database.CreateUserParams
	UpdateStarredCalls       []database.UpdateGithubStarredParams
	UpdateLoginStreakCalls   []database.UpdateLoginStreakParams
	UpdateUserAddIncomeCalls []database.UpdateUserAddIncomeParams
	UpdateUserStatsCalls     []database.UpdateUserStatsParams
}

func (m *MockUserRepo) GetUserByUsername(ctx context.Context, githubID string) (database.User, error) {
	m.GetUserCalls = append(m.GetUserCalls, githubID)
	return m.GetUserReturn, m.GetUserError
}

func (m *MockUserRepo) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	m.CreateUserCalls = append(m.CreateUserCalls, arg)
	return m.CreateUserReturn, m.CreateUserError
}

func (m *MockUserRepo) UpdateGithubStarred(ctx context.Context, arg database.UpdateGithubStarredParams) (database.User, error) {
	m.UpdateStarredCalls = append(m.UpdateStarredCalls, arg)
	return m.UpdateGithubStarredReturn, m.UpdateGithubStarredError
}

func (m *MockUserRepo) UpdateLoginStreak(ctx context.Context, arg database.UpdateLoginStreakParams) (database.User, error) {
	m.UpdateLoginStreakCalls = append(m.UpdateLoginStreakCalls, arg)
	return m.UpdateLoginStreakReturn, m.UpdateLoginStreakError
}

func (m *MockUserRepo) UpdateUserAddIncome(ctx context.Context, arg database.UpdateUserAddIncomeParams) (database.User, error) {
	m.UpdateUserAddIncomeCalls = append(m.UpdateUserAddIncomeCalls, arg)
	return m.UpdateUserAddIncomeReturn, m.UpdateUserAddIncomeError
}

func (m *MockUserRepo) UpdateUserStats(ctx context.Context, arg database.UpdateUserStatsParams) (database.User, error) {
	m.UpdateUserStatsCalls = append(m.UpdateUserStatsCalls, arg)
	return m.UpdateUserStatsReturn, m.UpdateUserStatsError
}
