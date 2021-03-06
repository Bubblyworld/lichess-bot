// Stolen shamelessly from dragontooth

package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	dragon "github.com/Bubblyworld/dragontoothmg"

	"clanpj/lisao/engine"
)

var VersionString = "0.0mga Pichu 1" + "CPU " + runtime.GOOS + "-" + runtime.GOARCH

func main() {
	uciLoop()
}

func uciLoop() {
	scanner := bufio.NewScanner(os.Stdin)
	board := dragon.ParseFen(dragon.Startpos) // the game board
	// used for communicating with search routine
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		if len(tokens) == 0 { // ignore blank lines
			continue
		}
		switch strings.ToLower(tokens[0]) {
		case "uci":
			fmt.Println("id name Lisao", VersionString)
			fmt.Println("id author Clan PJ")
			fmt.Println("option name SearchAlgorithm type combo default", engine.SearchAlgorithmString(), "var NegAlphaBeta")
			fmt.Println("option name SearchDepth type spin default", engine.SearchDepth, "min 1 max 1024")
			fmt.Println("option name SearchCutoffPercent type spin default", engine.SearchCutoffPercent, "min 1 max 100")
			fmt.Println("option name TimeLeftPerMoveDivisor type spin default", TimeLeftPerMoveDivisor, "min 2 max 200")
			fmt.Println("option name UseEarlyMoveHint type check default", engine.UseEarlyMoveHint)
			fmt.Println("option name HeurUseNullMove type check default", engine.HeurUseNullMove)
			fmt.Println("option name UseMoveOrdering type check default", engine.UseMoveOrdering)
			fmt.Println("option name UseIDMoveHint type check default", engine.UseIDMoveHint)
			fmt.Println("option name MinIDMoveHintDepth type spin default", engine.MinIDMoveHintDepth, "min 2 max 1024")
			fmt.Println("option name UseTT type check default", engine.UseTT)
			fmt.Println("option name HeurUseTTDeeperHits type check default", engine.HeurUseTTDeeperHits)
			fmt.Println("option name UseKillerMoves type check default", engine.UseKillerMoves)
			fmt.Println("option name UsePosRepetition type check default", engine.UsePosRepetition)
			fmt.Println("option name UseDeepKillerMoves type check default", engine.UseDeepKillerMoves)
			fmt.Println("option name UseQSearch type check default", engine.UseQSearch)
			fmt.Println("option name QSearchDepth type spin default", engine.QSearchDepth, "min 1 max 1024")
			fmt.Println("option name UseQSearchTT type check default", engine.UseQSearchTT)
			fmt.Println("option name UseQSearchMoveOrdering type check default", engine.UseQSearchMoveOrdering)
			fmt.Println("option name UseQSearchRampagePruning type check default", engine.UseQSearchRampagePruning)
			fmt.Println("option name QSearchRampagePruningDepth type spin default", engine.QSearchRampagePruningDepth, "min 0 max 1024")
			fmt.Println("option name UseQKillerMoves type check default", engine.UseQKillerMoves)
			fmt.Println("option name UseQDeepKillerMoves type check default", engine.UseQDeepKillerMoves)
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
		case "ucinewgame":
			// reset the board, in case the GUI skips 'position' after 'newgame'
			board = dragon.ParseFen(dragon.Startpos)
			// reset the history table
			ht = make(engine.HistoryTableT)
			// reset the TT
			engine.ResetTT()
			// reset the qsearch TT
			engine.ResetQtt()

		case "quit":
			return
		case "setoption":
			if len(tokens) != 5 || tokens[1] != "name" || tokens[3] != "value" {
				fmt.Println("info string Malformed setoption command")
				continue
			}
			switch strings.ToLower(tokens[2]) {
			case "searchalgorithm":
				switch strings.ToLower(tokens[4]) {
				case "negalphabeta":
					engine.SearchAlgorithm = engine.NegAlphaBeta
				default:
					fmt.Println("info string Unrecognised Search Algorithm:", tokens[4])
				}
			case "searchdepth":
				res, err := strconv.Atoi(tokens[4])
				if err != nil {
					fmt.Println("info string SearchDepth value is not an int (", err, ")")
					continue
				}
				engine.SearchDepth = res
			case "searchcutoffpercent":
				res, err := strconv.Atoi(tokens[4])
				if err != nil {
					fmt.Println("info string SearchCutoffPercent value is not an int (", err, ")")
					continue
				}
				engine.SearchCutoffPercent = res
			case "timeleftpermovedivisor":
				res, err := strconv.Atoi(tokens[4])
				if err != nil {
					fmt.Println("info string TimeLeftPerMoveDivisor value is not an int (", err, ")")
					continue
				}
				TimeLeftPerMoveDivisor = res
			case "usemoveordering":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseMoveOrdering = true
				case "false":
					engine.UseMoveOrdering = false
				default:
					fmt.Println("info string Unrecognised UseMoveOrdering option:", tokens[4])
				}
			case "useearlymovehint":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseEarlyMoveHint = true
				case "false":
					engine.UseEarlyMoveHint = false
				default:
					fmt.Println("info string Unrecognised UseEarlyMoveHint option:", tokens[4])
				}
			case "heurusenullmove":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.HeurUseNullMove = true
				case "false":
					engine.HeurUseNullMove = false
				default:
					fmt.Println("info string Unrecognised HeurUseNullMove option:", tokens[4])
				}
			case "useidmovehint":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseIDMoveHint = true
				case "false":
					engine.UseIDMoveHint = false
				default:
					fmt.Println("info string Unrecognised UseIDMoveHint option:", tokens[4])
				}
			case "minidmovehintdepth":
				res, err := strconv.Atoi(tokens[4])
				if err != nil {
					fmt.Println("info string MinIDMoveHintDepth value is not an int (", err, ")")
					continue
				}
				engine.MinIDMoveHintDepth = res
			case "usett":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseTT = true
				case "false":
					engine.UseTT = false
				default:
					fmt.Println("info string Unrecognised UseTT option:", tokens[4])
				}
			case "heurusettdeeperhits":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.HeurUseTTDeeperHits = true
				case "false":
					engine.HeurUseTTDeeperHits = false
				default:
					fmt.Println("info string Unrecognised HeurUseTTDeeperHits option:", tokens[4])
				}
			case "usekillermoves":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseKillerMoves = true
				case "false":
					engine.UseKillerMoves = false
				default:
					fmt.Println("info string Unrecognised UseKillerMoves option:", tokens[4])
				}
			case "useposrepetition":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UsePosRepetition = true
				case "false":
					engine.UsePosRepetition = false
				default:
					fmt.Println("info string Unrecognised UsePosRepetition option:", tokens[4])
				}
			case "usedeepkillermoves":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseDeepKillerMoves = true
				case "false":
					engine.UseDeepKillerMoves = false
				default:
					fmt.Println("info string Unrecognised UseDeepKillerMoves option:", tokens[4])
				}
			case "useqsearch":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseQSearch = true
				case "false":
					engine.UseQSearch = false
				default:
					fmt.Println("info string Unrecognised UseQSearch option:", tokens[4])
				}
			case "qsearchdepth":
				res, err := strconv.Atoi(tokens[4])
				if err != nil {
					fmt.Println("info string QSearchDepth value is not an int (", err, ")")
					continue
				}
				engine.QSearchDepth = res
			case "useqsearchtt":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseQSearchTT = true
				case "false":
					engine.UseQSearchTT = false
				default:
					fmt.Println("info string Unrecognised UseQSearchTT option:", tokens[4])
				}
			case "useqkillermoves":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseQKillerMoves = true
				case "false":
					engine.UseQKillerMoves = false
				default:
					fmt.Println("info string Unrecognised UseQKillerMoves option:", tokens[4])
				}
			case "useqdeepkillermoves":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseQDeepKillerMoves = true
				case "false":
					engine.UseQDeepKillerMoves = false
				default:
					fmt.Println("info string Unrecognised UseDeepKillerMoves option:", tokens[4])
				}
			case "useqsearchmoveordering":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseQSearchMoveOrdering = true
				case "false":
					engine.UseQSearchMoveOrdering = false
				default:
					fmt.Println("info string Unrecognised UseQSearchMoveOrdering option:", tokens[4])
				}
			case "useqsearchrampagepruning":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UseQSearchRampagePruning = true
				case "false":
					engine.UseQSearchRampagePruning = false
				default:
					fmt.Println("info string Unrecognised UseQSearchRampagePruning option:", tokens[4])
				}
			case "qsearchrampagepruningdepth":
				res, err := strconv.Atoi(tokens[4])
				if err != nil {
					fmt.Println("info string QSearchRampagePruningDepth value is not an int (", err, ")")
					continue
				}
				engine.QSearchRampagePruningDepth = res
			default:
				fmt.Println("info string Unknown UCI option", tokens[2])
			}
		case "go":
			goScanner := bufio.NewScanner(strings.NewReader(line))
			goScanner.Split(bufio.ScanWords)
			goScanner.Scan() // skip the first token
			var movetime, wtime, btime, winc, binc int
			var infinite bool
			var depth int // if 0 then we're searching on time
			var err error
			for goScanner.Scan() {
				nextToken := strings.ToLower(goScanner.Text())
				switch nextToken {
				case "infinite":
					infinite = true
					continue
				case "movetime":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option movetime")
						continue
					}
					movetime, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert wtime")
						continue
					}
				case "wtime":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option wtime")
						continue
					}
					wtime, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert wtime")
						continue
					}
				case "btime":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option btime")
						continue
					}
					btime, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert btime")
						continue
					}
				case "winc":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option winc")
						continue
					}
					winc, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert winc")
						continue
					}
				case "binc":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option binc")
						continue
					}
					binc, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert binc")
						continue
					}
				case "depth":
					if !goScanner.Scan() {
						fmt.Println("info string Malformed go command option depth")
						continue
					}
					depth, err = strconv.Atoi(goScanner.Text())
					if err != nil {
						fmt.Println("info string Malformed go command option; could not convert depth")
						continue
					}
				default:
					fmt.Println("info string Unknown go subcommand", nextToken)
					continue
				}
			}

			timeoutMs := 0
			if (movetime != 0 || (wtime != 0 && btime != 0)) && !infinite { // If times are specified
				timeoutMs = movetime
				if movetime == 0 {
					var ourtime, opptime, ourinc, oppinc int
					if board.Wtomove {
						ourtime, opptime, ourinc, oppinc = wtime, btime, winc, binc
					} else {
						ourtime, opptime, ourinc, oppinc = btime, wtime, binc, winc
					}
					timeoutMs = uciCalculateAllowedTimeMs(&board, ourtime, opptime, ourinc, oppinc)
				}
			}
			// Start the timeout timer...
			uciStartTimer(timeoutMs)
			// Run the search in another thread.
			go uciSearch(&board, depth, timeoutMs)
		// case "secretparam": // secret parameters used for optimizing the evaluation function
		// 	res, _ := strconv.Atoi(tokens[2])
		// 	switch tokens[1] {
		// 	case "BishopPairBonus":
		// 		eval.BishopPairBonus = res
		// 	case "DiagonalMobilityBonus":
		// 		eval.DiagonalMobilityBonus = res
		// 	case "OrthogonalMobilityBonus":
		// 		eval.OrthogonalMobilityBonus = res
		// 	case "DoubledPawnPenalty":
		// 		eval.DoubledPawnPenalty = res
		// 	case "PassedPawnBonus":
		// 		eval.PassedPawnBonus = res
		// 	case "IsolatedPawnPenalty":
		// 		eval.IsolatedPawnPenalty = res

		// 	default:
		// 		if tokens[1][0:14] == "PawnTableStart" {
		// 			idx := tokens[1][14:len(tokens[1])]
		// 			square, _ := strconv.Atoi(idx)
		// 			val, _ := strconv.Atoi(tokens[2])
		// 			eval.PawnTableStart[square] = val
		// 		} else if tokens[1][0:14] == "KingTableStart" {
		// 			idx := tokens[1][14:len(tokens[1])]
		// 			square, _ := strconv.Atoi(idx)
		// 			val, _ := strconv.Atoi(tokens[2])
		// 			eval.KingTableStart[square] = val
		// 		} else if tokens[1][0:15] == "CentralizeTable" {
		// 			idx := tokens[1][15:len(tokens[1])]
		// 			square, _ := strconv.Atoi(idx)
		// 			val, _ := strconv.Atoi(tokens[2])
		// 			eval.CentralizeTable[square] = val
		// 		} else if tokens[1][0:16] == "KnightTableStart" {
		// 			idx := tokens[1][16:len(tokens[1])]
		// 			square, _ := strconv.Atoi(idx)
		// 			val, _ := strconv.Atoi(tokens[2])
		// 			eval.KnightTableStart[square] = val
		// 		} else {
		// 			fmt.Println("Unknown secret param")
		// 		}
		// 	}
		case "stop":
			uciStop()
		case "position":
			posScanner := bufio.NewScanner(strings.NewReader(line))
			posScanner.Split(bufio.ScanWords)
			posScanner.Scan() // skip the first token
			if !posScanner.Scan() {
				fmt.Println("info string Malformed position command")
				continue
			}
			// reset the history map
			ht = make(engine.HistoryTableT)
			if strings.ToLower(posScanner.Text()) == "startpos" {
				board = dragon.ParseFen(dragon.Startpos)
				ht.Add(board.Hash()) // record that this state has occurred
				posScanner.Scan()    // advance the scanner to leave it in a consistent state
			} else if strings.ToLower(posScanner.Text()) == "fen" {
				fenstr := ""
				for posScanner.Scan() && strings.ToLower(posScanner.Text()) != "moves" {
					fenstr += posScanner.Text() + " "
				}
				if fenstr == "" {
					fmt.Println("info string Invalid fen position")
					continue
				}
				board = dragon.ParseFen(fenstr)
				ht.Add(board.Hash()) // record that this state has occurred
			} else {
				fmt.Println("info string Invalid position subcommand")
				continue
			}
			if strings.ToLower(posScanner.Text()) != "moves" {
				continue
			}
			for posScanner.Scan() { // for each move
				moveStr := strings.ToLower(posScanner.Text())
				legalMoves := board.GenerateLegalMoves()
				var nextMove dragon.Move
				found := false
				for _, mv := range legalMoves {
					if mv.String() == moveStr {
						nextMove = mv
						found = true
						break
					}
				}
				if !found { // we didn't find the move, but we will try to apply it anyway
					fmt.Println("info string Move", moveStr, "not found for position", board.ToFen())
					var err error
					nextMove, err = dragon.ParseMove(moveStr)
					if err != nil {
						fmt.Println("info string Contingency move parsing failed")
						continue
					}
				}
				board.Apply(nextMove)
				ht.Add(board.Hash()) // record that this state has occurred
			}
		default:
			fmt.Println("info string Unknown command:", line)
		}
	}
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
	bestMove, eval, stats, finalDepth, _ := engine.Search(board, ht, depth, timeoutMs, &timeout)

	elapsedSecs := time.Since(start).Seconds()

	// Stop the timer in case this was an early-out return
	uciStop()

	// Eval is expected from the engine's perspective, but we generate it from white's perspective
	if !board.Wtomove {
		eval = -eval
	}

	// Reverse order from which it appears in the UCI driver
	fmt.Println("info string   q-mates:", perC(stats.QMates, stats.QNonLeafs), "q-pat-cuts:", perC(stats.QPatCuts, stats.QNonLeafs), "q-rampage-prunes:", perC(stats.QRampagePrunes, stats.QNonLeafs), "q-killers:", perC(stats.QKillers, stats.QNonLeafs), "q-killer-cuts:", perC(stats.QKillerCuts, stats.QNonLeafs), "q-deep-killers:", perC(stats.QDeepKillers, stats.QNonLeafs), "q-deep-killer-cuts:", perC(stats.QDeepKillerCuts, stats.QNonLeafs))
	if engine.UseQSearchTT {
		fmt.Println("info string   qtt-hits:", perC(stats.QttHits, stats.QNonLeafs), "qtt-depth-hits:", perC(stats.QttDepthHits, stats.QNonLeafs), "qtt-beta-cuts:", perC(stats.QttBetaCuts, stats.QNonLeafs), "qtt-alpha-cuts:", perC(stats.QttAlphaCuts, stats.QNonLeafs), "qtt-late-cuts:", perC(stats.QttLateCuts, stats.QNonLeafs), "qtt-true-evals:", perC(stats.QttTrueEvals, stats.QNonLeafs))
	}
	fmt.Print("info string    q-non-leafs by depth:")
	for i := 0; i < engine.MaxQDepthStats && i < engine.QSearchDepth; i++ {
		fmt.Printf(" %d: %s", i, perC(stats.QNonLeafsAt[i], stats.QNonLeafs))
	}
	fmt.Println()
	fmt.Println("info string q-nodes:", stats.QNodes, "q-non-leafs:", stats.QNonLeafs, "q-all-nodes:", perC(stats.QAllChildrenNodes, stats.QNonLeafs), "q-1st-child-cuts:", perC(stats.QFirstChildCuts, stats.QNonLeafs), "q-pats:", perC(stats.QPats, stats.QNonLeafs), "q-quiesced:", perC(stats.QQuiesced, stats.QNonLeafs), "q-prunes:", perC(stats.QPrunes, stats.QNonLeafs))
	fmt.Println("info string   null-cuts:", perC(stats.NullMoveCuts, stats.NonLeafs), "valid-hint-moves:", perC(stats.ValidHintMoves, stats.NonLeafs), "hint-move-cuts:", perC(stats.HintMoveCuts, stats.NonLeafs), "mates:", perC(stats.Mates, stats.NonLeafs), "killers:", perC(stats.Killers, stats.NonLeafs), "killer-cuts:", perC(stats.KillerCuts, stats.NonLeafs), "deep-killers:", perC(stats.DeepKillers, stats.NonLeafs), "deep-killer-cuts:", perC(stats.DeepKillerCuts, stats.NonLeafs))
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

// 1/16th of the time left per move seems aggressive, but we bail early most of the time due to SearchCutoffPercent
var TimeLeftPerMoveDivisor = 16

// Simple strategy - use fixed percentage of the remaining time
func uciCalculateAllowedTimeMs(b *dragon.Board, ourtimeMs int, opptimeMs int, ourincMs int, oppincMs int) int {
	result := ourtimeMs / TimeLeftPerMoveDivisor
	if result <= 0 {
		return ourincMs
	}
	return result
}
