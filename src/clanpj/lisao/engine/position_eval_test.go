package engine

import (
	"testing"
	"fmt"
	"math/bits"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

var files = [8]string{ "a", "b", "c", "d", "e", "f", "g", "h" }
var ranks = [8]string{ "1", "2", "3", "4", "5", "6", "7", "8" }

var pieceTypeNothing = "."

var pieceTypes = [dragon.NColors][dragon.NPieces]string{
	{ ".", "P", "N", "B", "R", "Q", "K"},
	{ ".", "p", "n", "b", "r", "q", "k"}}

func posBitsToString(posBits uint64) string {
	s := "{"
	for posBits != 0 {
		pos := uint8(bits.TrailingZeros64(posBits))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		posBits = posBits ^ posBit

		s += " " + posToString(pos)
	}
	s += "}"
	return s
}

// Assumes nothing is white/black
func posColor(board *dragon.Board, pos uint8) dragon.ColorT {
	posBit := uint64(1) << uint(pos)
	if board.Bbs[dragon.White][dragon.All] & posBit == 0 {
		return dragon.White
	} else {
		return dragon.Black
	}
}

func posToString(pos uint8) string {
	return files[pos&7] + ranks[(pos >> 3)&7]
}

func DontTestPosEval(t *testing.T) {
	fmt.Println("Testing Pos Eval pos 0 is", posToString(0), "pos 63 is", posToString(63))

	fen := "r2q1rk1/2p1bppp/p2p1n2/1p2P3/4P1b1/1nP1BN2/PP3PPP/RN1QR1K1 w - - 1 12"
	board := dragon.ParseFen(fen)

	var posEval PosEvalT
	InitPosEval(&board, &posEval)

	fmt.Println("Fen", fen)
	fmt.Println()

	for pos := uint8(0); pos < 64; pos++ {
		posColor := posColor(&board, pos)
		pieceType := board.PieceAt(pos)
		posInfluence := posEval.influence[pos]
		fmt.Println(posToString(pos), pieceTypes[posColor][pieceType], "influence", posBitsToString(posInfluence))
	}
	
	fmt.Println()
	fmt.Println("Middle1", posBitsToString(middle1Bits))
	fmt.Println("Middle2", posBitsToString(middle2Bits))
	fmt.Println("Middle3", posBitsToString(middle3Bits))
}

func DontTestGenKingTropism(t *testing.T) {
	fmt.Println("var king2Masks = [64]uint64{")
	for pos := uint8(0); pos < 64; pos++ {
		if pos&3 == 0 {
			fmt.Printf("        ")
		}
		fmt.Printf("0x%016x, ", expandKingMask(pos))
		if pos&3 == 3 {
			fmt.Printf("\n")
		}
	}
}

func expandKingMask(pos1 uint8) uint64 {
	kingMask2 := uint64(0)
	kingMaskBits := kingMasks[pos1]

	for kingMaskBits != 0 {
		for kingMaskBits != 0 {
			pos := uint8(bits.TrailingZeros64(kingMaskBits))
			// (Could also use posBit-1 trick to clear the bit)
			posBit := uint64(1) << uint(pos)
			kingMaskBits = kingMaskBits ^ posBit
			
			kingMask2 |= kingMasks[pos]
		}
	}

	kingMask2 &= ^kingMasks[pos1]
	kingMask2 &= ^(uint64(1) << uint(pos1))
	return kingMask2
}

func TestKing2Mask(t *testing.T) {
	fmt.Println("King mask2", posToString(0), posBitsToString(king2Masks[0]))
	fmt.Println("King mask2", posToString(15), posBitsToString(king2Masks[15]))
	fmt.Println("King mask2", posToString(63), posBitsToString(king2Masks[63]))
	fmt.Println("King mask2", posToString(33), posBitsToString(king2Masks[33]))
	fmt.Println("King mask2", posToString(36), posBitsToString(king2Masks[36]))
}
