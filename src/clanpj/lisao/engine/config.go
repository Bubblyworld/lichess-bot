package engine

import dragon "github.com/Bubblyworld/dragontoothmg"

const MaxDepthStats = 16
const MaxQDepthStats = 16

// Configuration options
type SearchAlgorithmT int

const (
	NegAlphaBeta SearchAlgorithmT = iota
)

var SearchAlgorithm = NegAlphaBeta
var DumpSearchStats = false
var SearchDepth = 7          // Ignored now that time control is implemented
var SearchCutoffPercent = 30 // If we've used more than this percentage of the target time then we bail on the search instead of starting a new depth
var HeurUseNullMove = true
var HeurUseLMR = true // weaker than not using?
var UseMoveOrdering = true
var UseIDMoveHint = true
var UseIDMoveHintAlways = true
var MinIDMoveHintDepth = 2
var UseTT = true
var UseDeepTT = true // worse than not using?
var HeurUseTTDeeperHits = true // true iff we embrace deeper TT results as valid (heuristic!)
var UsePosRepetition = true
var QSearchDepth = 12
var UseQSearchTT = true
var HeurUseQTTDeeperHits = true // true iff we embrace deeper QTT results as valid (heuristic!)
var UseQSearchMoveOrdering = true
var UseQSearchRampagePruning = true // only valid if UseQSearchMoveOrdering == true
var QSearchRampagePruningDepth = 4  // only valid if UseQSearchRampagePruning == true
var UseQKillerMoves = true
var UseQDeepKillerMoves = true 

func SearchAlgorithmString() string {
	switch SearchAlgorithm {
	case NegAlphaBeta:
		return "NegAlphaBeta"
	default:
		SearchAlgorithm = NegAlphaBeta
		return "NegAlphaBeta"
	}
}

const MinDepth = 1
const MaxDepth = 254 // needs to fit in uint8 in some places (plus 1 cos 0 is used as a special value in the TT)
const NoMove dragon.Move = 0
