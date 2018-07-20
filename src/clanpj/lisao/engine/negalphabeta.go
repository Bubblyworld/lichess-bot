package engine

import (
	"math/bits"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Return the best eval attainable through alpha-beta from the given position (with killer-move hint), along with the move leading to the principal variation.
func (s *SearchT) NegAlphaBeta(depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, killer dragon.Move, parentNullMove bool) (dragon.Move, EvalCp) {

	// Bail if we've timed out
	if isTimedOut(s.timeout) {
		// Return the worst possible eval (opponent checkmate) to invalidate this incomplete search branch
		return NoMove, YourCheckMateEval
	}

	s.stats.Nodes++
	s.stats.NonLeafs++
	if depthFromRoot < MaxDepthStats {
		s.stats.NonLeafsAt[depthFromRoot]++
	}

	// Remember this to check whether our final eval is a lower or upper bound - for TT
	origBeta := beta
	origAlpha := alpha

	// Probe the Transposition Table
	var ttMove = NoMove
	if UseTT {
		ttEntry, isTTHit := probeTT(tt, s.board.Hash())

		if isTTHit {
			s.stats.TTHits++

			// Pick the right parity if it's available, else anything
			ttpEntry := &ttEntry.parityHits[depthToGoParity(depthToGo)]
			if ttpEntry.evalType == TTInvalid {
				ttpEntry = &ttEntry.parityHits[depthToGoParity(depthToGo)^1]
			}
			ttMove = ttpEntry.bestMove

			// If the TT hit is for exactly the same depth then use the eval; otherwise we just use the bestMove as a move hint.
			// We use a deeper TT hit only for the same parity since our eval in start-game is unstable between even/odd plies.
			// N.B. using deeper TT hit (eval)s changes the search tree, so disable HeurUseTTDeeperHits for correctness testing.
			canUseTTEval := false
			if depthToGo == int(ttpEntry.depthToGo) {
				s.stats.TTDepthHits++
				canUseTTEval = true
			} else if HeurUseTTDeeperHits && depthToGo < int(ttpEntry.depthToGo) && (depthToGo&1) == (int(ttpEntry.depthToGo)&1) {
				s.stats.TTDeeperHits++
				canUseTTEval = true
			}
			if canUseTTEval {
				ttEval := ttpEntry.eval
				// If the eval is exact then we're done
				if ttpEntry.evalType == TTEvalExact {
					s.stats.TTTrueEvals++
					return ttMove, ttpEntry.eval
				} else {
					var cutoffStats *uint64
					// We can have an alpha or beta cut-off depending on the eval type
					if ttpEntry.evalType == TTEvalLowerBound {
						cutoffStats = &s.stats.TTBetaCuts
						if alpha < ttEval {
							alpha = ttEval
						}
					} else {
						// TTEvalUpperBound
						cutoffStats = &s.stats.TTAlphaCuts
						if ttEval < beta {
							beta = ttEval
						}
					}
					// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
					if alpha >= beta {
						*cutoffStats++
						return ttMove, ttEval
					}
				}
			}
		}
	}

	// Maximise eval with beta cut-off
	bestMove := NoMove
	bestEval := YourCheckMateEval
	childKiller := NoMove

	// Anything after here interacts with the QTT - so single return location at the end of the func after writing back to QTT
	// We use a fake run-once loop so that we can break after each search step rather than indenting code arbitrarily deeper with each new feature/optimisation.
done:
	for once := true; once; once = false {

		isInCheck := isInCheckFast(s.board)

		// Early Null Move Heuristic
		if HeurUseNullMove {
			const nullMoveDepthSkip = 3 // must be odd to cope with our even/odd ply eval instability
			// Try null-move - but never 2 null moves in a row, and never in check otherwise king gets captured
			if !isInCheck && !parentNullMove && beta != MyCheckMateEval && depthToGo > nullMoveDepthSkip {
				// Use piece count to determine end-game for zugzwang avoidance - TODO improve this
				nNonPawns := bits.OnesCount64((s.board.White.All & ^s.board.White.Pawns) | (s.board.Black.All & ^s.board.Black.Pawns))
				// Proceed with null-move heuristic if there are at least 4 non-pawn pieces (note the count includes the two kings)
				if nNonPawns >= 6 {
					unapply := s.board.ApplyNullMove()
					_, nullMoveEval := s.NegAlphaBeta(depthToGo-nullMoveDepthSkip, depthFromRoot+1, -beta, -alpha, NoMove /*killer???*/ /*parentNullMove*/, true)
					nullMoveEval = -nullMoveEval // back to our perspective
					unapply()

					// Bail cleanly without polluting search results if we have timed out
					if depthToGo > 1 && isTimedOut(s.timeout) {
						break done
					}

					// Maximise our eval.
					// Note - this MUST be strictly > because we fail-soft AT the current best evel - beware!
					if nullMoveEval > bestEval {
						bestMove, bestEval = NoMove, nullMoveEval
					}

					if alpha < bestEval {
						alpha = bestEval
					}

					// Did null-move heuristic already provide a cut?
					if alpha >= beta {
						s.stats.NullMoveCuts++

						break done
					}

				}
			}
		}

		// TODO also use killer moves, but need to check them first for validity
		hintMove := ttMove

		// Try hint move before doing move-gen if we have a known valid move hint
		if UseEarlyMoveHint && !(UseIDMoveHint && depthToGo >= MinIDMoveHintDepth) {
			if hintMove != NoMove {
				s.stats.ValidHintMoves++
				// Make the move
				unapply := s.board.Apply(hintMove)
				// Add to the move history
				repetitions := s.ht.Add(s.board.Hash())

				// Get the (deep) eval
				var eval EvalCp
				// We consider 2-fold repetition to be a draw, since if a repeat can be forced then it can be forced again.
				// This reduces the search tree a bit and is common practice in chess engines.
				if UsePosRepetition && repetitions > 1 {
					s.stats.PosRepetitions++
					eval = DrawEval
				} else if depthToGo <= 1 {
					s.stats.Nodes++
					if UseQSearch {
						// Quiesce
						childKiller, eval, _ = s.QSearchNegAlphaBeta(QSearchDepth, depthFromRoot+1 /*depthFromQRoot*/, 0, -beta, -alpha, childKiller)
					} else {
						eval = NegaStaticEval(s.board)
					}
				} else {
					childKiller, eval = s.NegAlphaBeta(depthToGo-1, depthFromRoot+1, -beta, -alpha, childKiller, false)
				}
				eval = -eval // back to our perspective

				// Remove from the move history
				s.ht.Remove(s.board.Hash())
				// Take back the move
				unapply()

				// Bail cleanly without polluting search results if we have timed out
				if depthToGo > 1 && isTimedOut(s.timeout) {
					break done
				}

				// Maximise our eval.
				// Note - this MUST be strictly > because we fail-soft AT the current best evel - beware!
				if eval > bestEval {
					bestEval, bestMove = eval, hintMove
				}

				if alpha < bestEval {
					alpha = bestEval
				}

				// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
				if alpha >= beta {
					// beta cut-off
					s.stats.HintMoveCuts++
					s.stats.FirstChildCuts++
					if depthFromRoot < MaxDepthStats {
						s.stats.FirstChildCutsAt[depthFromRoot]++
					}

					break done
				}
			}
		}

		// Generate all legal moves
		legalMoves, isInCheck := s.board.GenerateLegalMoves2(false /*all moves*/)

		// Check for checkmate or stalemate
		if len(legalMoves) == 0 {
			s.stats.Mates++
			bestMove, bestEval = NoMove, negaMateEval(s.board, depthFromRoot) // TODO use isInCheck

			break done
		}

		// Try null-move heuristic
		if false/*done already*/ && HeurUseNullMove {
			const nullMoveDepthSkip = 3 // must be odd to cope with our even/odd ply eval instability
			// Try null-move - but never 2 null moves in a row, and never in check otherwise king gets captured
			if !isInCheck && !parentNullMove && beta != MyCheckMateEval && depthToGo > nullMoveDepthSkip {
				// Use piece count to determine end-game for zugzwang avoidance - TODO improve this
				nNonPawns := bits.OnesCount64((s.board.White.All & ^s.board.White.Pawns) | (s.board.Black.All & ^s.board.Black.Pawns))
				// Proceed with null-move heuristic if there are at least 4 non-pawn pieces (note the count includes the two kings)
				if nNonPawns >= 6 {
					unapply := s.board.ApplyNullMove()
					_, nullMoveEval := s.NegAlphaBeta(depthToGo-nullMoveDepthSkip, depthFromRoot+1, -beta, -alpha, NoMove /*killer???*/ /*parentNullMove*/, true)
					nullMoveEval = -nullMoveEval // back to our perspective
					unapply()

					// Bail cleanly without polluting search results if we have timed out
					if depthToGo > 1 && isTimedOut(s.timeout) {
						break done
					}

					// Maximise our eval.
					// Note - this MUST be strictly > because we fail-soft AT the current best evel - beware!
					if nullMoveEval > bestEval {
						bestMove, bestEval = NoMove, nullMoveEval
					}

					if alpha < bestEval {
						alpha = bestEval
					}

					// Did null-move heuristic already provide a cut?
					if alpha >= beta {
						s.stats.NullMoveCuts++

						break done
					}

				}
			}
		}

		killerMove := NoMove
		if UseKillerMoves {
			killerMove = killer
		}
		deepKiller := NoMove
		if UseDeepKillerMoves {
			deepKiller = s.deepKillers[depthFromRoot]
		}

		// Sort the moves heuristically
		if UseMoveOrdering {
			if len(legalMoves) > 1 {
				if UseIDMoveHint && depthToGo >= MinIDMoveHintDepth {
					idKiller := killerMove
					if idKiller == NoMove {
						idKiller = ttMove
					}
					// Get the best move for a search of depth-2.
					// We go 2 plies shallower since our eval is unstable between odd/even plies.
					// The result is effectively the (possibly new) ttMove.
					// TODO - weaken the beta bound (and alpha?) a bit?
					ttMove, _ = s.NegAlphaBeta(depthToGo-2, depthFromRoot, alpha, beta, idKiller, false)

				}
				orderMoves(s.board, legalMoves, ttMove, killerMove, deepKiller, &s.stats.Killers, &s.stats.DeepKillers)
			}
		} else if UseKillerMoves {
			// Place killer-move (or deep killer move) first if it's there
			prioritiseKillerMove(legalMoves, killer, UseDeepKillerMoves, s.deepKillers[depthFromRoot], &s.stats.Killers, &s.stats.DeepKillers)
		}

		for i, move := range legalMoves {
			// Don't repeat the hintMove
			if UseEarlyMoveHint && !(UseIDMoveHint && depthToGo >= MinIDMoveHintDepth) && move == hintMove {
				continue
			}

			// Make the move
			unapply := s.board.Apply(move)
			// Add to the move history
			repetitions := s.ht.Add(s.board.Hash())

			// Get the (deep) eval
			var eval EvalCp
			// We consider 2-fold repetition to be a draw, since if a repeat can be forced then it can be forced again.
			// This reduces the search tree a bit and is common practice in chess engines.
			if UsePosRepetition && repetitions > 1 {
				s.stats.PosRepetitions++
				eval = DrawEval
			} else if depthToGo <= 1 {
				s.stats.Nodes++
				if UseQSearch {
					// Quiesce
					childKiller, eval, _ = s.QSearchNegAlphaBeta(QSearchDepth, depthFromRoot+1 /*depthFromQRoot*/, 0, -beta, -alpha, childKiller)
				} else {
					eval = NegaStaticEval(s.board)
				}
			} else {
				childKiller, eval = s.NegAlphaBeta(depthToGo-1, depthFromRoot+1, -beta, -alpha, childKiller, false)
			}
			eval = -eval // back to our perspective

			// Remove from the move history
			s.ht.Remove(s.board.Hash())
			// Take back the move
			unapply()

			// Bail cleanly without polluting search results if we have timed out
			if depthToGo > 1 && isTimedOut(s.timeout) {
				break
			}

			// Maximise our eval.
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
				if bestMove == ttMove {
					s.stats.TTLateCuts++
				} else if bestMove == killerMove {
					s.stats.KillerCuts++
				} else if bestMove == deepKiller {
					s.stats.DeepKillerCuts++
				}
				if i == 0 {
					s.stats.FirstChildCuts++
					if depthFromRoot < MaxDepthStats {
						s.stats.FirstChildCutsAt[depthFromRoot]++
					}
				}
				break
			}
		}

		// If we didn't get a beta cut-off then we visited all children.
		if bestEval <= origBeta {
			s.stats.AllChildrenNodes++
		}
		s.deepKillers[depthFromRoot] = bestMove
	} // end of fake run-once loop

	if UseTT {
		// Update the TT - but only if the search was not truncated due to a time-out
		if !isTimedOut(s.timeout) {
			evalType := TTEvalExact
			if origBeta <= bestEval {
				evalType = TTEvalLowerBound
			} else if bestEval <= origAlpha {
				evalType = TTEvalUpperBound
			}
			// Write back the TT entry - this is an update if the TT already contains an entry for this hash
			writeTTEntry(tt, s.board.Hash(), bestEval, bestMove, depthToGo, evalType)
		}
	}

	// Regardless of a time-out this will still be valid - if bestMove == NoMove then we didn't complete any child branch
	return bestMove, bestEval
}

// Return the eval for stalemate or checkmate from curent mover's perspective
// Only valid if there are no legal moves.
func negaMateEval(board *dragon.Board, depthFromRoot int) EvalCp {
	if board.OurKingInCheck() {
		// checkmate - closer to root is better
		return YourCheckMateEval + EvalCp(depthFromRoot)
	}
	// stalemate
	return DrawEval
}

// Move the killer or deep-killer move to the front of the legal moves list, if it's in the legal moves list.
// Return true iff we're using the deep-killer
// TODO - install both killer and deepKiller if they're both valid and distinct
func prioritiseKillerMove(legalMoves []dragon.Move, killer dragon.Move, useDeepKillerMoves bool, deepKiller dragon.Move, killersStat *uint64, deepKillersStat *uint64) (dragon.Move, bool) {
	usingDeepKiller := false

	if killer == NoMove && useDeepKillerMoves {
		usingDeepKiller = true
		killer = deepKiller
	}
	// Place killer-move first if it's there
	if killer != NoMove {
		for i := 0; i < len(legalMoves); i++ {
			if legalMoves[i] == killer {
				legalMoves[0], legalMoves[i] = killer, legalMoves[0]
				break
			}
		}
	}
	if legalMoves[0] == killer {
		if usingDeepKiller {
			*deepKillersStat++
		} else {
			*killersStat++
		}
	}
	return killer, usingDeepKiller
}
