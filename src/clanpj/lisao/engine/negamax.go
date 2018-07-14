package engine

import dragon "github.com/Bubblyworld/dragontoothmg"

// Return the best eval attainable through negamax from the given
//   position, along with the move leading to the principal variation.
// Eval is given from current mover's perspective.
func (s *SearchT) NegaMax(depthToGo int, depthFromRoot int) (dragon.Move, EvalCp) {

	// Bail if we've timed out
	if isTimedOut(s.timeout) {
		// Return the worst possible eval (opponent checkmate) to invalidate this incomplete search branch
		return NoMove, YourCheckMateEval
	}

	s.stats.Nodes++
	s.stats.NonLeafs++

	// Generate all legal moves - thanks dragontoothmg!
	legalMoves := s.board.GenerateLegalMoves()

	// Check for checkmate or stalemate
	if len(legalMoves) == 0 {
		s.stats.Mates++
		return NoMove, negaMateEval(s.board, depthFromRoot)
	}

	bestMove := NoMove
	bestEval := YourCheckMateEval

	for _, move := range legalMoves {
		// Make the move
		unapply := s.board.Apply(move)

		// Get the (deep) eval
		var eval EvalCp
		if depthToGo <= 1 {
			s.stats.Nodes++
			// Ignore mate check to avoid generating moves at all leaf nodes
			eval = NegaStaticEval(s.board)
		} else {
			_, eval = s.NegaMax(depthToGo-1, depthFromRoot+1)
		}
		eval = -eval // back to our perspective

		// Take back the move
		unapply()

		// We try to maximise our eval.
		// Strictly > to match alphabeta
		if eval > bestEval {
			bestEval, bestMove = eval, move
		}
	}

	return bestMove, bestEval
}
