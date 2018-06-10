package bot

import (
	"errors"
	"math"

	dragon "github.com/dylhunn/dragontoothmg"
)

func Search(board dragon.Board) (*dragon.Move, error) {
	bestMove, _ := search(board, 4)

	if bestMove == nil {
		return nil, errors.New("bot: no legal move found in search")
	}

	return bestMove, nil
}

// search returns the best score attainable through minmax from the given
// position, along with the move leading to the principal variation.
func search(board dragon.Board, depth int) (*dragon.Move, float64) {
	legalMoves := board.GenerateLegalMoves()

	if depth <= 0 || len(legalMoves) == 0 {
		return nil, Evaluate(board, legalMoves)
	}

	var bestMove dragon.Move
	var bestScore float64 = math.MaxFloat64
	if board.Wtomove {
		bestScore = -math.MaxFloat64
	}

	for _, move := range legalMoves {
		unapply := board.Apply(move)
		_, score := search(board, depth-1)
		unapply()

		// If we're white, we try to maximise our score. If we're black, we try to
		// minimise our score.
		if board.Wtomove {
			if score >= bestScore {
				bestScore, bestMove = score, move
			}
		} else {
			if score <= bestScore {
				bestScore, bestMove = score, move
			}
		}
	}

	return &bestMove, bestScore
}
