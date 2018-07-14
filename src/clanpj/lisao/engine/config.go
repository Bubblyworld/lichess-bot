package engine

import dragon "github.com/Bubblyworld/dragontoothmg"

const MaxDepthStats = 16
const MaxQDepthStats = 16

// Configuration options
type SearchAlgorithmT int

const (
	MiniMax SearchAlgorithmT = iota
	NegaMax
	AlphaBeta
	NegAlphaBeta
)

var SearchAlgorithm = NegAlphaBeta
var SearchDepth = 7          // Ignored now that time control is implemented
var SearchCutoffPercent = 25 // If we've used more than this percentage of the target time then we bail on the search instead of starting a new depth
var HeurUseNullMove = true
var UseEarlyMoveHint = false // Try the hint move before doing movegen - worse until we can do early null-move heuristic (requires in-check test)
var UseMoveOrdering = true
var UseIDMoveHint = true
var MinIDMoveHintDepth = 3
var UseKillerMoves = true
var UseDeepKillerMoves = true // only valid if UseKillerMoves == true
var UseTT = true
var HeurUseTTDeeperHits = true // true iff we embrace deeper TT results as valid (heuristic!)
var UsePosRepetition = true
var UseQSearch = true
var QSearchDepth = 12
var UseQSearchTT = true
var UseQSearchMoveOrdering = true
var UseQSearchRampagePruning = true // only valid if UseQSearchMoveOrdering == true
var QSearchRampagePruningDepth = 4  // only valid if UseQSearchRampagePruning == true
var UseQKillerMoves = true
var UseQDeepKillerMoves = true // only valid if UseQKillerMoves == true

func SearchAlgorithmString() string {
	switch SearchAlgorithm {
	case MiniMax:
		return "MiniMax"
	case NegaMax:
		return "NegaMax"
	case AlphaBeta:
		return "AlphaBeta"
	case NegAlphaBeta:
		return "NegAlphaBeta"
	default:
		SearchAlgorithm = NegAlphaBeta
		return "NegAlphaBeta"
	}
}

const MinDepth = 1
const MaxDepth = 1024
const NoMove dragon.Move = 0
