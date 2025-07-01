package escoba

import (
	"fmt"
	"math/rand"
)

const (
	ORO    = "oro"
	COPA   = "copa"
	ESPADA = "espada"
	BASTO  = "basto"
)

// Card represents a Spanish deck card.
type Card struct {
	// Suit is the card's suit, which can be "oro", "copa", "espada" or "basto".
	Suit string `json:"suit"`

	// Number is the card's number, from 1 to 12.
	Number int `json:"number"`
}

func (c Card) String() string {
	return fmt.Sprintf("%d de %s", c.Number, c.Suit)
}

// GetEscobaValue returns the value of the card in Escoba (1-7 same, 10=8, 11=9, 12=10)
func (c Card) GetEscobaValue() int {
	if c.Number <= 7 {
		return c.Number
	}
	// 10, 11, 12 become 8, 9, 10
	return c.Number - 2
}

type deck struct {
	cards []Card
}

// Hand represents a player's hand.
type Hand struct {
	Cards []Card `json:"cards"`
}

func (h *Hand) String() string {
	if len(h.Cards) == 0 {
		return "empty hand"
	}
	result := "["
	for i, card := range h.Cards {
		if i > 0 {
			result += ", "
		}
		result += card.String()
	}
	result += "]"
	return result
}

func makeSpanishCards() []Card {
	cards := []Card{}
	suits := []string{ORO, COPA, ESPADA, BASTO}
	for _, suit := range suits {
		for i := 1; i <= 12; i++ {
			if i == 8 || i == 9 {
				continue
			}
			cards = append(cards, Card{Suit: suit, Number: i})
		}
	}

	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})

	return cards
}

func newDeck() *deck {
	return &deck{cards: makeSpanishCards()}
}

func (d *deck) dealHand() *Hand {
	hand := &Hand{Cards: []Card{}}
	for i := 0; i < 3; i++ {
		if len(d.cards) > 0 {
			hand.Cards = append(hand.Cards, d.cards[0])
			d.cards = d.cards[1:]
		}
	}
	return hand
}
