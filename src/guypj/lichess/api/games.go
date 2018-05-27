package api

import (
	"encoding/json"
	"log"
)

type GameType int

const (
	GameFullGameType  GameType = 1
	GameStateGameType GameType = 2
	ChatLineGameType  GameType = 3
)

type GameFullGameState struct {
}

type GameStateGameState struct {
}

type ChatLineGameState struct {
}

type GameStateMessage struct {
	Type GameType
	Data interface{}
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

	actionChannel := make(chan GameStateMessage)
	go func() {
		defer res.Body.Close()
		defer close(actionChannel)
		decoder := json.NewDecoder(res.Body)

		for decoder.More() {
			var buffer interface{}
			err = decoder.Decode(&buffer)
			if err != nil {
				// TODO(guy) return some kind of error type
				log.Printf("StreamGameState: %v", err)
				return
			}

			log.Printf("Got a message of type: %v", buffer)
		}
	}()

	return actionChannel, nil
}
