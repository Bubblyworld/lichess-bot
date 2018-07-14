package engine

import (
	"math"
	"math/bits"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Eval in centi-pawns, i.e. 100 === 1 pawn
type EvalCp int16

const WhiteCheckMateEval EvalCp = math.MaxInt16
const BlackCheckMateEval EvalCp = -math.MaxInt16 // don't use MinInt16 cos it's not symmetrical with MaxInt16

// For NegaMax and friends this naming is more accurate
const MyCheckMateEval EvalCp = math.MaxInt16
const YourCheckMateEval EvalCp = -math.MaxInt16 // don't use MinInt16 cos it's not symmetrical with MaxInt16

// Used to mark transposition (hash) tables entries as invalid
const InvalidEval EvalCp = math.MinInt16

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

// Static eval only - no mate checks - from the perspective of the player to move
func NegaStaticEval(board *dragon.Board) EvalCp {
	staticEval := StaticEval(board)

	if board.Wtomove {
		return staticEval
	}
	return -staticEval
}

// Static eval only - no mate checks - from white's perspective
func StaticEval(board *dragon.Board) EvalCp {
	whitePiecesEval := piecesEval(&board.White)
	blackPiecesEval := piecesEval(&board.Black)

	piecesEval := whitePiecesEval - blackPiecesEval

	endGameRatio := EndGameRatio(whitePiecesEval + blackPiecesEval)

	whitePiecesPosEval := piecesPosVal(&board.White, &whitePiecePosVals, &whiteKingEndgamePosVals, endGameRatio)
	blackPiecesPosEval := piecesPosVal(&board.Black, &blackPiecePosVals, &blackKingEndgamePosVals, endGameRatio)

	pawnExtrasEval := pawnExtrasVal(board)
	kingProtectionEval := kingProtectionVal(board, endGameRatio)

	piecesPosEval := whitePiecesPosEval - blackPiecesPosEval + pawnExtrasEval + kingProtectionEval

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
// Note these are totals of black and white pieces.
const EndGamePiecesValHi EvalCp = 6000
const EndGamePiecesValLo EvalCp = 2400

// TODO delta eval doesn't cope with end-game-aware king eval
const NeverInEndgame = true

// To what extent are we in end game; from 0.0 (not at all) to 1.0 (definitely)
func EndGameRatio(bAndWPiecesVal EvalCp) float64 {
	if NeverInEndgame {
		return 0.0
	}

	// Somewhat arbitrary
	if bAndWPiecesVal > EndGamePiecesValHi {
		return 0.0
	}

	if bAndWPiecesVal < EndGamePiecesValLo {
		return 1.0
	}

	return float64(EndGamePiecesValHi-bAndWPiecesVal) / float64(EndGamePiecesValHi-EndGamePiecesValLo)
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

// Passed pawn bonuses
// TODO rationalise these with pawn pos vals
const pp2 int8 = 7
const pp3 int8 = 13
const pp4 int8 = 20
const pp5 int8 = 28
const pp6 int8 = 37

var whitePassedPawnPosVals = [64]int8{
	0, 0, 0, 0, 0, 0, 0, 0,
	pp2, pp2, pp2, pp2, pp2, pp2, pp2, pp2,
	pp3, pp3, pp3, pp3, pp3, pp3, pp3, pp3,
	pp4, pp4, pp4, pp4, pp4, pp4, pp4, pp4,
	pp5, pp5, pp5, pp5, pp5, pp5, pp5, pp5,
	pp6, pp6, pp6, pp6, pp6, pp6, pp6, pp6,
	0, 0, 0, 0, 0, 0, 0, 0, // a 7th rank pawn is always passed, so covered by the pawn-pos-val
	0, 0, 0, 0, 0, 0, 0, 0}

var blackPassedPawnPosVals = [64]int8{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, // a 7th rank pawn is always passed, so covered by the pawn-pos-val
	pp6, pp6, pp6, pp6, pp6, pp6, pp6, pp6,
	pp5, pp5, pp5, pp5, pp5, pp5, pp5, pp5,
	pp4, pp4, pp4, pp4, pp4, pp4, pp4, pp4,
	pp3, pp3, pp3, pp3, pp3, pp3, pp3, pp3,
	pp2, pp2, pp2, pp2, pp2, pp2, pp2, pp2,
	0, 0, 0, 0, 0, 0, 0, 0}

// Bonus for pawns protecting pawns
const pProtPawnVal = 10

// Bonus for pawns protecting pieces
const pProtPieceVal = 7

// Penalty per doubled pawn
const doubledPawnPenalty = -15

// Pawn extras
func pawnExtrasVal(board *dragon.Board) EvalCp {
	wPawns := board.White.Pawns
	bPawns := board.Black.Pawns

	// Passed pawns
	wPawnScope := WPawnScope(wPawns)
	bPawnScope := BPawnScope(wPawns)

	wPassedPawns := wPawns & ^bPawnScope
	bPassedPawns := bPawns & ^wPawnScope

	wPPVal := pieceTypePiecesPosVal(wPassedPawns, &whitePassedPawnPosVals)
	bPPVal := pieceTypePiecesPosVal(bPassedPawns, &blackPassedPawnPosVals)

	// Pawns protected by pawns
	wPawnAtt := WPawnAttacks(wPawns)
	wPawnsProtectedByPawns := wPawnAtt & wPawns
	wPProtPawnsVal := bits.OnesCount64(wPawnsProtectedByPawns) * pProtPawnVal

	bPawnAtt := BPawnAttacks(bPawns)
	bPawnsProtectedByPawns := bPawnAtt & bPawns
	bPProtPawnsVal := bits.OnesCount64(bPawnsProtectedByPawns) * pProtPawnVal

	// Pieces protected by pawns
	wPieces := board.White.All & ^wPawns
	wPiecesProtectedByPawns := wPawnAtt & wPieces
	wPProtPiecesVal := bits.OnesCount64(wPiecesProtectedByPawns) * pProtPieceVal

	bPieces := board.Black.All & ^bPawns
	bPiecesProtectedByPawns := bPawnAtt & bPieces
	bPProtPiecesVal := bits.OnesCount64(bPiecesProtectedByPawns) * pProtPieceVal

	// Doubled pawns
	wPawnTelestop := NFill(N(wPawns))
	wDoubledPawns := wPawnTelestop & wPawns
	wDoubledPawnVal := bits.OnesCount64(wDoubledPawns) * doubledPawnPenalty

	bPawnTelestop := SFill(S(bPawns))
	bDoubledPawns := bPawnTelestop & bPawns
	bDoubledPawnVal := bits.OnesCount64(bDoubledPawns) * doubledPawnPenalty

	return (wPPVal - bPPVal) +
		EvalCp(wPProtPawnsVal-bPProtPawnsVal) +
		EvalCp(wPProtPiecesVal-bPProtPiecesVal) +
		EvalCp(wDoubledPawnVal-bDoubledPawnVal)
}

type KingProtectionT uint8

const (
	NoProtection KingProtectionT = iota
	QSideProtection
	KSideProtection
)

// Which white king positions qualify for protection eval - index 0 is square A1, index 63 is square H8
var wKingProtectionTypes = [64]KingProtectionT{
	QSideProtection, QSideProtection, QSideProtection, NoProtection, NoProtection, NoProtection, KSideProtection, KSideProtection,
	QSideProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, KSideProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection}

// Bitboard locations of white king protecting pieces indexes by protection type
var wKingProtectionBbs = [3]uint64{
	0x0,                // NoProtection
	0x0007070000000000, // QSideProtection
	0x00e0e00000000000} // KSideProtection

// Which black king positions qualify for protection eval
var bKingProtectionTypes = [64]KingProtectionT{
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection,
	QSideProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, NoProtection, KSideProtection,
	QSideProtection, QSideProtection, QSideProtection, NoProtection, NoProtection, NoProtection, KSideProtection, KSideProtection}

// Bitboard locations of black king protecting pieces indexes by protection type
var bKingProtectionBbs = [3]uint64{
	0x0,                // NoProtection
	0x0000000000070700, // QSideProtection
	0x0000000000e0e000} // KSideProtection

// Bonus for pieces that are protecting the king
const kingProtectorVal = 8

// Additional bonus for pawns that are protecting the king
const kingPawnProtectorVal = 11

// Naive king protection - count pieces around the king if the king is in the corner
// From White's perspective
func kingProtectionVal(board *dragon.Board, endGameRatio float64) EvalCp {
	if endGameRatio == 1.0 {
		return 0
	}

	wBbs := board.White
	wKingPos := bits.TrailingZeros64(wBbs.Kings)
	wKingProtectionType := wKingProtectionTypes[wKingPos]
	wKingProtectionBb := wKingProtectionBbs[wKingProtectionType]

	wNonKingPieces := wBbs.All & ^wBbs.Kings
	wKingProtectors := wNonKingPieces & wKingProtectionBb

	wKingPawnProtectors := wBbs.Pawns & wKingProtectionBb

	wKingProtectionVal := bits.OnesCount64(wKingProtectors)*kingProtectorVal + bits.OnesCount64(wKingPawnProtectors)*kingPawnProtectorVal

	bBbs := board.Black
	bKingPos := bits.TrailingZeros64(bBbs.Kings)
	bKingProtectionType := bKingProtectionTypes[bKingPos]
	bKingProtectionBb := bKingProtectionBbs[bKingProtectionType]

	bNonKingPieces := bBbs.All & ^bBbs.Kings
	bKingProtectors := bNonKingPieces & bKingProtectionBb

	bKingPawnProtectors := bBbs.Pawns & bKingProtectionBb

	bKingProtectionVal := bits.OnesCount64(bKingProtectors)*kingProtectorVal + bits.OnesCount64(bKingPawnProtectors)*kingPawnProtectorVal

	// King protection in end-game is irrelevant
	return EvalCp(float64(wKingProtectionVal-bKingProtectionVal) * (1.0 - endGameRatio))
}
