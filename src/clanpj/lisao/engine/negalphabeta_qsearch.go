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

			qDepthToGoPlus1 := uint8(qDepthToGo+1)
			
			//////// First try to find an exact eval at the same depth (or deeper if HeurUseQTTDeeperHits is configured)
			
			// We can use the lower bound entry if its same depth or deeper (and HeurUseQTTDeeperHits is configured),
			//    OR if the QTT entry is a quiesced value
			qttLbEntryUseable :=
				(qttEntry.lbIsQuiesced && /*why this extra depth condition? looks wrong?*/(qttEntry.lbQDepthToGoPlus1 == qDepthToGoPlus1 || HeurUseQTTDeeperHits)) ||
				qttEntry.lbQDepthToGoPlus1 == qDepthToGoPlus1 ||
				(HeurUseQTTDeeperHits && qttEntry.lbQDepthToGoPlus1 > qDepthToGoPlus1)
				
			// We can use the upper bound entry if its same depth or deeper (and HeurUseQTTDeeperHits is configured),
			//    OR if the QTT entry is a quiesced value
			qttUbEntryUseable :=
				(qttEntry.ubIsQuiesced && /*why this extra depth condition? looks wrong?*/(qttEntry.ubQDepthToGoPlus1 == qDepthToGoPlus1 || HeurUseQTTDeeperHits)) ||
				qttEntry.ubQDepthToGoPlus1 == qDepthToGoPlus1 ||
				(HeurUseQTTDeeperHits && qttEntry.ubQDepthToGoPlus1 > qDepthToGoPlus1)

			// See if we have an exact value
			if qttLbEntryUseable && qttUbEntryUseable && qttEntry.ubEval <= qttEntry.lbEval {
				s.stats.TTTrueEvals++
				return qttEntry.bestMove, qttEntry.ubEval, qttEntry.ubIsQuiesced, true
			}
	
			//////// See if we have a beta cut
			if qttLbEntryUseable {
				if beta <= qttEntry.lbEval {
					s.stats.QttBetaCuts++
					return qttEntry.bestMove, qttEntry.lbEval, qttEntry.lbIsQuiesced, true
				}
			}
			
			//////// See if we have an alpha cut
			if qttUbEntryUseable {
				if qttEntry.ubEval <= alpha {
					s.stats.QttAlphaCuts++
					return qttEntry.bestMove, qttEntry.ubEval, qttEntry.ubIsQuiesced, true
				}
			}

			//////// We can't use the eval, but use the bestMove anyhow
			return qttEntry.bestMove, YourCheckMateEval, false, false
		}
	}

	return NoMove, YourCheckMateEval, false, false
}


// Quiescence search - differs from full search as follows:
//   - we only look at captures, promotions and check evasion - we could/should also possibly look at checks, but check detection is currently expensive
//   - we consider 'standing pat' - i.e. do alpha/beta cutoff according to the node's static eval
// Return best-move, best-eval, isQuiesced
// TODO - better static eval if we bottom out without quiescing, e.g. static exchange evaluation (SEE)
// TODO - include moving away from attacks too?
func (s *SearchT) QSearchNegAlphaBeta(qDepthToGo int, depthFromRoot int, depthFromQRoot int, alpha EvalCp, beta EvalCp, eval0 EvalCp) (dragon.Move, EvalCp, bool) {

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

		// Sort the moves heuristically
		if UseQSearchMoveOrdering && len(legalMoves) > 1 {
			orderMoves(s.board, legalMoves, qttMove, s.qkt.killersForDepth(depthFromRoot)[:], s.stats.QKillers[depthFromRoot][:])
			if UseQSearchRampagePruning {
				nMovesToUse = pruneQueenRampages(s.board, legalMoves, depthFromQRoot, s.stats)
			}
		}

		// We're quiesced as long as all children (we visit) are quiesced.
		isQuiesced := true

		var boardSave dragon.BoardSaveT
		nChildrenVisited := 0

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
				_, eval, isChildQuiesced = s.QSearchNegAlphaBeta(qDepthToGo-1, depthFromRoot+1, depthFromQRoot+1, -beta, -alpha, childEval0)
				isQuiesced = isQuiesced && isChildQuiesced
			}
			eval = -eval // back to our perspective

			// Take back the move
			s.board.Restore(&boardSave)

			nChildrenVisited++

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

		s.qkt.addKillerMove(bestMove, depthFromRoot)

		// Is it a (beta) cut
		if origBeta <= bestEval {
			s.stats.QCutNodes++
			s.stats.QCutNodeChildren += uint64(nChildrenVisited)

			if nChildrenVisited == 1 {
				s.stats.QFirstChildCuts++
			}

			if bestMove != NoMove {
				if bestMove == qttMove {
					s.stats.QttMoveCuts++
				}

				killers := s.qkt.killersForDepth(depthFromRoot)
				for i := 0; i < NKillersPerDepth; i++ {
					if bestMove == killers[i] {
						s.stats.QKillerCuts[depthFromRoot][i]++
					}
				}
			}
		}
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
