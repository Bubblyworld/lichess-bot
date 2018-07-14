package engine

import dragon "github.com/Bubblyworld/dragontoothmg"

// Quiescence search - differs from full search as follows:
//   - we only look at captures, promotions and check evasion - we could/should also possibly look at checks, but check detection is currently expensive
//   - we consider 'standing pat' - i.e. do alpha/beta cutoff according to the node's static eval (TODO)
// Return best-move, best-eval, isQuiesced
// TODO - better static eval if we bottom out without quiescing, e.g. static exchange evaluation (SEE)
// TODO - include moving away from attacks too?
func (s *SearchT) QSearchNegAlphaBeta(qDepthToGo int, depthFromRoot int, depthFromQRoot int, alpha EvalCp, beta EvalCp, killer dragon.Move) (dragon.Move, EvalCp, bool) {

	s.stats.QNodes++
	s.stats.QNonLeafs++
	if depthFromQRoot < MaxQDepthStats {
		s.stats.QNonLeafsAt[depthFromQRoot]++
	}

	// Remember this to check whether our final eval is a lower or upper bound
	origBeta := beta
	origAlpha := alpha

	staticNegaEval := NegaStaticEval(s.board)

	// Stand pat - equivalent to considering the null move as a valid move.
	// Essentially the player to move doesn't _have_ to make a 'noisy' move - assuming that there is a quiet move available.
	if alpha < staticNegaEval {
		alpha = staticNegaEval
	}

	// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
	if alpha >= beta {
		s.stats.QPats++
		s.stats.QPatCuts++

		return NoMove, staticNegaEval, false // TODO - not sure what to return here for isQuiesced - this is playing safe
	}

	// Anything after here interacts with the QTT - so single return location at the end of the func after writing back to QTT

	// Probe the Quiescence Transposition Table
	var qttMove = NoMove
	if UseQSearchTT {
		qttEntry, isQttHit := probeQtt(qtt, s.board.Hash())

		if isQttHit {
			s.stats.QttHits++

			qttMove = qttEntry.bestMove

			// If the QTT hit is for exactly the same depth then use the eval; otherwise we just use the bestMove as a move hint
			// Note that most engines will use the TT eval if the TT entry is a deeper search; however this requires a 'stable' static eval
			//   and changes behaviour between TT-enabled/disabled. For rigourous testing it's better to be consistent.
			isExactHit := qDepthToGo == int(qttEntry.qDepthToGo) ||
				//    ... or if this is a fully quiesced result
				qDepthToGo > int(qttEntry.qDepthToGo) && qttEntry.isQuiesced

			if isExactHit {
				s.stats.QttDepthHits++
				qttEval := qttEntry.eval
				// If the eval is exact then we're done
				if qttEntry.evalType == TTEvalExact {
					s.stats.QttTrueEvals++
					return qttMove, qttEntry.eval, qttEntry.isQuiesced
				} else {
					var cutoffStats *uint64
					// We can have an alpha or beta cut-off depending on the eval type
					if qttEntry.evalType == TTEvalLowerBound {
						cutoffStats = &s.stats.QttBetaCuts
						if alpha < qttEval {
							alpha = qttEval
						}
					} else {
						// TTEvalUpperBound
						cutoffStats = &s.stats.QttAlphaCuts
						if qttEval < beta {
							beta = qttEval
						}
					}
					// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
					if alpha >= beta {
						*cutoffStats++
						return qttMove, qttEval, qttEntry.isQuiesced
					}
				}
			}
		}
	}

	// Maximise eval with beta cut-off
	bestMove := NoMove
	bestEval := staticNegaEval // stand pat value

	// Did we reach quiescence at all leaves?
	isQuiesced := false

	// Generate all noisy legal moves
	legalMoves, isInCheck := s.board.GenerateLegalMoves2( /*onlyCapturesPromosCheckEvasion*/ true)

	if len(legalMoves) == 0 {
		// No noisy moves - checkmate or stalemate or just quiesced
		isQuiesced = true
		if isInCheck {
			s.stats.QMates++
			bestMove, bestEval = NoMove, negaMateEval(s.board, depthFromRoot) // TODO checks for mate again expensively
		} else {
			// Already quiesced - just return static eval
			bestMove, bestEval = NoMove, staticNegaEval
		}
	} else {
		// Usually same as len(legalMoves) unless we prune the move list, for example queen rampage pruning
		nMovesToUse := len(legalMoves)

		killerMove := NoMove
		if UseQKillerMoves {
			killerMove = killer
		}
		deepKiller := NoMove
		if UseQDeepKillerMoves {
			deepKiller = s.deepKillers[depthFromRoot]
		}

		// Sort the moves heuristically
		if UseQSearchMoveOrdering {
			if len(legalMoves) > 1 {
				orderMoves(s.board, legalMoves, qttMove, killerMove, deepKiller, &s.stats.QKillers, &s.stats.QDeepKillers)
				if UseQSearchRampagePruning {
					nMovesToUse = pruneQueenRampages(s.board, legalMoves, depthFromQRoot, s.stats)
				}
			}
		} else if UseQKillerMoves {
			// Place killer-move (or deep killer move) first if it's there
			prioritiseKillerMove(legalMoves, killer, UseQDeepKillerMoves, s.deepKillers[depthFromRoot], &s.stats.QKillers, &s.stats.QDeepKillers)
		}

		// We're quiesced as long as all children (we visit) are quiesced.
		isQuiesced := true

		childKiller := NoMove

		for i := 0; i < nMovesToUse; i++ {
			move := legalMoves[i]

			// Get the (deep) eval
			var eval EvalCp

			// Make the move
			unapply := s.board.Apply(move)

			if qDepthToGo <= 1 {
				s.stats.QNodes++
				s.stats.QPrunes++

				// We hit max depth before quiescing
				isQuiesced = false
				// Ignore mate check to avoid generating moves at all leaf nodes
				eval = NegaStaticEval(s.board)
			} else {
				var isChildQuiesced bool
				childKiller, eval, isChildQuiesced = s.QSearchNegAlphaBeta(qDepthToGo-1, depthFromRoot+1, depthFromQRoot+1, -beta, -alpha, childKiller)
				isQuiesced = isQuiesced && isChildQuiesced
			}
			eval = -eval // back to our perspective

			// Take back the move
			unapply()

			// Note - this MUST be strictly > because we fail-soft AT the current best evel - beware!
			if eval > bestEval {
				bestEval, bestMove = eval, move
			}

			if alpha < bestEval {
				alpha = bestEval
			}

			// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
			if alpha >= beta {
				// beta cut-off
				if bestMove == qttMove {
					s.stats.QttLateCuts++
				} else if bestMove == killerMove {
					s.stats.QKillerCuts++
				} else if bestMove == deepKiller {
					s.stats.QDeepKillerCuts++
				}
				if i == 0 {
					s.stats.QFirstChildCuts++
				}
				break
			}
		}

		// Was stand-pat the best after all?
		if bestEval == staticNegaEval {
			s.stats.QPats++
		}

		// If we didn't get a beta cut-off then we visited all children.
		if bestEval <= origBeta {
			s.stats.QAllChildrenNodes++
		}
		s.deepKillers[depthFromRoot] = bestMove

	}

	if isQuiesced {
		s.stats.QQuiesced++
	}

	// Update the QTT
	if UseQSearchTT {
		evalType := TTEvalExact
		if origBeta <= bestEval {
			evalType = TTEvalLowerBound
		} else if bestEval <= origAlpha {
			evalType = TTEvalUpperBound
		}
		// Write back the QTT entry - this is an update if the TT already contains an entry for this hash
		writeQttEntry(qtt, s.board.Hash(), bestEval, bestMove, qDepthToGo, evalType, isQuiesced)
	}

	return bestMove, bestEval, isQuiesced
}

// Do rampage move pruning.
// Note: assumes queen captures appear first in the moves list which is true for MVV-LVA.
// Returns the number of moves to look at.
func pruneQueenRampages(board *dragon.Board, moves []dragon.Move, depthFromQRoot int, stats *SearchStatsT) int {
	nMovesToUse := len(moves)
	if depthFromQRoot >= QSearchRampagePruningDepth {
		victim0 := board.PieceAt(moves[0].To())
		// If the top-rated move is not a queen capture, likely a promo, then delay rampage pruning
		if victim0 == dragon.Queen {
			stats.QRampagePrunes++
			var move dragon.Move
			for nMovesToUse, move = range moves {
				victim := board.PieceAt(move.To())
				if victim != dragon.Queen {
					break
				}
			}
		}
	}

	return nMovesToUse
}
