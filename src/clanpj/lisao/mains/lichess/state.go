package main

import (
	"sync"

	"clanpj/lisao/lichess"
)

type State struct {
	client  *lichess.LichessClient
	stateMu sync.Mutex

	challenges  []Challenge
	activeGames []*Game
}

func NewState(client *lichess.LichessClient) *State {
	return &State{
		client: client,
	}
}
