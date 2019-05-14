// Static board evaluation using positional piece influence

package engine

import (
	// "fmt"
	"math/bits"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

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

	// Squares attacked by pawns of each color, east and west respectively.
	pawnAttacks [dragon.NColors][NAttackDirs]uint64;

	// What kind of piece pwns each square
	// 2 extra entries for connected-bishops-behind-queen and connected-rooks-behind-queen respectively
	squareInfluence [64][dragon.NColors][dragon.NPieces+2]uint8
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

func (p *PositionalEvalT) initColor(color dragon.ColorT) {
	p.initPawns(color)
	p.initKnights(color)
	p.initBishops(color)
	p.initRooks(color)
	p.initQueens(color)
	p.initKings(color)
}

func (p *PositionalEvalT) initPawns(color dragon.ColorT) {
	p.pawnAttacks[color][AttackEast], p.pawnAttacks[color][AttackWest] =
		dragon.CalculatePawnsCaptureBitboard(p.board.Bbs[color][dragon.Pawn], color)
}

// TODO - this could be vastly collapsed using interfaces for per-piece type computation.

func (p *PositionalEvalT) initKnights(color dragon.ColorT) {
	knights := p.board.Bbs[color][dragon.Knight]

	for knights != 0 {
		pos := uint8(bits.TrailingZeros64(knights))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		knights = knights ^ posBit

		p.initKnight(color, pos)
	}
}

func (p *PositionalEvalT) initKnight(color dragon.ColorT, pos uint8) {
	influence := dragon.KnightMovesBitboard(pos)
	
	p.influenceByPiece[pos] = influence
}

func (p *PositionalEvalT) initBishops(color dragon.ColorT) {
	bishops := p.board.Bbs[color][dragon.Bishop]

	for bishops != 0 {
		pos := uint8(bits.TrailingZeros64(bishops))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		bishops = bishops ^ posBit

		p.initBishop(color, pos)
	}
}

func (p *PositionalEvalT) initBishop(color dragon.ColorT, pos uint8) {
	influence := dragon.CalculateBishopMoveBitboard(uint8(pos), p.allPieces)
	
	p.influenceByPiece[pos] = influence
}

func (p *PositionalEvalT) initRooks(color dragon.ColorT) {
	rooks := p.board.Bbs[color][dragon.Rook]

	for rooks != 0 {
		pos := uint8(bits.TrailingZeros64(rooks))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		rooks = rooks ^ posBit

		p.initRook(color, pos)
	}
}

func (p *PositionalEvalT) initRook(color dragon.ColorT, pos uint8) {
	influence := dragon.CalculateRookMoveBitboard(uint8(pos), p.allPieces)
	
	p.influenceByPiece[pos] = influence
}


func (p *PositionalEvalT) initQueens(color dragon.ColorT) {
	queens := p.board.Bbs[color][dragon.Queen]

	for queens != 0 {
		pos := uint8(bits.TrailingZeros64(queens))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		queens = queens ^ posBit

		p.initQueen(color, pos)
	}
}

func (p *PositionalEvalT) initQueen(color dragon.ColorT, pos uint8) {
	influence := dragon.CalculateBishopMoveBitboard(uint8(pos), p.allPieces) |
		dragon.CalculateRookMoveBitboard(uint8(pos), p.allPieces)
	
	p.influenceByPiece[pos] = influence
}

func (p *PositionalEvalT) initKings(color dragon.ColorT) {
	kings := p.board.Bbs[color][dragon.King]

	for kings != 0 {
		pos := uint8(bits.TrailingZeros64(kings))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		kings = kings ^ posBit

		p.initKing(color, pos)
	}
}

func (p *PositionalEvalT) initKing(color dragon.ColorT, pos uint8) {
	influence := dragon.KingMovesBitboard(pos)
	
	p.influenceByPiece[pos] = influence
}

// Invert the piece-wise influence bitmaps to produce a square-wise influence map of the board.
func (p *PositionalEvalT) initSquareInflence() {
	// Direct influence
	p.initSquareInflenceForColor(dragon.White)
	p.initSquareInflenceForColor(dragon.Black)

	// Add connected piece influence
	p.initConnectedPieceInfluence()
}

// Add the influence of connected pieces which influence 'through' their connected partners (Alekhine's gun)
func (p *PositionalEvalT) initConnectedPieceInfluence() {
	// TODO
}


func (p *PositionalEvalT) initSquareInflenceForColor(color dragon.ColorT) {
	allPieces := p.board.Bbs[color][dragon.All]

	for allPieces != 0 {
		pos := uint8(bits.TrailingZeros64(allPieces))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		allPieces = allPieces ^ posBit

		p.initSquareInflenceForPiece(color, pos)
	}
	
}
	
func (p *PositionalEvalT) initSquareInflenceForPiece(color dragon.ColorT, pos uint8) {
	influenceBits := p.influenceByPiece[pos]
	piece := p.board.PieceAt(pos)

	for influenceBits != 0 {
		pos := uint8(bits.TrailingZeros64(influenceBits))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		influenceBits = influenceBits ^ posBit

		p.squareInfluence[pos][color][piece] ++
	}
}

// Bonus for a side dominating a square - just a simple first metric
const squarePwnedBonus = 0.1

func squarePwnedBonusForColor(pwningColor dragon.ColorT) float64 {
	if pwningColor == dragon.White {
		return squarePwnedBonus
	} else {
		return -squarePwnedBonus
	}
}

func isPwnedByColor(nWhite uint8, nBlack uint8) (isPwned bool, pwningColor dragon.ColorT) {
	pwningColor = dragon.White
	
	if nWhite == nBlack {
		isPwned = false
	} else {
		isPwned = true
		if nWhite < nBlack {
			pwningColor = dragon.Black
		}
	}
	return
}

func (p *PositionalEvalT) squareEval(pos uint8) float64 {
	var infl *[dragon.NColors][dragon.NPieces+2]uint8 = &p.squareInfluence[pos]
	var inflW *[dragon.NPieces+2]uint8 = &infl[dragon.White]
	var inflB *[dragon.NPieces+2]uint8 = &infl[dragon.Black]

	isPwned, pwningColor := isPwnedByColor(inflW[dragon.Pawn], inflB[dragon.Pawn])

	if isPwned {
		return squarePwnedBonusForColor(pwningColor)
	}

	// Treat bishops and knights as equal
	isPwned, pwningColor = isPwnedByColor(inflW[dragon.Knight] + inflW[dragon.Bishop], inflW[dragon.Knight] + inflW[dragon.Bishop])

	if isPwned {
		return squarePwnedBonusForColor(pwningColor)
	}
	
	isPwned, pwningColor = isPwnedByColor(inflW[dragon.Rook], inflB[dragon.Rook])

	if isPwned {
		return squarePwnedBonusForColor(pwningColor)
	}

	isPwned, pwningColor = isPwnedByColor(inflW[dragon.Queen], inflB[dragon.Queen])

	if isPwned {
		return squarePwnedBonusForColor(pwningColor)
	}

	// TODO connnected weak pieces
	
	isPwned, pwningColor = isPwnedByColor(inflW[dragon.King], inflB[dragon.King])

	if isPwned {
		return squarePwnedBonusForColor(pwningColor)
	}

	return 0.0
}

// Evaluation in centi-pawns of the positional influence matrix
func (p *PositionalEvalT) Eval() EvalCp {
	pwnedBonus := 0.0
	for pos := uint8(0); pos < 64; pos++ {
		pwnedBonus += p.squareEval(pos)
	}

	// Round to centipawns
	return EvalCp(pwnedBonus*100.0) // Rounding?
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

// Expensive part - O(n)+ - of static eval from white's perspective.
func StaticPositionalEvalOrderN(board *dragon.Board) EvalCp {
	var positionalEval PositionalEvalT
	InitPositionalEval(board, &positionalEval)

	return positionalEval.Eval()
}

