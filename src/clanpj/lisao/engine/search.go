package engine

import (
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

// MUST be a power of 2 cos we use & instead of % for fast hash table index
const TTSize = 2 * 1024 * 1024
// Main TT
var tt TtT = make([]TTEntryT, TTSize)

// MUST be a power of 2 cos we use & instead of % for fast hash table index
const TT2Size = 256 * 1024
// Secondary TT for deep nodes (far from leaves)
var tt2 TtT = make([]TTEntryT, TT2Size)

func ResetTT() {
	tt = make([]TTEntryT, TTSize)
	tt2 = make([]TTEntryT, TT2Size)
}

// MUST be a power of 2 cos we use & instead of % for fast hash table index
const QttSize = 256 * 1024

// Want this to be per-thread, but for now we're single-threaded so global is ok
var qtt []QSearchTTEntryT = make([]QSearchTTEntryT, QttSize)

func ResetQtt() {
	qtt = make([]QSearchTTEntryT, QttSize)
}

// The maximum difference between odd and even depth evals - a bit less than 50 at starting position which is almost certainly the worst case.
const MaxOddEvenEvalDiff = EvalCp(50)

// Search tree encapsulation
type SearchT struct {
	board       *dragon.Board
	ht          HistoryTableT
	deepKillers []dragon.Move
	evalByDepth []EvalCp
	stats       *SearchStatsT
	timeout     *uint32
	depth       int        // max search tree depth
	oddEvenEvalDiff EvalCp // An estimate of the absolute eval difference between even and odd depths - used for eval estimates in null-move heuristic, for example
}

func (s *SearchT) deepTtMinDepth() int {
	return s.depth/2
}

func NewSearchT(board *dragon.Board, ht HistoryTableT, deepKillers []dragon.Move, evalByDepth []EvalCp, stats *SearchStatsT, timeout *uint32, oddEvenEvalDiff EvalCp) *SearchT {
	return &SearchT{
		board:       board,
		ht:          ht,
		deepKillers: deepKillers,
		evalByDepth: evalByDepth,
		stats:       stats,
		timeout:     timeout,
		oddEvenEvalDiff: oddEvenEvalDiff,
	}
}

// Estimate an eval for the opposite depth parity
func (s *SearchT) paritySwapEval(eval EvalCp, depth int) EvalCp {
	return s.depthSwitchEval(eval, depth, depth-1)
}

// Make an attempt to translate an eval from one depth to another
func (s *SearchT) depthSwitchEval(eval EvalCp, fromDepth int, toDepth int) EvalCp {
	// Mate is mate, mate
	if eval < YourCheckMateEval + MaxDepth || MyCheckMateEval - MaxDepth < eval {
		return eval
	}
	// Otherwise use the depth eval difference for the root position to adjust to the previous depth
	return eval //+ (s.evalByDepth[toDepth] - s.evalByDepth[fromDepth])
}

// Construct a pv string from the pv line, where available, defaulting to just the bet move otherwise
func MkPvString(bestMove dragon.Move, pvLine []dragon.Move) string {
	pv := ""
	if pvLine[0] == NoMove {
		pv = bestMove.String()
	} else {
		pv = pvLine[0].String()
		for _, move := range pvLine[1:] {
			if move == NoMove {
				break
			}
			pv += " " + move.String()
		}
	}
	return pv
}

func absEvalCp(eval EvalCp) EvalCp {
	if EvalCp(0) <= eval {
		return eval
	} else {
		return -eval
	}
}

// Return eval from white's perspective, and the best move plus some search stats
// Does iterative deepening until depth or timeout
// If depth param != 0 then we do fixed depth search.
// If targetTimeMs != 0 then we try to limit tame waste by returning early from a full search at some depth when
//   we reckon there is not enough time to do the full next-level search.
// Return best-move, eval, stats, final-depth, pv, error
func Search(board *dragon.Board, ht HistoryTableT, depth int, targetTimeMs int, timeout *uint32) (dragon.Move, EvalCp, SearchStatsT, int, []dragon.Move, error) {
	var deepKillers [MaxDepth]dragon.Move
	var evalByDepth [MaxDepth]EvalCp
	var stats SearchStatsT
	var bestMove = NoMove
	var eval EvalCp = 0

	// Results from last full search or last valid partial search.
	var fullDepth = 0
	var fullBestMove = NoMove
	var fullPvLine []dragon.Move
	var fullEval EvalCp = 0
	// TODO our eval is somewhat unstable between odd/even plies, so we smooth this by returning our
	//   final eval as the average of the evals for the last two plies.
	var prevFullEval EvalCp = 0

	const minNEvalDiffs = 4 // the minimum number of eval diffs we need to override the default even/odd difference with empirical values
	var nEvalDiffs      int // the number of eval diffs we have seen, between adjacent depth levels
	var sumEvalAbsDiffs EvalCp // the sum of absolute diffs of adjacent depth level evals

	var maxDepthToGo = MaxDepth
	if depth > 0 {
		maxDepthToGo = depth
	}

	originalStart := time.Now()

	fmt.Println("info string using", SearchAlgorithmString(), "max depth", maxDepthToGo)

	s := NewSearchT(board, ht, deepKillers[:], evalByDepth[:], &stats, timeout, 50)

	// previous and previous previous depth timings
	// prevElapsedSecs, pprevElapsedSecs := float64(0), float64(0)
	
	var depthToGo int
	// Iterative deepening - note we do need to go depth by depth cos we use previous depth evals to fudge even/odd parity TT evals
	for depthToGo = MinDepth; depthToGo <= maxDepthToGo; depthToGo++ {
		s.depth = depthToGo
		// pvLine (where supported) - +1 cos [0] is unused, +1 more for tt NoMove crop
		pvLine := make([]dragon.Move, depthToGo+2)
		
		// Time the search
		start := time.Now()

		switch SearchAlgorithm {
		case NegAlphaBeta:
			eval0 := NegaStaticEvalOrder0(board)
			// Use the best move from the previous depth as the killer move for this depth
			var negaEval EvalCp
			bestMove, negaEval = s.NegAlphaBeta(depthToGo /*depthFromRoot*/, 0, YourCheckMateEval, MyCheckMateEval, fullBestMove, false, eval0, pvLine)
			eval = negaEval
			if board.Colortomove == dragon.Black {
				eval = -negaEval
			}

		default:
			return NoMove, 0, stats, 0, fullPvLine, errors.New("bot: unrecognised search algorithm")
		}

		elapsedSecs := time.Since(start).Seconds()

		// Reduce the output noise
		if maxDepthToGo <= 4 || depthToGo > 0 && bestMove != NoMove {
			// UCI wants eval always from white perspective
			evalForWhite := eval
			if board.Colortomove == dragon.Black {
				evalForWhite = -eval
			}
			// Print summary stats for the depth - slightly inaccurate because it includes accumulation of previous depths
			fmt.Println("info depth", depthToGo, "score cp", evalForWhite, "nodes", stats.Nodes, "time", uint64(elapsedSecs*1000), "nps", uint64(float64(stats.Nodes)/elapsedSecs), "pv", MkPvString(bestMove, pvLine[1:]))
			// if prevElapsedSecs != 0.0 {
			// 	fmt.Printf("info timing ratios d-1 %.3f", elapsedSecs/prevElapsedSecs)
			// 	if pprevElapsedSecs != 0.0 {
			// 		fmt.Printf(" d-2 %.3f", elapsedSecs/pprevElapsedSecs)
			// 	}
			// }
			// fmt.Println()
		}

		// pprevElapsedSecs = prevElapsedSecs
		// prevElapsedSecs = elapsedSecs

		// Have we timed out? If so, then ignore the results for this depth unless we got a valid partial result
		if isTimedOut(timeout) {
			fmt.Println("info string timed out in search for depth", depthToGo)
			if bestMove == NoMove {
				fmt.Println("info string no useful result before time-out at depth", depthToGo)
				break
			} else if SearchAlgorithm != NegAlphaBeta || !UseKillerMoves {
				// Only NegAlphaBeta supports a valid partial result and only if UseKillerMoves is enabled
				fmt.Println("info string ignoring partial search result - only supported for NegAlphaBeta with UseKillerMoves enabled", depthToGo)
				break
			}
		}

		fullBestMove = bestMove
		fullPvLine = pvLine[1:]
		prevFullEval = fullEval
		fullEval = eval
		fullDepth = depthToGo

		// Update even/odd eval diff stats
		if MinDepth < depthToGo {
			nEvalDiffs++
			sumEvalAbsDiffs += absEvalCp(fullEval - prevFullEval)
			if nEvalDiffs >= minNEvalDiffs {
				// Even/odd difference tends to shrink with increasing depth, so this is conservative
				s.oddEvenEvalDiff = (sumEvalAbsDiffs + EvalCp(nEvalDiffs-1))/EvalCp(nEvalDiffs)
			}
		}

		// Then always bail on time-out
		if isTimedOut(timeout) {
			break
		}

		evalByDepth[depthToGo] = eval

		// Bail early if we don't think we can get another full search level done
		if targetTimeMs > 0 {
			totalElapsedSecs := time.Since(originalStart).Seconds()
			totalElapsedMs := int(totalElapsedSecs * 1000)
			cutoffMs := targetTimeMs * SearchCutoffPercent / 100
			if totalElapsedMs > cutoffMs {
				break
			}
		}
		// copy the PV into the deep killers
		for i, move := range pvLine[1:] {
			if move == NoMove {
				break
			}
			deepKillers[i] = move
		}
	}

	// If we didn't get a move at all then barf
	if fullBestMove == NoMove {
		return NoMove, 0, stats, fullDepth, fullPvLine, errors.New("bot: no legal move found in search")
	}

	return fullBestMove, fullEval, stats, fullDepth, fullPvLine, nil
}

func isTimedOut(timeout *uint32) bool {
	return atomic.LoadUint32(timeout) != 0
}

// TT move is prefered to all others
const ttMoveValue uint8 = 255

// ...then the killer move
const killerValue uint8 = 254

// ...then the second (deep) killer
const killer2Value uint8 = 253

// Indexed by promo piece type - only N, B, R, Q valid
var promoMOValue = [8]uint8{0, 0 /*N*/, 105 /*B*/, 103 /*R*/, 104 /*Q*/, 109, 0, 0}

// Indexed by [victim][atacker]
// Basically MVV-LVA with king attacker rated high(!)
// TODO play with king ordering, and bishop-vs-knight ordering
// TODO boost moves that have danger of take-back, e.g. rook takes rook
var captureMOValue = [8][8]uint8{
	/*mover*/
	/*Nothing*/ {0, 9, 7, 6, 5, 4, 8, 0}, // non-capture move ordering - TODO???
	/*Pawn*/ {0, 19, 17, 16, 15, 14, 18, 0},
	/*Knight*/ {0, 39, 37, 36, 35, 34, 38, 0},
	/*Bishop*/ {0, 49, 47, 46, 45, 44, 48, 0},
	/*Rook*/ {0, 59, 57, 56, 55, 54, 58, 0},
	/*Queen*/ {0, 99, 97, 96, 95, 94, 98, 0},
	/*King*/ {0, 0, 0, 0, 0, 0, 0, 0}, // invalid king capture
	/*Invalid*/ {0, 0, 0, 0, 0, 0, 0, 0}}

// Sorting interface
type byMoValueDesc struct {
	moves  []dragon.Move
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
	return mo.values[i] > mo.values[j]
}

func mvvLvaEvalMoves(board *dragon.Board, moves []dragon.Move, values []uint8, ttMove dragon.Move, killer dragon.Move, killer2 dragon.Move, killersStat *uint64, deepKillersStat *uint64) {
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
}

// Order q-search moves heuristically.
// TODO - Spending a lot of time here - maybe choose the best move (only) first time round,
//          then do a full sort only if the first move doesn't cut.
// Preference is:
// 1. Promotions by promo type
// 2. MMV-LVA for captures
//     (most valuable victim first, then least-valuable attacker second)
func orderMoves(board *dragon.Board, moves []dragon.Move, ttMove dragon.Move, killer dragon.Move, killer2 dragon.Move, killersStat *uint64, deepKillersStat *uint64) {
	// Value of each move - nothing to do with any other eval, just a local ordering metric
	values := make([]uint8, len(moves))
	mvvLvaEvalMoves(board, moves, values, ttMove, killer, killer2, killersStat, deepKillersStat)
	mo := byMoValueDesc{moves, values}
	sort.Sort(&mo)
}

// TODO actually make fast(er)
// We can do fast heuristics based on the previous move for example -
//   only the previous move could have placed you in check, either directly or by discovery
func isInCheckFast(board *dragon.Board) bool {
       return board.OurKingInCheck()
}
