package bot

import (
	"errors"
	//"fmt"

	dragon "github.com/dylhunn/dragontoothmg"
)

const MaxPlies = 6

func Search(board dragon.Board) (*dragon.Move, error) {
	bestMove, _ /*eval*/ := search(board, MaxPlies-1)

	if bestMove == nil {
		return nil, errors.New("bot: no legal move found in search")
	}

	//fmt.Printf("Best move %6s eval %3.2f\n", bestMove.String(), eval)

	return bestMove, nil
}

// search returns the best eval attainable through minmax from the given
// position, along with the move leading to the principal variation.
func search(board dragon.Board, depth int) (*dragon.Move, int) {
	// Mate and stalemate ignored at leaf
	if depth <= 0 {
		return nil, Evaluate(board)
	}
	
	legalMoves := board.GenerateLegalMoves()

	// Check for mate/stalemate
	if len(legalMoves) == 0 {
		if board.OurKingInCheck() {
			// checkmate
			if board.Wtomove {
				return nil, -MateIn0 // black wins
			}
			return nil, MateIn0 // white wins
		} else {
			//stalemate
			return nil, 0
		}
	}

	//fmt.Printf("Eval: %3.2f\n", Evaluate(board, legalMoves))

	var bestMove dragon.Move
	bestEval := BestEval(board.Wtomove)

	for _, move := range legalMoves {
		unapply := board.Apply(move)
		_, eval := search(board, depth-1)
		unapply()

		//fmt.Printf("         move %6s eval %3.2f\n", move.String(), eval)

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

	return &bestMove, bestEval
}
