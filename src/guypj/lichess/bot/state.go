package bot

import (
	"guypj/lichess/api"
	"sync"
)

type State struct {
	client  *api.LichessClient
	stateMu sync.Mutex

	challenges  []Challenge
	activeGames []*Game
}

func NewState(client *api.LichessClient) *State {
	return &State{
		client: client,
	}
}
