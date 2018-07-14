package engine

import dragon "github.com/Bubblyworld/dragontoothmg"

// Return the best eval attainable through alpha-beta from the given position (with killer-move hint), along with the move leading to the principal variation.
func (s *SearchT) AlphaBeta(depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, killer dragon.Move) (dragon.Move, EvalCp) {

	// Bail if we've timed out
	if isTimedOut(s.timeout) {
		worstEval := WhiteCheckMateEval
		if s.board.Wtomove {
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

	usingDeepKiller := false
	if UseKillerMoves {
		// Place killer-move (or deep killer move) first if it's there
		killer, usingDeepKiller = prioritiseKillerMove(legalMoves, killer, UseDeepKillerMoves, s.deepKillers[depthFromRoot], &s.stats.Killers, &s.stats.DeepKillers)
	}

	// Would be smaller with negalpha-beta but this is simple
	if s.board.Wtomove {
		// White to move - maximise eval with beta cut-off
		var bestMove = NoMove
		var bestEval EvalCp = BlackCheckMateEval
		childKiller := NoMove

		for _, move := range legalMoves {
			// Make the move
			unapply := s.board.Apply(move)

			// Get the (deep) eval
			var eval EvalCp
			if depthToGo <= 1 {
				s.stats.Nodes++
				if UseQSearch {
					// Quiesce
					childKiller, eval = s.QSearchAlphaBeta(QSearchDepth, depthFromRoot+1, alpha, beta, childKiller)
				} else {
					eval = StaticEval(s.board)
				}
			} else {
				childKiller, eval = s.AlphaBeta(depthToGo-1, depthFromRoot+1, alpha, beta, childKiller)
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
				if UseKillerMoves && bestMove == killer {
					if usingDeepKiller {
						s.stats.DeepKillerCuts++
					} else {
						s.stats.KillerCuts++
					}
				}
				// beta cut-off
				s.deepKillers[depthFromRoot] = bestMove
				return bestMove, bestEval
			}
		}

		s.deepKillers[depthFromRoot] = bestMove
		return bestMove, bestEval
	} else {
		// Black to move - minimise eval with alpha cut-off
		var bestMove = NoMove
		var bestEval EvalCp = WhiteCheckMateEval
		childKiller := NoMove

		for _, move := range legalMoves {
			// Make the move
			unapply := s.board.Apply(move)

			// Get the (deep) eval
			var eval EvalCp
			if depthToGo <= 1 {
				s.stats.Nodes++
				if UseQSearch {
					// Quiesce
					childKiller, eval = s.QSearchAlphaBeta(QSearchDepth, depthFromRoot+1, alpha, beta, childKiller)
				} else {
					eval = StaticEval(s.board)
				}
			} else {
				childKiller, eval = s.AlphaBeta(depthToGo-1, depthFromRoot+1, alpha, beta, childKiller)
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
				if UseKillerMoves && bestMove == killer {
					if usingDeepKiller {
						s.stats.DeepKillerCuts++
					} else {
						s.stats.KillerCuts++
					}
				}
				// alpha cut-off
				s.deepKillers[depthFromRoot] = bestMove
				return bestMove, bestEval
			}
		}

		s.deepKillers[depthFromRoot] = bestMove
		return bestMove, bestEval
	}
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
