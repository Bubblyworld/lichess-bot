package engine

import (
	"fmt"
	"math/bits"
	"os"
	"strings"
	
	dragon "github.com/Bubblyworld/dragontoothmg"
)

func f() { strings.Repeat(" ", 1) }

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

// Return null-move eval
func (s *SearchT) nullMove(depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, pNullOrLMR bool, pIsFirstMove bool, eval0 EvalCp, isInCheck bool) EvalCp {
	// Default to returning alpha (as always)
	nullMoveEval := YourCheckMateEval
	
	if HeurUseNullMove {
		const nullMoveDepthSkip = 2 // must be odd to cope with our even/odd ply eval instability
		// Try null-move - but never 2 null moves in a row, and never in check otherwise king gets captured
		if !isInCheck && !pNullOrLMR && !pIsFirstMove && beta != MyCheckMateEval && depthToGo >= 4/*nullMoveDepthSkip*/ {
			// Use piece count to determine end-game for zugzwang avoidance - TODO improve this
			nNonPawns := bits.OnesCount64((s.board.Bbs[dragon.White][dragon.All] & ^s.board.Bbs[dragon.White][dragon.Pawn]) | (s.board.Bbs[dragon.Black][dragon.All] & ^s.board.Bbs[dragon.Black][dragon.Pawn]))
			// Proceed with null-move heuristic if there are at least 4 non-pawn pieces (note the count includes the two kings)
			if nNonPawns >= 6 {
				unapply := s.board.ApplyNullMove()
				_, possibleNullMoveEval := s.NegAlphaBeta(depthToGo-nullMoveDepthSkip, depthFromRoot+1, -beta, -alpha, NoMove, true, true, -eval0, dummyPvLine)
				possibleNullMoveEval = -possibleNullMoveEval // back to our perspective
				unapply()
				
				// Bail cleanly without polluting search results if we have timed out
				if !isTimedOut(s.timeout) {
					nullMoveEval = possibleNullMoveEval
				}
			}
		}
	}

	return nullMoveEval
}

func (s *SearchT) probeTT(depthToGo int, alpha EvalCp, beta EvalCp) (dragon.Move, EvalCp, bool) {
	if UseTT {
		// Try the deep TT
		if UseDeepTT && depthToGo >= s.deepTtMinDepth() {
			ttEntry, isTTHit := probeTT(tt2, s.board.Hash())
			// We check that the TT entry is not a relic from a shallower search (in which case it is likely to be worse than the main TT entry)
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
	
	return NoMove, YourCheckMateEval, false
}

func (s *SearchT) ttHitIsDeep(ttEntry *TTEntryT) bool {
	deepTtMinDepth := uint8(s.deepTtMinDepth())
	
	return ttEntry.parityHits[0].lbEntry.depthToGo >= deepTtMinDepth ||
		ttEntry.parityHits[0].ubEntry.depthToGo >= deepTtMinDepth ||
		ttEntry.parityHits[1].lbEntry.depthToGo >= deepTtMinDepth ||
		ttEntry.parityHits[1].ubEntry.depthToGo >= deepTtMinDepth
}

func (s *SearchT) processTtHit(ttEntry *TTEntryT, depthToGo int, alpha EvalCp, beta EvalCp) (dragon.Move, EvalCp, bool) {
	s.stats.TTHits++
	
	//////// First try to find an exact eval at the same depth (or deeper if HeurUseTTDeeperHits is configured)
	
	// Try the same depth parity first
	ttpEntry := &ttEntry.parityHits[depthToGoParity(depthToGo)]
	
	// Try the lower bound of the same depth parity (it might be exact)
	ttpLbEntry := &ttpEntry.lbEntry
	// We can use this value if its same depth or deeper (and HeurUseTTDeeperHits is configured)
	ttpLbEntryUseable := ttpLbEntry.evalType != TTInvalid && (ttpLbEntry.depthToGo == uint8(depthToGo) || (HeurUseTTDeeperHits && ttpLbEntry.depthToGo > uint8(depthToGo)))
	var ttpLbEntryEval EvalCp
	if ttpLbEntryUseable {
		ttpLbEntryEval = s.depthSwitchEval(ttpLbEntry.eval, int(ttpLbEntry.depthToGo), depthToGo)
	}
	
	if ttpLbEntryUseable && ttpLbEntry.evalType == TTEvalExact {
		s.stats.TTTrueEvals++
		return ttpLbEntry.bestMove, ttpLbEntryEval, true
	}
	
	// Try the upper bound of the same depth parity (it might be exact)
	ttpUbEntry := &ttpEntry.ubEntry
	// We can use this value if its same depth or deeper (and HeurUseTTDeeperHits is configured)
	ttpUbEntryUseable := ttpUbEntry.evalType != TTInvalid && (ttpUbEntry.depthToGo == uint8(depthToGo) || (HeurUseTTDeeperHits && ttpUbEntry.depthToGo > uint8(depthToGo)))
	var ttpUbEntryEval EvalCp
	if ttpUbEntryUseable {
		ttpUbEntryEval = s.depthSwitchEval(ttpUbEntry.eval, int(ttpUbEntry.depthToGo), depthToGo)
	}
	
	if ttpUbEntryUseable && ttpUbEntry.evalType == TTEvalExact {
		s.stats.TTTrueEvals++
		return ttpUbEntry.bestMove, ttpUbEntryEval, true
	}
	
	// ...then try the opposite parity entry (and we have to fudge the eval to correct for even/odd parity eval differences)
	ttpEntry2 := &ttEntry.parityHits[depthToGoParity(depthToGo)^1]
	
	// Try the lower bound of the opposite depth parity (it might be exact)
	ttpLbEntry2 := &ttpEntry2.lbEntry
	// We can use this value if its deeper (and HeurUseTTDeeperHits is configured)
	ttpLbEntry2Useable := ttpLbEntry2.evalType != TTInvalid && (HeurUseTTDeeperHits && ttpLbEntry2.depthToGo > uint8(depthToGo))
	var ttpLbEntry2Eval EvalCp
	if ttpLbEntry2Useable {
		ttpLbEntry2Eval = s.depthSwitchEval(ttpLbEntry2.eval, int(ttpLbEntry2.depthToGo), depthToGo)
	}
	
	if ttpLbEntry2Useable && ttpLbEntry2.evalType == TTEvalExact {
		s.stats.TTTrueEvals++
		return ttpLbEntry2.bestMove, ttpLbEntry2Eval, true
	}
	
	// Try the upper bound of the opposite depth parity (it might be exact)
	ttpUbEntry2 := &ttpEntry2.ubEntry
	// We can use this value if its deeper (and HeurUseTTDeeperHits is configured)
	ttpUbEntry2Useable := ttpUbEntry2.evalType != TTInvalid && (HeurUseTTDeeperHits && ttpUbEntry2.depthToGo > uint8(depthToGo))
	var ttpUbEntry2Eval EvalCp
	if ttpUbEntry2Useable {
		ttpUbEntry2Eval = s.depthSwitchEval(ttpUbEntry2.eval, int(ttpUbEntry2.depthToGo), depthToGo)
	}
	
	if ttpUbEntry2Useable && ttpUbEntry2.evalType == TTEvalExact {
		s.stats.TTTrueEvals++
		return ttpUbEntry2.bestMove, s.depthSwitchEval(ttpUbEntry2.eval, int(ttpUbEntry2.depthToGo), depthToGo), true
	}
	
	//////// See if we have a beta cut
	
	// First for TT entry of the same parity
	if ttpLbEntryUseable {
		if beta <= ttpLbEntryEval {
			s.stats.TTBetaCuts++
			return ttpLbEntry.bestMove, ttpLbEntryEval, true
		}
	}
	
	// ... then for TT entry of the opposite parity (and we have to fudge the eval to correct for even/odd parity eval differences)
	if ttpLbEntry2Useable {
		if beta <= ttpLbEntry2Eval {
			s.stats.TTBetaCuts++
			return ttpLbEntry2.bestMove, ttpLbEntry2Eval, true
		}
	}
	
	//////// See if we have an alpha cut
	
	// First for TT entry of the same parity
	if ttpUbEntryUseable {
		if ttpUbEntryEval <= alpha {
			s.stats.TTAlphaCuts++
			return ttpUbEntry.bestMove, ttpUbEntryEval, true
		}
	}
	
	// ... then for TT entry of the opposite parity (and we have to fudge the eval to correct for even/odd parity eval differences)
	if ttpUbEntry2Useable {
		if ttpUbEntry2Eval <= alpha {
			s.stats.TTAlphaCuts++
			return ttpUbEntry2.bestMove, ttpUbEntry2Eval, true
		}
	}
	
	//////// Set the ttMove
	ttMove, ttMoveDepthToGo := NoMove, uint8(0)
	if ttpLbEntry.evalType != TTInvalid {
		ttMove = ttpLbEntry.bestMove
		ttMoveDepthToGo = ttpLbEntry.depthToGo
	} else if ttpUbEntry.evalType != TTInvalid && ttMoveDepthToGo < ttpUbEntry.depthToGo {
		ttMove = ttpUbEntry.bestMove
		ttMoveDepthToGo = ttpUbEntry.depthToGo
	} else if ttpLbEntry2.evalType != TTInvalid && ttMoveDepthToGo < ttpLbEntry2.depthToGo {
		ttMove = ttpLbEntry2.bestMove
		ttMoveDepthToGo = ttpLbEntry2.depthToGo
	} else if ttpUbEntry2.evalType != TTInvalid && ttMoveDepthToGo < ttpUbEntry2.depthToGo {
		ttMove = ttpUbEntry2.bestMove
		ttMoveDepthToGo = ttpUbEntry2.depthToGo
	}
	
	return ttMove, YourCheckMateEval, false
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

// Node eval at depth 1
func (s *SearchT) NegAlphaBetaDepthM1(depthFromRoot int, alpha EvalCp, beta EvalCp, killer dragon.Move, eval0 EvalCp) (dragon.Move, EvalCp) {
	killerMove := NoMove
	if UseKillerMoves {
		killerMove = killer
	}
	deepKiller := NoMove
	if UseDeepKillerMoves {
		deepKiller = s.deepKillers[depthFromRoot]
	}
		
	// Maximise eval with beta cut-off
	bestMoveM1 := NoMove
	bestEvalM1 := YourCheckMateEval
	childKiller := NoMove

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
		if UseMoveOrdering {
			if len(legalMoves) > 1 {
				orderMoves(s.board, legalMoves, NoMove, killerMove, deepKiller, &s.stats.Killers, &s.stats.DeepKillers)
			}
		} else if UseKillerMoves {
			// Place killer-move (or deep killer move) first if it's there
			prioritiseKillerMove(legalMoves, killer, UseDeepKillerMoves, s.deepKillers[depthFromRoot], &s.stats.Killers, &s.stats.DeepKillers)
		}

		for i, move := range legalMoves {
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
				childKiller, eval, _ = s.QSearchNegAlphaBeta(QSearchDepth, depthFromRoot+1, /*depthFromQRoot*/0, -beta, -alpha, childKiller, childEval0)
			}
			eval = -eval // back to our perspective

			// Remove from the move history
			s.ht.Remove(s.board.Hash())
			// Take back the move
			s.board.Restore(&boardSave)

			if(false) {
				fmt.Printf("                           %smove %s alpha %6d beta %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &move, alpha, beta, eval)
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
				if bestMoveM1 == killerMove {
					s.stats.KillerCuts++
				} else if bestMoveM1 == deepKiller {
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
		if bestEvalM1 < beta {
			s.stats.AllChildrenNodes++
		}
		s.deepKillers[depthFromRoot] = bestMoveM1
	} // end of fake run-once loop

	return bestMoveM1, bestEvalM1
}

// Leaf node eval
func (s *SearchT) NegAlphaBetaDepth0(depthFromRoot int, alpha EvalCp, beta EvalCp, killer dragon.Move, eval0 EvalCp) (dragon.Move, EvalCp) {
	NodesD0++
	
	// Quiessence eval at this node
	bestMoveD0, rawEvalD0, _ := s.QSearchNegAlphaBeta(QSearchDepth, depthFromRoot, /*depthFromQRoot*/0, alpha, beta, killer, eval0)

	// Early out checkmate
	if isCheckmateEval(rawEvalD0) {
		return bestMoveD0, rawEvalD0
	}

	// alphaDM1 := YourCheckMateEval
	// betaDM1 := MyCheckMateEval

	// Search eval 1 extra ply
	bestMoveM1, bestEvalM1 := s.NegAlphaBetaDepthM1(depthFromRoot, alpha, beta, killer, eval0)
	
	// Early out checkmate
	if isCheckmateEval(bestEvalM1) {
		return bestMoveM1, bestEvalM1
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
		bestMoveD0, rawEvalD0, _ = s.QSearchNegAlphaBeta(QSearchDepth, depthFromRoot, /*depthFromQRoot*/0, YourCheckMateEval, MyCheckMateEval, killer, eval0)
		
		// Search eval 1 extra ply
		bestMoveM1, bestEvalM1 = s.NegAlphaBetaDepthM1(depthFromRoot, YourCheckMateEval, MyCheckMateEval, killer, eval0)
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
		//fmt.Printf("                                                                                    %salpha %6d / beta %6d - eval0: %6d  evalM1: %6d\n", strings.Repeat("  ", depthFromRoot), alpha, beta, rawEvalD0, bestEvalM1)
	}

	return bestMoveD0, EvalCp((int(rawEvalD0) + int(bestEvalM1) + (int(bestEvalM1)&1))/2)
}

// Return the best eval attainable through alpha-beta from the given position (with killer-move hint), along with the move leading to the principal variation.
func (s *SearchT) NegAlphaBeta(depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, killer dragon.Move, pNullOrLMR bool, pIsFirstMove bool, eval0 EvalCp, ppvLine []dragon.Move) (dragon.Move, EvalCp) {

	// Sanity check the eval0
	if false {
		eval0Check := NegaStaticEvalOrder0(s.board)
		if eval0 != eval0Check {
			fmt.Println("               eval0", eval0, "eval0Check", eval0Check, "fen", s.board.ToFen())
			os.Exit(1)
		}
	}

	// Bail if we've timed out
	if isTimedOut(s.timeout) {
		// Return the worst possible eval (opponent checkmate) to invalidate this incomplete search branch
		return NoMove, YourCheckMateEval
	}

	s.stats.Nodes++
	
	// Leaf node eval eval - note that low-depthToGo null moves come in with depthToGo < 0
	if depthToGo <= 0 {
		return s.NegAlphaBetaDepth0(depthFromRoot, alpha, beta, killer, eval0)
	}

	s.stats.NonLeafs++
	if depthFromRoot < MaxDepthStats {
		s.stats.NonLeafsAt[depthFromRoot]++
	}

	// Remember this to check whether our final eval is a lower or upper bound - for TT
	origBeta := beta
	origAlpha := alpha

	// Probe the Transposition Table
	ttMove, ttEval, ttIsCut := s.probeTT(depthToGo, alpha, beta)
	if ttIsCut {
		return ttMove, ttEval
	}
	
	// Maximise eval with beta cut-off
	bestMove := NoMove
	bestEval := YourCheckMateEval
	childKiller := NoMove

	// Maintain the PV line - 1 extra element for NoMove cropping with tt hit
	pvLine := make([]dragon.Move, depthToGo+1)

	// Anything after here interacts with the QTT - so single return location at the end of the func after writing back to QTT
	// We use a fake run-once loop so that we can break after each search step rather than indenting code arbitrarily deeper with each new feature/optimisation.
done:
	for once := true; once; once = false {
		var boardSave dragon.BoardSaveT

		isInCheck := isInCheckFast(s.board)

		// Null-move heuristic.
		// Note we miss stalemate here, but that should be a vanishingly small case
		nullMoveEval := s.nullMove(depthToGo, depthFromRoot, alpha, beta, pNullOrLMR, pIsFirstMove,  -eval0, isInCheck)
		bestEval, bestMove, alpha = updateEval(bestEval, bestMove, alpha, nullMoveEval, NoMove, nil, nil)
		// Did null-move heuristic already provide a cut?
		// TODO - I'm not sure we want to raise alpha using null move ever!
		if alpha >= beta {
			s.stats.NullMoveCuts++
			break done
		}

		killerMove := NoMove
		if UseKillerMoves {
			killerMove = killer
		}
		deepKiller := NoMove
		if UseDeepKillerMoves {
			deepKiller = s.deepKillers[depthFromRoot]
		}
		
		if UseIDMoveHint && depthToGo >= MinIDMoveHintDepth {
			// Get the best move for a search of depth-2.
			// We go 2 plies shallower since our eval is unstable between odd/even plies.
			// The result is effectively the (possibly new) ttMove.
			// We weaken the alpha and beta bounds bit to get a more accurate best-move (particularly in null-windows).
			// const minIdGap = EvalCp(5)
			// idGap := EvalCp(12 - depthToGo/2)
			// if idGap < minIdGap { idGap = minIdGap }
			idGap := EvalCp(5) // best at depth 12 at start pos but it's quite sensitive
			idAlpha := widenAlpha(alpha, idGap)
			idBeta := widenBeta(beta, idGap)
			depthSkip := 2
			if depthToGo >= 5/*MinIDMoveHintDepth*/ {
				depthSkip = 3
				if depthToGo >= 9 {
					depthSkip = 4
				}
			}
			idMove, _ := s.NegAlphaBeta(depthToGo-depthSkip, depthFromRoot, idAlpha, idBeta, killer, pNullOrLMR, pIsFirstMove/*???*/, eval0, dummyPvLine)
			if idMove != NoMove {
				ttMove = idMove
			}
		}
		
		// TODO also use killer moves, but need to check them first for validity
		hintMove := ttMove
		// We don't do null-window or depth-reduction for the first child
		isFirstMove := true

		// Try hint move before doing move-gen if we have a known valid move hint
		if UseEarlyMoveHint {
			if hintMove != NoMove {
				s.stats.ValidHintMoves++
				// Make the move
				s.board.MakeMove(hintMove, &boardSave)
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
					childKiller, eval = s.NegAlphaBeta(depthToGo-1, depthFromRoot+1, -beta, -alpha, childKiller, false, isFirstMove, childEval0, pvLine)
				}
				eval = -eval // back to our perspective
				// Remove from the move history
				s.ht.Remove(s.board.Hash())
				// Take back the move
				s.board.Restore(&boardSave)

				//fmt.Printf("                           %smove %s alpha %6d beta %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &hintMove, alpha, beta, eval)
				
				isFirstMove = false

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
					// Update the PV line
					pvLine[0] = hintMove
					copy(ppvLine[1:], pvLine)
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
		legalMoves, _ := s.board.GenerateLegalMoves2(false /*all moves*/)

		// Check for checkmate or stalemate
		if len(legalMoves) == 0 {
			s.stats.Mates++
			bestMove, bestEval = NoMove, negaMateEval(s.board, depthFromRoot) // TODO use isInCheck

			break done
		}

		// Sort the moves heuristically
		if UseMoveOrdering {
			if len(legalMoves) > 1 {
				orderMoves(s.board, legalMoves, ttMove, killerMove, deepKiller, &s.stats.Killers, &s.stats.DeepKillers)
			}
		} else if UseKillerMoves {
			// Place killer-move (or deep killer move) first if it's there
			prioritiseKillerMove(legalMoves, killer, UseDeepKillerMoves, s.deepKillers[depthFromRoot], &s.stats.Killers, &s.stats.DeepKillers)
		}

		for i, move := range legalMoves {
			// Don't repeat the hintMove
			if UseEarlyMoveHint && move == hintMove {
				continue
			}

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
				// LMR and null window probe - don't bother if we're on the PV
				if !isFirstMove {
					// Late Move Reduction - null-window probe at reduced depth with heuristicly wider alpha
					if HeurUseLMR && !pNullOrLMR && depthToGo >= 4 {
						// LMR probe
						lmrAlphaPad := EvalCp(60-10*depthToGo)
						if lmrAlphaPad < 15 {
							lmrAlphaPad = EvalCp(15)
						}
						depthSkip := 2
						if depthToGo >= 7 {
							depthSkip = 3
							if depthToGo >= 11 {
								depthSkip = 4
							}
						}
						
						if eval < -alpha {
							lmrAlpha := widenAlpha(alpha, lmrAlphaPad)
							childKiller, eval = s.NegAlphaBeta(depthToGo-depthSkip, depthFromRoot+1, -lmrAlpha-1, -lmrAlpha, childKiller, true, isFirstMove, childEval0, dummyPvLine)
							// If LMR probe fails to raise lmrAlpha then avoid full depth probe by fiddling eval appropriately
							if eval > -lmrAlpha-1 {
								eval = -alpha
							} else {
								eval = -alpha-1
							}
						}
					}
					
					// Null window probe (PV-search) with aggressive null/lmr - don't bother if the LMR probe failed
					const nwAlphaPad = EvalCp(5)
					if eval < -alpha {
						nwAlpha := widenAlpha(alpha, nwAlphaPad)
						// TODO(rpj) pvLine???
						childKiller, eval = s.NegAlphaBeta(depthToGo-1, depthFromRoot+1, -nwAlpha-1, -nwAlpha, childKiller, false, false, childEval0, dummyPvLine)
						if eval > -nwAlpha-1 {
							eval = -alpha
						} else {
							eval = -alpha-1
						}
					}
				}

				// If we get here then we might have a new PV, so treat this as the 'first move'
				if eval < -alpha && alpha < beta-1/*ignore if it's already a null window*/ {
					// Null window probe
					childKiller, eval = s.NegAlphaBeta(depthToGo-1, depthFromRoot+1, -alpha-1, -alpha, childKiller, false, true, childEval0, dummyPvLine)
				}
					
				if eval < -alpha {
					// Full search
					childKiller, eval = s.NegAlphaBeta(depthToGo-1, depthFromRoot+1, -beta, -alpha, childKiller, false, true, childEval0, pvLine)
				}
			}
			eval = -eval // back to our perspective

			//fmt.Printf("                           %smove %s alpha %6d beta %6d eval %6d \n", strings.Repeat("  ", depthFromRoot), &move, alpha, beta, eval)

			// Remove from the move history
			s.ht.Remove(s.board.Hash())
			// Take back the move
			s.board.Restore(&boardSave)

			isFirstMove = false

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
				// Update the PV line
				pvLine[0] = move
				copy(ppvLine[1:], pvLine)
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

	// Update the TT - but only if the search was not truncated due to a time-out
	if !isTimedOut(s.timeout) {
		s.updateTt(depthToGo, origAlpha, origBeta, bestEval, bestMove)
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
