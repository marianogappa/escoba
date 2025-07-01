package server

import (
	"encoding/json"

	"github.com/marianogappa/escoba/escoba"
)

const (
	MessageTypeHello = iota
	MessageTypeHeresGameState
	MessageTypeAction
	MessageTypeGimmeGameState
)

type IWebsocketMessage[T any] interface {
	GetType() int
	Deserialize() (T, error)
}

type WebsocketMessage struct {
	Type int `json:"type"`
}

func (m WebsocketMessage) GetType() int {
	return m.Type
}

type MessageHello struct {
	WebsocketMessage
	PlayerID int `json:"playerID"`
}

func NewMessageHello(playerID int) MessageHello {
	return MessageHello{WebsocketMessage: WebsocketMessage{Type: MessageTypeHello}, PlayerID: playerID}
}

func (m MessageHello) Deserialize() (int, error) {
	return m.PlayerID, nil
}

type MessageHeresGameState struct {
	WebsocketMessage
	GameState json.RawMessage `json:"playerID"`
}

func NewMessageHeresGameState(gameState escoba.GameState) (MessageHeresGameState, error) {
	bs, err := json.Marshal(gameState)
	return MessageHeresGameState{WebsocketMessage: WebsocketMessage{Type: MessageTypeHeresGameState}, GameState: bs}, err
}

func (gs MessageHeresGameState) Deserialize() (escoba.GameState, error) {
	var gameState escoba.GameState
	err := json.Unmarshal(gs.GameState, &gameState)
	return gameState, err
}

type MessageGimmeGameState struct {
	WebsocketMessage
}

func NewMessageGimmeGameState() MessageGimmeGameState {
	return MessageGimmeGameState{WebsocketMessage: WebsocketMessage{Type: MessageTypeGimmeGameState}}
}

type MessageAction struct {
	WebsocketMessage
	Action json.RawMessage `json:"action"`
}

func NewMessageAction(action escoba.Action) (MessageAction, error) {
	bs, err := json.Marshal(action)
	return MessageAction{WebsocketMessage: WebsocketMessage{Type: MessageTypeAction}, Action: bs}, err
}

func (a MessageAction) Deserialize() (escoba.Action, error) {
	return escoba.DeserializeAction(a.Action)
}
