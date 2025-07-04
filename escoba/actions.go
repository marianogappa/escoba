package escoba

import (
	"fmt"
	"slices"
)

const (
	THROW_CARD = "throw_card"
)

type act struct {
	Name string `json:"name"`
}

func (a act) GetName() string {
	return a.Name
}

func (a act) YieldsTurn(g GameState) bool {
	return true
}

// ActionThrowCard represents throwing a card and potentially capturing table cards
type ActionThrowCard struct {
	act
	Card               Card   `json:"card"`
	CapturedTableCards []Card `json:"capturedTableCards"`
}

func newActionThrowCard(card Card, capturedTableCards []Card) Action {
	return ActionThrowCard{
		act:                act{Name: THROW_CARD},
		Card:               card,
		CapturedTableCards: capturedTableCards,
	}
}

func (a ActionThrowCard) IsPossible(g GameState) bool {
	// Check if player has this card
	hasCard := false
	for _, card := range g.Hands[g.TurnPlayerID].Cards {
		if card == a.Card {
			hasCard = true
			break
		}
	}
	if !hasCard {
		return false
	}

	// Check if the captured table cards are valid
	if len(a.CapturedTableCards) == 0 {
		// This is a simple throw - valid only if no combinations are possible
		return len(g.findAllValidCombinations(a.Card, g.TableCards)) == 0
	}

	// Check if this specific combination is valid
	expectedSum := 15 - a.Card.GetEscobaValue()
	actualSum := 0
	for _, tableCard := range a.CapturedTableCards {
		actualSum += tableCard.GetEscobaValue()
	}

	if actualSum != expectedSum {
		return false
	}

	// Check if all captured cards are actually on the table
	tableCopy := make([]Card, len(g.TableCards))
	copy(tableCopy, g.TableCards)

	for _, capturedCard := range a.CapturedTableCards {
		found := false
		for i, tableCard := range tableCopy {
			if tableCard == capturedCard {
				// Remove this card from the copy to handle duplicates
				tableCopy = append(tableCopy[:i], tableCopy[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (a ActionThrowCard) Run(g *GameState) error {
	playerID := g.TurnPlayerID

	// Remove card from player's hand
	for i, card := range g.Hands[playerID].Cards {
		if card == a.Card {
			g.Hands[playerID].Cards = slices.Delete(g.Hands[playerID].Cards, i, i+1)
			break
		}
	}

	if len(a.CapturedTableCards) > 0 {
		// Capture the combination
		cardsToCapture := make([]Card, 0, len(a.CapturedTableCards)+1)
		cardsToCapture = append(cardsToCapture, a.CapturedTableCards...)
		cardsToCapture = append(cardsToCapture, a.Card)
		g.Piles[playerID] = append(g.Piles[playerID], cardsToCapture...)

		// Remove captured cards from table
		g.TableCards = g.removeCardsFromTable(g.TableCards, cardsToCapture)

		g.LastCapturerPlayerID = playerID

		// Check if this is an escoba (table was cleared)
		if len(g.TableCards) == 0 && len(a.CapturedTableCards) > 0 {
			g.Escobas[playerID]++
		}
	} else {
		// No combination, card goes to table
		g.TableCards = append(g.TableCards, a.Card)
	}

	return nil
}

func (a ActionThrowCard) String() string {
	if len(a.CapturedTableCards) == 0 {
		return fmt.Sprintf("throw %s to table", a.Card.String())
	}

	captured := "["
	for i, card := range a.CapturedTableCards {
		if i > 0 {
			captured += ", "
		}
		captured += card.String()
	}
	captured += "]"

	return fmt.Sprintf("throw %s and capture %s", a.Card.String(), captured)
}

// findAllValidCombinations finds all possible combinations of table cards that sum to (15 - thrownCardValue)
func findAllValidCombinations(thrownCard Card, tableCards []Card) [][]Card {
	targetSum := 15
	thrownValue := thrownCard.GetEscobaValue()

	if thrownValue >= targetSum {
		return [][]Card{} // Card value too high to make 15
	}

	remainingSum := targetSum - thrownValue
	var allCombinations [][]Card

	// Use DFS to find all valid combinations
	findCombinationsDFS(tableCards, remainingSum, []Card{}, 0, &allCombinations)

	return allCombinations
}

// findCombinationsDFS recursively finds all combinations that sum to targetSum
func findCombinationsDFS(tableCards []Card, targetSum int, currentCombination []Card, startIndex int, allCombinations *[][]Card) {
	// Base case: if target sum is 0, we found a valid combination
	if targetSum == 0 {
		// Make a copy of the current combination and add it to results
		combination := make([]Card, len(currentCombination))
		copy(combination, currentCombination)
		*allCombinations = append(*allCombinations, combination)
		return
	}

	// If target sum is negative or we've exhausted all cards, return
	if targetSum < 0 || startIndex >= len(tableCards) {
		return
	}

	// For each remaining card, try including it in the combination
	for i := startIndex; i < len(tableCards); i++ {
		card := tableCards[i]
		cardValue := card.GetEscobaValue()

		// Create a proper copy to avoid slice sharing issues
		newCombination := make([]Card, len(currentCombination), len(currentCombination)+1)
		copy(newCombination, currentCombination)
		newCombination = append(newCombination, card)

		// Include this card and recurse
		findCombinationsDFS(tableCards, targetSum-cardValue, newCombination, i+1, allCombinations)
	}
}

// Update the GameState method to use the standalone function
func (g *GameState) findAllValidCombinations(thrownCard Card, tableCards []Card) [][]Card {
	return findAllValidCombinations(thrownCard, tableCards)
}

// removeCardsFromTable removes the specified cards from the table
func (g *GameState) removeCardsFromTable(tableCards []Card, toRemove []Card) []Card {
	// Create set from tableCards
	tableSet := make(map[Card]struct{})
	for _, card := range tableCards {
		tableSet[card] = struct{}{}
	}

	// Remove cards that are in toRemove
	for _, card := range toRemove {
		delete(tableSet, card)
	}

	// Convert back to slice
	result := make([]Card, 0, len(tableSet))
	for card := range tableSet {
		result = append(result, card)
	}

	return result
}
