# This is my TODO

## Server ToDo

- Make values parameters with a config file at some point
  - Make these configs shared... context?
- Start sending error messages back to the client from the server (use the popups)
- BUG: player joining mid-round causes game to crash
- Player should be BOOTED when they run out of funds
- Shutdown cleanup procedure

### Priority

- Server health checks /healthz

## UI TODO

- settings menu to change username or something like that
- Revamp login/server menu. Should probably transition pages so we're not accidentally sending multiple login requests
- Tables show actual amount of players in TUI
- Left Bar

## Features

- Login page and database that keeps track of users
- DB keeps track of player stats (lifetime earnings, hands played, etc)
- searchable table names?
- Host it somewhere/ actually play with some friends
- displaying messages from the game like "waiting for all players to finish betting"

## Done... I think

- table.go make sure that game isn't started already when start command is sent
- update package message to return a transport message like the other package method does
- BUG: when a player leaves a game and there are no active players, the server crashes
- BUG: commands are not updating for each page. Will need to fix that
- allow user to change bet amount
- selecting a server
- allow user to leave a table
- Might be nice to have a map of [uuid]*Client in the table

- BUG: Players are not being kicked from server. Or removed from tables correctly
