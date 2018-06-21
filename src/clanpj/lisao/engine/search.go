package engine

import (
	"errors"
	// "fmt"
	// "time"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

const NoMove dragon.Move = 0

func Search(board *dragon.Board) (dragon.Move, error) {
	//bestMove, eval := minimax(board, /*depthToGo*/4, /*depthFromRoot*/0, StaticEval(board))
	bestMoveAB, _/*evalAB*/ := alphabeta(board, /*depthToGo*/5, /*depthFromRoot*/0, BlackCheckMateEval, WhiteCheckMateEval, StaticEval(board))

	// if eval != evalAB {
	// 	fmt.Printf("Boooo mm eval %d move %s ab eval %d move %s\n", eval, &bestMove, evalAB, &bestMoveAB)
	// 	time.Sleep(time.Duration(1) * time.Minute)
	// }
	// if bestMove != bestMoveAB {
	// 	fmt.Printf("Boooo tooo mm eval %d move %s ab eval %d move %s\n", eval, &bestMove, evalAB, &bestMoveAB)
	// 	time.Sleep(time.Duration(1) * time.Minute)
	// }
	
	if bestMoveAB == NoMove {
		return NoMove, errors.New("bot: no legal move found in search")
	}

	return bestMoveAB, nil
}

// Return the eval for stalemate or checkmate from white perspective.
// Only valid if there are no legal moves.
func mateEval(board *dragon.Board, depthFromRoot int) EvalCp {
	if board.OurKingInCheck() {
		// checkmate - closer to root is better
		if board.Wtomove {
			return BlackCheckMateEval + EvalCp(depthFromRoot)
		}
		
		return WhiteCheckMateEval - EvalCp(depthFromRoot)
	}
	// stalemate
	return DrawEval
}

// Return the best eval attainable through minmax from the given
//   position, along with the move leading to the principal variation.
func minimax(board *dragon.Board, depthToGo int, depthFromRoot int, staticEval EvalCp) (dragon.Move, EvalCp) {

	// Generate all legal moves - thanks dragontoothmg!
	legalMoves := board.GenerateLegalMoves()

	// Check for checkmate or stalemate
	if len(legalMoves) == 0 {
		return NoMove, mateEval(board, depthFromRoot)
	}
	
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

		// Get the (deep) eval
		var eval EvalCp
		if depthToGo <= 1 {
			// Ignore mate check to avoid generating moves at all leaf nodes
			eval = newStaticEval
		} else {
			_, eval = minimax(board, depthToGo-1, depthFromRoot+1, newStaticEval)
		}

		// if depthFromRoot == 0 {
		// 	fmt.Printf("          mm move %s eval %d\n", &move, eval)
		// }
			
		// Take back the move
		moveInfo.Unapply()

		// If we're white, we try to maximise our eval. If we're black, we try to
		// minimise our eval.
		if isWhiteMove {
			// Strictly > to match alphabeta
			if eval > bestEval {
				bestEval, bestMove = eval, move
				// if depthFromRoot == 0 {
				// 	fmt.Printf("            mm white best move %s eval %d\n", &bestMove, bestEval)
				// }
			}
		} else {
			// Strictly < to match alphabeta
			if eval < bestEval {
				bestEval, bestMove = eval, move
				// if depthFromRoot == 0 {
				// 	fmt.Printf("            mm black best move %s eval %d\n", &bestMove, bestEval)
				// }
			}
		}
	}

	return bestMove, bestEval
}

// Return the best eval attainable through alpha-beta from the given
//   position, along with the move leading to the principal variation.
func alphabeta(board *dragon.Board, depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, staticEval EvalCp) (dragon.Move, EvalCp) {

	// Generate all legal moves - thanks dragontoothmg!
	legalMoves := board.GenerateLegalMoves()

	// Check for checkmate or stalemate
	if len(legalMoves) == 0 {
		return NoMove, mateEval(board, depthFromRoot)
	}

	// Would be smaller with negalpha-beta but this is simple
	if board.Wtomove {
		// White to move - maximise eval with beta cut-off
		var bestMove dragon.Move
		var bestEval EvalCp = BlackCheckMateEval
		
		for _, move := range legalMoves {
			// Make the move
			moveInfo := board.Apply2(move)
			
			newStaticEval := staticEval + EvalDelta(move, moveInfo, /*isWhiteMove*/true)
			
			// Get the (deep) eval
			var eval EvalCp
			if depthToGo <= 1 {
				// Ignore mate check to avoid generating moves at all leaf nodes
				eval = newStaticEval
			} else {
				_, eval = alphabeta(board, depthToGo-1, depthFromRoot+1, alpha, beta, newStaticEval)
			}

			// if depthFromRoot == 0 {
			// 	fmt.Printf("          ab move %s eval %d\n", &move, eval)
			// }
			
			// Take back the move
			moveInfo.Unapply()
			
			// We're white - maximise our eval.
			// Note - this MUST be strictly > because we fail-soft AT the current best evel - beware!
			if eval > bestEval {
				bestEval, bestMove = eval, move
				// if depthFromRoot == 0 {
				// 	fmt.Printf("            white best move %s eval %d\n", &bestMove, bestEval)
				// }
			}

			if alpha < bestEval {
				alpha = bestEval
			}

			// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
			if beta <= alpha {
				// beta cut-off
				// if depthFromRoot == 0 {
				// 	fmt.Printf("              beta cut best move %s eval %d\n", &bestMove, bestEval)
				// }
				return bestMove, bestEval
			}
		}
		
		// if depthFromRoot == 0 {
		// 	fmt.Printf("              no beta cut best move %s eval %d\n", &bestMove, bestEval)
		// }
		return bestMove, bestEval
	} else {
		// Black to move - minimise eval with alpha cut-off
		var bestMove dragon.Move
		var bestEval EvalCp = WhiteCheckMateEval
		
		for _, move := range legalMoves {
			// Make the move
			moveInfo := board.Apply2(move)
			
			newStaticEval := staticEval + EvalDelta(move, moveInfo, /*isWhiteMove*/false)
			
			// Get the (deep) eval
			var eval EvalCp
			if depthToGo <= 1 {
				// Ignore mate check to avoid generating moves at all leaf nodes
				eval = newStaticEval
			} else {
				_, eval = alphabeta(board, depthToGo-1, depthFromRoot+1, alpha, beta, newStaticEval)
			}
			
			// if depthFromRoot == 0 {
			// 	fmt.Printf("          move %s eval %d\n", &move, eval)
			// }
			
			// Take back the move
			moveInfo.Unapply()
			
			// We're black - minimise our eval.
			// Note - this MUST be strictly < because we fail-soft AT the current best evel - beware!
			if eval < bestEval {
				bestEval, bestMove = eval, move
				// if depthFromRoot == 0 {
				// 	fmt.Printf("            black best move %s eval %d\n", &bestMove, bestEval)
				// }
			}

			if bestEval < beta {
				beta = eval
			}

			// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
			if beta <= alpha {
				// alpha cut-off
				// if depthFromRoot == 0 {
				// 	fmt.Printf("              alpha cut best move %s eval %d\n", &bestMove, bestEval)
				// }
				return bestMove, bestEval
			}
		}

		// if depthFromRoot == 0 {
		// 	fmt.Printf("              no alpha cut best move %s eval %d\n", &bestMove, bestEval)
		// }
		return bestMove, bestEval
	}
}
