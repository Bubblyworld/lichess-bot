package engine

import dragon "github.com/Bubblyworld/dragontoothmg"

// Return the best eval attainable through minmax from the given
//   position, along with the move leading to the principal variation.
// Eval is given from white's perspective.
func (s *SearchT) MiniMax(depthToGo int, depthFromRoot int) (dragon.Move, EvalCp) {

	isWhiteMove := s.board.Wtomove

	// Bail if we've timed out
	if isTimedOut(s.timeout) {
		worstEval := WhiteCheckMateEval
		if isWhiteMove {
			worstEval = BlackCheckMateEval
		}
		// Return the worst possible eval (opponent checkmate) to invalidate this incomplete search branch
		return NoMove, worstEval
	}

	s.stats.Nodes++
	s.stats.NonLeafs++

	// Generate all legal moves - thanks dragontoothmg!
	legalMoves := s.board.GenerateLegalMoves()

	// Check for checkmate or stalemate
	if len(legalMoves) == 0 {
		s.stats.Mates++
		return NoMove, mateEval(s.board, depthFromRoot)
	}

	var bestMove = NoMove
	var bestEval EvalCp = WhiteCheckMateEval
	if isWhiteMove {
		bestEval = BlackCheckMateEval
	}

	for _, move := range legalMoves {
		// Make the move
		unapply := s.board.Apply(move)

		// Get the (deep) eval
		var eval EvalCp
		if depthToGo <= 1 {
			s.stats.Nodes++
			// Ignore mate check to avoid generating moves at all leaf nodes
			eval = StaticEval(s.board)
		} else {
			_, eval = s.MiniMax( /*board,*/ depthToGo-1, depthFromRoot+1 /*, stats, timeout*/)
		}

		// Take back the move
		unapply()

		// If we're white, we try to maximise our eval. If we're black, we try to
		// minimise our eval.
		if isWhiteMove {
			// Strictly > to match alphabeta
			if eval > bestEval {
				bestEval, bestMove = eval, move
			}
		} else {
			// Strictly < to match alphabeta
			if eval < bestEval {
				bestEval, bestMove = eval, move
			}
		}
	}

	return bestMove, bestEval
}
