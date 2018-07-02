package engine

import (
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

const MaxDepthStats = 16
const MaxQDepthStats = 16

type SearchStatsT struct {
	Nodes uint64             // #nodes visited
	Mates uint64             // #true terminal nodes
	NonLeafs uint64          // #non-leaf nodes
	FirstChildCuts uint64    // #non-leaf nodes that (beta-)cut on the first child searched
	AllChildrenNodes uint64  // #non-leaf nodes with no beta cut
	Killers uint64           // #nodes with killer move available
	KillerCuts uint64        // #nodes with killer move cut
	DeepKillers uint64       // #nodes with deep killer move available
	DeepKillerCuts uint64    // #nodes with deep killer move cut
	PosRepetitions uint64    // #nodes with repeated position
	TTHits uint64            // #nodes with successful TT probe
	TTDepthHits uint64       // #nodes where TT hit was at the same depth
	TTBetaCuts uint64        // #nodes with beta cutoff from TT hit
	TTAlphaCuts uint64       // #nodes with alpha cutoff from TT hit
	TTLateCuts uint64        // #nodes with beta cutoff from TT hit
	TTTrueEvals uint64       // #nodes with QQT hits that are the same depth and are not a lower bound
	QNodes uint64            // #nodes visited in qsearch
	QMates uint64            // #true terminal nodes in qsearch
	QNonLeafs uint64         // #non-leaf qnodes
	QFirstChildCuts uint64   // #non-leaf qnodes that (beta-)cut on the first child searched
	QAllChildrenNodes uint64 // #non-leaf qnodes with no beta cut
	QKillers uint64          // #qnodes with killer move available
	QKillerCuts uint64       // #qnodes with killer move cut
	QDeepKillers uint64      // #qnodes with deep killer move available
	QDeepKillerCuts uint64   // #qnodes with deep killer move cut
	QRampagePrunes uint64    // #qnodes where we did queen rampage pruning
	QPats uint64             // #qnodes with stand pat best
	QPatCuts uint64          // #qnodes with stand pat cut
	QQuiesced uint64         // #qnodes where we successfully quiesced
	QPrunes uint64           // #qnodes where we reached full depth - i.e. likely failed to quiesce
	QttHits uint64           // #qnodes with successful QTT probe
	QttDepthHits uint64      // #qnodes where QTT hit was at the same depth
	QttBetaCuts uint64       // #qnodes with beta cutoff from QTT hit
	QttAlphaCuts uint64      // #qnodes with beta cutoff from QTT hit
	QttLateCuts uint64       // #qnodes with beta cutoff from QTT hit
	QttTrueEvals uint64      // #qnodes with QQT hits that are the same depth and are not a lower bound

	NonLeafsAt [MaxDepthStats]uint64      // non-leafs by depth
	FirstChildCutsAt[MaxDepthStats]uint64 // first-child cuts by depth
	QNonLeafsAt [MaxQDepthStats]uint64    // q-search non-leafs by depth
}
	

// Configuration options
type SearchAlgorithmT int
const (
	MiniMax SearchAlgorithmT = iota
	NegaMax
	AlphaBeta
	NegAlphaBeta
)
var SearchAlgorithm = NegAlphaBeta
var SearchDepth = 7                 // Ignored now that time control is implemented
var SearchCutoffPercent = 25        // If we've used more than this percentage of the target time then we bail on the search instead of starting a new depth
var UseDeltaEval = true
var UseMoveOrdering = true
var UseKillerMoves = true
var UseDeepKillerMoves = true       // only valid if UseKillerMoves == true
var UseTT = false //true
var UsePosRepetition = true
var UseQSearch = true
var QSearchDepth = 12
var UseQSearchTT = false //true
var UseQSearchMoveOrdering = true
var UseQSearchRampagePruning = true // only valid if UseQSearchMoveOrdering == true
var QSearchRampagePruningDepth = 4  // only valid if UseQSearchRampagePruning == true
var UseQKillerMoves = true
var UseQDeepKillerMoves = true     // only valid if UseQKillerMoves == true

func SearchAlgorithmString() string {
	switch SearchAlgorithm {
	case MiniMax:
		return "MiniMax"
	case NegaMax:
		return "NegaMax"
	case AlphaBeta:
		return "AlphaBeta"
	case NegAlphaBeta:
		return "NegAlphaBeta"
	default:
		SearchAlgorithm = NegAlphaBeta
		return "NegAlphaBeta"
	}
}

const MinDepth = 1
const MaxDepth = 1024
const NoMove dragon.Move = 0

func isTimedOut(timeout *uint32) bool {
	return atomic.LoadUint32(timeout) != 0
}

// Return eval from white's perspective, and the best move plus some search stats
// Does iterative deepening until depth or timeout
// If depth param != 0 then we do fixed depth search.
// If targetTimeMs != 0 then we try to limit tame waste by returning early from a full search at some depth when
//   we reckon there is not enough time to do the full next-level search.
// Return best-move, eval, stats, final-depth, error
func Search(board *dragon.Board, ht HistoryTableT, depth int, targetTimeMs int, timeout *uint32) (dragon.Move, EvalCp, SearchStatsT, int, error) {
	var deepKillers [MaxDepth]dragon.Move
	var stats SearchStatsT
	var bestMove = NoMove
	var eval EvalCp = 0
	var staticEval = StaticEval(board)
	var staticNegaEval = staticEval
	if !board.Wtomove {
		staticNegaEval = -staticEval
	}

	// Results from last full search.
	// TODO - it should be possible to get valid results from a partial search but I'm too dumb to work it out at the moment.
	var fullDepth = 0
	var fullBestMove = NoMove
	var fullEval EvalCp = 0
	// TODO our eval is somewhat unstable between odd/even plies, so we smooth this by returning our
	//   final eval as the average of the evals for the last two plies.
	var prevFullEval EvalCp = 0
	
	// Best results from previous depth in case the timeout depth didn't get as far as returning a result

	var maxDepthToGo = MaxDepth
	if depth > 0 {
		maxDepthToGo = depth
	}

	originalStart := time.Now()

	fmt.Println("info string using", SearchAlgorithmString(), "max depth", maxDepthToGo)

	var depthToGo int
	// Iterative deepening
	for depthToGo = MinDepth; depthToGo <= maxDepthToGo; depthToGo++ {
		// Time the search
		start := time.Now()

		switch SearchAlgorithm {
		case MiniMax:
			bestMove, eval = miniMax(board, depthToGo, /*depthFromRoot*/0, staticEval, &stats, timeout)
			
		case NegaMax:
			var negaEval EvalCp
			bestMove, negaEval = negaMax(board, depthToGo, /*depthFromRoot*/0, staticNegaEval, &stats, timeout)
			eval = negaEval
			if !board.Wtomove {
				eval = -negaEval
			}
			
		case AlphaBeta:
			// Use the best move from the previous depth as the killer move for this depth
			bestMove, eval = alphaBeta(board, depthToGo, /*depthFromRoot*/0, BlackCheckMateEval, WhiteCheckMateEval, staticEval, fullBestMove, deepKillers[:], &stats, timeout)
			
		case NegAlphaBeta:
			// Use the best move from the previous depth as the killer move for this depth
			var negaEval EvalCp
			bestMove, negaEval = negAlphaBeta(board, ht, depthToGo, /*depthFromRoot*/0, YourCheckMateEval, MyCheckMateEval, staticNegaEval, fullBestMove, deepKillers[:], &stats, timeout)
			eval = negaEval
			if !board.Wtomove {
				eval = -negaEval
			}
			
		default:
		return NoMove, 0, stats, 0, errors.New("bot: unrecognised search algorithm")
		}

		elapsedSecs := time.Since(start).Seconds()

		// Have we timed out? - if so, then ignore the results for this depth - it's a partial search
		// TODO - it should be possible to get valid results from a partial search but I'm too dumb at the moment :(
		if isTimedOut(timeout) {
			break
		}

		// Reduce the output noise
		if maxDepthToGo <=4 || depthToGo > 4 {
			// UCI wants eval always from white perspective
			evalForWhite := eval
			if !board.Wtomove {
				evalForWhite = -eval
			}
			// Print summary stats for the depth - slightly inaccurate because it includes accumulation of previous depths
			fmt.Println("info depth", depthToGo, "score cp", evalForWhite, "nodes", stats.Nodes, "time", uint64(elapsedSecs*1000), "nps", uint64(float64(stats.Nodes)/elapsedSecs), "pv", &bestMove)
		}

		fullBestMove = bestMove
		prevFullEval = fullEval
		fullEval = eval
		fullDepth = depthToGo

		// Bail early if we don't think we can get another full search level done
		if targetTimeMs > 0 {
			totalElapsedSecs := time.Since(originalStart).Seconds()
			totalElapsedMs := int(totalElapsedSecs*1000)
			cutoffMs := targetTimeMs*SearchCutoffPercent/100
			if totalElapsedMs > cutoffMs {
				break
			}
		}
	}

	// If we didn't get a move at all then barf
	if bestMove == NoMove {
		return NoMove, 0, stats, depthToGo, errors.New("bot: no legal move found in search")
	}

	// We smooth the odd/even instability by using the average eval of the last two depths
	return fullBestMove, (fullEval + prevFullEval)/2, stats, fullDepth, nil
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

// Return the new static eval from white's perspective - either by fast delta, or by full evaluation, depending on configuration
func getStaticEval(board* dragon.Board, oldStaticEval EvalCp, move dragon.Move, moveInfo *dragon.MoveApplication) EvalCp {
	if UseDeltaEval {
		// Much faster
		return oldStaticEval + EvalDelta(move, moveInfo, !board.Wtomove)
	} else {
		// For sanity check :P
		return StaticEval(board)
	}
}

// Return the new static eval from current mover's perspective - either by fast delta, or by full evaluation, depending on configuration
func getNegaStaticEval(board* dragon.Board, oldNegaStaticEval EvalCp, move dragon.Move, moveInfo *dragon.MoveApplication) EvalCp {
	if UseDeltaEval {
		// Much faster
		return oldNegaStaticEval + NegaEvalDelta(move, moveInfo, !board.Wtomove)
	} else {
		// For sanity check :P
		return NegaStaticEval(board)
	}
}

// Return the best eval attainable through minmax from the given
//   position, along with the move leading to the principal variation.
// Eval is given from white's perspective.
func miniMax(board *dragon.Board, depthToGo int, depthFromRoot int, staticEval EvalCp, stats *SearchStatsT, timeout *uint32) (dragon.Move, EvalCp) {

	// Bail if we've timed out
	if isTimedOut(timeout) {
		worstEval := WhiteCheckMateEval
		if board.Wtomove {
			worstEval = BlackCheckMateEval
		}
		// Return the worst possible eval (opponent checkmate) to invalidate this incomplete search branch
		return NoMove, worstEval
	}

	stats.Nodes++
	stats.NonLeafs++

	// Generate all legal moves - thanks dragontoothmg!
	legalMoves := board.GenerateLegalMoves()

	// Check for checkmate or stalemate
	if len(legalMoves) == 0 {
		stats.Mates++
		return NoMove, mateEval(board, depthFromRoot)
	}
	
	isWhiteMove := board.Wtomove
	var bestMove = NoMove
	var bestEval EvalCp = WhiteCheckMateEval
	if isWhiteMove {
		bestEval = BlackCheckMateEval
	}

	for _, move := range legalMoves {
		// Make the move
		moveInfo := board.Apply2(move)

		newStaticEval := getStaticEval(board, staticEval, move, moveInfo)

		// Get the (deep) eval
		var eval EvalCp
		if depthToGo <= 1 {
			stats.Nodes++
			// Ignore mate check to avoid generating moves at all leaf nodes
			eval = newStaticEval
		} else {
			_, eval = miniMax(board, depthToGo-1, depthFromRoot+1, newStaticEval, stats, timeout)
		}

		// Take back the move
		moveInfo.Unapply()

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

// Return the best eval attainable through negamax from the given
//   position, along with the move leading to the principal variation.
// Eval is given from current mover's perspective.
func negaMax(board *dragon.Board, depthToGo int, depthFromRoot int, staticNegaEval EvalCp, stats *SearchStatsT, timeout *uint32) (dragon.Move, EvalCp) {

	// Bail if we've timed out
	if isTimedOut(timeout) {
		// Return the worst possible eval (opponent checkmate) to invalidate this incomplete search branch
		return NoMove, YourCheckMateEval
	}

	stats.Nodes++
	stats.NonLeafs++

	// Generate all legal moves - thanks dragontoothmg!
	legalMoves := board.GenerateLegalMoves()

	// Check for checkmate or stalemate
	if len(legalMoves) == 0 {
		stats.Mates++
		return NoMove, negaMateEval(board, depthFromRoot)
	}
	
	bestMove := NoMove
	bestEval := YourCheckMateEval

	for _, move := range legalMoves {
		// Make the move
		moveInfo := board.Apply2(move)

		newNegaStaticEval := getNegaStaticEval(board, staticNegaEval, move, moveInfo)

		// Get the (deep) eval
		var eval EvalCp
		if depthToGo <= 1 {
			stats.Nodes++
			// Ignore mate check to avoid generating moves at all leaf nodes
			eval = newNegaStaticEval
		} else {
			_, eval = negaMax(board, depthToGo-1, depthFromRoot+1, -newNegaStaticEval, stats, timeout)
			eval = -eval // back to our perspective
		}

		// Take back the move
		moveInfo.Unapply()

		// We try to maximise our eval.
		// Strictly > to match alphabeta
		if eval > bestEval {
			bestEval, bestMove = eval, move
		}
	}

	return bestMove, bestEval
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

// Return the best eval attainable through alpha-beta from the given position (with killer-move hint), along with the move leading to the principal variation.
func alphaBeta(board *dragon.Board, depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, staticEval EvalCp, killer dragon.Move, deepKillers []dragon.Move, stats *SearchStatsT, timeout *uint32) (dragon.Move, EvalCp) {

	// Bail if we've timed out
	if isTimedOut(timeout) {
		worstEval := WhiteCheckMateEval
		if board.Wtomove {
			worstEval = BlackCheckMateEval
		}
		// Return the worst possible eval (opponent checkmate) to invalidate this incomplete search branch
		return NoMove, worstEval
	}

	stats.Nodes++
	stats.NonLeafs++

	// Generate all legal moves - thanks dragontoothmg!
	legalMoves := board.GenerateLegalMoves()

	// Check for checkmate or stalemate
	if len(legalMoves) == 0 {
		stats.Mates++
		return NoMove, mateEval(board, depthFromRoot)
	}

	usingDeepKiller := false
	if(UseKillerMoves) {
		// Place killer-move (or deep killer move) first if it's there
		killer, usingDeepKiller = prioritiseKillerMove(legalMoves, killer, UseDeepKillerMoves, deepKillers[depthFromRoot], &stats.Killers, &stats.DeepKillers)
	}
		
	// Would be smaller with negalpha-beta but this is simple
	if board.Wtomove {
		// White to move - maximise eval with beta cut-off
		var bestMove = NoMove
		var bestEval EvalCp = BlackCheckMateEval
		childKiller := NoMove
		
		for _, move := range legalMoves {
			// Make the move
			moveInfo := board.Apply2(move)
			
			newStaticEval := getStaticEval(board, staticEval, move, moveInfo)
			
			// Get the (deep) eval
			var eval EvalCp
			if depthToGo <= 1 {
				stats.Nodes ++
				if UseQSearch {
					// Quiesce
					childKiller, eval = qsearchAlphaBeta(board, QSearchDepth, depthFromRoot+1, alpha, beta, newStaticEval, childKiller, deepKillers, stats)
				} else {
					eval = newStaticEval
				}
			} else {
				childKiller, eval = alphaBeta(board, depthToGo-1, depthFromRoot+1, alpha, beta, newStaticEval, childKiller, deepKillers, stats, timeout)
			}

			// Take back the move
			moveInfo.Unapply()
			
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
						stats.DeepKillerCuts++
					} else {
						stats.KillerCuts++
					}
				}
				// beta cut-off
 				deepKillers[depthFromRoot] = bestMove
				return bestMove, bestEval
			}
		}
		
		deepKillers[depthFromRoot] = bestMove
		return bestMove, bestEval
	} else {
		// Black to move - minimise eval with alpha cut-off
		var bestMove = NoMove
		var bestEval EvalCp = WhiteCheckMateEval
		childKiller := NoMove
		
		for _, move := range legalMoves {
			// Make the move
			moveInfo := board.Apply2(move)
			
			newStaticEval := getStaticEval(board, staticEval, move, moveInfo)
			
			// Get the (deep) eval
			var eval EvalCp
			if depthToGo <= 1 {
				stats.Nodes++
				if UseQSearch {
					// Quiesce
					childKiller, eval = qsearchAlphaBeta(board, QSearchDepth, depthFromRoot+1, alpha, beta, newStaticEval, childKiller, deepKillers, stats)
				} else {
					eval = newStaticEval
				}
			} else {
				childKiller, eval = alphaBeta(board, depthToGo-1, depthFromRoot+1, alpha, beta, newStaticEval, childKiller, deepKillers, stats, timeout)
			}
			
			// Take back the move
			moveInfo.Unapply()
			
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
						stats.DeepKillerCuts++
					} else {
						stats.KillerCuts++
					}
				}
				// alpha cut-off
				deepKillers[depthFromRoot] = bestMove
				return bestMove, bestEval
			}
		}

		deepKillers[depthFromRoot] = bestMove
		return bestMove, bestEval
	}
}

// Quiescence search - differs from full search as follows:
//   - we only look at captures and promotions - we could/should also possibly look at check evasion, but check detection is currently expensive
//   - we consider 'standing pat' - i.e. do alpha/beta cutoff according to the node's static eval (TODO)
// TODO - implement more efficient generation of 'noisy' moves in dragontoothmg
// TODO - better static eval if we bottom out without quescing, e.g. static exchange evaluation (SEE)
func qsearchAlphaBeta(board *dragon.Board, qDepthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, staticEval EvalCp, killer dragon.Move, deepKillers []dragon.Move, stats *SearchStatsT) (dragon.Move, EvalCp) {

	stats.QNodes++
	stats.QNonLeafs++

	// Stand pat - equivalent to considering the null move as a valid move.
	// Essentially the player to move doesn't _have_ to make a capture - (we assume that there is a non-capture move available.)
	if board.Wtomove {
		if alpha < staticEval {
			alpha = staticEval
		}

		// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
		if alpha >= beta {
			stats.QPats++
			stats.QPatCuts++
			return NoMove, staticEval
		}

	} else {
		if staticEval < beta {
			beta = staticEval
		}
		
		// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
		if alpha >= beta {
			stats.QPats++
			stats.QPatCuts++
			return NoMove, staticEval
		}
	}

	// Generate all noisy legal moves - thanks dragontoothmg!
	legalMoves, isInCheck := board.GenerateLegalMoves2(/*onlyCapturesPromosCheckEvasion*/true)

	// No noisy mvoes
	if len(legalMoves) == 0 {
		// Check for checkmate or stalemate
		if isInCheck {
			stats.QMates++
			return NoMove, mateEval(board, depthFromRoot) // TODO checks for mate again expensively
		} else {
			// Already quiesced - just return static eval
			stats.QQuiesced++
			return NoMove, staticEval
		}
	}

	usingDeepKiller := false
	if(UseQKillerMoves) {
		// Place killer-move (or deep killer move) first if it's there
		killer, usingDeepKiller = prioritiseKillerMove(legalMoves, killer, UseQDeepKillerMoves, deepKillers[depthFromRoot], &stats.QKillers, &stats.QDeepKillers)
	}

	// Would be smaller with negalpha-beta but this is simple
	if board.Wtomove {
		// White to move - maximise eval with beta cut-off
		var bestMove = NoMove
		var bestEval EvalCp = BlackCheckMateEval
		childKiller := NoMove
		
		for _, move := range legalMoves {
			// Make the move
			moveInfo := board.Apply2(move)
			
			newStaticEval := getStaticEval(board, staticEval, move, moveInfo)
			
			// Get the (deep) eval
			var eval EvalCp
			if qDepthToGo <= 1 {
				stats.QNodes++
				stats.QPrunes++
				// Ignore mate check to avoid generating moves at all leaf nodes
				eval = newStaticEval
			} else {
				childKiller, eval = qsearchAlphaBeta(board, qDepthToGo-1, depthFromRoot+1, alpha, beta, newStaticEval, childKiller, deepKillers, stats)
			}

			// Take back the move
			moveInfo.Unapply()
			
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
						stats.QDeepKillerCuts++
					} else {
						stats.QKillerCuts++
					}
				}
				// beta cut-off
 				deepKillers[depthFromRoot] = bestMove
				return bestMove, bestEval
			}
		}

		stats.QPats++
		deepKillers[depthFromRoot] = bestMove
		return bestMove, bestEval
	} else {
		// Black to move - minimise eval with alpha cut-off
		var bestMove = NoMove
		var bestEval EvalCp = WhiteCheckMateEval
		childKiller := NoMove
		
		for _, move := range legalMoves {
			// Make the move
			moveInfo := board.Apply2(move)
			
			newStaticEval := getStaticEval(board, staticEval, move, moveInfo)
			
			// Get the (deep) eval
			var eval EvalCp
			if qDepthToGo <= 1 {
				stats.QNodes++
				stats.QPrunes++
				// Ignore mate check to avoid generating moves at all leaf nodes
				eval = newStaticEval
			} else {
				childKiller, eval = qsearchAlphaBeta(board, qDepthToGo-1, depthFromRoot+1, alpha, beta, newStaticEval, childKiller, deepKillers, stats)
			}
			
			// Take back the move
			moveInfo.Unapply()
			
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
						stats.QDeepKillerCuts++
					} else {
						stats.QKillerCuts++
					}
				}
				// alpha cut-off
 				deepKillers[depthFromRoot] = bestMove
				return bestMove, bestEval
			}
		}

		stats.QPats++
		deepKillers[depthFromRoot] = bestMove
		return bestMove, bestEval
	}
}

// TT move is prefered to all others
const ttMoveValue uint8 = 255
// ...then the killer move
const killerValue uint8 = 254
// ...then the second (deep) killer
const killer2Value uint8 = 253

// Indexed by promo piece type - only N, B, R, Q valid
var promoMOValue = [8]uint8 {0, 0, /*N*/105, /*B*/103, /*R*/104, /*Q*/109, 0, 0}

// Indexed by [victim][atacker]
// Basically MVV-LVA with king attacker rated high(!)
// TODO play with king ordering, and bishop-vs-knight ordering
// TODO boost moves that have danger of take-back, e.g. rook takes rook
var captureMOValue = [8][8]uint8 {
	/*mover*/
 	/*Nothing*/ {0,  9,  7,  6,  5,  4,  8, 0}, // non-capture move ordering - TODO???
	/*Pawn*/    {0, 19, 17, 16, 15, 14, 18, 0},
	/*Knight*/  {0, 39, 37, 36, 35, 34, 38, 0},
	/*Bishop*/  {0, 49, 47, 46, 45, 44, 48, 0},
	/*Rook*/    {0, 59, 57, 56, 55, 54, 58, 0},
	/*Queen*/   {0, 99, 97, 96, 95, 94, 98, 0},
	/*King*/    {0, 0, 0, 0, 0, 0, 0, 0},       // invalid king capture
	/*Invalid*/ {0, 0, 0, 0, 0, 0, 0, 0}}

// Sorting interface
type byMoValueDesc struct {
	moves []dragon.Move
	values []uint8
}

func (mo *byMoValueDesc) Len() int {
	return len(mo.moves)
}

func (mo *byMoValueDesc) Swap(i, j int) {
	mo.moves[i], mo.moves[j] = mo.moves[j], mo.moves[i]
	mo.values[i], mo.values[j] = mo.values[j], mo.values[i]
}

// Less is more for us
func (mo *byMoValueDesc) Less(i, j int) bool {
	return mo.values[i] >  mo.values[j]
}

// Order q-search moves heuristically.
// Preference is:
// 1. Promotions by promo type
// 2. MMV-LVA for captures
//     (most valuable victim first, then least-valuable attacker second)
func orderMoves(board *dragon.Board, moves []dragon.Move, ttMove dragon.Move, killer dragon.Move, killer2 dragon.Move, killersStat *uint64, deepKillersStat *uint64) {
	// Value of each move - nothing to do with any other eval, just a local ordering metric
	values := make([]uint8, len(moves))
	for i, move := range moves {
		if move == ttMove {
			values[i] = ttMoveValue
		} else if move == killer {
			*killersStat++
			values[i] = killerValue
		} else if move == killer2 {
			*deepKillersStat++
			values[i] = killer2Value
		} else {
			from, to := move.From(), move.To()
			attacker := board.PieceAt(from)
			// We miss en-passant but it's not worth the effort to do properly
			victim := board.PieceAt(to)
			promoPiece := move.Promote()
			
			values[i] = promoMOValue[promoPiece] + captureMOValue[victim][attacker]
		}
	}

	mo := byMoValueDesc{ moves, values}
	sort.Sort(&mo)
}

// MUST be a power of 2 cos we use & instead of % for fast hash table index
const TTSize = 1024*1024
//const QttSize = 256*1024

// Want this to be per-thread, but for now we're single-threaded so global is ok
var tt []TTEntryT = make([]TTEntryT, TTSize)

func ResetTT() {
	tt = make([]TTEntryT, TTSize)
}

// Return the best eval attainable through alpha-beta from the given position (with killer-move hint), along with the move leading to the principal variation.
func negAlphaBeta(board *dragon.Board, ht HistoryTableT, depthToGo int, depthFromRoot int, alpha EvalCp, beta EvalCp, staticNegaEval EvalCp, killer dragon.Move, deepKillers []dragon.Move, stats *SearchStatsT, timeout *uint32) (dragon.Move, EvalCp) {

	// Bail if we've timed out
	if isTimedOut(timeout) {
		// Return the worst possible eval (opponent checkmate) to invalidate this incomplete search branch
		return NoMove, YourCheckMateEval
	}

	stats.Nodes++
	stats.NonLeafs++
	if depthFromRoot < MaxDepthStats {
		stats.NonLeafsAt[depthFromRoot]++
	}

	// Remember this to check whether our final eval is a lower or upper bound
	origBeta := beta
	origAlpha := alpha

	// Anything after here interacts with the QTT - so single return location at the end of the func after writing back to QTT

	// Probe the Quiescence Transposition Table
	var ttEntry *TTEntryT = nil
	var ttMove = NoMove
	
	if UseTT {
		ttEntry = probeTT(tt, board.Hash(), depthToGo)

		if ttEntry != nil {
			stats.TTHits++

			ttMove = ttEntry.bestMove

			// If the TT hit is for exactly the same depth then use the eval; otherwise we just use the bestMove as a move hint
			// Note that most engines will use the TT eval if the TT is a deeper search; however this requires a 'stable' static eval
			//   and changes behaviour between TT-enabled/disabled. For rigourous testing it's better to be consistent.
			if depthToGo == int(ttEntry.depthToGo) {
				stats.TTDepthHits++
				ttEval := ttEntry.eval
				// If the eval is exact then we're done
				if ttEntry.evalType == TTEvalExact {
					stats.TTTrueEvals++
					return ttMove, ttEntry.eval
				} else {
					var cutoffStats *uint64
					// We can have an alpha or beta cut-off depending on the eval type
					if ttEntry.evalType == TTEvalLowerBound {
						cutoffStats = &stats.TTBetaCuts
						if alpha < ttEval {
							alpha = ttEval
						}
					} else {
						// TTEvalUpperBound
						cutoffStats = &stats.TTAlphaCuts
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

	// Generate all legal moves - thanks dragontoothmg!
	legalMoves := board.GenerateLegalMoves()

	// Check for checkmate or stalemate
	if len(legalMoves) == 0 {
		stats.Mates++
		bestMove, bestEval = NoMove, negaMateEval(board, depthFromRoot)
	} else {

		killerMove := NoMove
		if UseKillerMoves {
			killerMove = killer
		}
		deepKiller := NoMove
		if UseDeepKillerMoves {
			deepKiller = deepKillers[depthFromRoot]
		}

		// Sort the moves heuristically
		if UseQSearchMoveOrdering {
			if len(legalMoves) > 1 {
				orderMoves(board, legalMoves, ttMove, killerMove, deepKiller, &stats.Killers, &stats.DeepKillers)
			}
		} else if UseKillerMoves {
			// Place killer-move (or deep killer move) first if it's there
			prioritiseKillerMove(legalMoves, killer, UseDeepKillerMoves, deepKillers[depthFromRoot], &stats.Killers, &stats.DeepKillers)
		}

		childKiller := NoMove

		for i, move := range legalMoves {
			// Make the move
			moveInfo := board.Apply2(move)
			// Add to the move history
			repetitions := ht.Add(board.Hash())
			
			newNegaStaticEval := getNegaStaticEval(board, staticNegaEval, move, moveInfo)
			
			// Get the (deep) eval
			var eval EvalCp
			// We consider 2-fold repetition to be a draw, since if a repeat can be forced then it can be forced again.
			// This reduces the search tree a bit and is common practice in chess engines.
			if UsePosRepetition && repetitions > 1 {
				stats.PosRepetitions++
				eval = DrawEval
			} else if depthToGo <= 1 {
				stats.Nodes ++
				if UseQSearch {
					// Quiesce
					childKiller, eval, _ = qsearchNegAlphaBeta(board, QSearchDepth, depthFromRoot+1, /*depthFromQRoot*/0, -beta, -alpha, -newNegaStaticEval, childKiller, deepKillers, stats)
					eval = -eval // back to our perspective
				} else {
					eval = newNegaStaticEval
				}
			} else {
				childKiller, eval = negAlphaBeta(board, ht, depthToGo-1, depthFromRoot+1, -beta, -alpha, -newNegaStaticEval, childKiller, deepKillers, stats, timeout)
				eval = -eval // back to our perspective
			}
			
			// Remove from the move history
			ht.Remove(board.Hash())
			// Take back the move
			moveInfo.Unapply()
		
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
					stats.TTLateCuts++
				} else if bestMove == killerMove {
					stats.KillerCuts++
				} else if bestMove == deepKiller {
					stats.DeepKillerCuts++
				}
				if i == 0 {
					stats.FirstChildCuts++
					if depthFromRoot < MaxDepthStats {
						stats.FirstChildCutsAt[depthFromRoot]++
					}
				}
				break
			}
		}

		// If we didn't get a beta cut-off then we visited all children.
		if bestEval <= origBeta {
			stats.AllChildrenNodes++
		}
		deepKillers[depthFromRoot] = bestMove
	}
	
	// Update the TT
	if UseTT {
		evalType := TTEvalExact
		if origBeta <= bestEval {
			evalType = TTEvalLowerBound
		} else if bestEval <= origAlpha {
			evalType = TTEvalUpperBound
		}
		if ttEntry == nil {
			// Write a new TT entry
			writeTTEntry(tt, board.Hash(), bestEval, bestMove, depthToGo, evalType)
		} else {
			// Update the existing QTT entry
			updateTTEntry(ttEntry, bestEval, bestMove, depthToGo, evalType)
		}
	}
	
	return bestMove, bestEval
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
					break;
				}
			}
		}
	}

	return nMovesToUse
}

// MUST be a power of 2 cos we use & instead of % for fast hash table index
const QttSize = 64*1024
//const QttSize = 256*1024

// Want this to be per-thread, but for now we're single-threaded so global is ok
var qtt []QSearchTTEntryT = make([]QSearchTTEntryT, QttSize)

func ResetQtt() {
	qtt = make([]QSearchTTEntryT, QttSize)
}

// Quiescence search - differs from full search as follows:
//   - we only look at captures and promotions - we could/should also possibly look at check evasion, but check detection is currently expensive
//   - we consider 'standing pat' - i.e. do alpha/beta cutoff according to the node's static eval (TODO)
// Return best-move, best-eval, isQuiesced
// TODO - better static eval if we bottom out without quiescing, e.g. static exchange evaluation (SEE)
// TODO - include moving away from attacks too?
func qsearchNegAlphaBeta(board *dragon.Board, qDepthToGo int, depthFromRoot int, depthFromQRoot int, alpha EvalCp, beta EvalCp, staticNegaEval EvalCp, killer dragon.Move, deepKillers []dragon.Move, stats *SearchStatsT) (dragon.Move, EvalCp, bool) {

	stats.QNodes++
	stats.QNonLeafs++
	if depthFromQRoot < MaxQDepthStats {
		stats.QNonLeafsAt[depthFromQRoot]++
	}

	// Remember this to check whether our final eval is a lower or upper bound
	origBeta := beta
	origAlpha := alpha

	// Stand pat - equivalent to considering the null move as a valid move.
	// Essentially the player to move doesn't _have_ to make a 'noisy' move - assuming that there is a quiet move available.
	if alpha < staticNegaEval {
		alpha = staticNegaEval
	}

	// Note that this is aggressive, and we fail-soft AT the parent's best eval - be very ware!
	if alpha >= beta {
		stats.QPats++
		stats.QPatCuts++

		return NoMove, staticNegaEval, false // TODO - not sure what to return here for isQuiesced - this is playing safe
	}

	// Anything after here interacts with the QTT - so single return location at the end of the func after writing back to QTT

	// Probe the Quiescence Transposition Table
	var qttEntry *QSearchTTEntryT = nil
	var qttMove = NoMove
	
	if UseQSearchTT {
		qttEntry = probeQtt(qtt, board.Hash(), qDepthToGo)

		if qttEntry != nil {
			stats.QttHits++

			qttMove = qttEntry.bestMove

			// If the QTT hit is for exactly the same depth then use the eval; otherwise we just use the bestMove as a move hint
			// Note that most engines will use the TT eval if the TT entry is a deeper search; however this requires a 'stable' static eval
			//   and changes behaviour between TT-enabled/disabled. For rigourous testing it's better to be consistent.
			isExactHit := qDepthToGo == int(qttEntry.qDepthToGo) ||
				//    ... or if this is a fully quiesced result
				qDepthToGo > int(qttEntry.qDepthToGo) && qttEntry.isQuiesced
			
			if isExactHit {
				stats.QttDepthHits++
				qttEval := qttEntry.eval
				// If the eval is exact then we're done
				if qttEntry.evalType == TTEvalExact {
					stats.QttTrueEvals++
					return qttMove, qttEntry.eval, qttEntry.isQuiesced
				} else  {
					var cutoffStats *uint64
					// We can have an alpha or beta cut-off depending on the eval type
					if qttEntry.evalType == TTEvalLowerBound {
						cutoffStats = &stats.QttBetaCuts
						if alpha < qttEval {
							alpha = qttEval
						}
					} else {
						// TTEvalUpperBound
						cutoffStats = &stats.QttAlphaCuts
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
	bestEval := YourCheckMateEval

	// Did we reach quiescence at all leaves?
	isQuiesced := false
	
	// Generate all noisy legal moves
	legalMoves, isInCheck := board.GenerateLegalMoves2(/*onlyCapturesPromosCheckEvasion*/true)
	
	if len(legalMoves) == 0 {
		// No noisy moves - checkmate or stalemate or just quiesced
		isQuiesced = true
		if isInCheck {
			stats.QMates++
			bestMove, bestEval = NoMove, negaMateEval(board, depthFromRoot) // TODO checks for mate again expensively
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
			deepKiller = deepKillers[depthFromRoot]
		}
		
		// Sort the moves heuristically
		if UseQSearchMoveOrdering {
			if len(legalMoves) > 1 {
				orderMoves(board, legalMoves, qttMove, killerMove, deepKiller, &stats.QKillers, &stats.QDeepKillers)
				if UseQSearchRampagePruning {
					nMovesToUse = pruneQueenRampages(board, legalMoves, depthFromQRoot, stats)
				}
			}
		} else if UseQKillerMoves {
			// Place killer-move (or deep killer move) first if it's there
			prioritiseKillerMove(legalMoves, killer, UseQDeepKillerMoves, deepKillers[depthFromRoot], &stats.QKillers, &stats.QDeepKillers)
		}
		
		// We're quiesced as long as all children (we visit) are quiesced.
		isQuiesced := true

		childKiller := NoMove
		
		for i := 0; i < nMovesToUse; i++ {
			move := legalMoves[i]

			// Get the (deep) eval
			var eval EvalCp
			newNegaStaticEval := staticNegaEval + NegaFastEvalDelta(board, move)

			if qDepthToGo <= 1 {
				stats.QNodes++
				stats.QPrunes++
				// We hit max depth before quiescing
				isQuiesced = false
				// Ignore mate check to avoid generating moves at all leaf nodes
				eval = newNegaStaticEval
			} else {
				// Make the move
				moveInfo := board.Apply2(move)
				
				var isChildQuiesced bool
				childKiller, eval, isChildQuiesced = qsearchNegAlphaBeta(board, qDepthToGo-1, depthFromRoot+1, depthFromQRoot+1, -beta, -alpha, -newNegaStaticEval, childKiller, deepKillers, stats)
				eval = -eval // back to our perspective
				isQuiesced = isQuiesced && isChildQuiesced
			
				// Take back the move
				moveInfo.Unapply()
			}
			
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
					stats.QttLateCuts++
				} else if bestMove == killerMove {
					stats.QKillerCuts++
				} else if bestMove == deepKiller {
					stats.QDeepKillerCuts++
				}
				if i == 0 {
					stats.QFirstChildCuts++
				}
				break
			}
		}

		// Was stand-pat the best after all?
		if bestEval == staticNegaEval {
			stats.QPats++
		}

		// If we didn't get a beta cut-off then we visited all children.
		if bestEval <= origBeta {
			stats.QAllChildrenNodes++
		}
		deepKillers[depthFromRoot] = bestMove
		
	}

	if isQuiesced {
		stats.QQuiesced++
	}

	// Update the QTT
	if UseQSearchTT {
		evalType := TTEvalExact
		if origBeta <= bestEval {
			evalType = TTEvalLowerBound
		} else if bestEval <= origAlpha {
			evalType = TTEvalUpperBound
		}
		if qttEntry == nil {
			// Write a new QTT entry
			writeQttEntry(qtt, board.Hash(), bestEval, bestMove, qDepthToGo, evalType, isQuiesced)
		} else {
			// Update the existing QTT entry
			updateQttEntry(qttEntry, bestEval, bestMove, qDepthToGo, evalType, isQuiesced)
		}
	}
	
	return bestMove, bestEval, isQuiesced
}
