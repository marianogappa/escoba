//go:build tinygo
// +build tinygo

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/marianogappa/escoba/escoba"
)

func main() {
	js.Global().Set("escobaNew", js.FuncOf(escobaNew))
	js.Global().Set("escobaRunAction", js.FuncOf(escobaRunAction))
	js.Global().Set("escobaBotRunAction", js.FuncOf(escobaBotRunAction))
	select {}
}

var (
	state *escoba.GameState
	bot   escoba.Bot
)

func escobaNew(this js.Value, p []js.Value) interface{} {
	// Escoba doesn't have configurable rules like truco, so we create a standard game
	state = escoba.New()
	bot = escoba.NewBot()

	nbs, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}

	buffer := js.Global().Get("Uint8Array").New(len(nbs))
	js.CopyBytesToJS(buffer, nbs)
	return buffer
}

func escobaRunAction(this js.Value, p []js.Value) interface{} {
	jsonBytes := make([]byte, p[0].Length())
	js.CopyBytesToGo(jsonBytes, p[0])

	newBytes := _runAction(jsonBytes)

	buffer := js.Global().Get("Uint8Array").New(len(newBytes))
	js.CopyBytesToJS(buffer, newBytes)
	return buffer
}

func escobaBotRunAction(this js.Value, p []js.Value) interface{} {
	if !state.IsEnded {
		action := bot.ChooseAction(*state)
		// fmt.Println("Action chosen by bot:", action)

		err := state.RunAction(action)
		if err != nil {
			panic(fmt.Errorf("running action: %w", err))
		}
	}

	nbs, err := json.Marshal(state)
	if err != nil {
		panic(fmt.Errorf("marshalling game state: %w", err))
	}

	buffer := js.Global().Get("Uint8Array").New(len(nbs))
	js.CopyBytesToJS(buffer, nbs)
	return buffer
}

func _runAction(bs []byte) []byte {
	action, err := escoba.DeserializeAction(bs)
	if err != nil {
		panic(err)
	}
	err = state.RunAction(action)
	if err != nil {
		panic(err)
	}
	nbs, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	return nbs
}
