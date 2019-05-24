// Static board evaluation using positional piece influence

package engine

import (
	// "fmt"
	"math"
	"math/bits"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

func rank(pos uint8) uint8 { return pos >> 3; }
func file(pos uint8) uint8 { return pos & 7; }

var rankBits = [8]uint64 {
	0x00000000000000ff,
	0x000000000000ff00,
	0x0000000000ff0000,
	0x00000000ff000000,
	0x000000ff00000000,
	0x0000ff0000000000,
	0x00ff000000000000,
	0xff00000000000000}

var fileBits = [8]uint64 {
	0x0101010101010101,
	0x0202020202020202,
	0x0404040404040404,
	0x0808080808080808,
	0x1010101010101010,
	0x2020202020202020,
	0x4040404040404040,
	0x8080808080808080}

func rankBitsForPos(pos uint8) uint64 {
	r := rank(pos)
	return rankBits[r]
}

func fileBitsForPos(pos uint8) uint64 {
	f := file(pos)
	return fileBits[f]
}

var nwDiagBits [16]uint64
var neDiagBits [16]uint64

func nwIndexForPos(pos uint8) uint8 {
	r := rank(pos)
	f := file(pos)

	return r + f
}

func neIndexForPos(pos uint8) uint8 {
	r := rank(pos)
	f := file(pos)

	return 8 + r - f
}

func initNwDiag(i uint8) {
	for pos := uint8(0); pos < 64; pos++ {
		nwIndex := nwIndexForPos(pos)

		if nwIndex == i {
			nwDiagBits[nwIndex] |= (uint64(1) << pos)
		}
	}
}

func initNeDiag(i uint8) {
	for pos := uint8(0); pos < 64; pos++ {
		neIndex := neIndexForPos(pos)

		if neIndex == i {
			neDiagBits[i] |= (uint64(1) << pos)
		}
	}
}

func init() {
	for i := uint8(0); i < 16; i++ {
		initNwDiag(i)
		initNeDiag(i)
	}
}

func nwDiagBitsForPos(pos uint8) uint64 {
	nwIndex := nwIndexForPos(pos)
	return nwDiagBits[nwIndex]
}

func neDiagBitsForPos(pos uint8) uint64 {
	neIndex := neIndexForPos(pos)
	return neDiagBits[neIndex]
}

// Pawn attack directions
type PawnAttackDirT uint8
const (
	AttackEast PawnAttackDirT = iota
	AttackWest
	NAttackDirs
)

// Connected pieces behind a Queen
const BishopBehindQueen = dragon.NPieces
const RookBehindQueen = dragon.NPieces+1

// Synopsis of the board position, including individual piece's influence, attack/defence bitmaps and much more.
// We ignore side to move.
type PositionalEvalT struct {
	// The underlying board.
	board *dragon.Board

	// Just a cache to simply the code - basically bitmap of occupied squares of both colors.
	allPieces uint64
	
	// The 'influence' of each piece, indexed by the piece's position.
	// Includes defence of pieces of the same color, so do '& ^MyAll' to get actual possible moves.
	// Also disregards pinning, moving into check etc.
	// NB: Does NOT include pawns (because we calculate pawn influence en-masse by color).
	influenceByPiece [64]uint64;

	// The 'influence' of each piece, through a connected queen (only valid for Bishops and Rooks)
	influenceBehindQueenByPiece [64]uint64;
	
	// Squares attacked by pawns of each color, east and west respectively.
	pawnAttacks [dragon.NColors][NAttackDirs]uint64;

	// What kind of piece pwns each square
	// 2 extra entries for connected-bishops-behind-queen and connected-rooks-behind-queen respectively
	squareInfluence [64][dragon.NColors][dragon.NPieces+2]int
}

// Initialise the positional evaluation for the given board.
func InitPositionalEval(board *dragon.Board, posEval *PositionalEvalT) {
	posEval.board = board

	posEval.allPieces = board.Bbs[dragon.White][dragon.All] | board.Bbs[dragon.Black][dragon.All]

	posEval.initColor(dragon.White)
	posEval.initColor(dragon.Black)

	// Invert the piece-wise influencers to provide square-wise influence 
	posEval.initSquareInflence()
}

func (p *PositionalEvalT) InfluenceOfPos(pos uint8) uint64 { return p.influenceByPiece[pos] }
func (p *PositionalEvalT) InfluenceBehindQueenOfPos(pos uint8) uint64 { return p.influenceBehindQueenByPiece[pos] }
func (p *PositionalEvalT) SquareInfluencedBy(pos uint8) [dragon.NColors][dragon.NPieces+2]int  { return p.squareInfluence[pos] }

func (p *PositionalEvalT) initColor(color dragon.ColorT) {
	p.initPawns(color)
	p.initPieceType(color, dragon.Knight, dragon.KnightMovesBitboard)
	p.initBishops(color)
	p.initRooks(color)
	p.initPieceType(color, dragon.Queen, func (pos uint8) uint64 {
		return dragon.CalculateBishopMoveBitboard(pos, p.allPieces) |
			dragon.CalculateRookMoveBitboard(pos, p.allPieces)
	})
	// Queen influence through Bishops and Rooks
	//   and Bishop/Rook influence through (behind) Queens
	p.initConnectedQueensAndMinorSliders(color)
	p.initPieceType(color, dragon.King, dragon.KingMovesBitboard)
}

func (p *PositionalEvalT) initPawns(color dragon.ColorT) {
	p.pawnAttacks[color][AttackEast], p.pawnAttacks[color][AttackWest] =
		dragon.CalculatePawnsCaptureBitboard(p.board.Bbs[color][dragon.Pawn], color)
}

func (p *PositionalEvalT) initPieceType(color dragon.ColorT, piece dragon.Piece, influenceFn func (pos uint8) (influence uint64)) {
	pieces := p.board.Bbs[color][piece]

	for pieces != 0 {
		pos := uint8(bits.TrailingZeros64(pieces))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		pieces = pieces ^ posBit

		p.influenceByPiece[pos] = influenceFn(pos)
	}
}

func (p *PositionalEvalT) accumulateConnectorInfluence(lineConnectors uint64, line uint64, influence uint64) uint64 {
	lineInfluence := influence & line
		
	linePieces := lineConnectors
	for linePieces != 0 {
		pos := uint8(bits.TrailingZeros64(linePieces))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		linePieces = linePieces ^ posBit
		
		lineInfluence |= p.influenceByPiece[pos] & line
	}

	return lineInfluence
}

func (p *PositionalEvalT) processConnectors(pos uint8, lowerPieces uint64, line uint64, influence uint64) {
	lineConnectors := lowerPieces & line & influence

	if lineConnectors != 0 {
		// Merge all line influence bits
		lineInfluence := p.accumulateConnectorInfluence(lineConnectors, line, influence)

		// Add back merged line influence to this piece
		p.influenceByPiece[pos] |= lineInfluence
		
		// Add back merged line influence bits to all connected pieces
		linePieces := lineConnectors
		for linePieces != 0 {
			pos := uint8(bits.TrailingZeros64(linePieces))
			// (Could also use posBit-1 trick to clear the bit)
			posBit := uint64(1) << uint(pos)
			linePieces = linePieces ^ posBit
			
			p.influenceByPiece[pos] |= lineInfluence
		}
	}
}

// Initialise bishop influence including connected bishops
func (p *PositionalEvalT) initBishops(color dragon.ColorT) {
	allBishops := p.board.Bbs[color][dragon.Bishop]
	bishops := allBishops

	for bishops != 0 {
		pos := uint8(bits.TrailingZeros64(bishops))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		bishops = bishops ^ posBit

		influence := dragon.CalculateBishopMoveBitboard(pos, p.allPieces)
		p.influenceByPiece[pos] = influence

		// Look for connected Bishops - this can only happen if there was a promotion to Bishop :D

		lowerPosBits := posBit - 1
		lowerBishops := allBishops & lowerPosBits

		// Connected bishops on the NW diagonal
		nwDiag := nwDiagBitsForPos(pos)
		p.processConnectors(pos, lowerBishops, nwDiag, influence)

		// Connected bishops on the NE diagonal
		neDiag := neDiagBitsForPos(pos)
		p.processConnectors(pos, lowerBishops, neDiag, influence)
	}
}

// Initialise rook influence including connected rooks
func (p *PositionalEvalT) initRooks(color dragon.ColorT) {
	allRooks := p.board.Bbs[color][dragon.Rook]
	rooks := allRooks

	for rooks != 0 {
		pos := uint8(bits.TrailingZeros64(rooks))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		rooks = rooks ^ posBit

		influence := dragon.CalculateRookMoveBitboard(pos, p.allPieces)
		p.influenceByPiece[pos] = influence

		lowerPosBits := posBit - 1
		lowerRooks := allRooks & lowerPosBits

		// Connected rooks on the same rank
		rankBits := rankBitsForPos(pos)
		p.processConnectors(pos, lowerRooks, rankBits, influence)

		// Connected rooks on the same file
		fileBits := fileBitsForPos(pos)
		p.processConnectors(pos, lowerRooks, fileBits, influence)
	}
}

// Initialise queen influence including connected queens, bishops and rooks
func (p *PositionalEvalT) initQueens(color dragon.ColorT) {
	allQueens := p.board.Bbs[color][dragon.Queen]
	queens := allQueens

	for queens != 0 {
		pos := uint8(bits.TrailingZeros64(queens))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		queens = queens ^ posBit

		influence := dragon.CalculateQueenMoveBitboard(pos, p.allPieces)
		p.influenceByPiece[pos] = influence

		// Look for connected Queens - only possible if there are multiple queens on the board

		lowerPosBits := posBit - 1
		lowerQueens := allQueens & lowerPosBits

		// Connected queens on the same rank
		rankBits := rankBitsForPos(pos)
		p.processConnectors(pos, lowerQueens, rankBits, influence)

		// Connected queens on the same file
		fileBits := fileBitsForPos(pos)
		p.processConnectors(pos, lowerQueens, fileBits, influence)

		// Connected queens on the NW diagonal
		nwDiag := nwDiagBitsForPos(pos)
		p.processConnectors(pos, lowerQueens, nwDiag, influence)

		// Connected queens on the NE diagonal
		neDiag := neDiagBitsForPos(pos)
		p.processConnectors(pos, lowerQueens, neDiag, influence)
	}
}

func (p *PositionalEvalT) processMinorConnectors(pos uint8, allMinorPieces uint64, line uint64, influence uint64) {
	lineConnectors := allMinorPieces & line & influence

	if lineConnectors != 0 {
		// Merge all line influence bits
		lineInfluence := p.accumulateConnectorInfluence(lineConnectors, line, influence)

		// Add the accumulated influence to this queen
		p.influenceByPiece[pos] |= lineInfluence

		// The minor connectors get additional influence as 'behind queen' influence
		linePieces := lineConnectors
		for linePieces != 0 {
			pos := uint8(bits.TrailingZeros64(linePieces))
			// (Could also use posBit-1 trick to clear the bit)
			posBit := uint64(1) << uint(pos)
			linePieces = linePieces ^ posBit
			
			p.influenceBehindQueenByPiece[pos] |= lineInfluence & ^p.influenceByPiece[pos]
		}
	}
}


// Initialise queen influence through bishops and rooks (and vice versa)
func (p *PositionalEvalT) initConnectedQueensAndMinorSliders(color dragon.ColorT) {
	allQueens := p.board.Bbs[color][dragon.Queen]
	allRooks := p.board.Bbs[color][dragon.Rook]
	allBishops := p.board.Bbs[color][dragon.Bishop]
	
	queens := allQueens

	for queens != 0 {
		pos := uint8(bits.TrailingZeros64(queens))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		queens = queens ^ posBit

		influence := dragon.CalculateQueenMoveBitboard(pos, p.allPieces)
		p.influenceByPiece[pos] = influence

		// Look for connected minor sliders

		// Connected rooks on the same rank
		rankBits := rankBitsForPos(pos)
		p.processMinorConnectors(pos, allRooks, rankBits, influence)

		// Connected rooks on the same file
		fileBits := fileBitsForPos(pos)
		p.processMinorConnectors(pos, allRooks, fileBits, influence)

		// Connected bishops on the NW diagonal
		nwDiag := nwDiagBitsForPos(pos)
		p.processConnectors(pos, allBishops, nwDiag, influence)

		// Connected queens on the NE diagonal
		neDiag := neDiagBitsForPos(pos)
		p.processConnectors(pos, allBishops, neDiag, influence)
	}
}


// Invert the piece-wise influence bitmaps to produce a square-wise influence map of the board.
func (p *PositionalEvalT) initSquareInflence() {
	// Direct influence and influence through connected pieces
	p.initSquareInflenceForColor(dragon.White)
	p.initSquareInflenceForColor(dragon.Black)
}

func (p *PositionalEvalT) initSquareInflenceForColor(color dragon.ColorT) {
	allPieces := p.board.Bbs[color][dragon.All]

	for allPieces != 0 {
		pos := uint8(bits.TrailingZeros64(allPieces))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		allPieces = allPieces ^ posBit

		p.initSquareInfluenceForPos(color, pos)
	}
	
}
	
func (p *PositionalEvalT) initSquareInfluenceForPos(color dragon.ColorT, pos uint8) {
	influenceBits := p.influenceByPiece[pos]
	piece := p.board.PieceAt(pos)

	p.initSquareInfluenceForPiece(color, piece, influenceBits)

	behindQueenInfluenceBits := p.influenceBehindQueenByPiece[pos]
	if behindQueenInfluenceBits != 0 {
		behindQueenPiece := BishopBehindQueen
		if piece == dragon.Rook {
			behindQueenPiece = RookBehindQueen
		}
		p.initSquareInfluenceForPiece(color, behindQueenPiece, behindQueenInfluenceBits)
	}
}

func (p *PositionalEvalT) initSquareInfluenceForPiece(color dragon.ColorT, piece dragon.Piece, influenceBits uint64) {
	for influenceBits != 0 {
		pos := uint8(bits.TrailingZeros64(influenceBits))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		influenceBits = influenceBits ^ posBit

		p.squareInfluence[pos][color][piece]++
	}
}

// Bonus for a side dominating a square - just a simple first metric
var squarePwnedByBonus = [9/*dragon.NPieces+2*/]float64 {
	0.00, // Nothing
	1.00, // Pawn
	0.80, // Knight
	0.80, // Bishop
	0.65, // Rook
	0.50, // Queen
	0.30, // King
	0.40, // BishopBehindQueen
	0.40} // RookBehindQueen

// Reduction in bonus for each level of non-dominant piece types
const pieceTypeReduction = 0.25

// Maximum excess of black pieces that can be attacking a square (basically worst case is 10 rooks after 8 pawns promote to rooks :D )
const maxAbsDiff = 10

// Bonus for a side dominating a square
var squarePwnedByDiffNBonus = [maxAbsDiff + 1 + maxAbsDiff]float64 {
        -1.375, // -10
        -1.375, // -9
        -1.375, // -8
        -1.375, // -7
        -1.375, // -6
        -1.375, // -5
        -1.375, // -4
        -1.375, // -3
        -1.25, // -2
        -1.0, // -1
        -0.0, // 0.0 unused
        1.0, // -1
        1.25, // -2
        1.375, // -3
        1.375, // -4
        1.375, // -5
        1.375, // -6
        1.375, // -7
        1.375, // -8
        1.375, // -9
	1.375} // -10


// diff is nWhite-nBlack
func squarePwnedBonus(diff int, pieceCategory uint8, reduction float64) float64 {
	baseBonus := squarePwnedByBonus[pieceCategory]
	diffBonus := squarePwnedByDiffNBonus[diff + maxAbsDiff]

	return baseBonus*diffBonus
}

const attackDefenseEvalScale = 0.0

const useBoardZone = true

const z1 = 1.2
const z2 = 1.125
const z3 = 1.0
const z4 = 0.9375

// Focus on the center of the board
var zoneFactor = [64]float64 {
	z4, z4, z4, z4, z4, z4, z4, z4,
	z4, z3, z3, z3, z3, z3, z3, z4,
	z4, z3, z2, z2, z2, z2, z3, z4,
	z4, z3, z2, z1, z1, z2, z3, z4,
	z4, z3, z2, z1, z1, z2, z3, z4,
	z4, z3, z2, z2, z2, z2, z3, z4,
	z4, z3, z3, z3, z3, z3, z3, z4,
	z4, z4, z4, z4, z4, z4, z4, z4}

func (p *PositionalEvalT) squareEval(pos uint8) float64 {
	eval := p.squarePwnEval(pos)

	piece := p.board.PieceAt(pos)
	if piece != dragon.Nothing {
		eval += p.pieceAttackDefenceEval(pos, piece) * attackDefenseEvalScale
	}

	if useBoardZone {
		eval *= zoneFactor[pos]
	}
	
	return eval
}

const protectedPieceEval = 0.5
const hangingPieceEval = -0.1
const lostPieceEval = -0.5

func isLostPieceTrivial(piece dragon.Piece, inflThem *[dragon.NPieces+2]int) bool {
	for pieceType := dragon.Pawn; pieceType < piece; pieceType++ {
		if inflThem[pieceType] > 0 {
			return true
		}
	}
	return false
}

func isHangingPieceTrivial(piece dragon.Piece, inflUs *[dragon.NPieces+2]int, inflThem *[dragon.NPieces+2]int) bool {
	for pieceType := dragon.Pawn; pieceType < dragon.NPieces; pieceType++ {
		if inflUs[pieceType] > 0 || inflThem[pieceType] > 0 {
			return false
		}
	}
	return true
}

// Is this piece lost, protected or hanging and how serious is this
func (p *PositionalEvalT) pieceAttackDefenceEval(pos uint8, piece dragon.Piece) float64 {
	isWhite := p.board.Bbs[dragon.White][dragon.All] & (uint64(1) << pos) != 0 //p.board.isWhitePieceAt(pos)

	var infl *[dragon.NColors][dragon.NPieces+2]int = &p.squareInfluence[pos]
	var inflW *[dragon.NPieces+2]int = &infl[dragon.White]
	var inflB *[dragon.NPieces+2]int = &infl[dragon.Black]

	var inflUs = inflW
	var inflThem = inflB
	if !isWhite {
		inflUs = inflB
		inflThem = inflW
	}

	// Is this piece protected?
	if isLostPieceTrivial(piece, inflThem) {
		if isWhite {
			return lostPieceEval
		} else {
			return -lostPieceEval
		}
	}

	if isHangingPieceTrivial(piece, inflUs, inflThem) {
		if isWhite {
			return hangingPieceEval
		} else {
			return -hangingPieceEval
		}
	}
	
	return 0.0
}

// Which side controls this square and to what extent?
func (p *PositionalEvalT) squarePwnEval(pos uint8) float64 {
	var infl *[dragon.NColors][dragon.NPieces+2]int = &p.squareInfluence[pos]
	var inflW *[dragon.NPieces+2]int = &infl[dragon.White]
	var inflB *[dragon.NPieces+2]int = &infl[dragon.Black]

	eval := 0.0
	reduction := 1.0

	// Pawns
	pawnDiff := inflW[dragon.Pawn] - inflB[dragon.Pawn]

	if pawnDiff != 0 {
		eval += squarePwnedBonus(pawnDiff, uint8(dragon.Pawn), reduction)
		reduction *= pieceTypeReduction
	}

	// Bishops and knights as equal parties
	knightAndBishopDiff := (inflW[dragon.Knight] + inflW[dragon.Bishop]) - (inflB[dragon.Knight] + inflB[dragon.Bishop])

	if knightAndBishopDiff != 0 {
		eval += squarePwnedBonus(knightAndBishopDiff, uint8(dragon.Knight), reduction)
		reduction *= pieceTypeReduction
	}

	// Rooks
	rookDiff := inflW[dragon.Rook] - inflB[dragon.Rook]

	if rookDiff != 0 {
		eval += squarePwnedBonus(rookDiff, uint8(dragon.Rook), reduction)
		reduction *= pieceTypeReduction
	}

	// Queens
	queenDiff := inflW[dragon.Queen] - inflB[dragon.Queen]

	if queenDiff != 0 {
		eval += squarePwnedBonus(queenDiff, uint8(dragon.Queen), reduction)
		reduction *= pieceTypeReduction
	}

	// Bishops connected through queens
	bishopBehindQueenDiff := inflW[BishopBehindQueen] - inflB[BishopBehindQueen]

	if bishopBehindQueenDiff != 0 {
		eval += squarePwnedBonus(bishopBehindQueenDiff, uint8(BishopBehindQueen), reduction)
		reduction *= pieceTypeReduction
	}

	// Rooks connected through queens
	rookBehindQueenDiff := inflW[RookBehindQueen] - inflB[RookBehindQueen]

	if rookBehindQueenDiff != 0 {
		eval += squarePwnedBonus(rookBehindQueenDiff, uint8(RookBehindQueen), reduction)
		reduction *= pieceTypeReduction
	}

	// Kings
	kingDiff := inflW[dragon.King] - inflB[dragon.King]

	if kingDiff != 0 {
		eval += squarePwnedBonus(kingDiff, uint8(dragon.King), reduction)
		reduction *= pieceTypeReduction
	}

	return eval
}

// 0.25 seems reasonable 0.1 and 0.3 outperformed 0.5
const posEvalScale = 0.25

// Evaluation in centi-pawns of the positional influence matrix
func (p *PositionalEvalT) Eval() EvalCp {
	eval := 0.0
	for pos := uint8(0); pos < 64; pos++ {
		eval += p.squareEval(pos)
	}

	eval *= posEvalScale

	// Round to centipawns
	return EvalCp(math.Round(eval*100.0))
}

// Pawn rank bonuses (from white's perspective)
var pawnRankBonus = [8]float64 {
	0.0,
	-0.15,
	-0.07,
	0.04,
	0.11,
	0.29,
	0.85,
	0.0}

const pawnRankBonusScale = 1.0

func pawnRankEval(wPawns uint64, bPawns uint64) EvalCp {
	eval := 0.0

	for wPawns != 0 {
		pos := uint8(bits.TrailingZeros64(wPawns))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		wPawns = wPawns ^ posBit

		eval += pawnRankBonus[rank(pos)]
	}
	
	for bPawns != 0 {
		pos := uint8(bits.TrailingZeros64(bPawns))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		bPawns = bPawns ^ posBit

		eval -= pawnRankBonus[7-rank(pos)]
	}
	
	// Round to centipawns
	return EvalCp(math.Round(eval*pawnRankBonusScale*100.0))
}


	
// Cheap part of static eval by opportunistic delta eval.
// We don't do anything here (for now).
func NegaStaticPositionalEvalOrder0Fast(board *dragon.Board, prevEval0 EvalCp, moveInfo *dragon.BoardSaveT) EvalCp {
	return DrawEval
}

// Full evaluation of the cheap delta part of the eval - O(0) with delta eval.
// We don't do anything here (for now).
func StaticPositionalEvalOrder0(board *dragon.Board) EvalCp {
	return DrawEval
}

const includePiecesEval = true
const includePawnRankEval = true

// Expensive part - O(n)+ - of static eval from white's perspective.
func StaticPositionalEvalOrderN(board *dragon.Board) EvalCp {
	piecesVal := EvalCp(0)

	if includePiecesEval {
		whitePiecesVal := piecesEval(&board.Bbs[dragon.White])
		blackPiecesVal := piecesEval(&board.Bbs[dragon.Black])
		
		piecesVal = whitePiecesVal - blackPiecesVal
	}

	pawnRankVal := EvalCp(0)
	if includePawnRankEval {
		pawnRankVal = pawnRankEval(board.Bbs[dragon.White][dragon.Pawn], board.Bbs[dragon.Black][dragon.Pawn])
	}
	
	var positionalEval PositionalEvalT
	InitPositionalEval(board, &positionalEval)

	return piecesVal + pawnRankVal + positionalEval.Eval()
}

