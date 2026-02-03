-- +goose Up
CREATE TABLE users (
	github_id TEXT PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	wallet INT NOT NULL DEFAULT 1000,
	amount_bet_lifetime INT NOT NULL DEFAULT 0,
	amount_won_lifetime INT NOT NULL DEFAULT 0,
	amount_lost_lifetime INT NOT NULL DEFAULT 0,
	hands_played INT NOT NULL DEFAULT 0,
	hands_won INT NOT NULL DEFAULT 0,
	hands_lost INT NOT NULL DEFAULT 0,
	github_starred BOOL NOT NULL DEFAULT FALSE,
	last_login TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	login_streak INT NOT NULL DEFAULT 0,
	blackjacks INT not NULL DEFAULT 0
);

-- +goose Down
DROP TABLE IF EXISTS users;
