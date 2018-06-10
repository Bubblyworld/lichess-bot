package bot

import (
	"math"

	dragon "github.com/dylhunn/dragontoothmg"
)

func Evaluate(board dragon.Board, legalMoves []dragon.Move) float64 {
	if isStalemate(board, legalMoves) {
		return 0 // draw
	}

	if isMate(board, legalMoves) {
		if board.Wtomove {
			return -math.MaxFloat64 // black wins
		}

		return math.MaxFloat64 // white wins
	}

	score := 0

	score += 1 * countSetBits(board.White.Pawns)
	score += 3 * countSetBits(board.White.Bishops)
	score += 3 * countSetBits(board.White.Knights)
	score += 5 * countSetBits(board.White.Rooks)
	score += 8 * countSetBits(board.White.Queens)

	score -= 1 * countSetBits(board.Black.Pawns)
	score -= 3 * countSetBits(board.Black.Bishops)
	score -= 3 * countSetBits(board.Black.Knights)
	score -= 5 * countSetBits(board.Black.Rooks)
	score -= 8 * countSetBits(board.Black.Queens)

	return float64(score)
}

func countSetBits(n uint64) int {
	count := 0
	for i := uint64(0); i < 64; i++ {
		count += int(n & 0x1)
		n >>= 1
	}

	return count
}

// If there are no legal moves, there are two possibilities - either our king
// is in check, or it isn't. In the first case it's mate and we've lost, and
// in the second case it's stalemate and therefore a draw.
func isStalemate(board dragon.Board, legalMoves []dragon.Move) bool {
	return len(legalMoves) == 0 && !board.OurKingInCheck()
}

func isMate(board dragon.Board, legalMoves []dragon.Move) bool {
	return len(legalMoves) == 0 && board.OurKingInCheck()
}
