package main

import (
	"log"
	"sync"
	"time"

	"clanpj/lisao/lichess"
)

type Challenge struct {
	ID         string
	Challenger lichess.User
	Variant    lichess.Variant

	Retries int
}

func (state *State) PushChallenge(challenge Challenge) {
	state.stateMu.Lock()
	defer state.stateMu.Unlock()

	state.challenges = append(state.challenges, challenge)
}

func (state *State) PopChallenge() *Challenge {
	state.stateMu.Lock()
	defer state.stateMu.Unlock()

	if len(state.challenges) == 0 {
		return nil
	}

	challenge := state.challenges[0]
	state.challenges = state.challenges[1:]
	return &challenge
}

// TODO(guy) accept a finite number of games at once
func AcceptChallengesForever(state *State, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	var challenge *Challenge
	for {
		if challenge = state.PopChallenge(); challenge == nil {
			time.Sleep(time.Second)
			continue
		}

		if challenge.Variant.Key != "standard" {
			continue
		}

		if challenge.Retries >= 3 {
			continue
		}

		_, err := state.client.AcceptChallenge(challenge.ID)
		if err != nil {
			log.Printf("bot: Error accepting challenge %s: %v", challenge.ID, err)

			challenge.Retries += 1
			state.PushChallenge(*challenge)
		}
	}
}
