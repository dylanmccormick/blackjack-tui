# This is my TODO

## Server ToDo

- Make values paramaters with a config file at some point
- generate id from websocket to make sure that it is sticky.
- BUG: player joining mid-round causes game to crash
- Player should be BOOTED when they run out of funds

## UI TODO

- selecting a server
- settings menu to change username or something like that
- allow user to change bet amount
- BUG: commands are not updating for each page. Will need to fix that
- allow user to leave a table

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
