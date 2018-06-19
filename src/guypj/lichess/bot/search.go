package bot

import (
	"errors"
	//"fmt"
	"math"

	dragon "github.com/dylhunn/dragontoothmg"
)

const NoMove dragon.Move = 0

func Search(board *dragon.Board) (dragon.Move, error) {
	bestMove, _ /*eval*/ := search(board, 4)

	if bestMove == NoMove {
		return NoMove, errors.New("bot: no legal move found in search")
	}

	//fmt.Printf("Best move %6s eval %3.2f\n", bestMove.String(), eval)

	return bestMove, nil
}

// search returns the best score attainable through minmax from the given
// position, along with the move leading to the principal variation.
func search(board *dragon.Board, depth int) (dragon.Move, float64) {
	legalMoves := board.GenerateLegalMoves()

	if depth <= 0 || len(legalMoves) == 0 {
		return NoMove, Evaluate(board, legalMoves)
	}

	//fmt.Printf("Eval: %3.2f\n", Evaluate(board, legalMoves))

	var bestMove dragon.Move
	var bestScore float64 = math.MaxFloat64
	if board.Wtomove {
		bestScore = -math.MaxFloat64
	}

	for _, move := range legalMoves {
		unapply := board.Apply(move)
		_, score := search(board, depth-1)
		unapply()

		//fmt.Printf("         move %6s eval %3.2f\n", move.String(), score)

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

	return bestMove, bestScore
}
