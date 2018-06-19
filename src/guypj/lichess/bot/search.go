package bot

import (
	"errors"

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

// search returns the best eval attainable through minmax from the given
// position, along with the move leading to the principal variation.
func search(board *dragon.Board, depth int) (dragon.Move, EvalCp) {
	legalMoves := board.GenerateLegalMoves()

	if depth <= 0 || len(legalMoves) == 0 {
		return NoMove, Evaluate(board, legalMoves)
	}

	//fmt.Printf("Eval: %3.2f\n", Evaluate(board, legalMoves))

	var bestMove dragon.Move
	var bestEval EvalCp = WhiteCheckMateEval
	if board.Wtomove {
		bestEval = BlackCheckMateEval
	}

	for _, move := range legalMoves {
		unapply := board.Apply(move)
		_, eval := search(board, depth-1)
		unapply()

		//fmt.Printf("         move %6s eval %3d\n", move.String(), eval)

		// If we're white, we try to maximise our eval. If we're black, we try to
		// minimise our eval.
		if board.Wtomove {
			if eval >= bestEval {
				bestEval, bestMove = eval, move
			}
		} else {
			if eval <= bestEval {
				bestEval, bestMove = eval, move
			}
		}
	}

	return bestMove, bestEval
}
