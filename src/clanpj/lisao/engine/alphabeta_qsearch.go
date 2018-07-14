package engine

import dragon "github.com/Bubblyworld/dragontoothmg"

// Quiescence search - differs from full search as follows:
//   - we only look at captures, promotions and check evasion - we could/should also possibly look at checks, but check detection is currently expensive
//   - we consider 'standing pat' - i.e. do alpha/beta cutoff according to the node's static eval (TODO)
// TODO - better static eval if we bottom out without quescing, e.g. static exchange evaluation (SEE)
func (s *SearchT) QSearchAlphaBeta(qDepthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, killer dragon.Move) (dragon.Move, EvalCp) {

	s.stats.QNodes++
	s.stats.QNonLeafs++

	staticEval := StaticEval(s.board)

	// Stand pat - the player to move doesn't _have_ to make a capture (assuming that there is a non-capture move available.)
	if s.board.Wtomove {
		if alpha < staticEval {
			alpha = staticEval
		}

		// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
		if alpha >= beta {
			s.stats.QPats++
			s.stats.QPatCuts++
			return NoMove, staticEval
		}

	} else {
		if staticEval < beta {
			beta = staticEval
		}

		// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
		if alpha >= beta {
			s.stats.QPats++
			s.stats.QPatCuts++
			return NoMove, staticEval
		}
	}

	// Generate all noisy legal moves
	legalMoves, isInCheck := s.board.GenerateLegalMoves2( /*onlyCapturesPromosCheckEvasion*/ true)

	// No noisy mvoes
	if len(legalMoves) == 0 {
		// Check for checkmate or stalemate
		if isInCheck {
			s.stats.QMates++
			return NoMove, mateEval(s.board, depthFromRoot) // TODO checks for mate again expensively
		} else {
			// Already quiesced - just return static eval
			s.stats.QQuiesced++
			return NoMove, staticEval
		}
	}

	usingDeepKiller := false
	if UseQKillerMoves {
		// Place killer-move (or deep killer move) first if it's there
		killer, usingDeepKiller = prioritiseKillerMove(legalMoves, killer, UseQDeepKillerMoves, s.deepKillers[depthFromRoot], &s.stats.QKillers, &s.stats.QDeepKillers)
	}

	// Would be smaller with negalpha-beta but this is simple
	if s.board.Wtomove {
		// White to move - maximise eval with beta cut-off
		var bestMove = NoMove
		var bestEval EvalCp = staticEval // stand pat value
		childKiller := NoMove

		for _, move := range legalMoves {
			// Make the move
			unapply := s.board.Apply(move)

			// Get the (deep) eval
			var eval EvalCp
			if qDepthToGo <= 1 {
				s.stats.QNodes++
				s.stats.QPrunes++
				// Ignore mate check to avoid generating moves at all leaf nodes
				eval = StaticEval(s.board)
			} else {
				childKiller, eval = s.QSearchAlphaBeta(qDepthToGo-1, depthFromRoot+1, alpha, beta, childKiller)
			}

			// Take back the move
			unapply()

			// We're white - maximise our eval.
			// Note - this MUST be strictly > because we fail-soft AT the current best evel - beware!
			if eval > bestEval {
				bestEval, bestMove = eval, move
			}

			if alpha < bestEval {
				alpha = bestEval
			}

			// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
			if alpha >= beta {
				if UseQKillerMoves && bestMove == killer {
					if usingDeepKiller {
						s.stats.QDeepKillerCuts++
					} else {
						s.stats.QKillerCuts++
					}
				}
				// beta cut-off
				s.deepKillers[depthFromRoot] = bestMove
				return bestMove, bestEval
			}
		}

		s.stats.QPats++
		s.deepKillers[depthFromRoot] = bestMove
		return bestMove, bestEval
	} else {
		// Black to move - minimise eval with alpha cut-off
		var bestMove = NoMove
		var bestEval EvalCp = staticEval // stand pat value
		childKiller := NoMove

		for _, move := range legalMoves {
			// Make the move
			unapply := s.board.Apply(move)

			// Get the (deep) eval
			var eval EvalCp
			if qDepthToGo <= 1 {
				s.stats.QNodes++
				s.stats.QPrunes++
				// Ignore mate check to avoid generating moves at all leaf nodes
				eval = StaticEval(s.board)
			} else {
				childKiller, eval = s.QSearchAlphaBeta(qDepthToGo-1, depthFromRoot+1, alpha, beta, childKiller)
			}

			// Take back the move
			unapply()

			// We're black - minimise our eval.
			// Note - this MUST be strictly < because we fail-soft AT the current best evel - beware!
			if eval < bestEval {
				bestEval, bestMove = eval, move
			}

			if bestEval < beta {
				beta = bestEval
			}

			// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
			if alpha >= beta {
				if UseQKillerMoves && bestMove == killer {
					if usingDeepKiller {
						s.stats.QDeepKillerCuts++
					} else {
						s.stats.QKillerCuts++
					}
				}
				// alpha cut-off
				s.deepKillers[depthFromRoot] = bestMove
				return bestMove, bestEval
			}
		}

		s.stats.QPats++
		s.deepKillers[depthFromRoot] = bestMove
		return bestMove, bestEval
	}
}
