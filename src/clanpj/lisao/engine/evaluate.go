package engine

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

// Piece values
const nothingVal = 0
const pawnVal = 100
const knightVal = 300
const bishopVal = 300
const rookVal = 500
const queenVal = 900
const kingVal = 0

var pieceVals = [7]EvalCp{
	nothingVal,
	pawnVal,
	knightVal,
	bishopVal,
	rookVal,
	queenVal,
	kingVal}

var nothingPosVals = [64]int8{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0}

// Stolen from SunFish (tables inverted to reflect dragon pos ordering)
var whitePawnPosVals = [64]int8{
	0, 0, 0, 0, 0, 0, 0, 0,
	-31, 8, -7, -37, -36, -14, 3, -31,
	-22, 9, 5, -11, -10, -2, 3, -19,
	-26, 3, 10, 9, 6, 1, 0, -23,
	-17, 16, -2, 15, 14, 0, 15, -13,
	7, 29, 21, 44, 40, 31, 44, 7,
	78, 83, 86, 73, 102, 82, 85, 90,
	0, 0, 0, 0, 0, 0, 0, 0}

var whiteKnightPosVals = [64]int8{
	-74, -23, -26, -24, -19, -35, -22, -69,
	-23, -15, 2, 0, 2, 0, -23, -20,
	-18, 10, 13, 22, 18, 15, 11, -14,
	-1, 5, 31, 21, 22, 35, 2, 0,
	24, 24, 45, 37, 33, 41, 25, 17,
	10, 67, 1, 74, 73, 27, 62, -2,
	-3, -6, 100, -36, 4, 62, -4, -14,
	-66, -53, -75, -75, -10, -55, -58, -70}

var whiteBishopPosVals = [64]int8{
	-7, 2, -15, -12, -14, -15, -10, -10,
	19, 20, 11, 6, 7, 6, 20, 16,
	14, 25, 24, 15, 8, 25, 20, 15,
	13, 10, 17, 23, 17, 16, 0, 7,
	25, 17, 20, 34, 26, 25, 15, 10,
	-9, 39, -32, 41, 52, -10, 28, -14,
	-11, 20, 35, -42, -39, 31, 2, -22,
	-59, -78, -82, -76, -23, -107, -37, -50}

var whiteRookPosVals = [64]int8{
	-30, -24, -18, 5, -2, -18, -31, -32,
	-53, -38, -31, -26, -29, -43, -44, -53,
	-42, -28, -42, -25, -25, -35, -26, -46,
	-28, -35, -16, -21, -13, -29, -46, -30,
	0, 5, 16, 13, 18, -4, -9, -6,
	19, 35, 28, 33, 45, 27, 25, 15,
	55, 29, 56, 67, 55, 62, 34, 60,
	35, 29, 33, 4, 37, 33, 56, 50}

var whiteQueenPosVals = [64]int8{
	-39, -30, -31, -13, -31, -36, -34, -42,
	-36, -18, 0, -19, -15, -15, -21, -38,
	-30, -6, -13, -11, -16, -11, -16, -27,
	-14, -15, -2, -5, -1, -10, -20, -22,
	1, -16, 22, 17, 25, 20, -13, -6,
	-2, 43, 32, 60, 72, 63, 43, 2,
	14, 32, 60, -10, 20, 76, 57, 24,
	6, 1, -8, -104, 69, 24, 88, 26}

var whiteKingPosVals = [64]int8{
	17, 30, -3, -14, 6, -1, 40, 18,
	-4, 3, -14, -50, -57, -18, 13, 4,
	-47, -42, -43, -79, -64, -32, -29, -32,
	-55, -43, -52, -28, -51, -47, -8, -50,
	-55, 50, 11, -4, -19, 13, 0, -49,
	-62, 12, -57, 44, -67, 28, 37, -31,
	-32, 10, 55, 56, 56, 55, 10, 3,
	4, 54, 47, -99, -99, 60, 83, -62}

// From - https://chessprogramming.wikispaces.com/Simplified+evaluation+function - (tables inverted to reflect dragon pos ordering)
var whiteKingEndgamePosVals = [64]int8{
	-50, -30, -30, -30, -30, -30, -30, -50,
	-30, -30, 0, 0, 0, 0, -30, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -20, -10, 0, 0, -10, -20, -30,
	-50, -40, -30, -20, -20, -30, -40, -50}

var whitePiecePosVals = [7]*[64]int8{
	&nothingPosVals,
	&whitePawnPosVals,
	&whiteKnightPosVals,
	&whiteBishopPosVals,
	&whiteRookPosVals,
	&whiteQueenPosVals,
	&whiteKingPosVals}

// Stolen from SunFish
var blackPawnPosVals = [64]int8{
	0, 0, 0, 0, 0, 0, 0, 0,
	78, 83, 86, 73, 102, 82, 85, 90,
	7, 29, 21, 44, 40, 31, 44, 7,
	-17, 16, -2, 15, 14, 0, 15, -13,
	-26, 3, 10, 9, 6, 1, 0, -23,
	-22, 9, 5, -11, -10, -2, 3, -19,
	-31, 8, -7, -37, -36, -14, 3, -31,
	0, 0, 0, 0, 0, 0, 0, 0}

var blackKnightPosVals = [64]int8{
	-66, -53, -75, -75, -10, -55, -58, -70,
	-3, -6, 100, -36, 4, 62, -4, -14,
	10, 67, 1, 74, 73, 27, 62, -2,
	24, 24, 45, 37, 33, 41, 25, 17,
	-1, 5, 31, 21, 22, 35, 2, 0,
	-18, 10, 13, 22, 18, 15, 11, -14,
	-23, -15, 2, 0, 2, 0, -23, -20,
	-74, -23, -26, -24, -19, -35, -22, -69}

var blackBishopPosVals = [64]int8{
	-59, -78, -82, -76, -23, -107, -37, -50,
	-11, 20, 35, -42, -39, 31, 2, -22,
	-9, 39, -32, 41, 52, -10, 28, -14,
	25, 17, 20, 34, 26, 25, 15, 10,
	13, 10, 17, 23, 17, 16, 0, 7,
	14, 25, 24, 15, 8, 25, 20, 15,
	19, 20, 11, 6, 7, 6, 20, 16,
	-7, 2, -15, -12, -14, -15, -10, -10}

var blackRookPosVals = [64]int8{
	35, 29, 33, 4, 37, 33, 56, 50,
	55, 29, 56, 67, 55, 62, 34, 60,
	19, 35, 28, 33, 45, 27, 25, 15,
	0, 5, 16, 13, 18, -4, -9, -6,
	-28, -35, -16, -21, -13, -29, -46, -30,
	-42, -28, -42, -25, -25, -35, -26, -46,
	-53, -38, -31, -26, -29, -43, -44, -53,
	-30, -24, -18, 5, -2, -18, -31, -32}

var blackQueenPosVals = [64]int8{
	6, 1, -8, -104, 69, 24, 88, 26,
	14, 32, 60, -10, 20, 76, 57, 24,
        -2, 43, 32, 60, 72, 63, 43, 2,
	1, -16, 22, 17, 25, 20, -13, -6,
	-14, -15, -2, -5, -1, -10, -20, -22,
	-30, -6, -13, -11, -16, -11, -16, -27,
	-36, -18, 0, -19, -15, -15, -21, -38,
	-39, -30, -31, -13, -31, -36, -34, -42}

var blackKingPosVals = [64]int8{
	4, 54, 47, -99, -99, 60, 83, -62,
	-32, 10, 55, 56, 56, 55, 10, 3,
	-62, 12, -57, 44, -67, 28, 37, -31,
	-55, 50, 11, -4, -19, 13, 0, -49,
	-55, -43, -52, -28, -51, -47, -8, -50,
	-47, -42, -43, -79, -64, -32, -29, -32,
	-4, 3, -14, -50, -57, -18, 13, 4,
	17, 30, -3, -14, 6, -1, 40, 18}

// From - https://chessprogramming.wikispaces.com/Simplified+evaluation+function
var blackKingEndgamePosVals = [64]int8{
	-50, -40, -30, -20, -20, -30, -40, -50,
	-30, -20, -10, 0, 0, -10, -20, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -30, 0, 0, 0, 0, -30, -30,
	-50, -30, -30, -30, -30, -30, -30, -50}

var blackPiecePosVals = [7]*[64]int8{
	&nothingPosVals,
	&blackPawnPosVals,
	&blackKnightPosVals,
	&blackBishopPosVals,
	&blackRookPosVals,
	&blackQueenPosVals,
	&blackKingPosVals}

// Eval delta due to a move - from white perspective
func EvalDelta(move dragon.Move, moveInfo *dragon.MoveApplication, isWhiteMove bool) EvalCp {
	myPiecePosVals := blackPiecePosVals
	yourPiecePosVals := whitePiecePosVals
	if isWhiteMove {
		myPiecePosVals = whitePiecePosVals
		yourPiecePosVals = blackPiecePosVals
	}
	
	fromEval := pieceVals[moveInfo.FromPieceType] + EvalCp(myPiecePosVals[moveInfo.FromPieceType][move.From()])
	toEval := pieceVals[moveInfo.ToPieceType] + EvalCp(myPiecePosVals[moveInfo.ToPieceType][move.To()])
	captureEval := pieceVals[moveInfo.CapturedPieceType] + EvalCp(yourPiecePosVals[moveInfo.CapturedPieceType][moveInfo.CaptureLocation])

	var castlingRookDelta EvalCp = 0
	if moveInfo.IsCastling {
		myRookPosVals := myPiecePosVals[dragon.Rook]
		castlingRookDelta = EvalCp(myRookPosVals[moveInfo.RookCastleTo] - myRookPosVals[moveInfo.RookCastleFrom])
	}

	evalDelta := toEval - fromEval + captureEval + castlingRookDelta

	if !isWhiteMove {
		evalDelta = -evalDelta
	}

	return evalDelta
}

// Static eval only - no mate checks - from white perspective
func StaticEval(board *dragon.Board) EvalCp {
	whitePiecesEval := piecesEval(&board.White)
	blackPiecesEval := piecesEval(&board.Black)

	piecesEval := whitePiecesEval - blackPiecesEval

	whitePiecesPosEval := piecesPosVal(&board.White, &whitePiecePosVals, &whiteKingEndgamePosVals, EndGameRatio(whitePiecesEval))
	blackPiecesPosEval := piecesPosVal(&board.Black, &blackPiecePosVals, &blackKingEndgamePosVals, EndGameRatio(blackPiecesEval))

	piecesPosEval := whitePiecesPosEval - blackPiecesPosEval

	return piecesEval + piecesPosEval
}

// Sum of individual piece evals
func piecesEval(bitboards *dragon.Bitboards) EvalCp {
	eval := pawnVal * bits.OnesCount64(bitboards.Pawns)
	eval += bishopVal * bits.OnesCount64(bitboards.Bishops)
	eval += knightVal * bits.OnesCount64(bitboards.Knights)
	eval += rookVal * bits.OnesCount64(bitboards.Rooks)
	eval += queenVal * bits.OnesCount64(bitboards.Queens)

	return EvalCp(eval)
}

// Transition smoothly from King starting pos table to king end-game table between these total piece values.
const EndGamePiecesValHi EvalCp = 3000
const EndGamePiecesValLo EvalCp = 1200

// TODO delta eval doesn't cope with end-game-aware king eval
const NeverInEndgame = true

// To what extent are we in end game; from 0.0 (not at all) to 1.0 (definitely)
func EndGameRatio(piecesVal EvalCp) float64 {
	if NeverInEndgame {
		return 0.0
	}
	
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
func piecesPosVal(bitboards *dragon.Bitboards, piecePosVals *[7]*[64]int8, kingEndgamePosVals *[64]int8, endGameRatio float64) EvalCp {
	eval := pieceTypePiecesPosVal(bitboards.Pawns, piecePosVals[dragon.Pawn])
	eval += pieceTypePiecesPosVal(bitboards.Bishops, piecePosVals[dragon.Bishop])
	eval += pieceTypePiecesPosVal(bitboards.Knights, piecePosVals[dragon.Knight])
	eval += pieceTypePiecesPosVal(bitboards.Rooks, piecePosVals[dragon.Rook])
	eval += pieceTypePiecesPosVal(bitboards.Queens, piecePosVals[dragon.Queen])

	kingStartEval := pieceTypePiecesPosVal(bitboards.Kings, piecePosVals[dragon.King])
	kingEndgameEval := pieceTypePiecesPosVal(bitboards.Kings, kingEndgamePosVals)

	kingEval := (1.0-endGameRatio)*float64(kingStartEval) + endGameRatio*float64(kingEndgameEval)

	return eval + EvalCp(kingEval)
}

// Sum of piece position values for a particular type of piece
func pieceTypePiecesPosVal(bitmask uint64, piecePosVals *[64]int8) EvalCp {
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

