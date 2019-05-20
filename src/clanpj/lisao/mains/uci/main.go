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

var VersionString = "0.0POS2.4.2pa-ad Kung Pow " + "CPU " + runtime.GOOS + "-" + runtime.GOARCH

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
			fmt.Println("option name HeurUseNullMove type check default", engine.HeurUseNullMove)
			fmt.Println("option name UseMoveOrdering type check default", engine.UseMoveOrdering)
			fmt.Println("option name UseIDMoveHint type check default", engine.UseIDMoveHint)
			fmt.Println("option name MinIDMoveHintDepth type spin default", engine.MinIDMoveHintDepth, "min 2 max 1024")
			fmt.Println("option name UseTT type check default", engine.UseTT)
			fmt.Println("option name HeurUseTTDeeperHits type check default", engine.HeurUseTTDeeperHits)
			fmt.Println("option name UsePosRepetition type check default", engine.UsePosRepetition)
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
			// reset the killer move tables
			kt = emptyKt
			qkt = emptyKt
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
			case "useposrepetition":
				switch strings.ToLower(tokens[4]) {
				case "true":
					engine.UsePosRepetition = true
				case "false":
					engine.UsePosRepetition = false
				default:
					fmt.Println("info string Unrecognised UsePosRepetition option:", tokens[4])
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
					if board.Colortomove == dragon.White {
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
		case "getevalconfig": // secret parameters for evaluation
			configParams := engine.GetConfigParams()
			configParamStrings := make([]string, len(configParams))
			for i := 0; i < len(configParams); i++ {
				param := configParams[i]
				configParamStrings[i] = fmt.Sprintf("{%s,%d,%d,%d,%d}", param.Descr, param.Min, param.Max, param.Delta, param.Get())
			}
			fmt.Printf("evalconfig %s\n", strings.Join(configParamStrings, " "))
		case "setevalconfig": // secret parameters for evaluation
			paramStrings := tokens[1:]
			params := make([]int, len(paramStrings))
			for i := 0; i < len(paramStrings); i++ {
				params[i], _ = strconv.Atoi(paramStrings[i])
			}
			engine.SetConfigParams(params)
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
			// reset the killer move table - TODO (rpj) we should really just shift it up relative to the last position
			kt = emptyKt
			qkt = emptyKt
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

var emptyKt engine.KillerMoveTableT
var kt engine.KillerMoveTableT
var qkt engine.KillerMoveTableT

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
	bestMove, eval, stats, finalDepth, pvLine, _ := engine.Search(board, ht, &kt, &qkt,  depth, timeoutMs, &timeout)

	elapsedSecs := time.Since(start).Seconds()

	// Stop the timer in case this was an early-out return
	uciStop()

	// Eval is expected from the engine's perspective, but we generate it from white's perspective
	if board.Colortomove == dragon.Black {
		eval = -eval
	}

	// Reverse order from which it appears in the UCI driver
	if engine.DumpSearchStats {
		stats.Dump(finalDepth)
	}
	// TODO proper checkmate score string
	fmt.Println("info depth", finalDepth, "score cp", eval, "nodes", stats.Nodes, "time", uint64(elapsedSecs*1000), "nps", uint64(float64(stats.Nodes)/elapsedSecs), "pv", engine.MkPvString(bestMove, pvLine))

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
