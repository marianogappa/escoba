package escoba

import (
	"math/rand"
)

// Bot interface for choosing actions in escoba
type Bot interface {
	ChooseAction(gameState GameState) Action
}

// SimpleBot is a basic bot that chooses actions randomly
type SimpleBot struct{}

// NewBot creates a new simple bot
func NewBot() Bot {
	return &SimpleBot{}
}

// ChooseAction chooses a random action from the possible actions
func (b *SimpleBot) ChooseAction(gameState GameState) Action {
	actions := gameState.CalculatePossibleActions()
	if len(actions) == 0 {
		return nil
	}

	// Choose a random action
	randomIndex := rand.Intn(len(actions))
	return actions[randomIndex]
}
