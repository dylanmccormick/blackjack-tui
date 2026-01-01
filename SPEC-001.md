# Project Specification: Multiplayer Blackjack over WebSockets

**Project Code**: SPEC-001  
**Version**: 1.1  
**Date**: December 13, 2025  
**Complexity Level**: Advanced  
**Estimated Timeline**: 60 hours (30 days @ 2 hours/day)

---

## 1. Project Overview


### Project Name & Description
**BlackjackHub** - A production-grade, real-time multiplayer blackjack server with terminal-based client interface. Players connect via WebSocket to join virtual tables, playing against an automated dealer. The system supports concurrent games across multiple tables with proper state management, resilience, and observability.

### Learning Objectives
By completing this project, you will:

- Design and implement a clean, maintainable architecture for a stateful real-time system
- Master Go concurrency patterns for managing multiple game sessions safely
- Build production-quality WebSocket communication with proper error handling and reconnection logic
- Implement comprehensive testing strategies (unit, integration, end-to-end)
- Create a responsive TUI application using Go
- Handle complex state machines (game flow, player actions, dealer logic)
- Deploy and operate a real service on a home server with monitoring
- Practice idiomatic Go throughout (error handling, interfaces, composition, channels)
- Build observability into a system from the ground up

### Target Complexity
**Advanced** - This project assumes strong Go fundamentals and focuses on production-quality patterns, testing, concurrency, and deployment.

### Estimated Timeline

| Phase | Focus | Hours | Weeks |
|-------|-------|-------|-------|
| Phase 1 | Core Foundation | 28 hours | Weeks 1-2 |
| Phase 2 | Production Features | 28 hours | Weeks 3-4 |
| Buffer | Testing, Polish, Docs | 4 hours | End of Week 4 |
| **Total** | | **60 hours** | **4 weeks** |

---

## 2. User Personas

### Persona 1: Casey - The Casual Player
- **Background**: Software developer who enjoys card games during breaks
- **Goals**: Quick, fun blackjack games without complicated setup
- **Pain Points**: Doesn't want to create accounts or deal with web browser overhead; prefers terminal tools
- **How BlackjackHub Serves Them**: Simple CLI client, join a table instantly, play a few hands, leave cleanly

### Persona 2: Morgan - The Competitive Player
- **Background**: Card game enthusiast who likes tracking performance
- **Goals**: Play multiple hands, test different strategies, see statistics
- **Pain Points**: Wants reliable connection, fair dealing, clear game state
- **How BlackjackHub Serves Them**: Stable WebSocket connection with reconnection, clear TUI display of all cards and actions, consistent dealer behavior

### Persona 3: Alex - The Self-Hoster
- **Background**: DevOps engineer who runs services on home infrastructure
- **Goals**: Host a blackjack server friends can connect to
- **Pain Points**: Needs easy deployment, monitoring, and maintenance
- **How BlackjackHub Serves Them**: Containerized deployment, structured logging, graceful shutdown, health checks

---

## 3. Non-Functional Requirements

### Performance Goals
- **Game Action Latency**: Player actions (hit/stand/etc.) should reflect in game state within 100ms
- **WebSocket Message Delivery**: Messages should be delivered to all players at a table within 200ms
- **Concurrent Tables**: System should support at least 20 concurrent tables without degradation
- **Memory Usage**: Server should run comfortably within 256MB RAM under normal load
- **TUI Responsiveness**: Client interface should update within 50ms of receiving server messages

### Scalability Considerations
- Architecture should support horizontal scaling (even if not implemented initially)
- Game state should be isolated per table to avoid global locks
- Connection handling should use efficient patterns for 100+ concurrent connections
- Consider how you'd add persistence/database later without major refactoring

### Maintainability Goals
- **Code Organization**: Clear separation of concerns (domain logic, network, UI)
- **Testing**: Minimum 80% coverage for game engine logic; integration tests for critical paths
- **Documentation**: README with setup/deployment, architecture decisions documented
- **Idiomatic Go**: Follow Go best practices; code should pass `go vet`, `staticcheck`
- **Dependencies**: Minimize external dependencies; justify each one

### Reliability/Availability
- **Uptime Target**: 99.9% uptime (important for a home server that should "just work")
- **Graceful Degradation**: Handle player disconnections without crashing table
- **Fault Isolation**: One table crash should not affect other tables
- **Recovery**: Automatic recovery from transient failures; clear error messages for permanent failures
- **Graceful Shutdown**: Server should finish in-progress games before shutting down

### Security Considerations
- **Input Validation**: All client messages must be validated and sanitized
- **Rate Limiting**: Prevent spam/abuse from malicious clients
- **Resource Limits**: Prevent resource exhaustion attacks
- **Future Auth**: Design with future authentication in mind (even if not implemented now)
- **Network Security**: Consider TLS for WebSocket connections (wss://)

---

## 4. Testing Strategy

### Testing Levels

**Unit Tests**
- Game engine logic (card dealing, hand evaluation, dealer AI, game rules)
- State transitions and edge cases
- Message validation and serialization
- Pure business logic isolated from I/O

**Integration Tests**
- WebSocket message flow between client and server
- Table management
- Multi-player scenarios
- Reconnection handling

**End-to-End Tests**
- Complete game flows from client connection to game completion
- Multiple clients playing simultaneously
- Server startup/shutdown behavior

### Coverage Goals
- **Game Engine**: 90%+ coverage (this is your core business logic)
- **Server Logic**: 80%+ coverage
- **Overall**: 75%+ coverage
- **Critical Paths**: 100% coverage for money/betting calculations, hand evaluation

### Testing Philosophy
**Test-Driven Development** recommended for game engine and core business logic. For infrastructure code (WebSocket handling, TUI), a hybrid approach is acceptable - write tests alongside or immediately after implementation.

### Key Testing Scenarios

**Game Logic**
- Dealer AI behaves correctly (hits on 16 or less, stands on 17+)
- Blackjack (21 with two cards) pays correctly
- Bust detection (over 21)
- Ace handling (soft 17 vs hard 17)
- Tie/push scenarios

**Concurrency**
- Multiple players acting on different tables simultaneously
- Race conditions in state management
- Concurrent connections/disconnections

**Edge Cases**
- Player disconnects mid-hand
- All players at table disconnect
- Server shutdown during active game
- Invalid player actions
- Message ordering and out-of-order delivery

**State Management**
- Game state transitions are correct
- Only valid actions allowed in each state
- State is correctly broadcast to all players

---

## 5. Architecture Hints (NOT Implementation)

### System Boundaries
Think of your system as having these major components:
- **Game Engine**: Pure domain logic for blackjack rules
- **Server**: WebSocket server managing connections and tables
- **Client**: TUI application that connects and displays game state
- **Communication Protocol**: Message format between client and server

### Key Architectural Decisions to Consider

**State Management**
- How do you ensure game state consistency when multiple goroutines are involved?
- Where does state live? (per table, centralized, distributed?)
- How do you handle concurrent reads and writes to game state?
- What happens to state when players disconnect?

**Concurrency Model**
- Do you use one goroutine per connection, per table, or something else?
- How do you safely communicate between goroutines?
- What patterns prevent deadlocks and race conditions?
- How do you gracefully shut down all goroutines?

**Error Handling**
- How do you distinguish between recoverable and fatal errors?
- What should happen when one table errors? Should it affect others?
- How do you communicate errors to clients in a user-friendly way?

**Protocol Design**
- What format for messages? (JSON, Protocol Buffers, custom binary?)
- How do you version your protocol for future changes?
- How do you handle message validation and malformed messages?
- Request/response vs event-driven vs hybrid?

**Persistence (Future)**
- Even if not implementing now, how would you add a database later?
- What would need to be persisted? (user accounts, game history, stats?)
- How would you design domain models to be persistence-agnostic?

### Relevant Patterns to Research

**Concurrency Patterns**
- Actor model for table management
- Worker pools for connection handling
- Fan-out/fan-in for broadcasting messages
- Context for cancellation and timeouts

**Domain Design**
- Domain-Driven Design for game logic
- State machines for game flow
- Repository pattern (even without persistence now)
- Hexagonal architecture (ports and adapters)

**Resilience Patterns**
- Circuit breaker for health checks
- Graceful degradation
- Backoff and retry logic
- Timeout patterns

### Go-Specific Considerations

**Standard Library Packages to Explore**
- `net/http` - WebSocket upgrade handling (via `gorilla/websocket` or `nhooyr.io/websocket`)
- `context` - Cancellation and request-scoped values
- `sync` - Mutexes, WaitGroups, atomic operations
- `encoding/json` - Message serialization
- `testing` - Test framework and table-driven tests
- `log/slog` - Structured logging (added in Go 1.21)
- `time` - Timeouts and tickers

**WebSocket Libraries**
Since WebSocket support isn't in stdlib, you'll need to choose:
- `gorilla/websocket` - Most popular, mature
- `nhooyr.io/websocket` - Modern, simpler API
- Research: What are the tradeoffs? Which fits your needs?

**TUI Libraries**
- `charmbracelet/bubbletea` - Elm architecture for TUIs, very popular
- `rivo/tview` - Widget-based TUI framework
- `gdamore/tcell` - Lower-level terminal handling
- Research: Which provides the right abstraction level for your UI needs?

**Testing Tools**
- `stretchr/testify` - Assertion library (optional, but popular)
- `google/go-cmp` - Deep comparison for tests
- Built-in `testing` package with table-driven tests

**Similar Projects for Inspiration**
- Look at multiplayer game servers in Go
- Study chat applications using WebSockets
- Examine TUI applications like `lazygit`, `k9s` for UI patterns
- **Do NOT copy code**, but study their architecture and patterns

---

## 6. Epics & Stories

### Epic 1: Game Engine Foundation
**Description**: Build the core blackjack game engine with all business logic, completely decoupled from network or UI concerns.

**Time Estimate**: 6-8 hours

**Learning Focus**: Domain modeling, state machines, test-driven development, pure business logic

---

#### Story 1.1: Card and Deck Management
**As a game engine**, I need to manage a deck of cards so that I can deal hands to players and the dealer.

**Description**: Implement the fundamental card and deck abstractions. A standard deck has 52 cards (4 suits × 13 ranks). The deck should support shuffling and dealing cards.

**Acceptance Criteria**:
1. Can create a standard 52-card deck
2. Can shuffle the deck randomly
3. Can deal a card from the deck (removes it from deck)
4. Deck tracks remaining cards
5. Can create a new deck when cards run low (configurable threshold)
6. Cards have rank (A, 2-10, J, Q, K) and suit (♠, ♥, ♦, ♣)

**Technical Considerations**:
- How do you represent a card? Struct with rank and suit?
- What data structure for the deck? Slice?
- How do you ensure shuffling is truly random?
- Should the deck be mutable or immutable?
- How do you handle when the deck runs out of cards?

**Questions to Think About Before Implementing**:
1. Should Rank and Suit be strings, ints, or custom types? What are the tradeoffs?
2. How will you test randomness in shuffling? What makes a good test for this?
3. Should dealing a card return a pointer or a value? Why?
4. In real blackjack, multiple decks are often used - how would your design accommodate this?
5. What's the idiomatic Go way to handle "deck is empty" scenarios?

**What You'll Learn**:
- Idiomatic Go type definitions
- Randomness and testing non-deterministic functions
- Value vs pointer semantics
- Designing simple, focused APIs

**Key Concepts to Research**:
- Go's `math/rand` or `crypto/rand` packages
- Fisher-Yates shuffle algorithm
- Custom types and type aliases in Go
- Table-driven testing patterns

**Time Estimate**: 1-1.5 hours

---

#### Story 1.2: Hand Evaluation
**As a game engine**, I need to evaluate blackjack hands so that I can determine hand values and compare them.

**Description**: Implement hand evaluation logic that correctly calculates hand values according to blackjack rules. Aces can count as 1 or 11 (whichever is more favorable), face cards count as 10.

**Acceptance Criteria**:
1. Calculate hand value with number cards (2-10) counting as face value
2. Face cards (J, Q, K) count as 10
3. Aces count as 11 if it doesn't bust the hand, otherwise 1
4. Correctly identify "soft" hands (Ace counted as 11) vs "hard" hands
5. Correctly identify a "bust" (value > 21)
6. Correctly identify "blackjack" (Ace + 10-value card as initial two cards)
7. Handle multiple Aces correctly (only one can be 11)

**Technical Considerations**:
- A Hand likely contains multiple cards - what's the relationship?
- How do you efficiently calculate value, especially with multiple Aces?
- Should hand evaluation be a method or a function?
- How do you represent the distinction between soft and hard hands?

**Questions to Think About Before Implementing**:
1. Should a Hand own its cards or just reference them?
2. How do you test all the edge cases (multiple Aces, soft 17, etc.)?
3. Should hand value be calculated on-demand or cached? What are the tradeoffs?
4. What's the clearest way to express the Ace logic without making it complex?
5. How would you represent the state "this hand has blackjack" - a bool field, a method, or something else?

**What You'll Learn**:
- Designing domain objects with behavior
- Handling complex conditional logic cleanly
- Comprehensive test case enumeration
- The difference between derived state and stored state

**Key Concepts to Research**:
- Value receivers vs pointer receivers in Go
- Table-driven tests for combinatorial cases
- Blackjack rules (soft 17, etc.)
- Getter methods - are they idiomatic in Go?

**Time Estimate**: 2 hours

---

#### Story 1.3: Game Flow State Machine
**As a game engine**, I need to manage the game flow through distinct states so that the game progresses correctly from betting to resolution.

**Description**: Implement a state machine that manages a single game round: betting → dealing → player actions → dealer actions → resolution. The engine should enforce that only valid actions can be taken in each state.

**Acceptance Criteria**:
1. Game starts in "WaitingForBets" state
2. After bets placed, transition to "Dealing" state
3. After dealing, transition to "PlayerTurn" state
4. Players can only take actions during their turn
5. After all players act, transition to "DealerTurn" state
6. After dealer completes, transition to "Resolution" state
7. After resolution, transition back to "WaitingForBets"
8. Invalid state transitions are prevented (return errors)
9. Game enforces turn order among multiple players

**Technical Considerations**:
- How do you represent state? (enum/const, state pattern, other?)
- How do you prevent invalid transitions?
- How do you track whose turn it is?
- What data needs to be associated with each state?

**Questions to Think About Before Implementing**:
1. Should you use the State pattern (OOP) or a simpler state enum with a switch statement?
2. How do you make state transitions clear and testable?
3. What happens if a player times out during their turn? (Consider, but maybe don't implement yet)
4. How would you log state transitions for debugging?
5. Should state transition logic be part of the Game type or separate?

**What You'll Learn**:
- State machine implementation in Go
- Encapsulation and invariant enforcement
- Error design for domain violations
- Modeling complex flows

**Key Concepts to Research**:
- State pattern vs simple state machines
- Go error wrapping and sentinel errors
- Finite state machines (FSM)
- `iota` for constant enumeration

**Time Estimate**: 2-2.5 hours

---

#### Story 1.4: Dealer AI and Game Resolution
**As a game engine**, I need automated dealer behavior and win/loss resolution so that games can complete correctly.

**Description**: Implement the dealer's automatic play strategy (hit on 16 or less, stand on 17 or higher) and game resolution logic that determines winners, losers, and pushes.

**Acceptance Criteria**:
1. Dealer automatically hits when hand value is 16 or less
2. Dealer automatically stands when hand value is 17 or more
3. Dealer stands on soft 17 (configurable, as rules vary)
4. Correctly determine winner: player > dealer (without bust), dealer busts, blackjack
5. Correctly identify "push" (tie)
6. Correctly handle player bust (loses regardless of dealer)
7. Blackjack (natural 21) beats regular 21
8. Calculate payouts: blackjack pays 3:2, regular win pays 1:1, push pays 0

**Technical Considerations**:
- Should dealer logic be a method on Game or a separate function?
- How do you represent payout calculations?
- What's the return value of resolution - a structure with results for each player?
- Should the game automatically play the dealer, or is it triggered?

**Questions to Think About Before Implementing**:
1. How do you test dealer behavior deterministically when it depends on random cards?
2. Should payout calculation be part of the game engine or separate?
3. How do you handle the case where all players bust before the dealer acts?
4. What data structure best represents "game results"?
5. How would you make dealer strategy configurable (for future house rules)?

**What You'll Learn**:
- Testing with controlled randomness (dependency injection of deck)
- Modeling calculation results
- Conditional logic organization
- Configurable behavior patterns

**Key Concepts to Research**:
- Dependency injection in Go
- Functional options pattern
- Table-driven tests with expected outcomes
- Blackjack payout rules and variations

**Time Estimate**: 2 hours

---

### Epic 2: WebSocket Server Infrastructure
**Description**: Build the WebSocket server that manages client connections and tables. This is the networking layer that coordinates multiple concurrent games.

**Time Estimate**: 8-10 hours

**Learning Focus**: WebSocket protocol, Go concurrency patterns, connection lifecycle management, graceful shutdown

---

#### Story 2.1: WebSocket Connection Handling
**As a server**, I need to accept and manage WebSocket connections so that clients can communicate with the game server.

**Description**: Implement basic WebSocket server that accepts connections, handles the upgrade from HTTP, and manages connection lifecycle (ping/pong, disconnection detection).

**Acceptance Criteria**:
1. Server listens on a configurable port
2. HTTP endpoint upgrades to WebSocket connection
3. Server accepts multiple concurrent connections
4. Implements ping/pong for connection health checks
5. Detects client disconnections (timeout, explicit close)
6. Cleans up resources when connections close
7. Handles graceful shutdown (stops accepting new connections, waits for existing)
8. Logs connection events (connect, disconnect, errors)

**Technical Considerations**:
- Which WebSocket library will you use and why?
- How do you manage the lifecycle of a connection (goroutine per connection?)
- How do you detect and handle broken connections?
- What context patterns ensure clean shutdown?

**Questions to Think About Before Implementing**:
1. Should each connection get its own goroutine, or use a worker pool pattern?
2. How do you test WebSocket connection handling without a real network?
3. What's the right ping/pong interval to detect disconnections promptly without overhead?
4. How do you propagate shutdown signals to all active connections?
5. What information should you log for observability without creating noise?

**What You'll Learn**:
- WebSocket protocol and Go WebSocket libraries
- Connection lifecycle management
- Context-based cancellation
- Graceful shutdown patterns
- Health check implementation (ping/pong)

**Key Concepts to Research**:
- `gorilla/websocket` or `nhooyr.io/websocket` documentation
- `context.Context` for cancellation
- `sync.WaitGroup` for coordinating goroutines
- WebSocket ping/pong frames
- Structured logging with `log/slog`

**Time Estimate**: 3-4 hours

---

#### Story 2.2: Message Protocol Definition
**As a server and client**, we need a well-defined message protocol so that we can communicate game state and actions reliably.

**Description**: Define the message format and protocol for all client-server communication. Messages should be typed, versioned, and validated.

**Acceptance Criteria**:
1. Define message types for all game actions (join table, place bet, hit, stand, etc.)
2. Define message types for all server events (game state update, error, etc.)
3. Messages are JSON-serializable
4. Messages include type/kind field for routing
5. Messages include version field for future compatibility
6. Define validation rules for each message type
7. Implement message serialization/deserialization
8. Implement message validation that returns clear errors

**Technical Considerations**:
- How do you structure messages to be extensible?
- How do you handle different message types - union types, type field, separate types?
- Where does validation logic live?
- How do you handle malformed JSON?

**Questions to Think About Before Implementing**:
1. Should you use one large union type or multiple specific types? Tradeoffs?
2. How do you version your protocol to allow future changes without breaking clients?
3. What level of validation happens at deserialization vs business logic layer?
4. How do you make message construction easy and type-safe?
5. Should you generate JSON schema or OpenAPI spec for your protocol?

**What You'll Learn**:
- Protocol design for real-time systems
- JSON marshaling/unmarshaling in Go
- Input validation patterns
- Type-safe message handling
- API versioning considerations

**Key Concepts to Research**:
- Go `encoding/json` struct tags
- JSON schema for validation
- Protocol versioning strategies
- Discriminated unions in Go
- `json.RawMessage` for flexible parsing

**Time Estimate**: 2-3 hours

---

#### Story 2.3: Table Management
**As a server**, I need to manage multiple tables so that players can find and join games.

**Description**: Implement table management system. Tables are the core unit where games happen. Each table has one active game and a maximum number of players (e.g., 5).

**Acceptance Criteria**:
1. Server maintains a registry of tables
2. Tables can be created dynamically when first player joins (or pre-created)
3. Tables have a maximum player capacity (configurable, default 5)
4. Players can list available tables
5. Players can join a specific table (if not full)
6. Players can leave a table
7. Empty tables are cleaned up after a timeout (or kept alive)
8. All operations are thread-safe for concurrent access
9. Tables can have optional metadata (name, min/max bet, etc.)

**Technical Considerations**:
- How do you represent tables and their registry?
- What concurrency patterns ensure thread-safety?
- How do you notify players when table state changes?
- Where does the game instance live in relation to a table?

**Questions to Think About Before Implementing**:
1. Should tables use locks, or should they use the actor pattern with channels?
2. How do you prevent race conditions when multiple players join the same table simultaneously?
3. What happens when the last player leaves a table - immediate cleanup or delayed?
4. How do you broadcast "table updated" events to all players at a table?
5. Should table IDs be auto-generated, human-readable, or UUID-based?

**What You'll Learn**:
- Concurrent data structure management
- Actor model vs mutex-based concurrency
- Resource lifecycle management
- Registry/manager patterns

**Key Concepts to Research**:
- Go `sync.RWMutex` vs channels for synchronization
- Actor model pattern in Go
- UUID generation libraries
- Map thread-safety in Go

**Time Estimate**: 3-4 hours

---

#### Story 2.4: Connection-to-Player Mapping and Session Management
**As a server**, I need to associate WebSocket connections with player sessions so that I can route messages and handle reconnections.

**Description**: Implement session management that maps connections to player identities, handles player state, and supports basic reconnection (player disconnects and comes back to same table).

**Acceptance Criteria**:
1. Each connection is associated with a unique player session
2. Player sessions persist briefly after disconnection (30-60 seconds)
3. Player can reconnect and resume their seat at a table
4. Player session includes: ID, current table, connection status, bet amount
5. Server broadcasts to all players at a table when someone joins/leaves/reconnects
6. Handle case where player tries to join multiple tables (prevent or allow?)
7. Clean up stale sessions after timeout
8. Thread-safe session management

**Technical Considerations**:
- How do you generate and manage session IDs?
- What data structure maps sessions to connections?
- How do you handle the "ghost" period between disconnect and session expiry?
- What happens to a player's bet and hand if they disconnect mid-game?

**Questions to Think About Before Implementing**:
1. Should session IDs be short-lived (per connection) or long-lived (support coming back hours later)?
2. How do you test reconnection logic?
3. What's the right timeout for session cleanup?
4. Should you allow a player to be connected from multiple devices/connections?
5. How do you communicate to other players that someone disconnected vs just left?

**What You'll Learn**:
- Session management patterns
- Stateful connection handling
- Timeout and cleanup patterns
- Player lifecycle modeling

**Key Concepts to Research**:
- Session token generation (UUID, random strings)
- `time.AfterFunc` for timeout-based cleanup
- Connection pooling and management
- Heartbeat patterns for connection health

**Time Estimate**: 2-3 hours

---

### Epic 3: TUI Client Application
**Description**: Build the terminal user interface client that connects to the server and provides an interactive game experience.

**Time Estimate**: 8-10 hours

**Learning Focus**: TUI framework usage, client-side state management, asynchronous UI updates, user input handling

---

#### Story 3.1: Basic TUI Framework Setup and Connection
**As a player**, I want to launch a TUI client that connects to the game server so that I can start playing.

**Description**: Set up the TUI framework and implement basic connection flow. Client should prompt for server address, connect via WebSocket, and display connection status.

**Acceptance Criteria**:
1. Client launches and displays a welcome screen
2. Client prompts for server address (or uses default/config)
3. Client attempts WebSocket connection to server
4. Client displays connection status (connecting, connected, failed)
5. Client handles connection errors gracefully with clear messages
6. Client can detect disconnection and attempt reconnection
7. Client exits cleanly on Ctrl+C or quit command
8. Basic logging for debugging (can be toggled)

**Technical Considerations**:
- Which TUI library will you use and why?
- How do you handle asynchronous events (messages from server) while updating UI?
- How do you structure the client's internal state?
- What happens if the server isn't reachable?

**Questions to Think About Before Implementing**:
1. How does your chosen TUI framework handle async updates (messages arriving from network)?
2. Should connection parameters be command-line flags, config file, or interactive prompt?
3. How do you make the UI responsive during connection attempts?
4. What's the right reconnection strategy (immediate, backoff, give up)?
5. How do you test a TUI application?

**What You'll Learn**:
- TUI framework patterns and event loops
- Async UI updates in terminal applications
- WebSocket client implementation
- Graceful error handling in user-facing applications

**Key Concepts to Research**:
- `charmbracelet/bubbletea` or `rivo/tview` tutorials
- The Elm architecture (if using bubbletea)
- WebSocket client in Go
- Terminal signal handling (SIGINT, SIGTERM)

**Time Estimate**: 2-3 hours

---

#### Story 3.2: Table Selection Interface
**As a player**, I want to browse available tables so that I can choose where to play.

**Description**: Implement UI screen for listing tables and joining a specific table.

**Acceptance Criteria**:
1. Display list of available tables with metadata (occupancy, min/max bet, etc.)
2. Allow player to select a table (keyboard navigation)
3. Send join-table message to server
4. Handle server response (success or error, e.g., "table full")
5. Display loading state while waiting for server response
6. Allow player to navigate back or refresh the list

**Technical Considerations**:
- How do you model the UI screens/views?
- How do you handle navigation between screens?
- How do you keep table lists up-to-date (polling, server push, hybrid)?
- What happens if table data changes while player is viewing it?

**Questions to Think About Before Implementing**:
1. Should table lists refresh automatically or require manual refresh?
2. How do you handle the case where a table fills up between listing and joining?
3. What's the best keyboard navigation pattern for your TUI framework?
4. How do you show rich information (e.g., table stakes, game in progress) in a compact terminal UI?
5. Should you cache table data or always fetch fresh?

**What You'll Learn**:
- Multi-screen TUI navigation
- List rendering and selection in terminal
- Handling server-driven state changes in UI
- User feedback for async operations

**Key Concepts to Research**:
- Navigation patterns in your chosen TUI framework
- Rendering tables/lists in terminal
- Optimistic UI updates vs server-confirmed state
- Keyboard shortcuts and accessibility in TUIs

**Time Estimate**: 2-3 hours

---

#### Story 3.3: Game Play Interface
**As a player**, I want to see the game state and take actions (bet, hit, stand) so that I can play blackjack.

**Description**: Implement the main game screen that displays dealer hand, player hands, chips/bets, and allows player to take actions.

**Acceptance Criteria**:
1. Display dealer's face-up cards (one hidden during play)
2. Display all players at the table with their hands and bets
3. Highlight the current player's turn
4. Display player's own hand prominently with value
5. Display available actions (bet, hit, stand, and later: split, double) based on game state
6. Allow player to select and submit actions via keyboard
7. Display chip count and current bet
8. Update display in real-time as game events arrive from server
9. Show game results (win/loss/push) clearly
10. Display errors if invalid action attempted

**Technical Considerations**:
- How do you layout multiple players' hands in limited terminal space?
- How do you represent cards visually (ASCII art, simple text)?
- How do you handle rendering updates without flicker?
- How do you map keyboard input to game actions?

**Questions to Think About Before Implementing**:
1. What's the visual design for showing cards - simple ("AS", "10H") or fancy ASCII art?
2. How do you show player state (waiting, active, busted, won, etc.) clearly?
3. Should animations happen (cards being dealt), or instant updates?
4. How do you handle multiple rapid updates from server?
5. What accessibility considerations matter for terminal UI (colorblind-friendly colors, etc.)?

**What You'll Learn**:
- Complex layout management in terminal
- Real-time UI updates from network events
- Input handling and validation
- Visual design in constrained environments

**Key Concepts to Research**:
- ASCII art generators for cards
- Terminal colors and styling
- Event-driven UI updates
- Keyboard input handling in your TUI framework

**Time Estimate**: 4-5 hours

---

### Epic 4: Game Integration and Core Gameplay Loop
**Description**: Integrate the game engine with the server and client to create a working end-to-end gameplay experience.

**Time Estimate**: 8-10 hours

**Learning Focus**: Integration patterns, state synchronization, event-driven architecture, debugging distributed systems

---

#### Story 4.1: Table-Level Game Instance Management
**As a server**, I need to run a game instance for each table and coordinate player actions with game state.

**Description**: Integrate the game engine into the table management system. Each table should have a game instance that processes player actions and progresses through game states.

**Acceptance Criteria**:
1. Each table creates a game instance when first player joins
2. Game accepts bets from players and transitions to dealing when all players ready
3. Game deals initial cards (2 to each player, 2 to dealer with one hidden)
4. Game enforces turn order among players
5. Game processes player actions (hit/stand) and validates them
6. Game automatically plays dealer hand when all players complete
7. Game calculates results and updates player chip counts
8. Game resets for next round
9. Game state is broadcast to all players at table after each change
10. Table handles concurrent action requests safely

**Technical Considerations**:
- How does the table receive player actions and forward them to the game?
- How does the game notify the table of state changes to broadcast?
- What's the concurrency model - one goroutine per table?
- How do you handle the timing of automatic actions (dealing, dealer play)?

**Questions to Think About Before Implementing**:
1. Should game engine be aware of networking, or should table act as an adapter?
2. How do you prevent race conditions when multiple players act "simultaneously"?
3. What's the right interface between table and game engine?
4. Should dealer actions be automatic or triggered by the table?
5. How do you test the integration without running the full server?

**What You'll Learn**:
- Adapter pattern for integrating components
- Observer pattern for state change notifications
- Goroutine-per-instance concurrency pattern
- Integration testing strategies

**Key Concepts to Research**:
- Adapter and observer patterns in Go
- Event sourcing (lightweight version for game events)
- Testing with goroutines and channels
- Time-based triggers (`time.After`, `time.Ticker`)

**Time Estimate**: 3-4 hours

---

#### Story 4.2: Client-Server Game State Synchronization
**As a system**, I need to keep all players at a table synchronized with current game state so that everyone sees the same game.

**Description**: Implement the full message flow for game state synchronization. Server broadcasts game events; clients update their UI based on received events.

**Acceptance Criteria**:
1. When game state changes, server sends update messages to all players at table
2. Messages include full game state (or state diffs, depending on design)
3. Clients receive messages and update UI immediately
4. Clients display dealer cards (with one hidden until dealer's turn)
5. Clients display all players' hands and bets
6. Clients show whose turn it is
7. Clients show game results after resolution
8. Messages are ordered correctly (or clients handle out-of-order messages)
9. New players joining mid-game receive current state
10. State synchronization is tested with multiple concurrent clients

**Technical Considerations**:
- Should you send full state snapshots or deltas?
- How do you ensure message ordering?
- What happens if a client misses a message?
- How do you represent "hidden" dealer card in messages?

**Questions to Think About Before Implementing**:
1. Full state vs delta updates - what are the tradeoffs for your use case?
2. How do you handle the case where clients get out of sync with server?
3. Should clients trust server completely, or validate state locally?
4. What's the performance impact of broadcasting to many players at a table?
5. How do you version game state messages for future compatibility?

**What You'll Learn**:
- State synchronization patterns in distributed systems
- Event-driven architecture
- Message ordering and reliability
- Handling eventual consistency in real-time applications

**Key Concepts to Research**:
- Event sourcing and CQRS patterns
- Optimistic UI updates
- Message queuing patterns
- WebSocket message ordering guarantees

**Time Estimate**: 3-4 hours

---

#### Story 4.3: End-to-End Gameplay Flow
**As a player**, I want to complete a full game from joining a table, placing bets, playing hands, seeing results, and playing again.

**Description**: Ensure the complete gameplay loop works end-to-end. This story is about polishing the integration and handling all the edge cases.

**Acceptance Criteria**:
1. Player can join table and see other players already seated
2. All players place bets, then game deals automatically
3. Each player takes turns (hit/stand) in order
4. Dealer plays automatically after all players
5. Results are calculated and displayed correctly
6. Chip counts update after each round
7. New round starts automatically (or after confirmation)
8. Players can leave table mid-game (their hand forfeits)
9. Handle "all players bust before dealer" scenario
10. Handle "one player at table" scenario
11. Integration test with 3+ simulated clients playing multiple rounds

**Technical Considerations**:
- What's the trigger for starting a new round?
- How long should players wait between rounds?
- How do you handle stragglers who don't bet in time?
- What's the behavior when a player leaves mid-hand?

**Questions to Think About Before Implementing**:
1. Should there be a countdown timer for placing bets, or wait indefinitely?
2. What happens if a player disconnects with active bet - does the hand play automatically or fold?
3. How do you test multi-player scenarios efficiently?
4. Should there be any rate limiting on actions to prevent abuse?
5. How do you make the pacing feel natural (not too fast, not too slow)?

**What You'll Learn**:
- End-to-end system testing
- Edge case handling in distributed systems
- User experience design for asynchronous systems
- Debugging complex interactions

**Key Concepts to Research**:
- Integration testing with multiple goroutines/processes
- Mock clients for testing
- User experience patterns in real-time multiplayer games
- Race condition debugging tools (`go test -race`)

**Time Estimate**: 2-3 hours

---

### Epic 5: Production Hardening and Testing
**Description**: Bring the system to production quality with comprehensive testing, error handling, observability, and resilience features.

**Time Estimate**: 10-12 hours

**Learning Focus**: Production-ready Go patterns, comprehensive testing strategies, observability, reliability

---

#### Story 5.1: Comprehensive Unit and Integration Testing
**As a developer**, I need comprehensive tests so that I can refactor confidently and catch regressions.

**Description**: Achieve high test coverage across all components with focus on critical paths and edge cases.

**Acceptance Criteria**:
1. Game engine has 90%+ test coverage
2. All game rules and edge cases are tested (blackjack, bust, soft hands, etc.)
3. State machine transitions are thoroughly tested
4. Message validation is thoroughly tested
5. Table management has tests for concurrent access
6. Integration tests cover multi-player scenarios
7. Tests run fast (entire suite under 10 seconds)
8. Tests are deterministic (no flaky tests)
9. `go test -race` passes without race conditions
10. CI-ready test setup (can run in automated pipeline)

**Technical Considerations**:
- How do you test concurrent code deterministically?
- How do you mock/stub WebSocket connections for testing?
- How do you structure tests for readability and maintainability?
- What test helpers reduce boilerplate?

**Questions to Think About Before Implementing**:
1. What's your strategy for testing goroutines and channels?
2. How do you test time-dependent behavior (timeouts, cleanup)?
3. Should you use mocking libraries or hand-roll mocks?
4. How do you balance unit tests vs integration tests?
5. What code coverage percentage is actually meaningful vs just chasing numbers?

**What You'll Learn**:
- Advanced Go testing techniques
- Testing concurrent code
- Test organization and structure
- Mocking and stubbing strategies
- Race detection

**Key Concepts to Research**:
- Table-driven tests in depth
- `testing.T` helper methods
- Interface-based dependency injection for testability
- `httptest` for testing HTTP/WebSocket servers
- Test fixtures and setup/teardown patterns

**Time Estimate**: 4-5 hours

---

#### Story 5.2: Structured Logging and Observability
**As an operator**, I need comprehensive logging and metrics so that I can monitor the server and debug issues.

**Description**: Implement structured logging throughout the application and add basic metrics for monitoring.

**Acceptance Criteria**:
1. All significant events are logged (connections, disconnections, game starts, errors)
2. Logs use structured format (JSON) with consistent fields
3. Logs include context (player ID, table ID, request ID)
4. Different log levels are used appropriately (debug, info, warn, error)
5. Log level is configurable (via flag or environment variable)
6. Sensitive data is not logged (if any exists)
7. Metrics exposed for: active connections, active tables, games per minute, errors
8. Metrics endpoint (e.g., Prometheus format) available for scraping
9. Logs are performant (don't impact game latency)
10. Log rotation is considered (external tool or built-in)

**Technical Considerations**:
- Which logging library will you use?
- How do you pass context through the system for logging?
- What's the performance impact of structured logging?
- How do you expose metrics - HTTP endpoint, push to external system?

**Questions to Think About Before Implementing**:
1. Should you use `log/slog` (stdlib) or a third-party library like `zerolog`?
2. How do you propagate request IDs through concurrent operations?
3. What's the right balance of logging detail vs noise?
4. Should logs be written to stdout/stderr or files?
5. How do you test that logging happens correctly without cluttering test output?

**What You'll Learn**:
- Structured logging best practices
- Context propagation patterns
- Metrics and observability
- Production debugging techniques
- Performance considerations for logging

**Key Concepts to Research**:
- `log/slog` package (Go 1.21+)
- Prometheus metrics format
- Request ID and correlation ID patterns
- Context values in Go
- Log levels and when to use each

**Time Estimate**: 2-3 hours

---

#### Story 5.3: Graceful Shutdown and Resource Cleanup
**As an operator**, I need the server to shut down gracefully so that players don't lose game state and resources are cleaned up.

**Description**: Implement graceful shutdown that stops accepting new connections, finishes in-progress games, disconnects clients cleanly, and releases all resources.

**Acceptance Criteria**:
1. Server listens for shutdown signals (SIGINT, SIGTERM)
2. On shutdown signal, server stops accepting new WebSocket connections
3. Server allows in-progress games to complete (with timeout)
4. Server notifies connected clients of impending shutdown
5. Server closes all WebSocket connections cleanly
6. Server waits for all goroutines to finish (with timeout)
7. Server logs shutdown progress
8. Server exits with appropriate exit code (0 for clean shutdown)
9. If shutdown times out, server logs warnings and force-exits
10. Integration test verifies graceful shutdown behavior

**Technical Considerations**:
- How do you propagate shutdown signal to all goroutines?
- What's the right timeout for in-progress games?
- How do you ensure no goroutine leaks?
- What happens to players' chips/bets if forced to shut down?

**Questions to Think About Before Implementing**:
1. How do you test graceful shutdown programmatically?
2. Should you allow a "drain" period before shutdown, or start shutdown immediately?
3. What's the user experience for players when server shuts down?
4. How do you prevent new players from joining during shutdown?
5. Should shutdown state be persisted so players can resume later?

**What You'll Learn**:
- Signal handling in Go
- Context cancellation patterns
- Goroutine lifecycle management
- Resource cleanup patterns
- Graceful degradation

**Key Concepts to Research**:
- `signal.Notify` and signal handling
- Context cancellation and `context.WithCancel`
- `sync.WaitGroup` for goroutine synchronization
- Timeout patterns with `context.WithTimeout`
- Idempotent cleanup operations

**Time Estimate**: 2-3 hours

---

#### Story 5.4: Error Handling and Resilience
**As a system**, I need robust error handling so that errors don't crash the server or corrupt game state.

**Description**: Review and harden error handling throughout the system. Ensure errors are handled gracefully, logged, and communicated to users appropriately.

**Acceptance Criteria**:
1. All error paths are handled (no ignored errors)
2. Errors are wrapped with context for debugging
3. Panics are recovered in goroutines (where appropriate)
4. Invalid client messages don't crash server
5. One table's errors don't affect other tables (fault isolation)
6. Clients receive user-friendly error messages
7. Critical errors are logged with full context
8. Transient errors trigger retries (with backoff)
9. Permanent errors fail fast with clear messages
10. Error handling is tested (inject errors, verify handling)

**Technical Considerations**:
- How do you distinguish between recoverable and fatal errors?
- Where should panics be recovered, and where should they propagate?
- How do you test error paths comprehensively?
- What's the boundary between "log and continue" vs "fail fast"?

**Questions to Think About Before Implementing**:
1. Should you use sentinel errors, error types, or error wrapping for different error categories?
2. How do you prevent error handling code from obscuring business logic?
3. What errors should result in connection termination vs just rejecting an action?
4. How do you make error messages helpful without exposing internal details?
5. Should you implement circuit breaker pattern anywhere?

**What You'll Learn**:
- Go error handling idioms
- Error wrapping and unwrapping
- Panic and recover patterns
- Fault isolation techniques
- Building resilient systems

**Key Concepts to Research**:
- `errors.Wrap` and `errors.Unwrap`
- Sentinel errors and `errors.Is`
- Custom error types and `errors.As`
- Panic recovery with `defer` and `recover()`
- Circuit breaker pattern

**Time Estimate**: 2-3 hours

---

### Epic 6: Enhanced Gameplay Features
**Description**: Add advanced blackjack features (split, double down) and betting system to make the game more complete.

**Time Estimate**: 6-8 hours

**Learning Focus**: Extending existing architecture, maintaining clean design as complexity grows, backward-compatible changes

---

#### Story 6.1: Betting System and Chip Management
**As a player**, I want to bet virtual chips so that the game has stakes and meaning.

**Description**: Implement a chip/betting system. Players start with a balance, place bets each round, and win/lose chips based on outcomes.

**Acceptance Criteria**:
1. Players start with a configurable initial chip balance (e.g., 1000)
2. Players can set bet amount before each round (within min/max limits)
3. Bets are deducted from balance when round starts
4. Wins add chips to balance (1:1 for regular win, 3:2 for blackjack)
5. Losses deduct chips (already deducted at bet time)
6. Pushes return bet to player
7. Players cannot bet more than their balance
8. Players with zero balance can receive a "rebuy" (configurable)
9. Chip balances persist during session (lost on disconnect)
10. UI displays chip balance and current bet prominently

**Technical Considerations**:
- How do you handle chip arithmetic (int or float? precision issues?)
- Where does balance state live (player session, game engine, both?)
- How do you handle rounding for 3:2 blackjack payouts?
- What happens if a player runs out of chips mid-session?

**Questions to Think About Before Implementing**:
1. Should chips be integers (cents) to avoid floating-point errors?
2. How do you test payout calculations comprehensively?
3. Should there be table minimum/maximum bets?
4. What's the user experience when a player goes broke?
5. If you add persistence later, what needs to change in your design?

**What You'll Learn**:
- Financial calculation patterns
- State management across components
- User experience for resource management
- Validation and constraints

**Key Concepts to Research**:
- Avoiding floating-point errors in financial calculations
- Decimal libraries in Go
- Blackjack payout rules (3:2 vs 6:5)
- UX patterns for betting interfaces

**Time Estimate**: 2 hours

---

#### Story 6.2: Split Functionality
**As a player**, I want to split pairs so that I can play optimal blackjack strategy.

**Description**: Implement the split action. When a player has two cards of the same rank, they can split into two separate hands, each with its own bet.

**Acceptance Criteria**:
1. Split option is available only when player has pair (same rank)
2. Split option is available only on initial two cards
3. Player must have enough chips to match original bet for second hand
4. Split creates two separate hands from the original pair
5. Each hand receives one additional card immediately
6. Player plays each hand separately (turn order: hand 1, then hand 2)
7. Each hand is resolved independently (can win one, lose the other)
8. Cannot re-split (or limit to one re-split, configurable)
9. Aces split to one card each (cannot hit further) - traditional rule
10. UI shows both hands clearly and indicates which is active

**Technical Considerations**:
- How does the game engine represent multiple hands for one player?
- How do you track which hand is currently active?
- How does splitting affect turn order?
- What data structures change to accommodate splits?

**Questions to Think About Before Implementing**:
1. Should your Hand model support being part of a split, or is that table/game state?
2. How do you handle betting for split hands in your chip system?
3. What's the UI pattern for showing multiple hands in limited space?
4. How do you test all split scenarios (win/lose combinations)?
5. Should you allow re-splitting? What are the tradeoffs?

**What You'll Learn**:
- Extending domain models without breaking existing code
- Handling complex game rule variations
- UI challenges with dynamic layouts
- Testing combinatorial scenarios

**Key Concepts to Research**:
- Blackjack split rules and variations
- Dynamic UI layouts in TUI
- Refactoring existing code safely with tests
- State modeling for complex scenarios

**Time Estimate**: 3-4 hours

---

#### Story 6.3: Double Down Functionality
**As a player**, I want to double down on favorable hands so that I can maximize winning opportunities.

**Description**: Implement the double down action. Player doubles their bet, receives exactly one more card, then turn ends.

**Acceptance Criteria**:
1. Double down option available only on initial two cards (before hitting)
2. Player must have enough chips to double bet
3. Bet is doubled when double down is chosen
4. Player receives exactly one more card
5. Player's turn ends immediately after receiving card
6. Payout is calculated on doubled bet amount
7. Cannot double down after hitting
8. UI shows doubled bet amount clearly
9. Works correctly in combination with splits (can double each split hand)
10. Integration test covers double down win/loss scenarios

**Technical Considerations**:
- How do you enforce "one more card only" after double down?
- Where do you track that a hand has doubled down?
- How does this interact with the state machine?
- Does doubling the bet happen immediately or at resolution?

**Questions to Think About Before Implementing**:
1. Should double down be a state on the hand, or just an action that triggers specific behavior?
2. How do you ensure you can't hit after doubling down?
3. What's the UI feedback when a player tries to double without enough chips?
4. How do you test double down doesn't break existing game flow?
5. Some casinos restrict double down to certain hand values (9-11) - should this be configurable?

**What You'll Learn**:
- Adding features to existing state machines
- Action validation and constraints
- Maintaining backward compatibility
- Incremental feature development

**Key Concepts to Research**:
- Blackjack double down rules and variations
- Feature flags for optional rules
- Refactoring with tests as safety net
- State machine modification patterns

**Time Estimate**: 2 hours

---

### Epic 7: Deployment and Operations
**Description**: Prepare the application for deployment to a home server with proper packaging, configuration, and operational tooling.

**Time Estimate**: 6-8 hours

**Learning Focus**: Go application deployment, containerization, systemd, operational best practices, configuration management

---

#### Story 7.1: Configuration Management
**As an operator**, I need flexible configuration so that I can tune the server without recompiling.

**Description**: Implement configuration system supporting flags, environment variables, and config files for all tunable parameters.

**Acceptance Criteria**:
1. Server port is configurable
2. WebSocket ping interval is configurable
3. Table size (max players) is configurable
4. Initial chip balance is configurable
5. Bet limits (min/max) are configurable
6. Session timeout is configurable
7. Log level is configurable
8. Configuration via: command-line flags, environment variables, config file (YAML or TOML)
9. Configuration precedence is clear (flags > env > config file > defaults)
10. `--help` shows all configuration options with defaults
11. Invalid configuration produces clear error messages
12. Sample configuration file is provided

**Technical Considerations**:
- Which configuration library will you use (if any)?
- How do you validate configuration values?
- How do you handle sensitive values (future: API keys, passwords)?
- Should configuration be reloadable without restart?

**Questions to Think About Before Implementing**:
1. Should you use a library like `viper` or `cobra`, or hand-roll config parsing?
2. What's the right balance between configurability and sensible defaults?
3. How do you prevent invalid configurations (e.g., negative timeouts)?
4. Should configuration be logged at startup (for debugging)?
5. How do you document all configuration options for users?

**What You'll Learn**:
- Configuration management patterns
- CLI flag parsing in Go
- Environment variable handling
- Configuration validation
- User-facing documentation

**Key Concepts to Research**:
- `flag` package vs `pflag` vs `cobra`
- `viper` for configuration management
- YAML/TOML parsing in Go
- 12-factor app configuration principles
- Secure handling of secrets

**Time Estimate**: 2 hours

---

#### Story 7.2: Containerization with Docker
**As an operator**, I want to run the server in Docker so that deployment is reproducible and isolated.

**Description**: Create a Dockerfile and docker-compose setup for building and running the server.

**Acceptance Criteria**:
1. Dockerfile builds server binary in multi-stage build
2. Final image is minimal (Alpine or distroless)
3. Image runs as non-root user
4. Health check endpoint is defined
5. docker-compose.yml for easy local testing
6. Configuration can be passed via environment variables
7. Logs go to stdout/stderr for container logging
8. Graceful shutdown works in container (handles SIGTERM)
9. Image is tagged with version
10. Build is reproducible (same inputs = same image)

**Technical Considerations**:
- What base image provides the best size/security tradeoff?
- How do you handle Go module dependencies in Docker build?
- How do you minimize image size?
- What user/permissions should the container run with?

**Questions to Think About Before Implementing**:
1. Should you use scratch, Alpine, or distroless base image?
2. How do you optimize Docker layer caching for faster builds?
3. Should you build the binary in the Dockerfile or use pre-built binary?
4. What healthcheck endpoint makes sense (HTTP /health that checks...what)?
5. How do you test the Docker image locally before deploying?

**What You'll Learn**:
- Docker multi-stage builds
- Go application containerization
- Container best practices
- Image optimization techniques
- Docker Compose for local development

**Key Concepts to Research**:
- Multi-stage Dockerfile patterns for Go
- Distroless images
- Container security best practices
- Docker healthchecks
- Optimizing Docker build cache

**Time Estimate**: 2-3 hours

---

#### Story 7.3: Systemd Service and Deployment
**As an operator**, I want to run the server as a systemd service so that it starts on boot and restarts on failure.

**Description**: Create systemd unit file and deployment documentation for running on a home server.

**Acceptance Criteria**:
1. Systemd unit file provided for server
2. Service starts automatically on boot
3. Service restarts on failure (with backoff)
4. Service logs to journald
5. Service runs as dedicated non-root user
6. Service environment variables can be configured
7. Service shutdown is graceful (uses SIGTERM)
8. Deployment documentation includes: installation, configuration, starting/stopping, logs
9. Simple update/rollback procedure documented
10. Service status shows health (using `Type=notify` or simple health check)

**Technical Considerations**:
- Where should the binary and config file live (`/opt`, `/usr/local/bin`, etc.)?
- What user/group should the service run as?
- How do you handle log rotation?
- What restart policy makes sense?

**Questions to Think About Before Implementing**:
1. Should the service use `Type=simple`, `Type=notify`, or `Type=forking`?
2. What's the right restart policy (`on-failure`, `always`, etc.)?
3. How do you manage configuration file updates?
4. Should you provide an install script or manual instructions?
5. How do you handle database/state persistence when you add it later?

**What You'll Learn**:
- Systemd unit file creation
- Service management on Linux
- Deployment automation
- Operational documentation
- Production deployment practices

**Key Concepts to Research**:
- Systemd unit file format
- `systemctl` commands
- `journalctl` for log viewing
- Linux service user creation
- File system hierarchy standard (FHS)

**Time Estimate**: 2 hours

---

#### Story 7.4: Monitoring and Health Checks
**As an operator**, I need health checks and monitoring so that I know when the server has problems.

**Description**: Implement health check endpoint and integrate with monitoring tools for observability.

**Acceptance Criteria**:
1. HTTP health check endpoint (e.g., `/health`) returns 200 when healthy
2. Health check verifies critical components (can accept connections, etc.)
3. Metrics endpoint (e.g., `/metrics`) exposes Prometheus-format metrics
4. Key metrics tracked: active connections, active games, errors, uptime
5. Health check used by Docker healthcheck
6. Health check used by systemd (if using `Type=notify`)
7. Simple monitoring setup documented (Prometheus + Grafana or similar)
8. Alerting rules suggested (service down, error rate high)
9. Dashboard example provided (Grafana JSON or similar)
10. Health check is tested in integration tests

**Technical Considerations**:
- Should health check be on the same port as WebSocket, or separate?
- What makes the server "unhealthy"?
- How do you expose metrics without adding complexity?
- What's the performance impact of metrics collection?

**Questions to Think About Before Implementing**:
1. What criteria determine "healthy" vs "unhealthy"?
2. Should health check be deep (verify database connection) or shallow (just responding)?
3. How granular should metrics be (per-table stats, or aggregate)?
4. Should you use a metrics library, or hand-roll Prometheus format?
5. How do you test that metrics are being collected correctly?

**What You'll Learn**:
- Health check patterns
- Prometheus metrics format
- Observability best practices
- Monitoring and alerting setup
- SRE practices for home services

**Key Concepts to Research**:
- HTTP health check conventions
- Prometheus metrics types (counter, gauge, histogram)
- Grafana dashboard creation
- Alerting best practices
- `promhttp` Go library

**Time Estimate**: 2 hours

---

## 7. Success Criteria

### Production-Ready Definition
This project is considered "production-ready" when:

1. **Functionality**: Core gameplay works end-to-end with multiple concurrent players
2. **Testing**: 75%+ code coverage, integration tests pass, no race conditions
3. **Reliability**: Runs for 24+ hours without crashes, handles disconnections gracefully
4. **Observability**: Comprehensive logs and metrics enable debugging and monitoring
5. **Deployment**: Can be deployed via Docker and systemd with clear documentation
6. **Performance**: Supports 20+ concurrent tables with <200ms action latency
7. **Security**: Input validation, rate limiting, runs as non-root
8. **Documentation**: README with architecture overview, setup instructions, API/protocol docs

### Overall Acceptance Criteria
- [ ] Multiple players can join different tables and play concurrent games
- [ ] All basic blackjack rules work correctly (hit, stand, blackjack, bust)
- [ ] Split and double down work correctly
- [ ] Betting and chip management works
- [ ] Server handles graceful shutdown without data loss
- [ ] Comprehensive test suite runs in CI
- [ ] No race conditions detected by `go test -race`
- [ ] Deployed to home server and accessible over network
- [ ] Monitoring dashboard shows key metrics
- [ ] Code follows Go idioms and passes `go vet`, `staticcheck`

### Learning Success Indicators
You'll know you've successfully learned from this project when you can:

1. **Explain Architectural Decisions**: Articulate why you chose specific concurrency patterns, state management approaches, and communication protocols
2. **Demonstrate Testing Strategy**: Walk through your testing pyramid and explain how you test concurrent code, edge cases, and integrations
3. **Discuss Production Concerns**: Explain your observability strategy, error handling philosophy, and deployment architecture
4. **Code Review Confidently**: Review your own code and identify areas for improvement, refactoring opportunities, and tradeoffs
5. **Extend the System**: Add new features (like insurance, side bets) without major refactoring
6. **Debug Effectively**: Use logs, metrics, and tests to identify and fix issues
7. **Interview Readiness**: Confidently discuss this project in technical interviews, explaining challenges, solutions, and learnings

---

## Appendix: Suggested Development Order

While epics are presented logically, here's a recommended implementation order:

### Week 1 (Days 1-7)
1. **Epic 1**: Complete game engine (Stories 1.1-1.4)
2. **Epic 2.1-2.2**: Basic WebSocket server and protocol
3. **Epic 3.1**: Basic TUI client connection

### Week 2 (Days 8-14)
4. **Epic 2.3-2.4**: Table management and sessions
5. **Epic 4.1**: Integrate game engine with tables
6. **Epic 3.2-3.3**: Complete TUI interface
7. **Epic 4.2-4.3**: Full gameplay integration

### Week 3 (Days 15-21)
8. **Epic 5.1**: Comprehensive testing
9. **Epic 5.2-5.4**: Production hardening (logging, shutdown, errors)
10. **Epic 6.1**: Betting system

### Week 4 (Days 22-30)
11. **Epic 6.2-6.3**: Split and double down
12. **Epic 7.1-7.4**: Deployment and operations
13. **Final Polish**: Documentation, cleanup, final testing

---

## Appendix: Key Technologies and Resources

### Go Learning Resources
- Official Go Tour: https://tour.golang.org
- Effective Go: https://golang.org/doc/effective_go
- Go Concurrency Patterns: https://www.youtube.com/watch?v=f6kdp27TYZs
- Advanced Go Concurrency Patterns: https://www.youtube.com/watch?v=QDDwwePbDtw

### WebSocket Resources
- gorilla/websocket examples
- WebSocket protocol RFC 6455
- Real-time systems design patterns

### TUI Development
- Charm Bracelet tutorials (if using bubbletea)
- Terminal escape codes and ANSI standards

### Testing Resources
- Go testing best practices
- Table-driven tests in Go
- Testing concurrent code
- Test coverage interpretation

### Deployment
- Docker documentation
- Systemd documentation
- Prometheus/Grafana setup guides
- 12-factor app methodology

---

**End of Specification**

This specification is designed to guide you through building a production-quality multiplayer blackjack system while learning idiomatic Go, concurrent programming, testing strategies, and deployment practices. Remember: the goal is not just to build the application, but to deeply understand the principles and patterns that make it production-ready.

Good luck, and enjoy the learning journey! 🎰♠️♥️♦️♣️
