# Escoba de 15

A Go implementation of the Spanish card game "Escoba de 15" with WebSocket-based multiplayer support.

## About Escoba de 15

Escoba de 15 is a traditional Spanish card game played with a Spanish deck of 40 cards (no 8s or 9s). The goal is to capture cards from the table by matching them with cards from your hand to sum exactly 15.

### Game Rules

- **Players**: 2 players
- **Deck**: Spanish deck of 40 cards (1-7, 10-12 in each of 4 suits: oro, copa, espada, basto)
- **Card Values**: 1-7 = face value, 10=8, 11=9, 12=10
- **Goal**: First player to reach 15 points wins

### Gameplay

1. **Setup**: 4 cards are dealt to the table, 3 cards to each player
2. **Turns**: Players take turns throwing one card from their hand
3. **Capturing**: If your thrown card + any combination of table cards = 15, you capture those cards
4. **Escoba**: If you capture ALL table cards, you get an "escoba" (1 point)
5. **Rounds**: After both players exhaust their hands, new hands are dealt
6. **Set End**: When the deck is exhausted, points are awarded

### Scoring (at end of each set)

1. **Escobas**: 1 point per escoba made during the set
2. **Most Cards**: 1 point to player with more captured cards
3. **Most Oro Cards**: 1 point to player with more oro (gold) suit cards
4. **Seven of Oro**: 1 point to player who captured the 7 of oro
5. **La Setenta**: 1 point to player with highest "setenta" score

**La Setenta** calculation: For each suit, take your highest card ≤ 7. Sum all four values. Must have at least one card of each suit to qualify.

## Installation

```bash
go build -o escoba-game .
```

## Usage

### Start the server
```bash
./escoba-game server
```

### Connect players
```bash
# Terminal 1
./escoba-game player1

# Terminal 2  
./escoba-game player2
```

### Custom server address
```bash
./escoba-game player1 localhost:8080
```

### Environment Variables
- `PORT`: Server port (default: 8080)

## Architecture

- **WebSocket Server**: Handles game state and player connections
- **Game Engine**: Pure Go implementation of Escoba rules
- **Terminal Client**: Interactive terminal-based UI using termbox-go

## Key Components

- `escoba/`: Core game logic and rules
- `server/`: WebSocket server and message handling  
- `exampleclient/`: Terminal-based game client
- `main.go`: CLI entry point

## Development

Run tests:
```bash
go test ./...
```

The game maintains complete state synchronization between server and clients through JSON-serialized game states sent over WebSockets.

## Game State

The server broadcasts the complete game state after each action, including:
- Current round and turn information
- Player hands and scores
- Table cards and captured piles
- Escoba counts and possible actions
- Set results and game end conditions
