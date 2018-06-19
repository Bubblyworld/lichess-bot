package lichess

import (
	"encoding/json"
	"log"
)

type EventType int

const (
	ChallengeEventType EventType = 1
	GameStartEventType EventType = 2
)

type ChallengeEvent struct {
	Type string

	Challenge struct {
		ID     string
		Status string
		Rated  bool

		Challenger User
		DestUser   User

		Variant Variant

		TimeControl struct {
			Type      string
			Limit     int64
			Increment int64
		}
	}
}

type GameStartEvent struct {
	Type string

	Game struct {
		ID string
	}
}

type EventMessage struct {
	Type EventType
	Data interface{}
}

func (msg *EventMessage) UnmarshalJSON(bytes []byte) error {
	var buffer interface{}
	err := json.Unmarshal(bytes, &buffer)
	if err != nil {
		return err
	}

	eventType := buffer.(map[string]interface{})["type"]
	switch eventType {
	case "challenge":
		var challenge ChallengeEvent
		err = json.Unmarshal(bytes, &challenge)
		if err != nil {
			return err
		}

		msg.Type = ChallengeEventType
		msg.Data = challenge

	case "gameStart":
		var gameStart GameStartEvent
		err = json.Unmarshal(bytes, &gameStart)
		if err != nil {
			return err
		}

		msg.Type = GameStartEventType
		msg.Data = gameStart
	}

	return nil
}

type UnknownEventAction struct{}

func (lc *LichessClient) StreamEvents() (chan EventMessage, error) {
	req, err := lc.newRequest("GET", "/api/stream/event", nil)
	if err != nil {
		return nil, err
	}

	res, err := lc.doRequest(req)
	if err != nil {
		return nil, err
	}

	eventChannel := make(chan EventMessage)
	go func() {
		defer res.Body.Close()
		defer close(eventChannel)
		decoder := json.NewDecoder(res.Body)

		for decoder.More() {
			var msg EventMessage
			err = decoder.Decode(&msg)
			if err != nil {
				// TODO(guy) return some kind of error type
				log.Printf("StreamEvents: %v", err)
				return
			}

			eventChannel <- msg
		}
	}()

	return eventChannel, nil
}
