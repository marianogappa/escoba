package exampleclient

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/marianogappa/escoba/escoba"
	"github.com/nsf/termbox-go"
)

type ui struct {
	wantKeyPressCh chan struct{}
	sendKeyPressCh chan rune
}

func NewUI() *ui {
	ui := &ui{
		wantKeyPressCh: make(chan struct{}),
		sendKeyPressCh: make(chan rune),
	}
	ui.startKeyEventLoop()
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	return ui
}

func (u *ui) Close() {
	termbox.Close()
}

func (u *ui) play(playerID int, gameState escoba.GameState) (escoba.Action, error) {
	if err := u.render(playerID, gameState, PRINT_MODE_NORMAL); err != nil {
		return nil, err
	}

	if gameState.IsEnded {
		return nil, nil
	}

	var (
		action          escoba.Action
		possibleActions = _deserializeActions(gameState.PossibleActions)
	)
	for {
		num := u.pressAnyNumber()
		if num > len(possibleActions) {
			continue
		}
		action = possibleActions[num-1]
		break
	}
	return action, nil
}

type printMode int

const (
	PRINT_MODE_NORMAL printMode = iota
	PRINT_MODE_SHOW_SET_RESULT
	PRINT_MODE_END
)

func (u *ui) render(playerID int, state escoba.GameState, mode printMode) error {
	err := termbox.Clear(termbox.ColorWhite, termbox.ColorBlack)
	if err != nil {
		return err
	}

	var (
		mx, my = termbox.Size()
		you    = playerID
		them   = state.OpponentOf(you)
	)

	// Display opponent's hand (face down)
	opponentHand := state.Hands[them]
	if opponentHand != nil {
		unrevealed := strings.Repeat("[] ", len(opponentHand.Cards))
		printAt(0, 0, unrevealed)
	}

	// Display round and game info
	printUpToAt(mx-1, 0, fmt.Sprintf("Ronda %d", state.RoundNumber))

	youMano := ""
	themMano := ""
	if state.RoundTurnPlayerID == you {
		youMano = " (mano)"
	} else {
		themMano = " (mano)"
	}

	printUpToAt(mx-1, 1, fmt.Sprintf("Vos%v: %v puntos", youMano, state.Scores[you]))
	printUpToAt(mx-1, 2, fmt.Sprintf("Oponente%v: %v puntos", themMano, state.Scores[them]))

	// Display table cards
	tableCardsStr := "Mesa: " + getCardsString(state.TableCards, false, false)
	printAt(0, my/2-2, tableCardsStr)

	// Display captured piles info
	youPileCount := len(state.Piles[you])
	themPileCount := len(state.Piles[them])
	youEscobas := state.Escobas[you]
	themEscobas := state.Escobas[them]

	pileInfo := fmt.Sprintf("Cartas capturadas - Vos: %d (escobas: %d), Oponente: %d (escobas: %d)",
		youPileCount, youEscobas, themPileCount, themEscobas)
	printAt(0, my/2-1, pileInfo)

	// Display your hand
	yourHand := state.Hands[you]
	if yourHand != nil {
		yourCards := getCardsString(yourHand.Cards, true, false)
		printAt(0, my-4, "Tu mano: "+yourCards)
	}

	switch mode {
	case PRINT_MODE_NORMAL:
		lastActionString := getLastActionString(you, state)
		printAt(0, my/2, lastActionString)
	case PRINT_MODE_SHOW_SET_RESULT:
		if state.LastSetResults != nil {
			result := state.LastSetResults
			resultStr := fmt.Sprintf("Terminó el mazo. Puntos obtenidos - Vos: %d, Oponente: %d",
				result.PointsAwarded[you], result.PointsAwarded[them])
			printAt(0, my/2, resultStr)
		}
	case PRINT_MODE_END:
		if playerID == state.WinnerPlayerID {
			printAt(0, my/2, "¡Ganaste la partida! 🥰")
		} else {
			printAt(0, my/2, "Perdiste la partida 😭")
		}
	}

	if mode == PRINT_MODE_SHOW_SET_RESULT || mode == PRINT_MODE_END {
		printAt(0, my-2, "Presioná cualquier tecla para continuar...")
		termbox.Flush()
		u.pressAnyKey()
		return nil
	}

	if state.TurnPlayerID == playerID {
		if mode == PRINT_MODE_NORMAL {
			actionsString := ""
			for i, action := range _deserializeActions(state.PossibleActions) {
				actionStr := spanishAction(action)
				actionsString += fmt.Sprintf("%d. %s   ", i+1, actionStr)
			}
			printAt(0, my-2, actionsString)
		}
	} else {
		printAt(0, my-2, "Esperando al otro jugador...")
	}

	termbox.Flush()
	return nil
}

func printAt(x, y int, s string) {
	_s := []rune(s)
	for i, r := range _s {
		termbox.SetCell(x+i, y, r, termbox.ColorDefault, termbox.ColorDefault)
	}
}

// Write so that the output ends at x, y
func printUpToAt(x, y int, s string) {
	_s := []rune(s)
	for i, r := range _s {
		termbox.SetCell(x-len(_s)+i, y, r, termbox.ColorDefault, termbox.ColorDefault)
	}
}

func getCardsString(cards []escoba.Card, withNumbers bool, withBack bool) string {
	var cs []string
	for i, card := range cards {
		if withNumbers {
			cs = append(cs, fmt.Sprintf("%v. %v", i+1, getCardString(card)))
		} else {
			cs = append(cs, getCardString(card))
		}
	}
	if withBack {
		cs = append(cs, "0. Volver")
	}
	return strings.Join(cs, "  ")
}

func getCardString(card escoba.Card) string {
	return fmt.Sprintf("[%v%v]", card.Number, suitEmoji(card.Suit))
}

func suitEmoji(suit string) string {
	switch suit {
	case escoba.ESPADA:
		return "🔪"
	case escoba.BASTO:
		return "🌿"
	case escoba.ORO:
		return "💰"
	case escoba.COPA:
		return "🍷"
	default:
		return "❓"
	}
}

func getLastActionString(playerID int, state escoba.GameState) string {
	if len(state.Actions) == 0 {
		return "¡Empezó el juego!"
	}
	if state.RoundJustStarted {
		return "¡Empezó la ronda!"
	}

	lastActionBs := state.Actions[len(state.Actions)-1]
	lastActionOwnerPlayerID := state.ActionOwnerPlayerIDs[len(state.ActionOwnerPlayerIDs)-1]
	return getActionString(lastActionBs, lastActionOwnerPlayerID, playerID)
}

func getActionString(lastActionBs json.RawMessage, lastActionOwnerPlayerID int, playerID int) string {
	lastAction, err := escoba.DeserializeAction(lastActionBs)
	if err != nil {
		return "Error deserializando acción"
	}

	threw := "tiraste"
	who := "Vos"
	if playerID != lastActionOwnerPlayerID {
		who = "Oponente"
		threw = "tiró"
	}

	var what string
	switch lastAction.GetName() {
	case escoba.THROW_CARD:
		action := lastAction.(*escoba.ActionThrowCard)
		cardStr := getCardString(action.Card)
		if len(action.CapturedTableCards) > 0 {
			capturedStr := getCardsString(action.CapturedTableCards, false, false)
			what = fmt.Sprintf("%v %v y capturó %v", threw, cardStr, capturedStr)
		} else {
			what = fmt.Sprintf("%v %v a la mesa", threw, cardStr)
		}
	default:
		what = "acción desconocida"
	}

	return fmt.Sprintf("%v %v", who, what)
}

func (u *ui) startKeyEventLoop() {
	keyPressesCh := make(chan termbox.Event)
	go func() {
		for {
			event := termbox.PollEvent()
			if event.Type != termbox.EventKey {
				continue
			}
			if event.Key == termbox.KeyEsc || event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyCtrlD || event.Key == termbox.KeyCtrlZ || event.Ch == 'q' {
				termbox.Close()
				log.Println("¡Chau!")
				os.Exit(0)
			}
			keyPressesCh <- event
		}
	}()

	go func() {
		for {
			select {
			case <-keyPressesCh:
			case <-u.wantKeyPressCh:
				event := <-keyPressesCh
				u.sendKeyPressCh <- event.Ch
			}
		}
	}()
}

func (u *ui) pressAnyKey() {
	u.wantKeyPressCh <- struct{}{}
	<-u.sendKeyPressCh
}

func (u *ui) pressAnyNumber() int {
	u.wantKeyPressCh <- struct{}{}
	r := <-u.sendKeyPressCh
	num, err := strconv.Atoi(string(r))
	if err != nil {
		return u.pressAnyNumber()
	}
	return num
}

func spanishAction(action escoba.Action) string {
	switch action.GetName() {
	case escoba.THROW_CARD:
		_action := action.(*escoba.ActionThrowCard)
		cardStr := getCardString(_action.Card)
		if len(_action.CapturedTableCards) > 0 {
			capturedStr := getCardsString(_action.CapturedTableCards, false, false)
			return fmt.Sprintf("%v (captura %v)", cardStr, capturedStr)
		} else {
			return fmt.Sprintf("%v (a mesa)", cardStr)
		}
	default:
		return "???"
	}
}

func _deserializeActions(as []json.RawMessage) []escoba.Action {
	_as := []escoba.Action{}
	for _, a := range as {
		_a, _ := escoba.DeserializeAction(a)
		_as = append(_as, _a)
	}
	return _as
}
