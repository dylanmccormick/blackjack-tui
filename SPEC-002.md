# SPEC-002: Production Deployable Blackjack TUI (Go)

## Intent
Harden this repo into a backend-credible, production-deployable Go service without adding new blackjack mechanics.

The end state is a public server reachable at `wss://blackjack.dylanjmccormick.com` (via Cloudflare Tunnel), with GitHub login, persisted wallet/stats, anti-abuse controls, and operational visibility.

## Goals
- Publicly reachable WebSocket server behind Cloudflare Tunnel.
- All access authenticated (GitHub only).
- Persist user identity, wallet balance, and aggregate stats in SQLite.
- Server is robust:
  - no panics on malformed input or disconnects
  - race-free (`go test -race ./...`)
  - graceful shutdown
- Operable:
  - structured logging
  - `/healthz`, `/readyz`, `/metrics`
- TUI is stable:
  - no stdout/stderr logging that corrupts the UI

## Non-goals (v1)
- Persisting in-progress games across server restarts.
- Multi-provider auth.
- CAPTCHA / browser bot challenges.
- “Star repo for credits” verification.
- Spectator mode.

## Definitions
- User: persisted account mapped 1:1 to a GitHub identity.
- Session token: opaque credential issued by this server; stored client-side; hashed server-side.
- Connection: a single WebSocket connection associated with exactly one authenticated user.

## Architecture (high level)
- `client/`: Bubble Tea TUI.
- `server/`: HTTP + WebSocket server.
- `game/`: domain/game rules and state machine.
- `protocol/`: message types and DTO mapping.
- SQLite: persisted to disk via Docker volume.

### Protocol note (HTTP vs WebSocket)
- WebSocket begins as an HTTP request and upgrades to a framed protocol.
- Auth is completed before opening WebSocket (device flow via HTTP endpoints).
- WebSocket upgrade requires `Authorization: Bearer <app_session_token>`.

## Authentication (GitHub Device Flow)
GitHub device flow is used because it is CLI/TUI-friendly and does not require browser callbacks to the client.

### Desired UX
1. User selects “Login with GitHub” in the TUI.
2. TUI displays GitHub verification URL and the short user code.
3. User completes auth in a browser.
4. TUI waits (polls) until authorized.
5. TUI stores the app session token locally and connects via WebSocket.

### Trust boundaries
- The GitHub OAuth client secret exists only on the server.
- The client never stores GitHub access tokens in v1; it only stores the app session token.
- Login attempts may be lost on server restart (acceptable v1).

### Login attempt binding (anti-hijack)
- The server creates a login attempt record and binds it to the request IP.
- Poll requests must come from the same IP; otherwise deny and log.
- The server must not reveal whether an attempt exists/authorized when IP mismatches.

## Session policy
### Token lifetime
- Fixed TTL: 7 days from issuance.
- No sliding extension in v1.

### One active session per user
- At most 1 active session per user.
- On successful login:
  - revoke any prior session for that user
  - issue a new session token

### One active WebSocket connection per user
- At most 1 active WebSocket connection per user.
- If a second connection attempts to upgrade while a connection is active:
  - deny the upgrade
  - log with `user_id`, derived IP, and reason

### Re-login behavior
- Re-login always succeeds.
- Re-login revokes the old session immediately.
- If the user has an active WebSocket connection under the old session:
  - server sends a `session_revoked` message (server -> client)
  - then closes the connection

### Client token storage
- Store token locally (e.g., `~/.config/blackjack-tui/config.json`).
- Token must never be printed to stdout/stderr or logged.
- File should be created with restrictive permissions.

## Reconnect semantics
- Track whether a user left a table intentionally vs unintentionally.
- Auto-rejoin rule:
  - if last table exists AND last disconnect was unintentional -> auto-rejoin
  - else -> lobby
- Turn behavior:
  - if it becomes the user’s turn while disconnected, server auto-stands.

## Persistence (SQLite)
Use a pure-Go SQLite driver.

### Persisted data (minimum)
- Users:
  - internal user id
  - github user id (unique)
  - github login
  - created/updated timestamps
- Sessions:
  - token hash (unique)
  - user id
  - created_at, expires_at
  - revoked_at (nullable)
- Wallet:
  - integer balance per user
- Stats (aggregate):
  - hands played
  - amount won
  - amount lost
  - optional breakdowns (wins/losses/push/blackjacks)

### Transactional updates
- Wallet and stats updates must be atomic (single DB transaction) at round resolution.

### Faucet (optional but recommended)
- Daily top-up (e.g., +1000 once per 24 hours).
- Enforced server-side.

## Anti-abuse controls
This server is public; it must be resilient to spam and malformed input.

### Pre-auth rate limiting (by IP)
- Login start: 1 request per IP per minute.
- Login poll/status: 60 requests per IP per minute.
- Violations: deny and log.

### Message validation
- Allow-list message types.
- Reject unknown types.
- Validate payload schema and size.
- Do not mutate game state or timers on invalid commands.

### Authenticated rate limiting
- Rate limit inbound WebSocket messages per connection.
- Optionally apply stricter limits to expensive actions (table creation, etc.).

## Observability
### Logging
- Structured logs in server.
- Include fields when available:
  - `user_id`, `conn_id`, `table_id`, `msg_type`, `remote_ip`, `error`
- Never log secrets (tokens, raw auth codes).

### Endpoints
- `GET /healthz`: liveness.
- `GET /readyz`: readiness (DB reachable, server ready to accept connections).
- `GET /metrics`: Prometheus metrics.

### Suggested metrics
- active connections
- active tables
- messages received / rejected
- auth successes / failures
- errors by category

## Deployment (Docker + Cloudflare Tunnel)
### Production trust model
- The origin server is reachable only through Cloudflare Tunnel.
- Docker publishes the server port bound to localhost (not `0.0.0.0`).
- In production, the server may trust Cloudflare client IP headers.

### Dev trust model
- In dev, do not trust forwarded headers.
- Use the TCP remote address only.

### Remote IP derivation
- Production (via tunnel): prefer `CF-Connecting-IP`, else TCP remote address.
- Dev: TCP remote address only.

### Login attempt binding (production + dev)
- Bind login attempts to derived IP.
- Deny and log if an attempt is polled from a different IP.

## Testing strategy
### Local requirements
- `go test ./...` passes.
- `go test -race ./...` passes.
- `go vet ./...` passes.

### CI requirements
- Run the above in CI for every PR.
- Optional: curated `golangci-lint`.

### Test focus
- Protocol serialization and validation.
- Session issuance + revocation behavior.
- Rate limiter behavior.
- Wallet/stats transaction integrity.

## 4-week milestone plan
### Week 1: Core hardening + CI
- Fix failing/stale tests.
- Add CI gates (test, race, vet, formatting, lint).
- Remove panics and fix concurrency hazards.
- Implement graceful shutdown.

### Week 2: Auth + session tokens
- Implement GitHub device flow (server-mediated).
- Implement app session tokens + hashing + expiry.
- Require auth for WebSocket upgrade.
- Enforce single connection per user.

### Week 3: SQLite persistence
- Add migrations.
- Persist users/sessions/wallet/stats.
- Transactional wallet/stats updates at round resolution.
- Optional daily faucet.

### Week 4: Deployment + ops polish
- Dockerize + persistent volume for DB.
- Cloudflare Tunnel routing to `blackjack.dylanjmccormick.com`.
- Add anti-abuse controls and confirm header trust rules.
- Add health + metrics endpoints.
- Ensure TUI never corrupts terminal output with logs.

## Acceptance criteria (demo)
- Fresh user can login via GitHub device flow and play.
- Re-login revokes old session; active connection receives `session_revoked` then closes.
- Second concurrent WebSocket connection for same user is denied.
- Wallet/stats persist across server restart.
- CI green and race detector passes.
- Deployed behind Cloudflare at `wss://blackjack.dylanjmccormick.com`.
