package engine

import (
	// "fmt"
	// "math"
	"math/bits"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Masks for attacks
// In order: knight on A1, B1, C1, ... F8, G8, H8
var knightMasks = [64]uint64{
	0x0000000000020400, 0x0000000000050800, 0x00000000000a1100, 0x0000000000142200,
	0x0000000000284400, 0x0000000000508800, 0x0000000000a01000, 0x0000000000402000,
	0x0000000002040004, 0x0000000005080008, 0x000000000a110011, 0x0000000014220022,
	0x0000000028440044, 0x0000000050880088, 0x00000000a0100010, 0x0000000040200020,
	0x0000000204000402, 0x0000000508000805, 0x0000000a1100110a, 0x0000001422002214,
	0x0000002844004428, 0x0000005088008850, 0x000000a0100010a0, 0x0000004020002040,
	0x0000020400040200, 0x0000050800080500, 0x00000a1100110a00, 0x0000142200221400,
	0x0000284400442800, 0x0000508800885000, 0x0000a0100010a000, 0x0000402000204000,
	0x0002040004020000, 0x0005080008050000, 0x000a1100110a0000, 0x0014220022140000,
	0x0028440044280000, 0x0050880088500000, 0x00a0100010a00000, 0x0040200020400000,
	0x0204000402000000, 0x0508000805000000, 0x0a1100110a000000, 0x1422002214000000,
	0x2844004428000000, 0x5088008850000000, 0xa0100010a0000000, 0x4020002040000000,
	0x0400040200000000, 0x0800080500000000, 0x1100110a00000000, 0x2200221400000000,
	0x4400442800000000, 0x8800885000000000, 0x100010a000000000, 0x2000204000000000,
	0x0004020000000000, 0x0008050000000000, 0x00110a0000000000, 0x0022140000000000,
	0x0044280000000000, 0x0088500000000000, 0x0010a00000000000, 0x0020400000000000}

var kingMasks = [64]uint64{
	0x0000000000000302, 0x0000000000000705, 0x0000000000000e0a, 0x0000000000001c14,
	0x0000000000003828, 0x0000000000007050, 0x000000000000e0a0, 0x000000000000c040,
	0x0000000000030203, 0x0000000000070507, 0x00000000000e0a0e, 0x00000000001c141c,
	0x0000000000382838, 0x0000000000705070, 0x0000000000e0a0e0, 0x0000000000c040c0,
	0x0000000003020300, 0x0000000007050700, 0x000000000e0a0e00, 0x000000001c141c00,
	0x0000000038283800, 0x0000000070507000, 0x00000000e0a0e000, 0x00000000c040c000,
	0x0000000302030000, 0x0000000705070000, 0x0000000e0a0e0000, 0x0000001c141c0000,
	0x0000003828380000, 0x0000007050700000, 0x000000e0a0e00000, 0x000000c040c00000,
	0x0000030203000000, 0x0000070507000000, 0x00000e0a0e000000, 0x00001c141c000000,
	0x0000382838000000, 0x0000705070000000, 0x0000e0a0e0000000, 0x0000c040c0000000,
	0x0003020300000000, 0x0007050700000000, 0x000e0a0e00000000, 0x001c141c00000000,
	0x0038283800000000, 0x0070507000000000, 0x00e0a0e000000000, 0x00c040c000000000,
	0x0302030000000000, 0x0705070000000000, 0x0e0a0e0000000000, 0x1c141c0000000000,
	0x3828380000000000, 0x7050700000000000, 0xe0a0e00000000000, 0xc040c00000000000,
	0x0203000000000000, 0x0507000000000000, 0x0a0e000000000000, 0x141c000000000000,
	0x2838000000000000, 0x5070000000000000, 0xa0e0000000000000, 0x40c0000000000000}

var king2Masks = [64]uint64{
        0x0000000000070404, 0x00000000000f0808, 0x00000000001f1111, 0x00000000003e2222,
        0x00000000007c4444, 0x0000000000f88888, 0x0000000000f01010, 0x0000000000e02020,
        0x0000000007040404, 0x000000000f080808, 0x000000001f111111, 0x000000003e222222,
        0x000000007c444444, 0x00000000f8888888, 0x00000000f0101010, 0x00000000e0202020,
        0x0000000704040407, 0x0000000f0808080f, 0x0000001f1111111f, 0x0000003e2222223e,
        0x0000007c4444447c, 0x000000f8888888f8, 0x000000f0101010f0, 0x000000e0202020e0,
        0x0000070404040700, 0x00000f0808080f00, 0x00001f1111111f00, 0x00003e2222223e00,
        0x00007c4444447c00, 0x0000f8888888f800, 0x0000f0101010f000, 0x0000e0202020e000,
        0x0007040404070000, 0x000f0808080f0000, 0x001f1111111f0000, 0x003e2222223e0000,
        0x007c4444447c0000, 0x00f8888888f80000, 0x00f0101010f00000, 0x00e0202020e00000,
        0x0704040407000000, 0x0f0808080f000000, 0x1f1111111f000000, 0x3e2222223e000000,
        0x7c4444447c000000, 0xf8888888f8000000, 0xf0101010f0000000, 0xe0202020e0000000,
        0x0404040700000000, 0x0808080f00000000, 0x1111111f00000000, 0x2222223e00000000,
        0x4444447c00000000, 0x888888f800000000, 0x101010f000000000, 0x202020e000000000,
        0x0404070000000000, 0x08080f0000000000, 0x11111f0000000000, 0x22223e0000000000,
        0x44447c0000000000, 0x8888f80000000000, 0x1010f00000000000, 0x2020e00000000000}

type RegionT uint8
const (
	Middle1 RegionT = iota
	Middle2
	Middle3
	Fullboard
	NRegions
)

const middle1Bits uint64   = 0x0000001818000000
const middle2Bits uint64   = 0x00003c3c3c3c0000
const middle3Bits uint64   = 0x007e7e7e7e7e7e00
const fullBoardBits uint64 = 0xffffffffffffffff

var Regions = [NRegions]uint64{ middle1Bits, middle2Bits, middle3Bits, fullBoardBits}

// Synopsis of the board position, including individual piece's influence, attack/defence bitmaps and much more.
// We ignore side to move.
// TODO - treat pinned pieces properly.
type PosEvalT struct {
	// The underlying board.
	board       *dragon.Board

	// Just a cache to simply the code
	allPieces uint64
	kingPos [dragon.NColors]uint8
	
	// The 'influence' of each piece, indexed by the piece's position.
	// Includes defence of pieces of the same color, so do '& ^MyAll' to get actual possible moves.
	// NB: Does NOT include pawns (because we calculate pawn influence en-masse by color).
	influence [64]uint64;

	// Squares attacked by pawns of each color.
	pawnAttacks [dragon.NColors]uint64;

	// Possible pawn moves of each color.
	// TODO - l8rs...
	//pawnMoves [NColors]uint64;

	// Aggregate (or'ed) influence of each piece-type of each color.
	// Total aggregate influence is in [color][All]
	influenceByPieceType [dragon.NColors][dragon.NPiecesWithAll]uint64;

	// Aggregate (or'ed) influence of pieces weaker than the given piece type.
	// We consider bishops and knights as equal for these purposes.
	influenceByWeakerPieceTypes [dragon.NColors][dragon.NPieces]uint64;

	// Total board influence by color
	influenceByColor [dragon.NColors]uint64
	
	// For each board position, the bitvector of all pieces (including pawns and kings) that influence it.
	// Includes influencers of both colors.
	// TODO - l8ters...
	//influencedBy [64]uint64;
}

// Initialise the position evaluation for the given board.
func InitPosEval(board *dragon.Board, posEval *PosEvalT) {
	posEval.board = board

	posEval.allPieces = board.Bbs[dragon.White][dragon.All] | board.Bbs[dragon.Black][dragon.All]

	posEval.kingPos[dragon.White] = uint8(bits.TrailingZeros64(board.Bbs[dragon.White][dragon.King]))
	posEval.kingPos[dragon.Black] = uint8(bits.TrailingZeros64(board.Bbs[dragon.Black][dragon.King]))

	posEval.initColor(dragon.White)
	posEval.initColor(dragon.Black)
}

func (p *PosEvalT) initColor(color dragon.ColorT) {
	p.initPawns(color)
	p.initKnights(color)
	p.initBishops(color)
	p.initRooks(color)
	p.initQueens(color)
	p.initKings(color)

	p.initInfluenceByWeakerPieceTypes(color)

	p.initInflenceByColor(color)
}

func (p *PosEvalT) initPawns(color dragon.ColorT) {
	if color == dragon.White {
		p.pawnAttacks[dragon.White] = WPawnAttacks(p.board.Bbs[dragon.White][dragon.Pawn])
	} else {
		p.pawnAttacks[dragon.Black] = BPawnAttacks(p.board.Bbs[dragon.Black][dragon.Pawn])
	}
}

func (p *PosEvalT) initKnights(color dragon.ColorT) {
	knights := p.board.Bbs[color][dragon.Knight]

	for knights != 0 {
		pos := uint8(bits.TrailingZeros64(knights))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		knights = knights ^ posBit

		p.initKnight(color, pos)
	}
}

func (p *PosEvalT) initKnight(color dragon.ColorT, pos uint8) {
	influence := knightMasks[pos]
	
	p.influence[pos] = influence

	p.influenceByPieceType[color][dragon.Knight] |= influence
}

func (p *PosEvalT) initBishops(color dragon.ColorT) {
	bishops := p.board.Bbs[color][dragon.Bishop]

	for bishops != 0 {
		pos := uint8(bits.TrailingZeros64(bishops))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		bishops = bishops ^ posBit

		p.initBishop(color, pos)
	}
}

func (p *PosEvalT) initBishop(color dragon.ColorT, pos uint8) {
	influence := dragon.CalculateBishopMoveBitboard(uint8(pos), p.allPieces)
	
	p.influence[pos] = influence

	p.influenceByPieceType[color][dragon.Bishop] |= influence
}

func (p *PosEvalT) initRooks(color dragon.ColorT) {
	rooks := p.board.Bbs[color][dragon.Rook]

	for rooks != 0 {
		pos := uint8(bits.TrailingZeros64(rooks))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		rooks = rooks ^ posBit

		p.initRook(color, pos)
	}
}

func (p *PosEvalT) initRook(color dragon.ColorT, pos uint8) {
	influence := dragon.CalculateRookMoveBitboard(uint8(pos), p.allPieces)
	
	p.influence[pos] = influence

	p.influenceByPieceType[color][dragon.Rook] |= influence
}


func (p *PosEvalT) initQueens(color dragon.ColorT) {
	queens := p.board.Bbs[color][dragon.Queen]

	for queens != 0 {
		pos := uint8(bits.TrailingZeros64(queens))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		queens = queens ^ posBit

		p.initQueen(color, pos)
	}
}

func (p *PosEvalT) initQueen(color dragon.ColorT, pos uint8) {
	influence := dragon.CalculateBishopMoveBitboard(uint8(pos), p.allPieces) |
		dragon.CalculateRookMoveBitboard(uint8(pos), p.allPieces)
	
	p.influence[pos] = influence

	p.influenceByPieceType[color][dragon.Queen] |= influence
}

func (p *PosEvalT) initKings(color dragon.ColorT) {
	kings := p.board.Bbs[color][dragon.King]

	for kings != 0 {
		pos := uint8(bits.TrailingZeros64(kings))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		kings = kings ^ posBit

		p.initKing(color, pos)
	}
}

func (p *PosEvalT) initKing(color dragon.ColorT, pos uint8) {
	influence := kingMasks[pos]
	
	p.influence[pos] = influence

	p.influenceByPieceType[color][dragon.King] |= influence
}

func (p *PosEvalT) initInfluenceByWeakerPieceTypes(color dragon.ColorT) {
	p.influenceByWeakerPieceTypes[color][dragon.Knight] = p.pawnAttacks[color]
	p.influenceByWeakerPieceTypes[color][dragon.Bishop] = p.pawnAttacks[color]

	p.influenceByWeakerPieceTypes[color][dragon.Rook] = p.pawnAttacks[color] |
		p.influenceByPieceType[color][dragon.Knight] |
		p.influenceByPieceType[color][dragon.Bishop]

	p.influenceByWeakerPieceTypes[color][dragon.Queen] =
		p.influenceByPieceType[color][dragon.Rook] |
		p.influenceByWeakerPieceTypes[color][dragon.Rook]
}

func (p *PosEvalT) initInflenceByColor(color dragon.ColorT) {
	p.influenceByColor[color] =
		p.influenceByWeakerPieceTypes[color][dragon.Queen] |
		p.influenceByPieceType[color][dragon.Queen] |
		p.influenceByPieceType[color][dragon.King] // Not sure if King should be included, but...
}


/** From White's perspective */
func (p *PosEvalT) calcInfluenceEval() EvalCp {
	influenceEval := EvalCp(0)

	// Iterate through allPieces rather...
	for piecesBits := p.allPieces; piecesBits !=0; {
		pos := uint8(bits.TrailingZeros64(piecesBits))
		posBit := uint64(1) << uint(pos)
		piecesBits = piecesBits ^ posBit

		color := dragon.White
		if (posBit & p.board.Bbs[dragon.Black][dragon.All]) != 0 {
			color = dragon.Black
		}
		pieceInfluenceEval := p.calcPieceInfluenceEval(pos, color)
		if color == dragon.Black {
			pieceInfluenceEval = -pieceInfluenceEval
		}
		influenceEval += pieceInfluenceEval
	}

	return influenceEval
}

const SpaceBonusPerSquare = EvalCp(7)

/** From White's perspective */
func (p *PosEvalT) calcSpaceEval() EvalCp {
	whiteSpaceBits := p.influenceByColor[dragon.White] & ^p.influenceByColor[dragon.Black]
	
	blackSpaceBits := p.influenceByColor[dragon.Black] & ^p.influenceByColor[dragon.White]

	return SpaceBonusPerSquare * EvalCp(bits.OnesCount64(whiteSpaceBits) - bits.OnesCount64(blackSpaceBits))
}


// Note that the regions are inclusive (so full-board includes middle3 includes middle2 includes middle1)
var NothingInfluenceByBoardRegion = [NRegions]EvalCp{ 0, 0, 0, 0}
var PawnInfluenceByBoardRegion    = [NRegions]EvalCp{ 0, 0, 0, 0}  // Already captured by pos-vals for Pawns
var KnightInfluenceByBoardRegion  = [NRegions]EvalCp{ 0, 0, 0, 0}  // Already captured by pos-vals for Knights
var BishopInfluenceByBoardRegion  = [NRegions]EvalCp{ 3, 3, 3, 3}
var RookInfluenceByBoardRegion    = [NRegions]EvalCp{ 2, 2, 2, 2}
var QueenInfluenceByBoardRegion   = [NRegions]EvalCp{ 1, 1, 1, 1}
var KingInfluenceByBoardRegion    = [NRegions]EvalCp{ 0, 0, 0, 0}  // Already captured by pos-vals for Knights

var PieceInfluenceByBoardRegion = [dragon.NPieces][NRegions]EvalCp{
	NothingInfluenceByBoardRegion,
	PawnInfluenceByBoardRegion,
	KnightInfluenceByBoardRegion,
	BishopInfluenceByBoardRegion,
	RookInfluenceByBoardRegion,
	QueenInfluenceByBoardRegion,
	KingInfluenceByBoardRegion}

var StuckPiecePenalty = [dragon.NPieces]EvalCp{
	0,    // Nothing
	0,    // Pawn
	3,    // Knight is never really stuck
	5,  // Bishop
	7,  // Rook
	7,  // Queen
	0}    // King
	
var SemiStuckPiecePenalty = [dragon.NPieces]EvalCp{
	0,    // Nothing
	0,    // Pawn
	2,   // Knight
	3,   // Bishop
	4,  // Rook
	5,  // Queen
	0}    // King

var King1AttackBonus = [dragon.NPieces]EvalCp{
	0,    // Nothing
	2,    // Pawn
	3,   // Knight
	3,   // Bishop
	5,  // Rook
	5,  // Queen
	0}    // King

var King2AttackBonus = [dragon.NPieces]EvalCp{
	0,    // Nothing
	1,   // Pawn
	2,    // Knight
	2,    // Bishop
	3,    // Rook
	3,    // Queen
	0}    // King

/** From Piece color's perspective */
func (p *PosEvalT) calcPieceInfluenceEval(pos uint8, color dragon.ColorT) EvalCp {
	eval := EvalCp(0)

	oppColor := dragon.Black ^ color
	piece := p.board.PieceAt(pos)

	// Board influence
	regionEvals := &PieceInfluenceByBoardRegion[piece]

	influence := p.influence[pos]
	eval += regionEvals[Fullboard] * EvalCp(bits.OnesCount64(influence))
	
	influenceMiddle1 := influence & Regions[Middle1]
	eval += regionEvals[Middle1] * EvalCp(bits.OnesCount64(influenceMiddle1))
		
	influenceMiddle2 := influence & Regions[Middle2]
	eval += regionEvals[Middle2] * EvalCp(bits.OnesCount64(influenceMiddle2))
		
	influenceMiddle3 := influence & Regions[Middle3]
	eval += regionEvals[Middle3] * EvalCp(bits.OnesCount64(influenceMiddle3))

	// King attack/defence - TODO defence
	oppKingPos := p.kingPos[oppColor]

	influenceOppKing1 := influence & kingMasks[oppKingPos]
	eval += King1AttackBonus[piece] * EvalCp(bits.OnesCount64(influenceOppKing1))

	influenceOppKing2 := influence & king2Masks[oppKingPos]
	eval += King2AttackBonus[piece] * EvalCp(bits.OnesCount64(influenceOppKing2))

	// Penalty for being stuck - exclude protection of own color pieces
	moves := influence & ^p.board.Bbs[color][dragon.All]
	if moves == 0 {
		eval += StuckPiecePenalty[piece]
	} else {
		// Semi-stuck - all potential targets are blocked by weaker opponents
		moves := moves & ^p.influenceByWeakerPieceTypes[oppColor][piece]
		if moves == 0 {
			eval += SemiStuckPiecePenalty[piece]
		}
	}

	return eval
}	
