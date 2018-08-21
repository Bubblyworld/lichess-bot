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

func main() {
	defer profile.Start().Stop()
	fmt.Println("Starting...")
	board := dragon.ParseFen(dragon.Startpos) // the game board
	uciSearch(&board, 10, 0)
}

func perC(n uint64, N uint64) string {
	return fmt.Sprintf("%d [%.2f%%]", n, float64(n)/float64(N)*100)
}

// This MUST be per-search-thread but for now we're single-threaded so global is fine.
var ht engine.HistoryTableT = make(engine.HistoryTableT)

// We use a shared variable using golang sync mechanisms for atomic shared operation.
// When timeOut != 0 then we bail on the search.
// The time-out is typically controled by a Timer, except when in infinite search mode,
//   or when explicitly cancelled with UCI stop command.
var timeout uint32

// Timer controlling the timeout variable
var timeoutTimer *time.Timer

// Lightweight wrapper around Lisao Search.
// Prints the results (bestmove) and various stats.
func uciSearch(board *dragon.Board, depth int, timeoutMs int) {
	// Reset the timeout
	atomic.StoreUint32(&timeout, 0)

	// Time the search
	start := time.Now()

	// Search for the winning move!
	bestMove, eval, stats, finalDepth, _, _ := engine.Search(board, ht, depth, timeoutMs, &timeout)

	elapsedSecs := time.Since(start).Seconds()

	// Stop the timer in case this was an early-out return
	uciStop()

	// Eval is expected from the engine's perspective, but we generate it from white's perspective
	if board.Colortomove == dragon.Black {
		eval = -eval
	}

	// Reverse order from which it appears in the UCI driver
	fmt.Println("info string   "/*q-moves:", stats.QMoves, "q-simple-moves:", perC(stats.QSimpleMoves, stats.QMoves), "q-simple-captures:", perC(stats.QSimpleCaptures, stats.QMoves), "q-nomoves:", stats.QNoMoves, "q-1moves:", stats.Q1Move, "q-movegens:", perC(stats.QMoveGens, stats.QNonLeafs)*/, "q-mates:", perC(stats.QMates, stats.QNonLeafs), "q-pat-cuts:", perC(stats.QPatCuts, stats.QNonLeafs), "q-rampage-prunes:", perC(stats.QRampagePrunes, stats.QNonLeafs), "q-killers:", perC(stats.QKillers, stats.QNonLeafs), "q-killer-cuts:", perC(stats.QKillerCuts, stats.QNonLeafs), "q-deep-killers:", perC(stats.QDeepKillers, stats.QNonLeafs), "q-deep-killer-cuts:", perC(stats.QDeepKillerCuts, stats.QNonLeafs))
	// if engine.UseEarlyMoveHint {
	// 	fmt.Println("info string   mv-all:", stats.MVAll, "mv-non-king:", perC(stats.MVNonKing, stats.MVAll), "mv-ours:", perC(stats.MVOurPiece, stats.MVAll), "mv-pawn:", perC(stats.MVPawn, stats.MVAll), "mv-pawn-push:", perC(stats.MVPawnPush, stats.MVPawn), "mv-pp-ok:", perC(stats.MVPawnPushOk, stats.MVPawnPush), "mv-pawnok:", perC(stats.MVPawnOk, stats.MVPawn), "mv-nonpawn:", perC(stats.MVNonPawn, stats.MVAll), "mv-nonpawn-ok:", perC(stats.MVNonPawnOk, stats.MVNonPawn), "mv-disc0:", perC(stats.MVDisc0, stats.MVAll), "mv-disc1:", perC(stats.MVDisc1, stats.MVAll), "mv-disc2:", perC(stats.MVDisc2, stats.MVAll), "mv-disc-no:", perC(stats.MVDiscMaybe, stats.MVAll))
	// }
	if engine.UseQSearchTT {
		fmt.Println("info string   qtt-hits:", perC(stats.QttHits, stats.QNonLeafs), "qtt-depth-hits:", perC(stats.QttDepthHits, stats.QNonLeafs), "qtt-beta-cuts:", perC(stats.QttBetaCuts, stats.QNonLeafs), "qtt-alpha-cuts:", perC(stats.QttAlphaCuts, stats.QNonLeafs), "qtt-late-cuts:", perC(stats.QttLateCuts, stats.QNonLeafs), "qtt-true-evals:", perC(stats.QttTrueEvals, stats.QNonLeafs))
	}
	fmt.Print("info string    q-non-leafs by depth:")
	for i := 0; i < engine.MaxQDepthStats && i < engine.QSearchDepth; i++ {
		fmt.Printf(" %d: %s", i, perC(stats.QNonLeafsAt[i], stats.QNonLeafs))
	}
	fmt.Println()
	fmt.Println("info string q-nodes:", stats.QNodes, "q-non-leafs:", stats.QNonLeafs, "q-all-nodes:", perC(stats.QAllChildrenNodes, stats.QNonLeafs), "q-1st-child-cuts:", perC(stats.QFirstChildCuts, stats.QNonLeafs), "q-pats:", perC(stats.QPats, stats.QNonLeafs), "q-quiesced:", perC(stats.QQuiesced, stats.QNonLeafs), "q-prunes:", perC(stats.QPrunes, stats.QNonLeafs))
	fmt.Println()
	fmt.Println("info string   "/*moves:", stats.Moves, "simple-moves:", perC(stats.SimpleMoves, stats.Moves), "simple-captures:", perC(stats.SimpleCaptures, stats.Moves), "move-gens:", perC(stats.MoveGens, stats.NonLeafs)*/, "null-cuts:", perC(stats.NullMoveCuts, stats.NonLeafs), "valid-hint-moves:", perC(stats.ValidHintMoves, stats.NonLeafs), /*"early-killers:", perC(stats.EarlyKillers, stats.NonLeafs), "valid-early-killers:", perC(stats.ValidEarlyKillers, stats.NonLeafs),*/ "hint-move-cuts:", perC(stats.HintMoveCuts, stats.NonLeafs), "mates:", perC(stats.Mates, stats.NonLeafs), "killers:", perC(stats.Killers, stats.NonLeafs), "killer-cuts:", perC(stats.KillerCuts, stats.NonLeafs), "deep-killers:", perC(stats.DeepKillers, stats.NonLeafs), "deep-killer-cuts:", perC(stats.DeepKillerCuts, stats.NonLeafs))
	if engine.UseTT {
		fmt.Println("info string   tt-hits:", perC(stats.TTHits, stats.NonLeafs), "tt-depth-hits:", perC(stats.TTDepthHits, stats.NonLeafs), "tt-deeper-hits:", perC(stats.TTDeeperHits, stats.NonLeafs), "tt-beta-cuts:", perC(stats.TTBetaCuts, stats.NonLeafs), "tt-alpha-cuts:", perC(stats.TTAlphaCuts, stats.NonLeafs), "tt-late-cuts:", perC(stats.TTLateCuts, stats.NonLeafs), "tt-true-evals:", perC(stats.TTTrueEvals, stats.NonLeafs))
	}
	fmt.Print("info string    1st-child-cuts by depth:")
	for i := 0; i < engine.MaxDepthStats && i < finalDepth; i++ {
		fmt.Printf(" %d: %s", i, perC(stats.FirstChildCutsAt[i], stats.NonLeafsAt[i]))
	}
	fmt.Println()
	fmt.Print("info string    non-leafs by depth:")
	for i := 0; i < engine.MaxDepthStats && i < finalDepth; i++ {
		fmt.Printf(" %d: %s", i, perC(stats.NonLeafsAt[i], stats.NonLeafs))
	}
	fmt.Println()
	fmt.Println("info string nodes:", stats.Nodes, "non-leafs:", stats.NonLeafs, "all-nodes:", perC(stats.AllChildrenNodes, stats.NonLeafs), "1st-child-cuts:", perC(stats.FirstChildCuts, stats.NonLeafs), "pos-repetitions:", perC(stats.PosRepetitions, stats.Nodes))
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
