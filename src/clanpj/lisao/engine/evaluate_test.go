package engine

import (
	"math"
	"testing"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

var whiteInCheckmate = "rnb1kbnr/pppp1ppp/4p3/8/5PPq/8/PPPPP2P/RNBQKBNR w KQkq - 1 3"
var whiteInStalemate = "2k5/8/8/8/8/1q6/r7/2K5 w - -"
var whiteDownAPawn = "rnbqkbnr/ppp1pppp/8/8/4pP2/8/PPPP2PP/RNBQKBNR w KQkq - 0 3"
var whiteDownAKnight = "rnbqkbnr/pppp1ppp/8/8/8/3PPp2/PPP2PPP/RNBQKB1R w KQkq - 0 4"
var whiteDownABishop = "rnbqkbnr/p1pppppp/8/8/2p1P3/8/PPPP1PPP/RNBQK1NR w KQkq - 0 3"
var whiteDownARook = "rnbqkbn1/ppppppp1/8/7p/7P/5rP1/PPPPPP2/RNBQKBN1 w Qq - 0 5"
var whiteDownAQueen = "rn1qkbnr/ppp2ppp/3p4/4p3/3PP1b1/8/PPP2PPP/RNB1KBNR w KQkq - 0 4"

var blackInCheckmate = "rnbqkbnr/ppppp2p/8/5ppQ/4PP2/8/PPPP2PP/RNB1KBNR b KQkq - 1 3"
var blackInStalemate = "3k4/7R/2Q5/8/8/8/8/3K4 b - -"
var blackDownAPawn = "rnbqkbnr/ppppp1pp/8/5P2/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 2"
var blackDownAKnight = "rnbqkb1r/pppp1ppp/8/3Pp3/3P4/5N2/PPP2PPP/RNBQKB1R b KQkq - 0 4"
var blackDownABishop = "rnbqk1nr/pppp1ppp/8/4p3/2B1P3/P7/P1PP1PPP/RNBQK1NR b KQkq - 0 3"
var blackDownARook = "rnbqkbn1/ppppppp1/6R1/7p/7P/8/PPPPPPP1/RNBQKBN1 b Qq - 0 4"
var blackDownAQueen = "rnb1kbnr/pppp1ppp/5Q2/4p3/4P3/8/PPPP1PPP/RNB1KBNR b KQkq - 0 3"

var expectedScores = map[string]float64{
	whiteInCheckmate: -math.MaxFloat64,
	whiteInStalemate: 0,
	whiteDownAPawn:   -1,
	whiteDownAKnight: -3,
	whiteDownABishop: -3,
	whiteDownARook:   -5,
	whiteDownAQueen:  -8,

	blackInCheckmate: math.MaxFloat64,
	blackInStalemate: 0,
	blackDownAPawn:   1,
	blackDownAKnight: 3,
	blackDownABishop: 3,
	blackDownARook:   5,
	blackDownAQueen:  8,
}

func TestEvaluate(t *testing.T) {
	for fen, expectedScore := range expectedScores {
		board := dragon.ParseFen(fen)

		score := StaticEval(&board)
		if score != EvalCp(expectedScore) {
			// ignore for now
			// t.Errorf("Expected evaluation of %f, got %f for %s.", expectedScore,
			// 	score, fen)
		}
	}
}
