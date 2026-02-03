-- name: CreateUser :one
INSERT INTO users(github_id, created_at, updated_at, last_login)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetUserByUsername :one
select *
from users
where github_id = ?
;

-- name: UpdateUserStats :one
UPDATE users
SET updated_at = CURRENT_TIMESTAMP,
wallet = ?,
amount_bet_lifetime = amount_bet_lifetime + ?,
amount_won_lifetime = amount_won_lifetime + ?,
amount_lost_lifetime = amount_lost_lifetime + ?,
hands_played = hands_played + 1,
hands_won = hands_won + ?,
hands_lost = hands_lost + ?,
blackjacks = blackjacks + ?
WHERE github_id = ?
RETURNING *
;
-- name: UpdateGithubStarred :one
UPDATE users 
SET updated_at = CURRENT_TIMESTAMP,
github_starred = ?
WHERE github_id = ?
RETURNING *
;

-- name: UpdateLoginStreak :one
UPDATE users
SET updated_at = CURRENT_TIMESTAMP,
last_login = CURRENT_TIMESTAMP,
login_streak = ?
WHERE github_id = ?
RETURNING *
;

-- name: UpdateUserAddIncome :one
UPDATE users
SET updated_at = CURRENT_TIMESTAMP,
last_login = CURRENT_TIMESTAMP,
wallet = wallet + ?
WHERE github_id = ?
RETURNING *
;
