# Architecture Decisions for Blackjack TUI

## Concurrency Model

The server uses goroutines and channels to handle concurrent gameplay

- **Lobby**: Single goroutine for the lobby to handle all messages through channels. (actor pattern)
  - Makes it so we don't need to have mutexes
  - Thread-safe access to table registry
  - All commands for table management (create/delete) come through the lobby

- **Table**: One goroutine per table
  - Manages the game state for a particular game
  - Players communicate through channels so there is no shared state
  - prevents multiple operations happening on the game at once

- **Client**: Two goroutines per websocket connection
  - `readPump`: Reads from the websocket and sends to the table
  - `writePump`: Reads from the table/lobby and sends to the client

- **Session Manager**: Single goroutine processing commands
  - No need to have a mutex. All access is serialized
  - all commands go through the same channel

** Why? **:

- Easier to understand the state without mutexes
- No possibility of deadlocks
- Usually not a lot of need to access the same records at once.

## Design Decisions

### Github Oauth Flow

**Why?**: TUI-Friendly. Most people in the terminal have a github account
**Possible Alternatives**: SSH key-based auth. Roll my own auth. API Tokens

### SQLite instead of postgres

**Why?**: Didn't need a complex distributed database. Makes deployment simple.
**Trade-Off**: Doesn't scale well. If I ran more than one server I couldn't access the same sqlite db. Since it is a single file
**Migration**: Could move to postgres if needed. Would need to have a lot of users

### In-Memory game state

**Why?**: Rounds are short. One hand of blackjack getting deleted is not the end of the world. It's all fake money.
**Trade-Off**: State will be lost if server crashes or restarts
**Future**: Could persist game state to a database, but really I'm not going to need that anytime soon

### Websocket for game communication

**Why?**: Real time updates for the game. Better than polling. Good foundation for other game servers as well
**Trade-off**: More complex than rest, but cool to have live game state

## Package structure

├── client/ # Bubbletea TUI (client-side only)
├── server/ # WebSocket server, lobby, table management
├── game/ # Pure game logic (no I/O, fully testable)
├── protocol/ # Shared message types (client + server)
├── auth/ # GitHub OAuth + session management
├── store/ # Database access layer (sqlc generated)
└── internal/
    └── database/ # sqlc models
