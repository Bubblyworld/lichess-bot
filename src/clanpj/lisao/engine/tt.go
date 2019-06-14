// Transposition table for Main Search

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Members ordered by descending size for better packing
type TTEntryT struct {
	// Could just store hi bits cos the hash index encodes the low bits implicitly
	zobrist uint64         // Zobrist hash from dragontoothmg
	bestMove  dragon.Move
	lbEval EvalCp
	ubEval EvalCp
	lbDepthToGoPlus1 uint8 // 0 means value is not valid
	ubDepthToGoPlus1 uint8 // 0 means value is not valid
}

// The eval for a TT entry can be exact, a lower bound, or an upper bound
type TTEvalT uint8

const (
	TTInvalid TTEvalT = iota // must be the 0 item
	TTEvalExact
	TTEvalLowerBound // from beta cut-off
	TTEvalUpperBound // from alpha cut-off
)

func ttIndex(tt []TTEntryT, zobrist uint64) int {
	// Note: assumes tt size is a power of 2!!!
	return int(zobrist) & (len(tt) - 1)
}

func isTTHit(entry *TTEntryT, zobrist uint64) bool {
	return entry.zobrist == zobrist
}

func depthToGoParity(depthToGo int) int { return depthToGo & 1 }

// Update a TT entry
// There is policy in here, because we need to decide whether to overwrite or not with different depths and eval types.
func updateTTEntry(entry *TTEntryT, eval EvalCp, bestMove dragon.Move, depthToGo int, evalType TTEvalT) {
	depthToGoPlus1 := uint8(depthToGo+1)

	// Try to update the lower-bound value
	if evalType == TTEvalExact || evalType == TTEvalLowerBound {
		// Replace if depth is greater or eval is more accurate
		if entry.lbDepthToGoPlus1 < depthToGoPlus1 || (entry.lbDepthToGoPlus1 == depthToGoPlus1 && entry.lbEval < eval) {
			// If the eval came from a null-move cut then the bestMove is NoMove - keep ther previous best move
			// Hrmmm - it's empirically WORSE to leave the previous move in place!
			if true || bestMove != NoMove {
				entry.bestMove = bestMove
			}
			entry.lbEval = eval
			entry.lbDepthToGoPlus1 = depthToGoPlus1
		}
	}

	// Try to update the lower-bound value
	if evalType == TTEvalExact || evalType == TTEvalUpperBound {
		// Replace if depth is greater or eval is more accurate
		if entry.ubDepthToGoPlus1 < depthToGoPlus1 || (entry.ubDepthToGoPlus1 == depthToGoPlus1 && eval < entry.ubEval) {
			// If the eval came from a null-move cut then the bestMove is NoMove - keep ther previous best move
			// Hrmmm - it's empirically WORSE to leave the previous move in place!
			if true || bestMove != NoMove {
				entry.bestMove = bestMove
			}
			entry.ubEval = eval
			entry.ubDepthToGoPlus1 = depthToGoPlus1
		}
	}
}

// Return a copy of the TT entry, and whether it is a hit
// We copy to avoid entry overwrite shenanigans
func probeTT(tt []TTEntryT, zobrist uint64) (TTEntryT, bool) {
	index := ttIndex(tt, zobrist)
	var entry TTEntryT = tt[index]

	return entry, isTTHit(&entry, zobrist)
}

type TtT []TTEntryT

// Initialise a TT entry
func (tt TtT) writeTTEntry(zobrist uint64, eval EvalCp, bestMove dragon.Move, depthToGo int, evalType TTEvalT) {
	var entry TTEntryT // use a full struct overwrite to obliterate old data

	// Do we already have an entry for the hash?
	oldTTEntry, isHit := probeTT(tt, zobrist)

	if isHit {
		entry = oldTTEntry
		updateTTEntry(&entry, eval, bestMove, depthToGo, evalType)
	} else {
		// initialise a new entry
		depthToGoPlus1 := uint8(depthToGo+1)
		entry.zobrist = zobrist

		if evalType == TTEvalExact || evalType == TTEvalLowerBound {
			entry.bestMove = bestMove
			entry.lbEval = eval
			entry.lbDepthToGoPlus1 = depthToGoPlus1
		}

		if evalType == TTEvalExact || evalType == TTEvalUpperBound {
			entry.bestMove = bestMove
			entry.ubEval = eval
			entry.ubDepthToGoPlus1 = depthToGoPlus1
		}
	}
	index := ttIndex(tt, zobrist)
	tt[index] = entry
}

