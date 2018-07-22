package engine

import (
	"errors"
	"fmt"
	"math/bits"
	"sort"
	"sync/atomic"
	"time"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

// MUST be a power of 2 cos we use & instead of % for fast hash table index
const TTSize = 1024 * 1024

//const QttSize = 256*1024

// Want this to be per-thread, but for now we're single-threaded so global is ok
var tt []TTEntryT = make([]TTEntryT, TTSize)

func ResetTT() {
	tt = make([]TTEntryT, TTSize)
}

// MUST be a power of 2 cos we use & instead of % for fast hash table index
const QttSize = 64 * 1024

//const QttSize = 256*1024

// Want this to be per-thread, but for now we're single-threaded so global is ok
var qtt []QSearchTTEntryT = make([]QSearchTTEntryT, QttSize)

func ResetQtt() {
	qtt = make([]QSearchTTEntryT, QttSize)
}

// Search tree encapsulation
type SearchT struct {
	board       *dragon.Board
	ht          HistoryTableT
	deepKillers []dragon.Move
	stats       *SearchStatsT
	timeout     *uint32
}

func NewSearchT(board *dragon.Board, ht HistoryTableT, deepKillers []dragon.Move, stats *SearchStatsT, timeout *uint32) *SearchT {
	return &SearchT{
		board:       board,
		ht:          ht,
		deepKillers: deepKillers,
		stats:       stats,
		timeout:     timeout,
	}
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

	// Results from last full search or last valid partial search.
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

	s := NewSearchT(board, ht, deepKillers[:], &stats, timeout)

	var depthToGo int
	// Iterative deepening
	for depthToGo = MinDepth; depthToGo <= maxDepthToGo; depthToGo++ {
		// Time the search
		start := time.Now()

		switch SearchAlgorithm {
		case NegAlphaBeta:
			// Use the best move from the previous depth as the killer move for this depth
			var negaEval EvalCp
			bestMove, negaEval = s.NegAlphaBeta(depthToGo /*depthFromRoot*/, 0, YourCheckMateEval, MyCheckMateEval, fullBestMove, false)
			eval = negaEval
			if !board.Wtomove {
				eval = -negaEval
			}

		default:
			return NoMove, 0, stats, 0, errors.New("bot: unrecognised search algorithm")
		}

		elapsedSecs := time.Since(start).Seconds()

		// Reduce the output noise
		if maxDepthToGo <= 4 || depthToGo > 0 && bestMove != NoMove {
			// UCI wants eval always from white perspective
			evalForWhite := eval
			if !board.Wtomove {
				evalForWhite = -eval
			}
			// Print summary stats for the depth - slightly inaccurate because it includes accumulation of previous depths
			fmt.Println("info depth", depthToGo, "score cp", evalForWhite, "nodes", stats.Nodes, "time", uint64(elapsedSecs*1000), "nps", uint64(float64(stats.Nodes)/elapsedSecs), "pv", &bestMove)
		}

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
		prevFullEval = fullEval
		fullEval = eval
		fullDepth = depthToGo

		// Then always bail on time-out
		if isTimedOut(timeout) {
			break
		}

		// Bail early if we don't think we can get another full search level done
		if targetTimeMs > 0 {
			totalElapsedSecs := time.Since(originalStart).Seconds()
			totalElapsedMs := int(totalElapsedSecs * 1000)
			cutoffMs := targetTimeMs * SearchCutoffPercent / 100
			if totalElapsedMs > cutoffMs {
				break
			}
		}
	}

	// If we didn't get a move at all then barf
	if fullBestMove == NoMove {
		return NoMove, 0, stats, fullDepth, errors.New("bot: no legal move found in search")
	}

	// We smooth the odd/even instability by using the average eval of the last two depths
	return fullBestMove, (fullEval + prevFullEval) / 2, stats, fullDepth, nil
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

	mo := byMoValueDesc{moves, values}
	sort.Sort(&mo)
}

// TODO actually make fast(er)
func isInCheckFast(board *dragon.Board) bool {
	return board.OurKingInCheck()
}

// true iff the given (typically 'killer') move is definitely valid, i.e.
//   1. Piece is there (not captured) and it's a valid move direction for the piece
//   2. Piece is not blocked by another piece
//   3. Move does not discover check
// It's ok to return a false negative (i.e. return false for a valid move), but we return true only
//  if the move is definitely valid.
func isValidMoveFast(board *dragon.Board, ourMove dragon.Move) bool {
	debug := false
	if board.ToFen() == "rnbqkbnr/pppppppp/8/8/Q1P5/8/PP1PPPPP/RNB1KBNR b KQkq - 1 1" && ourMove.String() == "d7d6" {
		fmt.Println("                   debuging isValidMoveFast")
		debug = true
	}
	
	ourFrom := ourMove.From()
	ourFromBit := uint64(1) << ourFrom

	yourAll, ourAll, ourKings := board.White.All, board.Black.All, board.Black.Kings
	if board.Wtomove {
		yourAll, ourAll, ourKings = board.Black.All, board.White.All, board.White.Kings
	}

	// If it's a king move then bail - too complicated to compute check after the move, castling through check, etc.
	if (ourKings & ourFromBit) != 0 {
		return false // TODO can we try harder?
	}
	
	// Are we moving our own piece?
	if (ourAll & ourFromBit) == 0 {
		return false
	}

	ourTo := ourMove.To()
	ourToBit := uint64(1) << ourTo

	// Get the move type - we assume that the move is not complete nonsense
	ourDirDist := dirDist(ourFrom, ourTo)
	ourDirFlag := dirFlag(ourDirDist.dir)

	bothAll := board.White.All | board.Black.All
	ourPiece := board.PieceAt(ourFrom)

	// Is this a valid move path for the piece - valid direction and not blocked
	if ourPiece == dragon.Pawn {
		ourPawnPushFlags, ourPawnCaptureFlags := blackPawnPushFlags, blackPawnCaptureFlags
		if board.Wtomove {
			ourPawnPushFlags, ourPawnCaptureFlags = whitePawnPushFlags, whitePawnCaptureFlags
		}
		// Pawns are special cos their direction has to match whether they are capturing or not.
		// TODO - must do this better/properly with forward directions only for each colour
		if ourDirFlag & ourPawnPushFlags != 0 {
			// Pawn push - target square must be open
			// TODO 2-square pawn push
			if ourDirDist.dist != 1 || (bothAll & ourToBit) != 0 {
				return false
			}
		} else if ourDirFlag & ourPawnCaptureFlags != 0 {
			// Capture - target square must be opposition piece
			// TODO en-passant
			if ourDirDist.dist != 1 || (yourAll & ourToBit) == 0 {
				return false
			}
		} else {
			// Invalid pawn move
			return false
		}
	} else {
		// Non-pawn - is this a valid direction for the move type
		if (PieceToDirFlags[ourPiece] & ourDirFlag) == 0 {
			return false
		}
		// Target square must be open or opposition piece
		if (ourToBit & ourAll) != 0 {
			return false
		}
		// Is there another piece blocking the path - only need to check multi-square slider moves
		if ourDirDist.dist != 1 {
			if isSliderPathBlocked(ourFrom, ourDirDist.dir, ourDirDist.dist, bothAll, false) {
				return false
			}
		}
	}

	// If we get here then the move itself looks good, so check for discovered check

	kingPos := uint8(bits.TrailingZeros64(ourKings))
	kingDirDist := dirDist(kingPos, ourFrom)
	kingDirFlag := dirFlag(kingDirDist.dir)

	if debug {
		fmt.Println("                   king pos", kingPos, "ourFrom", ourFrom, "king-dir", kingDirDist.dir, "king-dist", kingDirDist.dist)
	}

	// If king direction from piece starting position is not a slider direction
	//   then there is no opportunity for discovered check
	if kingDirFlag & queenDirFlags == 0 {
		if debug {
			fmt.Println("                   king dir not slider dir")
		}
		return true
	}

	// If the move is towards or away from the king then there is no opportunity for discovered check,
	//   because the piece itself is still blocking all possible discoveries.
	if kingDirFlag & (ourDirFlag | dirFlag(oppositeDir[ourDirDist.dir])) != 0 {
		if debug {
			fmt.Println("                   piece moved along discovery ray")
		}
		return true
	}

	// If there is another piece blocking the path to the king then we're good
	if isSliderPathBlocked(kingPos, kingDirDist.dir, kingDirDist.dist, bothAll, debug) {
		if debug {
			fmt.Println("                   blocked king-side of moved piece")
		}
		return true
	}

	// TODO - ok, we are exposed to discovered check, but is there an opponent piece taking advantage?
	// At the moment I don't have a good way to determine this other than brute force.
	// Need to generate bits for the full direction away from the king and determine the closest
	//   opponent piece (which is either the highest or lowest bit depending on direction)
	//   and then check whether that piece attacks along the same direction.
	return false
}

// Only for sanity checking
func isValidMoveSlow(board *dragon.Board, ourMove dragon.Move) bool {
	legalMoves, _ := board.GenerateLegalMoves2(false /*all moves*/)
	for _, move := range legalMoves {
		if move == ourMove {
			return true
		}
	}
	return false
}
