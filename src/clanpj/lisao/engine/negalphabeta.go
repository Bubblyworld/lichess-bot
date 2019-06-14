package engine

import (
	"fmt"
	"os"
	"strings"
	
	dragon "github.com/Bubblyworld/dragontoothmg"
)

const DEBUG = false

// Used for non-PV sub-searches
// TODO (rpj) rather just use nil on non-PV paths
var dummyPvLine = make([]dragon.Move, MaxDepth)

// Return the new best eval, best move and updated alpha (functional style for Guy)
func updateEval(bestEval EvalCp, bestMove dragon.Move, alpha EvalCp, eval EvalCp, move dragon.Move, ppvLine []dragon.Move, pvLine []dragon.Move) (EvalCp, dragon.Move, EvalCp) {
	// Maximise our eval.
	// Note - this MUST be strictly > because we fail-soft AT the current best evel - beware!
	if eval > bestEval {
		bestEval, bestMove = eval, move
	}
	
	if alpha < bestEval {
		alpha = bestEval
		// Update the PV line
		if pvLine != nil {
			pvLine[0] = move
			if ppvLine != nil {
				copy(ppvLine[1:], pvLine)
			}
		}
	}

	return bestEval, bestMove, alpha
}

// Do a shallow search to get a best move
func (s *SearchT) getShallowBestMove(depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, eval0 EvalCp, ttMove dragon.Move) dragon.Move {
	shallowBestMove := NoMove //ttMove

	if UseIDMoveHint && (UseIDMoveHintAlways || ttMove == NoMove) && depthToGo >= 2/*MinIDMoveHintDepth*/ { // Empirically doing depth 1 probe at skip 1 is worse; ditto depth 2 at skip 2
		s.stats.NShallowBestMoveCalcs++
		
		// Get the best move for a search of depth-2. TODO
		// We go 2 plies shallower since our eval is unstable between odd/even plies.
		// The result is effectively the (possibly new) ttMove.
		// We weaken the alpha and beta bounds bit to get a more accurate best-move (particularly in null-windows).
		idGap := EvalCp(20) // TODO (rpj) tune by depthToGo?
		if depthToGo < 3 { idGap = EvalCp(60) }
		idAlpha := widenAlpha(alpha, idGap)
		idBeta := widenBeta(beta, idGap)
		depthSkip := 1
		if depthToGo >= 3 {
			depthSkip = 2
			//if depthToGo >= 3 {
				//depthSkip = 3
				if depthToGo >= 20 {
					depthSkip = 4
				}
			//}
		}
		shallowBestMove, _ = s.NegAlphaBeta(depthToGo-depthSkip, depthFromRoot, idAlpha, idBeta, eval0, dummyPvLine, /*needMove*/true)

		if shallowBestMove == NoMove {
			s.stats.NNoMoveShallowBestMove++
		}
	}

	return shallowBestMove
}

// Return true iff we get a beta cut from null-move heuristic.
func (s *SearchT) nullMoveEval(depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, eval0 EvalCp, isInCheck bool, needMove bool) EvalCp {
	// Default to returning alpha (as always)
	nullMoveEval := YourCheckMateEval
	nullMoveBestResponse := NoMove
	
	if HeurUseNullMove {
		// Empirically doing a skip-depth 1 at depth-to-go 1 is worse than not
		const nullMoveDepthSkip = 2
		// Try null-move - but never in check otherwise king gets captured and never when we need a (shallow) move hint
		if !isInCheck && !needMove && beta != MyCheckMateEval && depthToGo >= nullMoveDepthSkip {
			depthSkip := nullMoveDepthSkip
			if depthToGo >= 5 {
				depthSkip++
				if depthToGo >= 9 {
					depthSkip++
				}
			}
			unapply := s.board.ApplyNullMove()
			// We use a null-window probe here (-beta+1) but then we can't use the null-move eval to raise alpha
			nullMoveBestResponse, nullMoveEval = s.NegAlphaBeta(depthToGo-depthSkip, depthFromRoot+1, -beta, -beta+1, -eval0, dummyPvLine, /*needMove*/false)
			nullMoveEval = -nullMoveEval // back to our perspective
			unapply()
			
			if(DEBUG) { fmt.Printf("                           %snull-move response %s alpha %6d beta %6d eval0 %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &nullMoveBestResponse, alpha, beta, eval0, nullMoveEval) }
			
			// If the null-move eval is a beta cut, then sanity check for zugzwang with shallow search to same depth as null-move search
			if beta <= nullMoveEval {
				_, shallowEval := s.NegAlphaBeta(depthToGo-depthSkip, depthFromRoot, -beta, -beta+1, eval0, dummyPvLine, /*needMove*/true) // TODO needMove == true to avoid shallower null but it's a hack
				if shallowEval < nullMoveEval {
					nullMoveEval = shallowEval
				}
			}
		}
	}

	return nullMoveEval
}

func (s *SearchT) probeTT(depthToGo int, alpha EvalCp, beta EvalCp) (dragon.Move, int, EvalCp, bool) {
	if UseTT {
		//var deepTtEntry TTEntryT
		//var isDeepTtHit bool
		
		// Try the deep TT
		if UseDeepTT && depthToGo >= s.deepTtMinDepth() {
			ttEntry, isTTHit := probeTT(tt2, s.board.Hash())
			// We check that the TT entry is not a relic from a shallower search (in which case it is likely to be worse than the main TT entry)
			// ??? What if we hit in the deep TT and not in the main TT ???
			if isTTHit && s.ttHitIsDeep(&ttEntry) {
				return s.processTtHit(&ttEntry, depthToGo, alpha, beta)
			}
		}

		// Try the general TT
		ttEntry, isTTHit := probeTT(tt, s.board.Hash())

		if isTTHit {
			return s.processTtHit(&ttEntry, depthToGo, alpha, beta)
		}
	}
	
	return NoMove, 0, YourCheckMateEval, false
}

func (s *SearchT) ttHitIsDeep(ttEntry *TTEntryT) bool {
	deepTtMinDepthPlus1 := uint8(s.deepTtMinDepth()+1)
	
	return ttEntry.lbDepthToGoPlus1 >= deepTtMinDepthPlus1 ||
		ttEntry.ubDepthToGoPlus1 >= deepTtMinDepthPlus1
}

func (s *SearchT) processTtHit(ttEntry *TTEntryT, depthToGo int, alpha EvalCp, beta EvalCp) (dragon.Move, int, EvalCp, bool) {
	s.stats.TTHits++
	
	depthToGoPlus1 := uint8(depthToGo+1)

	ttDepthToGo := int(ttEntry.lbDepthToGoPlus1)-1
	if ttEntry.lbDepthToGoPlus1 < ttEntry.ubDepthToGoPlus1 {
		ttDepthToGo = int(ttEntry.ubDepthToGoPlus1)-1
	}
	
	//////// First try to find an exact eval at the same depth (or deeper if HeurUseTTDeeperHits is configured)
	
	// We can use the LB value if its same depth or deeper (and HeurUseTTDeeperHits is configured)
	ttLbEntryUseable := ttEntry.lbDepthToGoPlus1 != 0 && (ttEntry.lbDepthToGoPlus1 == depthToGoPlus1 || (HeurUseTTDeeperHits && ttEntry.lbDepthToGoPlus1 > depthToGoPlus1))
	
	// We can use the UB value if its same depth or deeper (and HeurUseTTDeeperHits is configured)
	ttUbEntryUseable := ttEntry.ubDepthToGoPlus1 != 0 && (ttEntry.ubDepthToGoPlus1 == depthToGoPlus1 || (HeurUseTTDeeperHits && ttEntry.ubDepthToGoPlus1 > depthToGoPlus1))
	
	// See if we have an exact value
	if ttLbEntryUseable && ttUbEntryUseable && ttEntry.ubEval <= ttEntry.lbEval {
		s.stats.TTTrueEvals++
		return ttEntry.bestMove, ttDepthToGo, ttEntry.ubEval, true
	}
	
	//////// See if we have a beta cut
	if ttLbEntryUseable {
		if beta <= ttEntry.lbEval {
			s.stats.TTBetaCuts++
			return ttEntry.bestMove, ttDepthToGo, ttEntry.lbEval, true
		}
	}
	
	//////// See if we have an alpha cut
	if ttUbEntryUseable {
		if ttEntry.ubEval <= alpha {
			s.stats.TTAlphaCuts++
			return ttEntry.bestMove, ttDepthToGo, ttEntry.ubEval, true
		}
	}
	
	//////// We can't use the eval, but use the bestMove anyhow
	return ttEntry.bestMove, ttDepthToGo, YourCheckMateEval, false
}

func (s *SearchT) updateTt(depthToGo int, origAlpha EvalCp, origBeta EvalCp, bestEval EvalCp, bestMove dragon.Move) {
	if UseTT {
		evalType := TTEvalExact
		if origBeta <= bestEval {
			evalType = TTEvalLowerBound
		} else if bestEval <= origAlpha {
			evalType = TTEvalUpperBound
		}

		// Write back to the deep TT if this is a deep node
		if UseDeepTT && depthToGo >= s.deepTtMinDepth() {
			tt2.writeTTEntry(s.board.Hash(), bestEval, bestMove, depthToGo, evalType)
		}
		
		// Write back to the main TT
		tt.writeTTEntry(s.board.Hash(), bestEval, bestMove, depthToGo, evalType)
	}
}

func widenAlpha(alpha EvalCp, pad EvalCp) EvalCp {
	if alpha < YourCheckMateEval + pad {
		return YourCheckMateEval
	} else {
		return alpha - pad
	}
}

func widenBeta(beta EvalCp, pad EvalCp) EvalCp {
	if MyCheckMateEval - pad < beta {
		return MyCheckMateEval
	} else {
		return beta + pad
	}
}

var MaxD0DM1EvalDiff = -1000000
var MinD0DM1EvalDiff = 1000000
var NodesD0 = 0
var NodesDM1 = 0
var NodesD0FullWidth = 0
var NodesD0NegDiff = 0

const MaxD0DM1EvalDiffEstimate = EvalCp(100)

// Node eval at depth -1
// Returns bestMove, eval, nChildrenVisited
func (s *SearchT) NegAlphaBetaDepthM1(depthFromRoot int, alpha EvalCp, beta EvalCp, ttMove dragon.Move, eval0 EvalCp) (dragon.Move, EvalCp, int) {
	// Maximise eval with beta cut-off
	bestMoveM1 := NoMove
	bestEvalM1 := YourCheckMateEval
	nChildrenVisited := 0

	// TODO - why no TT here?

	// We use a fake run-once loop so that we can break after each search step rather than indenting code arbitrarily deeper with each new feature/optimisation.
done:
	for once := true; once; once = false {
		var boardSave dragon.BoardSaveT

		// Generate all legal moves
		legalMoves, _ := s.board.GenerateLegalMoves2(false /*all moves*/)

		// Check for checkmate or stalemate
		if len(legalMoves) == 0 {
			s.stats.Mates++
			bestMoveM1, bestEvalM1 = NoMove, negaMateEval(s.board, depthFromRoot) // TODO use isInCheck

			break done
		}

		// Sort the moves heuristically
		if UseMoveOrdering && len(legalMoves) > 1 {
			orderMoves(s.board, legalMoves, NoMove, ttMove, s.kt.killersForDepth(depthFromRoot)[:], s.stats.Killers[depthFromRoot][:])
		}

		for _, move := range legalMoves {
			NodesDM1++
			
			// Make the move
			s.board.MakeMove(move, &boardSave)
			// Add to the move history
			repetitions := s.ht.Add(s.board.Hash())
			
			childEval0 := NegaStaticEvalOrder0Fast(s.board, -eval0, &boardSave)
			
			// Get the (deep) eval
			var eval EvalCp
			// We consider 2-fold repetition to be a draw, since if a repeat can be forced then it can be forced again.
			// This reduces the search tree a bit and is common practice in chess engines.
			if UsePosRepetition && repetitions > 1 {
				s.stats.PosRepetitions++
				eval = DrawEval
			} else {
				_, eval, _ = s.QSearchNegAlphaBeta(QSearchDepth, depthFromRoot+1, /*depthFromQRoot*/0, -beta, -alpha, childEval0)
			}
			eval = -eval // back to our perspective

			// Remove from the move history
			s.ht.Remove(s.board.Hash())
			// Take back the move
			s.board.Restore(&boardSave)

			nChildrenVisited++

			if(true) {
				if(DEBUG) { fmt.Printf("                               %s(M1) move %s alpha %6d beta %6d eval0 %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &move, alpha, beta, eval0, eval) }
			}

			// Maximise our eval.
			// Note - this MUST be strictly > because we fail-soft AT the current best evel - beware!
			if eval > bestEvalM1 {
				bestEvalM1, bestMoveM1 = eval, move
			}

			if alpha < bestEvalM1 {
				alpha = bestEvalM1
			}

			// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
			if alpha >= beta {
 				// beta cut-off
				break done
			}
		}

	} // end of fake run-once loop

	s.kt.addKillerMove(bestMoveM1, depthFromRoot)
	
	return bestMoveM1, bestEvalM1, nChildrenVisited
}

// Leaf node eval - we do a split level eval to avoid even/odd eval zigzagging
// Returns bestMove, eval, nChildrenVisited (at depth M1)
func (s *SearchT) NegAlphaBetaDepth0(depthFromRoot int, alpha EvalCp, beta EvalCp, ttMove dragon.Move, eval0 EvalCp) (dragon.Move, EvalCp, int) {
	NodesD0++
	
	// Quiessence eval at this node
	bestMoveD0, rawEvalD0, _ := s.QSearchNegAlphaBeta(QSearchDepth, depthFromRoot, /*depthFromQRoot*/0, alpha, beta, eval0)
	if(DEBUG) { fmt.Printf("                             %sD0 move %s alpha %6d beta %6d eval0 %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &bestMoveD0, alpha, beta, eval0, rawEvalD0) }

	// Non-balanced eval
	if(!UseBalancedEval) {
		return bestMoveD0, rawEvalD0, 0
	}

	// Early out checkmate
	if isCheckmateEval(rawEvalD0) {
		return bestMoveD0, rawEvalD0, 0
	}

	// Search eval 1 extra ply
	bestMoveM1, bestEvalM1, nChildrenVisited := s.NegAlphaBetaDepthM1(depthFromRoot, alpha, beta, ttMove, eval0)
	if(DEBUG) { fmt.Printf("                             %sM1 move %s alpha %6d beta %6d eval0 %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &bestMoveM1, alpha, beta, eval0, bestEvalM1) }
	
	// Early out checkmate
	if isCheckmateEval(bestEvalM1) {
		return bestMoveM1, bestEvalM1, 0
	}

	// If the depth 0 and depth 1 results lie in the same ranges, then we're done.
	// If not, then we need to re-evaluate with wider bounds.
	if rawEvalD0 <= alpha && alpha < bestEvalM1 ||
		rawEvalD0 < beta && beta <= bestEvalM1 ||
		bestEvalM1 <= alpha && alpha < rawEvalD0 ||
		bestEvalM1 < beta && beta <= rawEvalD0 {

		NodesD0FullWidth++

		// TODO can do much better than this
		
		// Quiessence eval at this node
		bestMoveD0, rawEvalD0, _ = s.QSearchNegAlphaBeta(QSearchDepth, depthFromRoot, /*depthFromQRoot*/0, YourCheckMateEval, MyCheckMateEval, eval0)
		if(DEBUG) { fmt.Printf("                             %sD0#2 move %s alpha %6d beta %6d eval0 %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &bestMoveD0, alpha, beta, eval0, rawEvalD0) }
		
		// Search eval 1 extra ply
		bestMoveM1, bestEvalM1, nChildrenVisited = s.NegAlphaBetaDepthM1(depthFromRoot, YourCheckMateEval, MyCheckMateEval, ttMove, eval0)
		if(DEBUG) { fmt.Printf("                             %sM1#2 move %s alpha %6d beta %6d eval0 %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &bestMoveM1, alpha, beta, eval0, bestEvalM1) }
	}

	bestMove := bestMoveM1
	if bestMove == NoMove {
		bestMove = bestMoveD0
	}

	d0DM1EvalDiff := int(bestEvalM1) - int(rawEvalD0)
	if d0DM1EvalDiff < MinD0DM1EvalDiff {
		MinD0DM1EvalDiff = d0DM1EvalDiff
	}
	if MaxD0DM1EvalDiff < d0DM1EvalDiff {
		MaxD0DM1EvalDiff = d0DM1EvalDiff
	}
	if d0DM1EvalDiff < 0 {
		NodesD0NegDiff++
	}

	if true {//alpha <= rawEvalD0 && rawEvalD0 <= beta && alpha <= bestEvalM1 && bestEvalM1 <= beta {
		//if(DEBUG) { fmt.Printf("                                                                                    %salpha %6d / beta %6d - eval0: %6d  evalM1: %6d\n", strings.Repeat("  ", depthFromRoot), alpha, beta, rawEvalD0, bestEvalM1) }
	}

	// TODO (RPJ) this code is an attempt to round-to M1 but I don't think it's right
	return bestMoveD0, EvalCp((int(rawEvalD0) + int(bestEvalM1) + (int(bestEvalM1)&1))/2), nChildrenVisited
}

const DEBUG_EVAL0 = false

// Return the best eval attainable through alpha-beta from the given position, along with the move leading to the principal variation.
func (s *SearchT) NegAlphaBeta(depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, eval0 EvalCp, ppvLine []dragon.Move, needMove bool) (dragon.Move, EvalCp) {

	// Sanity check the eval0
	if DEBUG_EVAL0 {
		eval0Check := NegaStaticEvalOrder0(s.board)
		if eval0 != eval0Check {
			fmt.Println("!!!!!!!!               eval0", eval0, "eval0Check", eval0Check, "fen", s.board.ToFen())
			os.Exit(1)
		}
	}

	// Bail if we've timed out
	if isTimedOut(s.timeout) {
		// Return the worst possible eval (opponent checkmate) to invalidate this incomplete search branch
		return NoMove, YourCheckMateEval
	}

	s.stats.Nodes++
	
	// Remember this to check whether our final eval is a lower or upper bound - for TT
	origBeta := beta
	origAlpha := alpha

	if depthToGo <= 0 { s.stats.D0TTProbes++ }
	// Probe the Transposition Table
	ttMove, ttDepthToGo, ttEval, ttIsCut := s.probeTT(depthToGo, alpha, beta)
	if depthToGo <= 0 {
		if ttIsCut {
			s.stats.D0TTCuts++
		} else if ttMove != NoMove { s.stats.D0TTMoves++ }
	}
	if ttIsCut && (ttMove != NoMove || needMove == false) {
		return ttMove, ttEval
	}

	ttDepthToGoIsD1 := depthToGo <= ttDepthToGo+1
	s.stats.AfterTTNodes++

	// Maximise eval with beta cut-off
	bestMove := NoMove
	bestEval := YourCheckMateEval
	nChildrenVisited := 0

	// Maintain the PV line - 1 extra element for NoMove cropping with tt hit
	pvLine := make([]dragon.Move, depthToGo+1)

	shallowBestMove := NoMove

	// Anything after here updates the TT - so single return location at the end of the func after writing back to TT
	// We use a fake run-once loop so that we can break after each search step rather than indenting code arbitrarily deeper with each new feature/optimisation.
done:
	for once := true; once; once = false {
		// Leaf node eval
		if depthToGo <= 0 {
			s.stats.AfterNullNodesByD[0]++
			bestMove, bestEval, nChildrenVisited = s.NegAlphaBetaDepth0(depthFromRoot, alpha, beta, ttMove, eval0)
			break done
		}

		s.stats.NonLeafs++
		if depthFromRoot < MaxDepthStats {
			s.stats.NonLeafsAt[depthFromRoot]++
		}
		
		isInCheck := isInCheckFast(s.board)

		// Null-move heuristic.
		// Note we miss stalemate here, but that should be a vanishingly small case
		nullMoveEval := s.nullMoveEval(depthToGo, depthFromRoot, alpha, beta, eval0, isInCheck, needMove)

 		// Bail cleanly without polluting search results if we have timed out
		if isTimedOut(s.timeout) {
			break done
		}

		// [This comment seems semi-bogus/defunct]: We don't use the null-move eval to raise alpha because it's only null-window probe around beta in the null-move code, not a full [alpha, beta] window
		if beta <= nullMoveEval {
			bestEval, alpha = nullMoveEval, nullMoveEval
			break done
		}

		s.stats.AfterNullNodes++
		s.stats.AfterNullNodesByD[depthToGo]++
		if ttDepthToGoIsD1 {
			s.stats.AfterNullNodesByD1[depthToGo]++
		}

		shallowBestMove = s.getShallowBestMove(depthToGo, depthFromRoot, alpha, beta, eval0, ttMove)

		// Bail cleanly without polluting search results if we have timed out
		if isTimedOut(s.timeout) {
			break done
		}

		// Generate all legal moves
		legalMoves, _ := s.board.GenerateLegalMoves2(false /*all moves*/)

		// Check for checkmate or stalemate
		if len(legalMoves) == 0 {
			s.stats.Mates++
			bestMove, bestEval = NoMove, negaMateEval(s.board, depthFromRoot) // TODO use isInCheck

			break done
		}

		// Sort the moves heuristically
		if UseMoveOrdering && len(legalMoves) > 1 {
			orderMoves(s.board, legalMoves, shallowBestMove, ttMove, s.kt.killersForDepth(depthFromRoot)[:], s.stats.Killers[depthFromRoot][:])
		}

		for _, move := range legalMoves {
			var boardSave dragon.BoardSaveT

			// Make the move
			s.board.MakeMove(move, &boardSave)
			// Add to the move history
			repetitions := s.ht.Add(s.board.Hash())
			
			childEval0 := NegaStaticEvalOrder0Fast(s.board, -eval0, &boardSave)
			
			// Get the (deep) eval
			var eval EvalCp
			// We consider 2-fold repetition to be a draw, since if a repeat can be forced then it can be forced again.
			// This reduces the search tree a bit and is common practice in chess engines.
			if UsePosRepetition && repetitions > 1 {
				s.stats.PosRepetitions++
				eval = DrawEval
			} else {
				eval = -alpha-1
				
				// Late Move Reduction (LMR) - null-window probe at reduced depth with heuristicly wider alpha - never for the first move to avoid depth collapse
				if HeurUseLMR && nChildrenVisited != 0 && depthToGo >= 4 {
					// LMR probe
					lmrAlphaPad := EvalCp(60-10*depthToGo)
					if lmrAlphaPad < 20 {
						lmrAlphaPad = EvalCp(20)
					}
					depthSkip := 2
					//if depthToGo >= 7 {
						//depthSkip = 3
						if depthToGo >= 11 {
							depthSkip = 4
						}
					//}
					
					if eval < -alpha {
						lmrAlpha := widenAlpha(alpha, lmrAlphaPad)
						_, eval = s.NegAlphaBeta(depthToGo-depthSkip, depthFromRoot+1, -lmrAlpha-1, -lmrAlpha, childEval0, dummyPvLine, /*needMove*/false)
						// If LMR probe fails to raise lmrAlpha then avoid full depth probe by fiddling eval appropriately
						if eval > -lmrAlpha-1 {
							eval = -alpha
						} else {
							eval = -alpha-1
						}
					}
				}
					
				// Null window probe (PV-search) - don't bother if the LMR probe failed
				if eval < -alpha && alpha < beta-1/*no point if it's already a null window*/ {
					// Null window probe
					_, eval = s.NegAlphaBeta(depthToGo-1, depthFromRoot+1, -alpha-1, -alpha, childEval0, dummyPvLine, /*needMove*/false)
				}
					
				if eval < -alpha {
					// Full search
					_, eval = s.NegAlphaBeta(depthToGo-1, depthFromRoot+1, -beta, -alpha, childEval0, pvLine, /*needMove*/false)
				}
			}
			eval = -eval // back to our perspective

			if(DEBUG) { fmt.Printf("                           %smove %s alpha %6d beta %6d eval0 %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &move, alpha, beta, eval0, eval) }

			// Remove from the move history
			s.ht.Remove(s.board.Hash())
			// Take back the move
			s.board.Restore(&boardSave)

			// Bail cleanly without polluting search results if we have timed out
			if isTimedOut(s.timeout) {
				break done
			}
			
			nChildrenVisited++

			// Maximise our eval.
			// Note - this MUST be strictly > because we fail-soft AT the current best evel - beware!
			if eval > bestEval {
				bestEval, bestMove = eval, move
			}

			if alpha < bestEval {
				alpha = bestEval
				// Update the PV line
				pvLine[0] = move
				copy(ppvLine[1:], pvLine)
			}

			// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
			if alpha >= beta {
				// beta cut-off
				break done
			}
		}

		// If we didn't get a beta cut-off then we visited all children.
		s.stats.AllChildrenNodes++
		
	} // end of fake run-once loop

	// Update the TT and stats - but only if the search was not truncated due to a time-out
	if !isTimedOut(s.timeout) {
		if depthToGo <= 0 { s.stats.D0TTUpdates++ }
		s.updateTt(depthToGo, origAlpha, origBeta, bestEval, bestMove)

		s.kt.addKillerMove(bestMove, depthFromRoot)

		s.stats.AfterChildLoopNodes++

		// This is irritating - leaf nodes also get here hence 0 < depthToGo, hrm this excludes d0 stats - what was I thinking?
		if true || 0 < depthToGo {
			// Exclude null-move cuts from best move origin stats hence bestMove != NoMove
			if bestMove != NoMove {
				if shallowBestMove != NoMove {
					s.stats.NShallowBestMoves++
					if shallowBestMove == bestMove {
						s.stats.NShallowBestMovesBest++
						if origBeta <= bestEval {
							s.stats.NShallowBestMoveCuts++
						}
					} else {
						if origBeta <= bestEval {
							s.stats.NShallowBestMoveOtherCuts++
						}
					}
				}
				if ttMove != NoMove {
					s.stats.NTTMoves++
					s.stats.NTTMovesByD[depthToGo]++
					if ttMove == bestMove {
						s.stats.NTTMovesBest++
						s.stats.NTTMovesBestByD[depthToGo]++
						if origBeta <= bestEval {
							s.stats.NTTMoveCuts++
							s.stats.NTTMoveCutsByD[depthToGo]++
						}
					} else {
						if origBeta <= bestEval {
							s.stats.NTTMoveOtherCuts++
							s.stats.NTTMoveOtherCutsByD[depthToGo]++
						}
					}
					if ttDepthToGoIsD1 {
						s.stats.NTTMovesByD1[depthToGo]++
						if ttMove == bestMove {
							s.stats.NTTMovesBestByD1[depthToGo]++
							if origBeta <= bestEval {
								s.stats.NTTMoveCutsByD1[depthToGo]++
							}
						} else {
							if origBeta <= bestEval {
								s.stats.NTTMoveOtherCutsByD1[depthToGo]++
							}
						}
					}
				}
			}
			
			// Is it a (beta) cut
			if origBeta <= bestEval {
				s.stats.CutNodes++
				s.stats.CutNodeChildren += uint64(nChildrenVisited)
				
				if nChildrenVisited == 0 {
					s.stats.NullMoveCuts++
				} else if nChildrenVisited == 1 {
					s.stats.FirstChildCuts++
				}
				
				if depthToGo < MinIDMoveHintDepth {
					s.stats.ShallowCutNodes++
					s.stats.ShallowCutNodeChildren += uint64(nChildrenVisited)
					if nChildrenVisited == 0 {
						s.stats.ShallowNullMoveCuts++
					}
				} else {
					s.stats.DeepCutNodes++
					s.stats.DeepCutNodeChildren += uint64(nChildrenVisited)
					if nChildrenVisited == 0 {
						s.stats.DeepNullMoveCuts++
					}
				}
				
				if bestMove != NoMove { // exclude null move cut?
					if bestMove == ttMove {
						s.stats.TTMoveCuts++
					}
					
					killers := s.kt.killersForDepth(depthFromRoot)
					for i := 0; i < NKillersPerDepth; i++ {
						if bestMove == killers[i] {
							s.stats.KillerCuts[depthFromRoot][i]++
						}
					}
				}
			} else {
				s.stats.AllChildrenNodes2++
			}
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
