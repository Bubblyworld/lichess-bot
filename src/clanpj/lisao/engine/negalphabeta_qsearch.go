package engine

import (
	"fmt"
	"os"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

func (s *SearchT) probeQTT(qDepthToGo int, alpha EvalCp, beta EvalCp) (dragon.Move, EvalCp, bool, bool) {
	if UseQSearchTT {
		qttEntry, isQttHit := probeQtt(qtt, s.board.Hash())

		if isQttHit {
			s.stats.QttHits++

			//////// First try to find an exact eval at the same depth (or deeper if HeurUseQTTDeeperHits is configured)
			
			// Try the lower bound of the same depth parity (it might be exact)
			qttpLbEntry := &qttEntry.lbEntry
			// We can use this value if its same depth or deeper (and HeurUseQTTDeeperHits is configured),
			//    OR if the QTT entry is a quiesced value
			qttpLbEntryUseable :=
				(qttpLbEntry.isQuiesced && (qttpLbEntry.qDepthToGo == uint8(qDepthToGo) || HeurUseQTTDeeperHits)) ||
				qttpLbEntry.evalType != TTInvalid && (qttpLbEntry.qDepthToGo == uint8(qDepthToGo) || (HeurUseQTTDeeperHits && qttpLbEntry.qDepthToGo > uint8(qDepthToGo)))
			if qttpLbEntryUseable && qttpLbEntry.evalType == TTEvalExact {
				s.stats.QttTrueEvals++
				return qttpLbEntry.bestMove, qttpLbEntry.eval, qttpLbEntry.isQuiesced, true
			}
				
			// Try the upper bound of the same depth parity (it might be exact)
			qttpUbEntry := &qttEntry.ubEntry
			// We can use this value if its same depth or deeper (and HeurUseQTTDeeperHits is configured),
			//    OR if the QTT entry is a quiesced value
			qttpUbEntryUseable :=
				(qttpUbEntry.isQuiesced && (qttpUbEntry.qDepthToGo == uint8(qDepthToGo) || HeurUseQTTDeeperHits)) ||
				qttpUbEntry.evalType != TTInvalid && (qttpUbEntry.qDepthToGo == uint8(qDepthToGo) || (HeurUseQTTDeeperHits && qttpUbEntry.qDepthToGo > uint8(qDepthToGo)))
			if qttpUbEntryUseable && qttpUbEntry.evalType == TTEvalExact {
				s.stats.QttTrueEvals++
				return qttpUbEntry.bestMove, qttpUbEntry.eval, qttpUbEntry.isQuiesced, true
			}

			//////// See if we have a beta cut
			if qttpLbEntryUseable {
				if beta <= qttpLbEntry.eval {
					s.stats.QttBetaCuts++
					return qttpLbEntry.bestMove, qttpLbEntry.eval, qttpLbEntry.isQuiesced, true
				}
			}
			
			//////// See if we have an alpha cut
			if qttpUbEntryUseable {
				if qttpUbEntry.eval <= alpha {
					s.stats.QttAlphaCuts++
					return qttpUbEntry.bestMove, qttpUbEntry.eval, qttpUbEntry.isQuiesced, true
				}
			}

			//////// Set the qttMove
			qttMove, qttMoveDepthToGo := NoMove, uint8(0)
			if qttpLbEntry.evalType != TTInvalid {
				qttMove = qttpLbEntry.bestMove
				qttMoveDepthToGo = qttpLbEntry.depthToGo
			} else if qttpUbEntry.evalType != TTInvalid && qttMoveDepthToGo < qttpUbEntry.depthToGo {
				qttMove = qttpUbEntry.bestMove
				qttMoveDepthToGo = qttpUbEntry.depthToGo
			}

			return qttMove, YourCheckMateEval, false, false
			
		}
	}

	return NoMove, YourCheckMateEval, false, false
}


// Quiescence search - differs from full search as follows:
//   - we only look at captures, promotions and check evasion - we could/should also possibly look at checks, but check detection is currently expensive
//   - we consider 'standing pat' - i.e. do alpha/beta cutoff according to the node's static eval (TODO)
// Return best-move, best-eval, isQuiesced
// TODO - better static eval if we bottom out without quiescing, e.g. static exchange evaluation (SEE)
// TODO - include moving away from attacks too?
func (s *SearchT) QSearchNegAlphaBeta(qDepthToGo int, depthFromRoot int, depthFromQRoot int, alpha EvalCp, beta EvalCp, killer dragon.Move, eval0 EvalCp) (dragon.Move, EvalCp, bool) {

	// Sanity check the eval0
	if false {
		eval0Check := NegaStaticEvalOrder0(s.board)
		if eval0 != eval0Check {
			fmt.Println("               eval0", eval0, "eval0Check", eval0Check, "fen", s.board.ToFen())
			os.Exit(1)
		}
	}
	
	s.stats.QNodes++
	s.stats.QNonLeafs++
	if depthFromQRoot < MaxQDepthStats {
		s.stats.QNonLeafsAt[depthFromQRoot]++
	}

	// Remember this to check whether our final eval is a lower or upper bound
	origBeta := beta
	origAlpha := alpha

	staticNegaEval := NegaStaticEvalFast(s.board, eval0)

 	// Sanity check the fast eval
	if false {
		staticNegaEvalCheck := NegaStaticEval(s.board)
		if staticNegaEval != staticNegaEvalCheck {
			fmt.Println("               eval", staticNegaEval, "evalCheck", staticNegaEvalCheck, "fen", s.board.ToFen())
			os.Exit(1)
		}
	}
	
	// Stand pat - equivalent to considering the null move as a valid move.
	// Essentially the player to move doesn't _have_ to make a 'noisy' move - assuming that there is a quiet move available.
	if alpha < staticNegaEval {
		alpha = staticNegaEval
	}

	// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
	if alpha >= beta {
		s.stats.QPats++
		s.stats.QPatCuts++

		return NoMove, staticNegaEval, true
	}

	// Anything after here interacts with the QTT - so single return location at the end of the func after writing back to QTT

	// Probe the Quiescence Transposition Table
	qttMove, qttEval, qttIsQuiesced, qttIsCut := s.probeQTT(qDepthToGo, alpha, beta)
	if qttIsCut {
		return qttMove, qttEval, qttIsQuiesced
	}
	
	// Maximise eval with beta cut-off
	bestMove := NoMove
	bestEval := staticNegaEval // stand pat value

	// Did we reach quiescence at all leaves?
	isQuiesced := false

	// Generate all noisy legal moves
	legalMoves, isInCheck := s.board.GenerateLegalMoves2(/*onlyCapturesPromosCheckEvasion*/true)

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

		var boardSave dragon.BoardSaveT

		for i := 0; i < nMovesToUse; i++ {
			move := legalMoves[i]

			// Get the (deep) eval
			var eval EvalCp

			// Make the move
			s.board.MakeMove(move, &boardSave)
			childEval0 := NegaStaticEvalOrder0Fast(s.board, -eval0, &boardSave)

			if qDepthToGo <= 1 {
				s.stats.QNodes++
				s.stats.QPrunes++

				// We hit max depth before quiescing
				isQuiesced = false
				// Ignore mate check to avoid generating moves at all leaf nodes
				eval := NegaStaticEvalFast(s.board, childEval0)

				// Sanity check the fast eval
				if false {
					staticNegaEvalCheck := NegaStaticEval(s.board)
					if eval != staticNegaEvalCheck {
						fmt.Println("               eval", eval, "evalCheck", staticNegaEvalCheck, "i", i, "move", &move, "fen", s.board.ToFen())
						os.Exit(1)
					}
				}
				
			} else {
				var isChildQuiesced bool
				childKiller, eval, isChildQuiesced = s.QSearchNegAlphaBeta(qDepthToGo-1, depthFromRoot+1, depthFromQRoot+1, -beta, -alpha, childKiller, childEval0)
				isQuiesced = isQuiesced && isChildQuiesced
			}
			eval = -eval // back to our perspective

			// Take back the move
			s.board.Restore(&boardSave)

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
