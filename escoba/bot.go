package escoba

import (
	"sort"
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
	if len(actions) == 1 {
		return actions[0]
	}

	throwActions := make([]ActionThrowCard, 0)
	for _, action := range actions {
		if action.GetName() == THROW_CARD {
			throwActions = append(throwActions, action.(ActionThrowCard))
		}
	}
	if len(throwActions) == 0 {
		return nil
	}
	if len(throwActions) == 1 {
		return throwActions[0]
	}

	sort.Slice(throwActions, func(i, j int) bool {
		actionI := throwActions[i]
		actionJ := throwActions[j]

		// 1. if .IsEscoba(), it's priority 1
		isEscobaI := actionI.IsEscoba(&gameState)
		isEscobaJ := actionJ.IsEscoba(&gameState)
		if isEscobaI != isEscobaJ {
			return isEscobaI
		}

		// 2. if .CapturesSieteDeVelos(), it's priority 2
		hasSieteI := actionI.CapturesSieteDeVelos()
		hasSieteJ := actionJ.CapturesSieteDeVelos()
		if hasSieteI != hasSieteJ {
			return hasSieteI
		}

		// 3.
		return isLeftBetterThanRight(actionI, actionJ, gameState)
	})

	return throwActions[0]
}

func isLeftBetterThanRight(left ActionThrowCard, right ActionThrowCard, gameState GameState) bool {
	scoreLeft := caresAboutCardCount(gameState)*leftHasMoreCards(left, right) + caresAboutOroCount(gameState)*leftHasMoreOros(left, right) + caresAboutSetenta(gameState)*leftHasMoreSetenta(left, right)
	scoreRight := caresAboutCardCount(gameState)*leftHasMoreCards(right, left) + caresAboutOroCount(gameState)*leftHasMoreOros(right, left) + caresAboutSetenta(gameState)*leftHasMoreSetenta(right, left)
	return scoreLeft > scoreRight
}

func leftHasMoreCards(left ActionThrowCard, right ActionThrowCard) int {
	if left.CardCount() > right.CardCount() {
		return 1
	}
	return 0
}

func leftHasMoreOros(left ActionThrowCard, right ActionThrowCard) int {
	if left.OrosCount() > right.OrosCount() {
		return 1
	}
	return 0
}

func leftHasMoreSetenta(left ActionThrowCard, right ActionThrowCard) int {
	if setentaPotential(left) > setentaPotential(right) {
		return 1
	}
	return 0
}

func setentaPotential(action ActionThrowCard) int {
	return action.CardSetentaSum()
}

func caresAboutCardCount(gameState GameState) int {
	if cardCount(gameState) > 20 {
		return 0
	}
	return 1
}

func cardCount(gameState GameState) int {
	return len(gameState.Piles[1])
}

func caresAboutOroCount(gameState GameState) int {
	if oroCount(gameState) > 5 {
		return 0
	}
	return 1
}

func oroCount(gameState GameState) int {
	count := 0
	for _, card := range gameState.Piles[1] {
		if card.Suit == ORO {
			count++
		}
	}
	return count
}

func caresAboutSetenta(gameState GameState) int {
	if laSetentaScore(gameState) > 22 {
		return 0
	}
	return 1
}

func laSetentaScore(gameState GameState) int {
	return gameState.calculateSetenta(1)
}
