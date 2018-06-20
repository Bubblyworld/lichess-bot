package engine

import (
	"errors"

	dragon "github.com/dylhunn/dragontoothmg"
)

const NoMove dragon.Move = 0

func Search(board *dragon.Board) (dragon.Move, error) {
	bestMove, _ /*eval*/ := search(board, /*depthToGo*/4, /*depthFromRoot*/0)

	if bestMove == NoMove {
		return NoMove, errors.New("bot: no legal move found in search")
	}

	//fmt.Printf("Best move %6s eval %3.2f\n", bestMove.String(), eval)

	return bestMove, nil
}

// search returns the best eval attainable through minmax from the given
// position, along with the move leading to the principal variation.
func search(board *dragon.Board, depthToGo int, depthFromRoot int) (dragon.Move, EvalCp) {

	// Ignore mate to avoid generating moves at all leaf nodes
	if depthToGo <= 0 {
		return NoMove, Evaluate(board)
	}
	
	legalMoves := board.GenerateLegalMoves()

	// If there are no legal moves, there are two possibilities - either our king
	// is in check, or it isn't. In the first case it's mate and we've lost, and
	// in the second case it's stalemate and therefore a draw.
	if len(legalMoves) == 0 {
		if board.OurKingInCheck() {
			// checkmate - closer to root is better
			if board.Wtomove {
				return NoMove, BlackCheckMateEval + EvalCp(depthFromRoot)
			}
			
			return NoMove, WhiteCheckMateEval - EvalCp(depthFromRoot)
		}
		// stalemate
		return NoMove, DrawEval
	}
	
	//fmt.Printf("Eval: %3.2f\n", Evaluate(board, legalMoves))

	var bestMove dragon.Move
	var bestEval EvalCp = WhiteCheckMateEval
	if board.Wtomove {
		bestEval = BlackCheckMateEval
	}

	for _, move := range legalMoves {
		moveInfo := board.Apply2(move)
		_, eval := search(board, depthToGo-1, depthFromRoot+1)
		moveInfo.Unapply()

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
