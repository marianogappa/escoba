package escoba

import (
	"encoding/json"
	"errors"
	"fmt"
)

// GameState represents the state of an Escoba game.
type GameState struct {
	// RoundTurnPlayerID is the player ID of the player who starts the round, or "mano".
	RoundTurnPlayerID int `json:"roundTurnPlayerID"`

	// RoundNumber is the number of the current round, starting from 1.
	RoundNumber int `json:"roundNumber"`

	// TurnPlayerID is the player ID of the player whose turn it is to play an action.
	TurnPlayerID int `json:"turnPlayerID"`

	// Hands is a map of player IDs to their respective hands.
	Hands map[int]*Hand `json:"hands"`

	// TableCards are the cards currently on the table
	TableCards []Card `json:"tableCards"`

	// Piles are the captured cards for each player
	Piles map[int][]Card `json:"piles"`

	// Escobas count for each player (points from clearing the table)
	Escobas map[int]int `json:"escobas"`

	// Scores is a map of player IDs to their respective scores.
	// Scores go from 0 to 15.
	Scores map[int]int `json:"scores"`

	// PossibleActions is a list of possible actions that the current player can take.
	PossibleActions []json.RawMessage `json:"possibleActions"`

	// RoundFinished is true if the current round is finished.
	RoundFinished bool `json:"roundFinished"`

	// SetFinished is true if the current set of rounds is finished (deck exhausted).
	SetFinished bool `json:"setFinished"`

	// IsEnded is true if the whole game is ended.
	IsEnded bool `json:"isEnded"`

	// WinnerPlayerID is the player ID of the player who won the game.
	// Set to -1 if the game ended in a draw.
	WinnerPlayerID int `json:"winnerPlayerID"`

	// Actions is the list of actions that have been run in the game.
	Actions []json.RawMessage `json:"actions"`

	// ActionOwnerPlayerIDs is the list of player IDs who ran each action.
	ActionOwnerPlayerIDs []int `json:"actionOwnerPlayerIDs"`

	// RoundJustStarted is true if a round has just started.
	RoundJustStarted bool `json:"roundJustStarted"`

	// LastSetResults contains the results of the last completed set
	LastSetResults *SetResult `json:"lastSetResults"`

	// LastCapturerPlayerID tracks the last player who made a capture (for remaining table cards)
	LastCapturerPlayerID int `json:"lastCapturerPlayerID"`

	// SetJustStarted is true if a set has just started.
	SetJustStarted bool `json:"setJustStarted"`

	deck *deck `json:"-"`
}

// SetResult contains the scoring results for a completed set of rounds
type SetResult struct {
	CardCounts     map[int]int  `json:"cardCounts"`     // Number of cards in each player's pile
	OroCardCounts  map[int]int  `json:"oroCardCounts"`  // Number of oro cards in each player's pile
	HasSieteDeOro  map[int]bool `json:"hasSieteDeOro"`  // Whether each player has the 7 of oro
	SetentaScores  map[int]int  `json:"setentaScores"`  // La setenta scores for each player
	PointsAwarded  map[int]int  `json:"pointsAwarded"`  // Points awarded to each player
	EscobasThisSet map[int]int  `json:"escobasThisSet"` // Escobas made in this set
}

func (sr *SetResult) String() string {
	result := "Set Results:\n"
	for playerID := 0; playerID <= 1; playerID++ {
		result += fmt.Sprintf("  Player %d: %d cards (%d oro), escobas: %d, setenta: %d, points awarded: %d",
			playerID, sr.CardCounts[playerID], sr.OroCardCounts[playerID],
			sr.EscobasThisSet[playerID], sr.SetentaScores[playerID], sr.PointsAwarded[playerID])
		if sr.HasSieteDeOro[playerID] {
			result += " [has 7 de oro]"
		}
		result += "\n"
	}
	return result
}

func New(opts ...func(*GameState)) *GameState {
	gs := &GameState{
		RoundTurnPlayerID:    0, // Player 0 starts as mano
		RoundNumber:          0,
		LastCapturerPlayerID: 0, // Initialize to player 0 (mano) as default
		Scores:               map[int]int{0: 0, 1: 0},
		Hands:                map[int]*Hand{0: nil, 1: nil},
		TableCards:           []Card{},
		Piles:                map[int][]Card{0: {}, 1: {}},
		Escobas:              map[int]int{0: 0, 1: 0},
		IsEnded:              false,
		WinnerPlayerID:       -1,
		Actions:              []json.RawMessage{},
		deck:                 newDeck(),
	}

	for _, opt := range opts {
		opt(gs)
	}

	gs.startNewSet()
	return gs
}

func (g *GameState) startNewSet() {
	g.deck = newDeck() // Fresh deck for each set
	g.TableCards = []Card{}
	g.Piles = map[int][]Card{0: {}, 1: {}}
	g.Escobas = map[int]int{0: 0, 1: 0}
	g.LastCapturerPlayerID = g.RoundTurnPlayerID // Reset to current mano
	g.RoundNumber = 0
	g.SetFinished = false
	g.SetJustStarted = true
	g.startNewRound()
}

func (g *GameState) startNewRound() {
	g.RoundJustStarted = true
	g.RoundNumber++
	g.TurnPlayerID = g.RoundTurnPlayerID

	// Deal 3 cards to each player
	if len(g.deck.cards) >= 6 {
		g.Hands[0] = g.deck.dealHand()
		g.Hands[1] = g.deck.dealHand()
	} else {
		// No more cards, set is finished
		g.SetFinished = true
		g.scoreSet()
		return
	}

	// On first round, deal 4 cards to table
	if g.RoundNumber == 1 {
		for i := 0; i < 4; i++ {
			if len(g.deck.cards) > 0 {
				g.TableCards = append(g.TableCards, g.deck.cards[0])
				g.deck.cards = g.deck.cards[1:]
			}
		}

		// Check if table cards sum to 15 (dealer gets an escoba)
		if g.sumCards(g.TableCards) == 15 {
			g.Piles[g.RoundTurnPlayerID] = append(g.Piles[g.RoundTurnPlayerID], g.TableCards...)
			g.Escobas[g.RoundTurnPlayerID]++
			g.LastCapturerPlayerID = g.RoundTurnPlayerID // Track dealer as last capturer
			g.TableCards = []Card{}
		}
	}

	g.RoundFinished = false
	g.PossibleActions = _serializeActions(g.CalculatePossibleActions())
}

func (g *GameState) RunAction(action Action) error {
	if g.IsEnded {
		return errGameIsEnded
	}

	if !action.IsPossible(*g) {
		return errActionNotPossible
	}

	err := action.Run(g)
	if err != nil {
		return err
	}

	g.RoundJustStarted = false
	g.SetJustStarted = false
	bs := SerializeAction(action)
	g.Actions = append(g.Actions, bs)
	g.ActionOwnerPlayerIDs = append(g.ActionOwnerPlayerIDs, g.CurrentPlayerID())

	// Check if round is finished (both players have no cards)
	if len(g.Hands[0].Cards) == 0 && len(g.Hands[1].Cards) == 0 {
		g.RoundFinished = true
	}

	// Start new round if current round is finished
	if !g.IsEnded && g.RoundFinished && !g.SetFinished {
		g.startNewRound()
		return nil
	}

	// Handle set completion
	if g.SetFinished {
		return nil
	}

	// Switch player turn
	if !g.IsEnded && !g.RoundFinished && action.YieldsTurn(*g) {
		g.TurnPlayerID = g.OpponentOf(g.TurnPlayerID)
	}

	g.PossibleActions = _serializeActions(g.CalculatePossibleActions())
	return nil
}

func (g *GameState) scoreSet() {
	result := &SetResult{
		CardCounts:     make(map[int]int),
		OroCardCounts:  make(map[int]int),
		HasSieteDeOro:  make(map[int]bool),
		SetentaScores:  make(map[int]int),
		PointsAwarded:  make(map[int]int),
		EscobasThisSet: map[int]int{0: g.Escobas[0], 1: g.Escobas[1]},
	}

	// Remaining table cards go to the last player who captured
	if len(g.TableCards) > 0 {
		g.Piles[g.LastCapturerPlayerID] = append(g.Piles[g.LastCapturerPlayerID], g.TableCards...)
		g.TableCards = []Card{}
	}

	// Count total cards and oro cards for each player
	for playerID := 0; playerID <= 1; playerID++ {
		result.CardCounts[playerID] = len(g.Piles[playerID])
		result.OroCardCounts[playerID] = 0
		result.HasSieteDeOro[playerID] = false

		for _, card := range g.Piles[playerID] {
			if card.Suit == ORO {
				result.OroCardCounts[playerID]++
				if card.Number == 7 {
					result.HasSieteDeOro[playerID] = true
				}
			}
		}

		// Calculate la setenta
		result.SetentaScores[playerID] = g.calculateSetenta(playerID)
	}

	// Award points
	// 1. Escobas
	for playerID := 0; playerID <= 1; playerID++ {
		result.PointsAwarded[playerID] += g.Escobas[playerID]
	}

	// 2. Most cards
	if result.CardCounts[0] > result.CardCounts[1] {
		result.PointsAwarded[0]++
	} else if result.CardCounts[1] > result.CardCounts[0] {
		result.PointsAwarded[1]++
	}

	// 3. Most oro cards
	if result.OroCardCounts[0] > result.OroCardCounts[1] {
		result.PointsAwarded[0]++
	} else if result.OroCardCounts[1] > result.OroCardCounts[0] {
		result.PointsAwarded[1]++
	}

	// 4. Seven of oro
	if result.HasSieteDeOro[0] {
		result.PointsAwarded[0]++
	} else if result.HasSieteDeOro[1] {
		result.PointsAwarded[1]++
	}

	// 5. La setenta
	if result.SetentaScores[0] > result.SetentaScores[1] && result.SetentaScores[0] > 0 {
		result.PointsAwarded[0]++
	} else if result.SetentaScores[1] > result.SetentaScores[0] && result.SetentaScores[1] > 0 {
		result.PointsAwarded[1]++
	}

	// Apply points to scores
	for playerID := 0; playerID <= 1; playerID++ {
		g.Scores[playerID] += result.PointsAwarded[playerID]
	}

	g.LastSetResults = result

	// Check for game end
	if g.Scores[0] >= 15 || g.Scores[1] >= 15 {
		g.IsEnded = true
		if g.Scores[0] > g.Scores[1] {
			g.WinnerPlayerID = 0
		} else if g.Scores[1] > g.Scores[0] {
			g.WinnerPlayerID = 1
		} else {
			// Draw: both players have equal points >= 15
			g.WinnerPlayerID = -1
		}
	} else {
		// Start new set
		g.RoundTurnPlayerID = g.OpponentOf(g.RoundTurnPlayerID) // Switch mano
		g.startNewSet()
	}
}

func (g *GameState) calculateSetenta(playerID int) int {
	// For each suit, find the highest card <= 7
	suitBest := make(map[string]int)
	suitHasCard := make(map[string]bool)

	for _, card := range g.Piles[playerID] {
		value := card.GetEscobaValue()
		if value <= 7 {
			if !suitHasCard[card.Suit] || value > suitBest[card.Suit] {
				suitBest[card.Suit] = value
				suitHasCard[card.Suit] = true
			}
		}
	}

	// Must have at least one card of each suit
	if len(suitHasCard) < 4 {
		return 0
	}

	total := 0
	for _, value := range suitBest {
		total += value
	}
	return total
}

func (g *GameState) sumCards(cards []Card) int {
	sum := 0
	for _, card := range cards {
		sum += card.GetEscobaValue()
	}
	return sum
}

func (g GameState) CurrentPlayerID() int {
	return g.TurnPlayerID
}

func (g GameState) OpponentOf(playerID int) int {
	if playerID == 0 {
		return 1
	}
	return 0
}

// IsDraw returns true if the game ended in a draw
func (g GameState) IsDraw() bool {
	return g.IsEnded && g.WinnerPlayerID == -1
}

func (g GameState) CalculatePossibleActions() []Action {
	var actions []Action
	hasValidCombinations := false

	// For each card in player's hand, find all valid combinations
	for _, card := range g.Hands[g.TurnPlayerID].Cards {
		validCombinations := g.findAllValidCombinations(card, g.TableCards)

		if len(validCombinations) > 0 {
			hasValidCombinations = true
			// Create an action for each valid combination
			for _, combination := range validCombinations {
				actions = append(actions, newActionThrowCard(card, combination))
			}
		}
	}

	// If no valid combinations exist for any card, allow simple throws
	if !hasValidCombinations {
		for _, card := range g.Hands[g.TurnPlayerID].Cards {
			actions = append(actions, newActionThrowCard(card, []Card{}))
		}
	}

	return actions
}

var (
	errActionNotPossible = errors.New("action not possible")
	errGameIsEnded       = errors.New("game is ended")
)

func _serializeActions(as []Action) []json.RawMessage {
	_as := []json.RawMessage{}
	for _, a := range as {
		_as = append(_as, json.RawMessage(SerializeAction(a)))
	}
	return _as
}

func SerializeAction(action Action) []byte {
	bs, _ := json.Marshal(action)
	return bs
}

func DeserializeAction(bs []byte) (Action, error) {
	var actionName struct {
		Name string `json:"name"`
	}

	err := json.Unmarshal(bs, &actionName)
	if err != nil {
		return nil, err
	}

	var action Action
	switch actionName.Name {
	case THROW_CARD:
		action = &ActionThrowCard{}
	default:
		return nil, fmt.Errorf("unknown action type %v", actionName.Name)
	}

	err = json.Unmarshal(bs, action)
	if err != nil {
		return nil, err
	}

	return action, nil
}

type Action interface {
	IsPossible(g GameState) bool
	Run(g *GameState) error
	GetName() string
	YieldsTurn(g GameState) bool
	String() string
}

func (g *GameState) TableCardsString() string {
	if len(g.TableCards) == 0 {
		return "table: empty"
	}
	result := "table: ["
	for i, card := range g.TableCards {
		if i > 0 {
			result += ", "
		}
		result += card.String()
	}
	result += "]"
	return result
}

func (g *GameState) GameStateString() string {
	result := fmt.Sprintf("=== Round %d, Player %d's turn ===\n", g.RoundNumber, g.TurnPlayerID)
	result += fmt.Sprintf("Scores: P0=%d, P1=%d\n", g.Scores[0], g.Scores[1])
	result += fmt.Sprintf("Escobas: P0=%d, P1=%d\n", g.Escobas[0], g.Escobas[1])
	result += fmt.Sprintf("Player 0 hand: %s\n", g.Hands[0].String())
	result += fmt.Sprintf("Player 1 hand: %s\n", g.Hands[1].String())
	result += fmt.Sprintf("%s\n", g.TableCardsString())
	return result
}
