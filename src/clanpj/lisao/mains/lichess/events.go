package main

import (
	"log"
	"sync"

	"clanpj/lisao/lichess"
)

func ListenForEventsForever(state *State, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	eventsChannel, err := state.client.StreamEvents()
	if err != nil {
		log.Fatalf("bot: Error getting events stream: %v", err)
		return
	}

	for {
		msg, ok := <-eventsChannel
		if !ok {
			break
		}

		switch msg.Type {
		case lichess.ChallengeEventType:
			challenge := msg.Data.(lichess.ChallengeEvent)
			state.PushChallenge(Challenge{
				ID:         challenge.Challenge.ID,
				Challenger: challenge.Challenge.Challenger,
				Variant:    challenge.Challenge.Variant,
			})

		case lichess.GameStartEventType:
			gameStart := msg.Data.(lichess.GameStartEvent)
			state.PushGame(&Game{
				ID: gameStart.Game.ID,
			})

		default:
			log.Printf("bot: Received unknown event %v.", msg)
			return
		}
	}
}
