# Blackjack TUI

![Tests](https://github.com/dylanmccormick/blackjack-tui/actions/workflows/test.yml/badge.svg)

![demo](./demo.gif)

## Description

This is an implementation of the game blackjack in a Terminal UI format. The application has two parts, the server and the client. Eventually it will be finished so you can launch either with a CLI command and play blackjack over the internet with your friends. To start this project I used Claude to create a project spec which I could follow and implement to work on my programming chops. I have used Claude as well to try and think through problems, but all of the code in the project is grass-fed and hand-written by yours truly.

## Motivation

Have you ever thought that people are making games too complicated? Why do I need JAVASCRIPT when I can just run things in the terminal? Why do I need a WEBSITE when I can just have a server? These are the questions I'm trying to answer by building blackjack TUI. The terminal-based blackjack game that prevents you from having to interact with javascript ever again.

My goal in creating this project is to learn more about how to design software well. I'm learning that the key to good software is minimizing complexity. I intend to get this reviewed by some people in a discord chat when I'm done and see what kind of improvements I could make on it. Really the goal is to become the best software engineer I can and learn to do things excellently. Also I think building projects like this will help hone my instinct on how software should be built and designed.

## Quick Start

Play on Public Server
go install github.com/dylanmccormick/blackjack-tui@latest
blackjack-tui tui
Run Your Own Server

### 1. Get the code

```
git clone https://github.com/dylanmccormick/blackjack-tui
cd blackjack-tui
```

### 2. Set up environment

`cp .env.example .env`

### Edit .env with your GitHub OAuth Client ID

```bash
# GitHub OAuth Application Client ID
# Create one at: https://github.com/settings/developers
GIT_CLIENT_ID=your_client_id_here
# Database location
SQLITE_DB=./blackjack-db.db
```

### Run server

The server will automatically run on localhost:42069. You can change the `config.yaml` file to update what port you want the server to work on.
You could update config.yaml to change any of the settings for the server
To run the server you can run the command:

`go run . server`

In another terminal, run client

`go run . tui`

## Usage

### Arguments:

tui - This argument will run the TUI for the game
server - this argument will run the server for the game

Available Options: "tui", "server"

--mock -- run the TUI in mock mode to be able to see the changes you make without needing to connect to a server
`blackjack-tui tui --mock`

## How to play

You will need a github login (I assume you have one if you're reading this). To start you can select one of the servers in the server menu or host your own server. From that screen you will be able to log in to github to start playing blackjack! You will get income every day that you visit the application and there may be a bonus for streaks and a special hidden bonus (⭐?).

Then you can create a new table and start playing blackjack against the computer! The commands should be on screen to tell you what buttons to press :)

The server is hosted on my own hardware so if it is down... sorry probably Russian spies trying to steal my data.

## Contributing

Thanks for checking out my project! If you have any suggestions or tips for me feel free to send me a message or open an issue. I'd love to hear what you have to say.

## What I learned in this project

**Go Concurrency Patterns:**

- Actor model with channels for safe concurrent state management
- Goroutine lifecycle management and graceful shutdown
- Context propagation for cancellation and timeouts

**Backend Systems:**

- WebSocket protocol and real-time bidirectional communication
- GitHub OAuth Device Flow for CLI-friendly authentication
- SQLite with Go using sqlc for type-safe queries
- Session management and authentication middleware

**Production Deployment:**

- Docker multi-stage builds for minimal images
- Traefik reverse proxy configuration
- Structured logging for production debugging
- Health check endpoints and graceful shutdown

**TUI Development:**

- Bubble Tea (Elm architecture) for terminal applications
- Managing async updates in terminal UIs
- Command-based navigation patterns

## Why Go?

I chose to use go for this project because it is a common tool for command-line applications. I wanted to have a strongly-typed language that forced me to think about how my data is modeled. The language I have used most in my career is Python, but that does not reinforce good habits when it comes to modeling data and object-oriented design. Go also has idiomatic error checking and a good standard library. I also like that there are good packages for websockets (gorilla websockets) and for terminal UIs (bubbletea).

## Architecture

```
┌────────────┐   HTTP       ┌────────────┐ HTTP   ┌─────────────┐
│ TUI Client │◄────────────►│   Server   │◄─────► │ Github      │
│ (Bubbletea)│              │            │        │ Oauth       │
│            │◄────────────►│            │        │ Device Flow │
└────────────┘  WebSocket   └────────────┘        └─────────────┘
             (json protocol)      ▲
                                  │
                                  ▼
                            ┌────────────┐
                            │ SQLite DB  │
                            │            │
                            │ User Stats │
                            │            │
                            └────────────┘
```
