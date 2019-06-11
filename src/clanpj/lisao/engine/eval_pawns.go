// Passed pawn bonuses

package engine

import (
	// "fmt"
	"math/bits"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

func evalByRank(wPieces uint64, bPieces uint64, rankVal *[8]float64) float64 {
	eval := 0.0

	for wPieces != 0 {
		pos := uint8(bits.TrailingZeros64(wPieces))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		wPieces = wPieces ^ posBit

		eval += pawnRankBonus[Rank(pos)]
	}
	
	for bPieces != 0 {
		pos := uint8(bits.TrailingZeros64(bPieces))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		bPieces = bPieces ^ posBit

		eval -= pawnRankBonus[7-Rank(pos)]
	}
	
	return eval
}

// Pawn rank bonuses (from white's perspective)
var pawnRankBonus = [8]float64 {
	0.0,
	-0.15,
	-0.07,
	0.04,
	0.11,
	0.29,
	0.40,
	0.0}

const pawnRankBonusScale = 0.75

func PawnRankEval(wPawns uint64, bPawns uint64) float64 {
	return evalByRank(wPawns, bPawns, &pawnRankBonus) * pawnRankBonusScale
}
	
// Pawn rank bonuses (from white's perspective)
var passedPawnRankBonus = [8]float64 {
	0.0,
	0.13,
	0.20,
	0.28,
	0.37,
	0.45,
	0.0}

var connectedPasserScale = 0.5 // in additional to baseline passed pawn value

func connectedPassers(pawns uint64, passedPawns uint64) uint64 {
	ewConnectorSquares := E(pawns) | W(pawns)
	allConnectorSquares := S(ewConnectorSquares) | ewConnectorSquares | N(ewConnectorSquares)

	return allConnectorSquares & passedPawns
}

const passedPawnRankScale = 1.0

// Evaluate passed pawns
func PassedPawnsEval(wPawns uint64, bPawns uint64) float64 {
	wPawnScope := WPawnScope(wPawns)
	bPawnScope := BPawnScope(bPawns)

	wPassedPawns := wPawns & ^bPawnScope
	bPassedPawns := bPawns & ^wPawnScope

	eval := evalByRank(wPassedPawns, bPassedPawns, &passedPawnRankBonus) * passedPawnRankScale
	
	wConnectedPassers := connectedPassers(wPawns, wPassedPawns)
	bConnectedPassers := connectedPassers(bPawns, bPassedPawns)

	eval += evalByRank(wConnectedPassers, bConnectedPassers, &passedPawnRankBonus) * connectedPasserScale

	return eval
}

// Penalty per doubled pawn
var doubledPawnPenalty float64 = -0.13

// Evaluate doubled pawns
func DoubledPawnsEval(wPawns uint64, bPawns uint64) float64 {
	wPawnTelestop := NFill(N(wPawns))
	wDoubledPawns := wPawnTelestop & wPawns
	nWDoubledPawns := bits.OnesCount64(wDoubledPawns)

	bPawnTelestop := SFill(S(bPawns))
	bDoubledPawns := bPawnTelestop & bPawns
	nBDoubledPawns := bits.OnesCount64(bDoubledPawns)

	return float64(nWDoubledPawns - nBDoubledPawns) * doubledPawnPenalty
}

func countPawnIslandsAndIsolatedPawns(pawnFiles uint64) (nIslands int, nIsolatedPawns int) {
	nIslands = 0
	nIsolatedPawns = 0
	
	nPawnsInIsland := 0
	for i := uint(0); i < 8; i++ {
		bit := uint64(1) << i
		fileBit := pawnFiles & bit
		if fileBit == 0 {
			if nPawnsInIsland != 0 {
				nIslands++
				if nPawnsInIsland == 1 {
					nIsolatedPawns++
				}
				nPawnsInIsland = 0
			}
		} else {
			// join the island
			nPawnsInIsland++
		}
	}
	if nPawnsInIsland != 0 {
		nIslands++
		if nPawnsInIsland == 1 {
			nIsolatedPawns++
		}
	}
	return
}

var pawnIslandPenalty = -0.07
var isolatedPawnPenalty = -0.09 // In addition to the pawn island penalty

func pawnIslandsEvalForColor(pawns uint64) float64 {
	pawns32 := pawns | (pawns >> 32)
	pawns16 := pawns32 | (pawns >> 16)
	pawns8 := pawns16 | (pawns16 >> 8)

	pawnFiles := pawns8 & 0xff

	// TODO use init lookup array(s)
	nIslands, nIsolatedPawns := countPawnIslandsAndIsolatedPawns(pawnFiles)

	// Ignore the first island 0 there has to be at least one (if there are no pawns)
	if nIslands > 0 {
		nIslands--
	}

	return float64(nIslands)*pawnIslandPenalty + float64(nIsolatedPawns)*isolatedPawnPenalty
}

// Isolated pawns and pawn islands
func PawnIslandsEval(wPawns uint64, bPawns uint64) float64 {
	return pawnIslandsEvalForColor(wPawns) -
		pawnIslandsEvalForColor(bPawns)
}

var connectedPawnBonus = 0.15
var ewConnectedPawnBonus = 0.09 // on top of connectedPawnBonus

// Connected pawns
func connectedPawns(pawns uint64) (allConnectors uint64, ewConnectors uint64) {
	ewConnectorSquares := E(pawns) | W(pawns)
	allConnectorSquares := S(ewConnectorSquares) | ewConnectorSquares | N(ewConnectorSquares)

	allConnectors = allConnectorSquares & pawns
	ewConnectors = ewConnectorSquares & pawns
	return
}

func connectedPawnsEvalForColor(pawns uint64) float64 {
	allConnectedPawns, ewConnectedPawns := connectedPawns(pawns)

	nConnectedPawns := bits.OnesCount64(allConnectedPawns)
	nEwConnectedPawns := bits.OnesCount64(ewConnectedPawns)

	return float64(nConnectedPawns)*connectedPawnBonus + float64(nEwConnectedPawns)*ewConnectedPawnBonus	
}

func ConnectedPawnsEval(wPawns uint64, bPawns uint64) float64 {
	return connectedPawnsEvalForColor(wPawns) -
		connectedPawnsEvalForColor(bPawns)
}

var pawnStructureScale = 0.65

// Pawn structure valuation (from white's perspective) in pawns
func PawnStructureEval(board *dragon.Board) float64 {
	wPawns := board.Bbs[dragon.White][dragon.Pawn]
	bPawns := board.Bbs[dragon.Black][dragon.Pawn]

	pawnStructureVal := PawnRankEval(wPawns, bPawns) +
		PassedPawnsEval(wPawns, bPawns) +
		DoubledPawnsEval(wPawns, bPawns) +
		PawnIslandsEval(wPawns, bPawns) +
		ConnectedPawnsEval(wPawns, bPawns)

	return pawnStructureVal * pawnStructureScale
}

