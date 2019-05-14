package engine

import (
	// "fmt"
	"math"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Eval in centi-pawns, i.e. 100 === 1 pawn
type EvalCp int16

const WhiteCheckMateEval EvalCp = math.MaxInt16
const BlackCheckMateEval EvalCp = -math.MaxInt16 // don't use MinInt16 cos it's not symmetrical with MaxInt16

// For NegaMax and friends this naming is more accurate
const MyCheckMateEval EvalCp = math.MaxInt16
const YourCheckMateEval EvalCp = -math.MaxInt16 // don't use MinInt16 cos it's not symmetrical with MaxInt16

// Checkmate at depth N is represented by the eval N away from absolute checkmate val
func isCheckmateEval(eval EvalCp) bool {
	return eval <= YourCheckMateEval + MaxDepth ||
		eval >= MyCheckMateEval - MaxDepth
}

// Used to mark transposition (hash) tables entries as invalid
const InvalidEval EvalCp = math.MinInt16

const DrawEval EvalCp = 0

// Different evals
// Configuration options
type EvalAlgorithmT int

const (
	PiecePosEval EvalAlgorithmT = iota
	PositionalEval
)

// Absolute bound on expensive eval - so we can do cheap futility pruning on the (cheaper) O(0) eval
// TODO - work out what sensible clamp bounds are.
const MaxAbsStaticEvalOrderN = EvalCp(500)

// Static eval only - no mate checks - from the perspective of the player to move
func NegaStaticEval(board *dragon.Board) EvalCp {
	staticEval := StaticEval(board)

	if board.Colortomove == dragon.White {
		return staticEval
	}
	return -staticEval
}

// Static eval only - no mate checks - from the perspective of the player to move
// Using precomputed O(0) component.
func NegaStaticEvalFast(board *dragon.Board, negaEval0 EvalCp) EvalCp {
	staticEvalOrderN := StaticEvalOrderN(board)

	if board.Colortomove != dragon.White {
		staticEvalOrderN = -staticEvalOrderN
	}
	
	return negaEval0 + staticEvalOrderN
}

// Static eval only - no mate checks - from white's perspective
func StaticEval(board *dragon.Board) EvalCp {
	return StaticEvalOrder0(board) + StaticEvalOrderN(board)
}

// Cheap part  - O(0) by delta eval - of static eval from the perspective of the player to move
func NegaStaticEvalOrder0(board *dragon.Board) EvalCp {
	staticEval0 := StaticEvalOrder0(board)

	if board.Colortomove == dragon.White {
		return staticEval0
	}
	return -staticEval0
}
	
// Cheap part  - O(0) by delta eval - of static eval from white's perspective.
// This is full evaluation - we prefer to do much cheaper delta evaluation.
func StaticEvalOrder0(board *dragon.Board) EvalCp {
	if EvalAlgorithm == PiecePosEval {
		return StaticPiecePosEvalOrder0(board)
	} else if EvalAlgorithm == PositionalEval {
		return StaticPositionalEvalOrder0(board)
	} else {
		return DrawEval
	}
}

// Cheap part of static eval by opportunistic delta eval.
// Doing the easy case first and falling back to full eval until someone's more keen
func NegaStaticEvalOrder0Fast(board *dragon.Board, prevEval0 EvalCp, moveInfo *dragon.BoardSaveT) EvalCp {
	if EvalAlgorithm == PiecePosEval {
		return NegaStaticPiecePosEvalOrder0Fast(board, prevEval0, moveInfo)
	} else if EvalAlgorithm == PositionalEval {
		return NegaStaticPositionalEvalOrder0Fast(board, prevEval0, moveInfo)
	} else {
		return DrawEval
	}
}

// Expensive part - O(n) even with delta eval - of static eval from white's perspective.
func StaticEvalOrderN(board *dragon.Board) EvalCp {
	if EvalAlgorithm == PiecePosEval {
		return StaticPiecePosEvalOrderN(board)
	} else if EvalAlgorithm == PositionalEval {
		return StaticPositionalEvalOrderN(board)
	} else {
		return DrawEval
	}
}

