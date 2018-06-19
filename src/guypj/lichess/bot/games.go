package bot

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	dragon "github.com/dylhunn/dragontoothmg"

	"guypj/lichess/api"
)

// TODO(guy) should be in state
var botName = "Lisao"

type Game struct {
	ID         string
	Moves      []string // List of moves in UCI format.
	InitialFen string
	WeAreWhite bool

	isPlaying bool
	mutex     sync.Mutex
}

func (state *State) PushGame(game *Game) {
	state.stateMu.Lock()
	defer state.stateMu.Unlock()

	state.activeGames = append(state.activeGames, game)
}

func (state *State) RemoveGame(gameID string) {
	state.stateMu.Lock()
	defer state.stateMu.Unlock()

	var games []*Game
	for _, game := range state.activeGames {
		if game.ID != gameID {
			games = append(games, game)
		}
	}

	state.activeGames = games
}

func lockGame(game *Game) bool {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	acquiredLock := false
	if !game.isPlaying {
		acquiredLock = true
		game.isPlaying = true
	}

	return acquiredLock
}

func unlockGame(game *Game) {
	game.mutex.Lock()
	defer game.mutex.Unlock()

	game.isPlaying = false
}

func PlayGamesForever(state *State, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	for {
		for _, game := range state.activeGames {
			if !game.isPlaying {
				go playGame(state, game)
			}
		}

		time.Sleep(time.Second)
	}
}

func playGame(state *State, game *Game) {
	ok := lockGame(game)
	if !ok {
		return
	}
	defer unlockGame(game)

	gameStateCh, err := state.client.StreamGameState(game.ID)
	if err != nil {
		log.Printf("bot: Error getting update stream for game %s: %v", game.ID, err)
		return
	}

	// Listen to game updates as long as we can.
	for {
		msg, ok := <-gameStateCh
		if !ok {
			break
		}

		err := handleMessage(state, game, msg)
		if err != nil {
			log.Printf("bot: Error handling update message for game %s: %v",
				game.ID, err)
			return
		}

		isOver, err := gameIsOver(game)
		if err != nil {
			log.Printf("bot: Error determining if game %s is over.", game.ID)
			return
		}
		if isOver {
			log.Printf("bot: Game %s has finished.", game.ID)
			state.RemoveGame(game.ID)
			return
		}
	}
}

func handleMessage(state *State, game *Game, msg api.GameStateMessage) error {
	var anyErr error
	switch msg.Type {
	case api.GameFullGameStateType:
		anyErr = handleInitialGameState(game, msg.Data.(api.GameFullGameState))

	case api.GameStateGameStateType:
		anyErr = handleGameUpdate(game, msg.Data.(api.GameStateGameState))

	case api.ChatLineGameStateType:
		anyErr = handleChatEvent(game, msg.Data.(api.ChatLineGameState))

	default:
		errMsg := fmt.Sprintf("bot: Received unknown game update for game %s: %v",
			game.ID, msg)

		anyErr = errors.New(errMsg)
	}

	if anyErr != nil {
		return anyErr
	}

	// If the game is not finished and it's our turn, we should move.
	isOver, err := gameIsOver(game)
	if err != nil {
		return err
	}

	if isOurTurn(game) && !isOver {
		err := makeMove(state, game)
		if err != nil {
			return err
		}
	}

	return nil
}

func handleInitialGameState(game *Game, initialState api.GameFullGameState) error {
	game.InitialFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	game.Moves = []string{}
	if initialState.State.Moves != "" {
		game.Moves = strings.Split(initialState.State.Moves, " ")
	}

	if initialState.White.Name == botName {
		game.WeAreWhite = true
	} else if initialState.Black.Name == botName {
		game.WeAreWhite = false
	} else {
		errMsg := fmt.Sprintf(
			"bot: Error, expected one of the players in game %s to be %s.",
			game.ID, botName)

		return errors.New(errMsg)
	}

	return nil
}

func handleGameUpdate(game *Game, update api.GameStateGameState) error {
	game.Moves = []string{}
	if update.Moves != "" {
		game.Moves = strings.Split(update.Moves, " ")
	}

	return nil
}

func handleChatEvent(game *Game, chatEvent api.ChatLineGameState) error {
	log.Printf("bot: Received chat: %s", chatEvent.Text)

	return nil
}

func isOurTurn(game *Game) bool {
	whiteToPlay := len(game.Moves)%2 == 0

	return whiteToPlay == game.WeAreWhite
}

func getBoard(game *Game) (*dragon.Board, error) {
	board := dragon.ParseFen(game.InitialFen)
	for _, moveStr := range game.Moves {
		move, err := dragon.ParseMove(moveStr)
		if err != nil {
			return nil, err
		}

		board.Apply(move)
	}

	return &board, nil
}

func gameIsOver(game *Game) (bool, error) {
	board, err := getBoard(game)
	if err != nil {
		return false, err
	}

	return len(board.GenerateLegalMoves()) == 0, nil
}

func makeMove(state *State, game *Game) error {
	board, err := getBoard(game)
	if err != nil {
		return err
	}

	move, err := Search(board)
	if err != nil {
		return err
	}

	return state.client.PostMove(game.ID, move.String())
}
