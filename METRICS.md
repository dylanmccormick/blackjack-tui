# Metrics to Implement

1.  Connection Duration (histogram)

- **Name:** `blackjack_connection_duration_seconds`
- **Type:** Histogram
- **Buckets:** [60, 300, 600, 1800, 3600, 7200]
- **Where to start timer:** [server - serveWs()]
- **Where to observe:** [table/lobby? whenever a player logs out]
- **Question it answers:** How long do players stay connected?

2.  Bet Amounts (histogram or counter?)

- **Name:** `blackjack_bet_amount` or `blackjack_bets_total`
- **Type:** Histogram -- we can try to guess avg, mean, median bet 
- **Buckets/Labels:** 10, 50, 100, 200, 500, 1000
- **Where to observe:** [table.go]
- **Question it answers:** What's the distribution of bet sizes?

3.  Round Duration (histogram with labels)

- **Name:** `blackjack_round_duration_seconds`
- **Type:** Histogram
- **Buckets:** [seconds, 10, 20, 30, 60, 120]
- **Labels:** `player_count`
- **Where to start timer:** [table.go autoProgress()]
- **Where to observe:** [table.go autoProgress()]
- **Question it answers:** Does round time correlate with player count?

4.  Simple Gauges/Counters (easy wins)

- Active connections
- Active tables
- Total bets made
- Total payout (like if we were a real casino we'd want to know what we're making)
- "Game income"
- "Game loss"
