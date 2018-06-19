package bot

import (
	"math"
	"math/bits"

	dragon "github.com/dylhunn/dragontoothmg"
)

// Eval in centi-pawns, i.e. 100 === 1 pawn
type EvalCp int16

const WhiteCheckMateEval EvalCp = math.MaxInt16
const BlackCheckMateEval EvalCp = -math.MaxInt16 // don't use MinInt16 cos it's not symmetrical with MaxInt16

const DrawEval EvalCp = 0

const PawnVal = 100
const KnightVal = 300
const BishopVal = 300
const RookVal = 500
const QueenVal = 900

// Stolen from SunFish (tables inverted to reflect dragon pos ordering)
var PawnPosVals = []int8{
	0, 0, 0, 0, 0, 0, 0, 0,
	-31, 8, -7, -37, -36, -14, 3, -31,
	-22, 9, 5, -11, -10, -2, 3, -19,
	-26, 3, 10, 9, 6, 1, 0, -23,
	-17, 16, -2, 15, 14, 0, 15, -13,
	7, 29, 21, 44, 40, 31, 44, 7,
	78, 83, 86, 73, 102, 82, 85, 90,
	0, 0, 0, 0, 0, 0, 0, 0}

var KnightPosVals = []int8{
	-74, -23, -26, -24, -19, -35, -22, -69,
	-23, -15, 2, 0, 2, 0, -23, -20,
	-18, 10, 13, 22, 18, 15, 11, -14,
	-1, 5, 31, 21, 22, 35, 2, 0,
	24, 24, 45, 37, 33, 41, 25, 17,
	10, 67, 1, 74, 73, 27, 62, -2,
	-3, -6, 100, -36, 4, 62, -4, -14,
	-66, -53, -75, -75, -10, -55, -58, -70}

var BishopPosVals = []int8{
	-7, 2, -15, -12, -14, -15, -10, -10,
	19, 20, 11, 6, 7, 6, 20, 16,
	14, 25, 24, 15, 8, 25, 20, 15,
	13, 10, 17, 23, 17, 16, 0, 7,
	25, 17, 20, 34, 26, 25, 15, 10,
	-9, 39, -32, 41, 52, -10, 28, -14,
	-11, 20, 35, -42, -39, 31, 2, -22,
	-59, -78, -82, -76, -23, -107, -37, -50}

var RookPosVals = []int8{
	-30, -24, -18, 5, -2, -18, -31, -32,
	-53, -38, -31, -26, -29, -43, -44, -53,
	-42, -28, -42, -25, -25, -35, -26, -46,
	-28, -35, -16, -21, -13, -29, -46, -30,
	0, 5, 16, 13, 18, -4, -9, -6,
	19, 35, 28, 33, 45, 27, 25, 15,
	55, 29, 56, 67, 55, 62, 34, 60,
	35, 29, 33, 4, 37, 33, 56, 50}

var QueenPosVals = []int8{
	-39, -30, -31, -13, -31, -36, -34, -42,
	-36, -18, 0, -19, -15, -15, -21, -38,
	-30, -6, -13, -11, -16, -11, -16, -27,
	-14, -15, -2, -5, -1, -10, -20, -22,
	1, -16, 22, 17, 25, 20, -13, -6,
	-2, 43, 32, 60, 72, 63, 43, 2,
	14, 32, 60, -10, 20, 76, 57, 24,
	6, 1, -8, -104, 69, 24, 88, 26}

var KingPosVals = []int8{
	17, 30, -3, -14, 6, -1, 40, 18,
	-4, 3, -14, -50, -57, -18, 13, 4,
	-47, -42, -43, -79, -64, -32, -29, -32,
	-55, -43, -52, -28, -51, -47, -8, -50,
	-55, 50, 11, -4, -19, 13, 0, -49,
	-62, 12, -57, 44, -67, 28, 37, -31,
	-32, 10, 55, 56, 56, 55, 10, 3,
	4, 54, 47, -99, -99, 60, 83, -62}

// From - https://chessprogramming.wikispaces.com/Simplified+evaluation+function - (tables inverted to reflect dragon pos ordering)
var KingEndgamePosVals = []int8{
	-50, -30, -30, -30, -30, -30, -30, -50,
	-30, -30, 0, 0, 0, 0, -30, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -20, -10, 0, 0, -10, -20, -30,
	-50, -40, -30, -20, -20, -30, -40, -50}

func Evaluate(board *dragon.Board, legalMoves []dragon.Move) EvalCp {
	if isStalemate(board, legalMoves) {
		return DrawEval
	}

	if isMate(board, legalMoves) {
		if board.Wtomove {
			return BlackCheckMateEval
		}

		return WhiteCheckMateEval
	}

	whitePiecesVal := PiecesVal(&board.White)
	blackPiecesVal := PiecesVal(&board.Black)

	piecesEval := whitePiecesVal - blackPiecesVal

	whitePiecesPosVal := PiecesPosVal(&board.White, true, EndGameRatio(whitePiecesVal))
	blackPiecesPosVal := PiecesPosVal(&board.Black, false, EndGameRatio(blackPiecesVal))

	piecesPosEval := whitePiecesPosVal - blackPiecesPosVal

	return piecesEval + piecesPosEval
}

// Sum of individual piece evals
func PiecesVal(bitboards *dragon.Bitboards) EvalCp {
	eval := PawnVal * bits.OnesCount64(bitboards.Pawns)
	eval += BishopVal * bits.OnesCount64(bitboards.Bishops)
	eval += KnightVal * bits.OnesCount64(bitboards.Knights)
	eval += RookVal * bits.OnesCount64(bitboards.Rooks)
	eval += QueenVal * bits.OnesCount64(bitboards.Queens)

	return EvalCp(eval)
}

// Transition smoothly from King starting pos table to king end-game table between these total piece values.
const EndGamePiecesValHi EvalCp = 3000
const EndGamePiecesValLo EvalCp = 1200

// To what extent are we in end game; from 0.0 (not at all) to 1.0 (definitely)
func EndGameRatio(piecesVal EvalCp) float64 {
	// Somewhat arbitrary
	if piecesVal > EndGamePiecesValHi {
		return 0.0
	}

	if piecesVal < EndGamePiecesValLo {
		return 1.0
	}

	return float64(EndGamePiecesValHi-piecesVal) / float64(EndGamePiecesValHi-EndGamePiecesValLo)
}

// Sum of piece position values
//   endGameRatio is a number between 0.0 and 1.0 where 1.0 means we're in end-game
func PiecesPosVal(bitboards *dragon.Bitboards, isWhite bool, endGameRatio float64) EvalCp {
	eval := PieceTypePiecesPosVal(bitboards.Pawns, isWhite, PawnPosVals)
	eval += PieceTypePiecesPosVal(bitboards.Bishops, isWhite, BishopPosVals)
	eval += PieceTypePiecesPosVal(bitboards.Knights, isWhite, KnightPosVals)
	eval += PieceTypePiecesPosVal(bitboards.Rooks, isWhite, RookPosVals)
	eval += PieceTypePiecesPosVal(bitboards.Queens, isWhite, QueenPosVals)

	kingStartEval := PieceTypePiecesPosVal(bitboards.Kings, isWhite, KingPosVals)
	kingEndgameEval := PieceTypePiecesPosVal(bitboards.Kings, isWhite, KingEndgamePosVals)

	kingEval := (1.0-endGameRatio)*float64(kingStartEval) + endGameRatio*float64(kingEndgameEval)

	return eval + EvalCp(kingEval)
}

// Sum of piece position values for a particular type of piece
func PieceTypePiecesPosVal(bitmask uint64, isWhite bool, piecePosVals []int8) EvalCp {
	if !isWhite {
		// Flip the bitmask of Black pieces into White's perspective
		bitmask = bits.ReverseBytes64(bitmask)
	}

	var eval EvalCp = 0

	for bitmask != 0 {
		pos := bits.TrailingZeros64(bitmask)
		// (Could also use firstBit-1 trick to clear the bit)
		firstBit := uint64(1) << uint(pos)
		bitmask = bitmask ^ firstBit

		eval += EvalCp(piecePosVals[pos])
	}

	return eval
}

// If there are no legal moves, there are two possibilities - either our king
// is in check, or it isn't. In the first case it's mate and we've lost, and
// in the second case it's stalemate and therefore a draw.
func isStalemate(board *dragon.Board, legalMoves []dragon.Move) bool {
	return len(legalMoves) == 0 && !board.OurKingInCheck()
}

func isMate(board *dragon.Board, legalMoves []dragon.Move) bool {
	return len(legalMoves) == 0 && board.OurKingInCheck()
}
