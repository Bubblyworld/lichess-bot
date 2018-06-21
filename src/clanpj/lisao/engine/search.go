package engine

import (
	"errors"
	//"fmt"
	//"time"

	dragon "github.com/dylhunn/dragontoothmg"
)

const NoMove dragon.Move = 0

func Search(board *dragon.Board) (dragon.Move, error) {
	bestMove, _ /*eval*/ := search(board, /*depthToGo*/5, /*depthFromRoot*/0, StaticEval(board))

	if bestMove == NoMove {
		return NoMove, errors.New("bot: no legal move found in search")
	}

	//fmt.Printf("Best move %6s eval %3.2f\n", bestMove.String(), eval)

	return bestMove, nil
}

const CheckEval = true

// search returns the best eval attainable through minmax from the given
// position, along with the move leading to the principal variation.
func search(board *dragon.Board, depthToGo int, depthFromRoot int, staticEval EvalCp) (dragon.Move, EvalCp) {

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
	
	//fmt.Printf("Eval: %3.2f\n", StaticEval(board, legalMoves))

	isWhiteMove := board.Wtomove
	var bestMove dragon.Move
	var bestEval EvalCp = WhiteCheckMateEval
	if isWhiteMove {
		bestEval = BlackCheckMateEval
	}

	for _, move := range legalMoves {
		// Make the move
		moveInfo := board.Apply2(move)

		newStaticEval := staticEval + EvalDelta(move, moveInfo, isWhiteMove)
		// newStaticEval2 := StaticEval(board)

		// Get the (deep) eval
		var eval EvalCp
		if depthToGo <= 1 {
			// Ignore mate check to avoid generating moves at all leaf nodes
			eval = newStaticEval
		} else {
			_, eval = search(board, depthToGo-1, depthFromRoot+1, newStaticEval)
		}

		// Take back the move
		moveInfo.Unapply()


		// if newStaticEval != newStaticEval2 {
		// 	fmt.Printf("info %s move %s eval %4d eval2 %4d\n", board.ToFen(), &move, newStaticEval, newStaticEval2)
		// 	time.Sleep(time.Duration(1) * time.Hour)
		// }
		//fmt.Printf("         move %6s eval %3d\n", move.String(), eval)

		// If we're white, we try to maximise our eval. If we're black, we try to
		// minimise our eval.
		if isWhiteMove {
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
