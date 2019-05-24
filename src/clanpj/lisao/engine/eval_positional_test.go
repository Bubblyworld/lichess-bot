package engine

import (
	"fmt"
	"math/bits"
	"testing"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

var pieces = [dragon.NColors][dragon.NPieces]string { { ".", "P", "N", "B", "R", "Q", "K"}, { ".", "p", "n", "b", "r", "q", "k"}}

func printfPosBits(posBits uint64) {
	for posBits != 0 {
		pos := uint8(bits.TrailingZeros64(posBits))
		// (Could also use posBit-1 trick to clear the bit)
		posBit := uint64(1) << uint(pos)
		posBits = posBits ^ posBit
		
		fmt.Printf(" %s", dragon.IndexToAlgebraic(dragon.Square(pos)))
	}
}

func doFen(fen string, descr string) {
	fmt.Println()
	fmt.Printf("%s [%s]\n", fen, descr)
	fmt.Println()

	board := dragon.ParseFen(fen)

	var posEval PositionalEvalT
	InitPositionalEval(&board, &posEval)

	for pos := uint8(0); pos < 64; pos++ {
		color := dragon.White
		if (uint64(1) << pos) & board.Bbs[dragon.Black][dragon.All] != 0 {
			color = dragon.Black
		}
		piece := board.PieceAt(pos)
		influence := posEval.InfluenceOfPos(pos)
		influenceBehindQueen := posEval.InfluenceBehindQueenOfPos(pos)

		if piece != dragon.Nothing || influence != 0 || influenceBehindQueen != 0 {
			fmt.Printf("%d %s %s influence %016x:", pos, dragon.IndexToAlgebraic(dragon.Square(pos)), pieces[color][piece], influence)
			printfPosBits(influence)

			if influenceBehindQueen != 0 {
				fmt.Printf(" behind-queen:")
				printfPosBits(influenceBehindQueen)
			}

			fmt.Println()
		}
	}
	fmt.Println()
}

var connectedWBishopsNW = [2]string { "8/6B1/8/8/3B4/8/8/8 w KQkq - 0 1", "Bishops G7 D4"}
var disconnectedWBishopsNW = [2]string { "8/6B1/8/4N3/3B4/8/8/8 w KQkq - 0 1", "Bishops G7 D4 Knight E5"}
var connectedWBishopsAndQNW = [2]string { "8/6B1/8/8/3B4/8/1Q6/8 w KQkq - 0 1", "Bishops G7 D4 Queen B1"}
var connectedWBishopsNE = [2]string { "8/1B6/8/8/4B3/8/8/8 w KQkq - 0 1", "Bishops G2 D5"}
var connectedWBishopsAndQNE = [2]string { "8/1B6/8/3Q4/4B3/8/8/8 w KQkq - 0 1", "Bishops G2 E4 Queen D5"}

var connectedWRooksRank = [2]string { "8/1R3R2/8/8/8/8/8/8 w KQkq - 0 1", "Rooks B7 F7"}
var connectedWRooksFile = [2]string { "8/5R2/8/8/8/5R2/8/8 w KQkq - 0 1", "Rooks F7 F3"}
var connectedWRooksAndQRank = [2]string { "8/1RQ2R2/8/8/8/8/8/8 w KQkq - 0 1", "Rooks B7 F7 Queen B6"}

var connectedWQueensNW = [2]string { "8/6Q1/8/8/3Q4/8/8/8 w KQkq - 0 1", "Queens G7 D4"}

var connectedWBishopAndPawnNW = [2]string { "8/6P1/8/8/3B4/8/8/8 w KQkq - 0 1", "Bishop D4 Pawn G7"}
var connectedWBishopAndPawnNE = [2]string { "8/8/8/2P5/3B4/8/8/8 w KQkq - 0 1", "Bishop D4 Pawn C5"}

var notconnectedBBishopAndPawnNE = [2]string { "8/8/8/2p5/3b4/8/8/8 w KQkq - 0 1", "Black Bishop D4 Pawn C5"}
var connectedBBishopAndPawnNW = [2]string { "8/8/8/8/3b4/2p5/8/8 w KQkq - 0 1", "Black Bishop D4 Pawn C3"}

var connectedWBishopAndPawnNWEdge = [2]string { "8/8/7P/8/8/8/3B4/8 w KQkq - 0 1", "Bishop D2 Pawn H6"}
var connectedBBishopAndPawnNWEdge = [2]string { "8/8/8/8/1b6/p7/8/8 w KQkq - 0 1", "Black Bishop B4 Pawn A3"}

var ccrFens = [][2]string {
	{"rn1qkb1r/pp2pppp/5n2/3p1b2/3P4/2N1P3/PP3PPP/R1BQKBNR w KQkq - 0 1", "id 'CCR01'; bm Qb3"},
	{"rn1qkb1r/pp2pppp/5n2/3p1b2/3P4/1QN1P3/PP3PPP/R1B1KBNR b KQkq - 1 1", "id 'CCR02';bm Bc8"},
	{"r1bqk2r/ppp2ppp/2n5/4P3/2Bp2n1/5N1P/PP1N1PP1/R2Q1RK1 b kq - 1 10", "id 'CCR03'; bm Nh6; am Ne5"},
	{"r1bqrnk1/pp2bp1p/2p2np1/3p2B1/3P4/2NBPN2/PPQ2PPP/1R3RK1 w - - 1 12", "id 'CCR04'; bm b4"},
	{"rnbqkb1r/ppp1pppp/5n2/8/3PP3/2N5/PP3PPP/R1BQKBNR b KQkq - 3 5", "id 'CCR05'; bm e5"}, 
	{"rnbq1rk1/pppp1ppp/4pn2/8/1bPP4/P1N5/1PQ1PPPP/R1B1KBNR b KQ - 1 5", "id 'CCR06'; bm Bcx3+"},
	{"r4rk1/3nppbp/bq1p1np1/2pP4/8/2N2NPP/PP2PPB1/R1BQR1K1 b - - 1 12", "id 'CCR07'; bm Rfb8"},
	{"rn1qkb1r/pb1p1ppp/1p2pn2/2p5/2PP4/5NP1/PP2PPBP/RNBQK2R w KQkq c6 1 6", "id 'CCR08'; bm d5"},
	{"r1bq1rk1/1pp2pbp/p1np1np1/3Pp3/2P1P3/2N1BP2/PP4PP/R1NQKB1R b KQ - 1 9", "id 'CCR09'; bm Nd4"},
	{"rnbqr1k1/1p3pbp/p2p1np1/2pP4/4P3/2N5/PP1NBPPP/R1BQ1RK1 w - - 1 11", "id 'CCR10'; bm a4"},
	{"rnbqkb1r/pppp1ppp/5n2/4p3/4PP2/2N5/PPPP2PP/R1BQKBNR b KQkq f3 1 3", "id 'CCR11'; bm d5"},
	{"r1bqk1nr/pppnbppp/3p4/8/2BNP3/8/PPP2PPP/RNBQK2R w KQkq - 2 6", "id 'CCR12'; bm Bxf7+"},
	{"rnbq1b1r/ppp2kpp/3p1n2/8/3PP3/8/PPP2PPP/RNBQKB1R b KQ d3 1 5", "id 'CCR13'; am Ne4"}, 
	{"rnbqkb1r/pppp1ppp/3n4/8/2BQ4/5N2/PPP2PPP/RNB2RK1 b kq - 1 6", "id 'CCR14'; am Nxc4"},
	{"r2q1rk1/2p1bppp/p2p1n2/1p2P3/4P1b1/1nP1BN2/PP3PPP/RN1QR1K1 w - - 1 12", "id 'CCR15'; bm exf6"},
	{"r1bqkb1r/2pp1ppp/p1n5/1p2p3/3Pn3/1B3N2/PPP2PPP/RNBQ1RK1 b kq - 2 7", "id 'CCR16'; bm d5"},
	{"r2qkbnr/2p2pp1/p1pp4/4p2p/4P1b1/5N1P/PPPP1PP1/RNBQ1RK1 w kq - 1 8", "id 'CCR17'; am hxg4"},
	{"r1bqkb1r/pp3ppp/2np1n2/4p1B1/3NP3/2N5/PPP2PPP/R2QKB1R w KQkq e6 1 7", "id 'CCR18'; bm Bxf6+"},
	{"rn1qk2r/1b2bppp/p2ppn2/1p6/3NP3/1BN5/PPP2PPP/R1BQR1K1 w kq - 5 10", "id 'CCR19'; am Bxe6"},
	{"r1b1kb1r/1pqpnppp/p1n1p3/8/3NP3/2N1B3/PPP1BPPP/R2QK2R w KQkq - 3 8", "id 'CCR20'; am Ndb5"},
	{"r1bqnr2/pp1ppkbp/4N1p1/n3P3/8/2N1B3/PPP2PPP/R2QK2R b KQ - 2 11", "id 'CCR21'; am Kxe6"},
	{"r3kb1r/pp1n1ppp/1q2p3/n2p4/3P1Bb1/2PB1N2/PPQ2PPP/RN2K2R w KQkq - 3 11", "id 'CCR22'; bm a4"},
	{"r1bq1rk1/pppnnppp/4p3/3pP3/1b1P4/2NB3N/PPP2PPP/R1BQK2R w KQ - 3 7", "id 'CCR23'; bm Bxh7+"},
	{"r2qkbnr/ppp1pp1p/3p2p1/3Pn3/4P1b1/2N2N2/PPP2PPP/R1BQKB1R w KQkq - 2 6", "id 'CCR24'; bm Nxe5"},
	{"rn2kb1r/pp2pppp/1qP2n2/8/6b1/1Q6/PP1PPPBP/RNB1K1NR b KQkq - 1 6", "id 'CCR25'; am Qxb3"}}

func TestInfluence(t *testing.T) {
	//doFen(connectedWBishopsNW[0], connectedWBishopsNW[1])
	//doFen(disconnectedWBishopsNW[0], disconnectedWBishopsNW[1])
	//doFen(connectedWBishopsAndQNW[0], connectedWBishopsAndQNW[1])
	//doFen(connectedWBishopsNE[0], connectedWBishopsNE[1])
	//doFen(connectedWBishopsAndQNE[0], connectedWBishopsAndQNE[1])
	
	//doFen(connectedWRooksRank[0], connectedWRooksRank[1])
	//doFen(connectedWRooksFile[0], connectedWRooksFile[1])
	//doFen(connectedWRooksAndQRank[0], connectedWRooksAndQRank[1])

	//doFen(connectedWQueensNW[0], connectedWQueensNW[1])

	//doFen(connectedWBishopAndPawnNW[0], connectedWBishopAndPawnNW[1])
	//doFen(connectedWBishopAndPawnNE[0], connectedWBishopAndPawnNE[1])
	//doFen(notconnectedBBishopAndPawnNE[0], notconnectedBBishopAndPawnNE[1])
	//doFen(connectedBBishopAndPawnNW[0], connectedBBishopAndPawnNW[1])
	//doFen(connectedWBishopAndPawnNWEdge[0], connectedWBishopAndPawnNWEdge[1])
	doFen(connectedBBishopAndPawnNWEdge[0], connectedBBishopAndPawnNWEdge[1])
	
	//doFen(ccrFens[0][0], ccrFens[0][1])
}
