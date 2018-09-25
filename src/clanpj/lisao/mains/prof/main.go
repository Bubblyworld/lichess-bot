package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	dragon "github.com/Bubblyworld/dragontoothmg"

	"clanpj/lisao/engine"

	"github.com/pkg/profile"
)

var VersionString = "0.0eg Pichu 1" + "CPU " + runtime.GOOS + "-" + runtime.GOARCH

const Fine70Fen = "8/k7/3p4/p2P1p2/P2P1P2/8/8/K7 w - -"
const RandomFen = "r4k2/pp1b1p1Q/3pp1n1/7R/4P3/1B3q2/P1P1N3/1K6 b - - 0 1"

const NullMoveFalsePositive = "rnbqkbnr/1ppppppp/p7/8/3P4/5N2/PPP1PPPP/RNBQKB1R b KQkq d3 0 2" // dToGo 3 alpha -18 beta -17 null-eval -4 eval -20

var p = [][2]string { {"a", "b"} }

var CcrFens = [][2]string {
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

func doFen(fen string, descr string) {
	fmt.Println()
	fmt.Printf("%s [%s]\n", fen, descr)
	fmt.Println()

	board := dragon.ParseFen(fen)

	// reset the history table
	ht = make(engine.HistoryTableT)
	// reset the killer move table
	kt = emptyKt
	// reset the TT
	engine.ResetTT()
	// reset the qsearch TT
	engine.ResetQtt()
	
	uciSearch(&board, 10, 0, engine.YourCheckMateEval, engine.MyCheckMateEval)
	//fmt.Println("#nodes-d0", engine.NodesD0, "#full-width", engine.NodesD0FullWidth, "#neg", engine.NodesD0NegDiff, "#nodes-dm1", engine.NodesDM1, "Max d0/dm1 eval diff", engine.MaxD0DM1EvalDiff, "Min d0/dm1 eval diff", engine.MinD0DM1EvalDiff)
}

func main() {
	defer profile.Start().Stop()
	doFen(dragon.Startpos, "starting pos")
	//doFen(Fine70Fen)
	//doFen(RandomFen)
	for _, fenDescr := range CcrFens {
		doFen(fenDescr[0], fenDescr[1])
	}
}

// This MUST be per-search-thread but for now we're single-threaded so global is fine.
var ht engine.HistoryTableT = make(engine.HistoryTableT)

var emptyKt engine.KillerMoveTableT
var kt engine.KillerMoveTableT

// We use a shared variable using golang sync mechanisms for atomic shared operation.
// When timeOut != 0 then we bail on the search.
// The time-out is typically controled by a Timer, except when in infinite search mode,
//   or when explicitly cancelled with UCI stop command.
var timeout uint32

// Timer controlling the timeout variable
var timeoutTimer *time.Timer

// Lightweight wrapper around Lisao Search.
// Prints the results (bestmove) and various stats.
func uciSearch(board *dragon.Board, depth int, timeoutMs int, alpha engine.EvalCp, beta engine.EvalCp) {
	// Reset the timeout
	atomic.StoreUint32(&timeout, 0)

	// Time the search
	start := time.Now()

	// Search for the winning move!
	bestMove, eval, stats, finalDepth, _, _ := engine.Search2(board, ht, &kt, depth, timeoutMs, &timeout, alpha, beta)

	elapsedSecs := time.Since(start).Seconds()

	// Stop the timer in case this was an early-out return
	uciStop()

	// Eval is expected from the engine's perspective, but we generate it from white's perspective
	if board.Colortomove == dragon.Black {
		eval = -eval
	}

	stats.Dump(finalDepth)

	// TODO proper checkmate score string
	fmt.Println("info depth", finalDepth, "score cp", eval, "nodes", stats.Nodes, "time", uint64(elapsedSecs*1000), "nps", uint64(float64(stats.Nodes)/elapsedSecs), "pv", &bestMove)

	// Print the result
	fmt.Println("bestmove", &bestMove)
}

// Start the search timeout timer
func uciStartTimer(timeoutMs int) {
	if timeoutMs == 0 {
		return
	}
	// TODO - atomic!
	timeoutTimer = time.AfterFunc(time.Duration(timeoutMs)*time.Millisecond, func() { uciStop() })
}

// Explicitly stop the search by canceling the timer and setting the timeout shared memory address.
func uciStop() {
	if timeoutTimer != nil {
		// It may already have been stopped or timed out
		timeoutTimer.Stop()
		// TODO atomic!
		timeoutTimer = nil
	}

	// Notify search threads to bail
	atomic.StoreUint32(&timeout, 1)
}
