package api

import (
	"encoding/json"
	"log"
)

type GameStateType int

const (
	GameFullGameStateType  GameStateType = 1
	GameStateGameStateType GameStateType = 2
	ChatLineGameStateType  GameStateType = 3
)

type GameFullGameState struct {
	ID    string
	Type  string
	Rated bool

	White   User
	Black   User
	Variant Variant
	Clock   Clock

	InitialFen string
	State      GameStateGameState
}

type GameStateGameState struct {
	Type  string
	Moves string

	WTime int64 // ms
	WInc  int64

	BTime int64 // ms
	BInc  int64
}

type ChatLineGameState struct {
	Type     string
	Username string
	Text     string
	Room     string
}

type GameStateMessage struct {
	Type GameStateType
	Data interface{}
}

func (msg *GameStateMessage) UnmarshalJSON(bytes []byte) error {
	var buffer interface{}
	err := json.Unmarshal(bytes, &buffer)
	if err != nil {
		return err
	}

	gameStateType := buffer.(map[string]interface{})["type"]
	switch gameStateType {
	case "gameFull":
		var gameFull GameFullGameState
		err = json.Unmarshal(bytes, &gameFull)
		if err != nil {
			return err
		}

		msg.Type = GameFullGameStateType
		msg.Data = gameFull

	case "gameState":
		var gameState GameStateGameState
		err = json.Unmarshal(bytes, &gameState)
		if err != nil {
			return err
		}

		msg.Type = GameStateGameStateType
		msg.Data = gameState

	case "chatLine":
		var chatLine ChatLineGameState
		err = json.Unmarshal(bytes, &chatLine)
		if err != nil {
			return err
		}

		msg.Type = ChatLineGameStateType
		msg.Data = chatLine
	}

	return nil
}

func (lc *LichessClient) StreamGameState(id string) (chan GameStateMessage, error) {
	apiUrl := "/api/bot/game/stream/" + id
	req, err := lc.newRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	res, err := lc.doRequest(req)
	if err != nil {
		return nil, err
	}

	gameStateChannel := make(chan GameStateMessage)
	go func() {
		defer res.Body.Close()
		defer close(gameStateChannel)
		decoder := json.NewDecoder(res.Body)

		for decoder.More() {
			var msg GameStateMessage
			err = decoder.Decode(&msg)
			if err != nil {
				// TODO(guy) return some kind of error type
				log.Printf("StreamGameState: %v", err)
				return
			}

			gameStateChannel <- msg
		}
	}()

	return gameStateChannel, nil
}
