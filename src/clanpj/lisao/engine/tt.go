// Transposition table for Main Search

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

type TTBoundEntryT struct {
	eval      EvalCp
	bestMove  dragon.Move
	depthToGo uint8
 	evalType  TTEvalT
}

type TTParityEntryT struct {
	lbEntry TTBoundEntryT
	ubEntry TTBoundEntryT
}

// Members ordered by descending size for better packing
type TTEntryT struct {
	// Could just store hi bits cos the hash index encodes the low bits implicitly
	zobrist uint64         // Zobrist hash from dragontoothmg
	parityHits [2]TTParityEntryT
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

// Initialise a TT entry
func writeTTEntry(tt []TTEntryT, zobrist uint64, eval EvalCp, bestMove dragon.Move, depthToGo int, evalType TTEvalT) {
	var entry TTEntryT // use a full struct overwrite to obliterate old data

	// Do we already have an entry for the hash?
	oldTTEntry, isHit := probeTT(tt, zobrist)

	if isHit {
		entry = oldTTEntry
		updateTTEntry(&entry, eval, bestMove, depthToGo, evalType)
	} else {
		// initialise a new entry
		entry.zobrist = zobrist

		pEntry := &entry.parityHits[depthToGoParity(depthToGo)]

		if evalType == TTEvalExact || evalType == TTEvalLowerBound {
			pEntry.lbEntry.eval = eval
			pEntry.lbEntry.bestMove = bestMove
			pEntry.lbEntry.depthToGo = uint8(depthToGo)
			pEntry.lbEntry.evalType = evalType
		}

		if evalType == TTEvalExact || evalType == TTEvalUpperBound {
			pEntry.ubEntry.eval = eval
			pEntry.ubEntry.bestMove = bestMove
			pEntry.ubEntry.depthToGo = uint8(depthToGo)
			pEntry.ubEntry.evalType = evalType
		}
	}
	index := ttIndex(tt, zobrist)
	tt[index] = entry
}

// Update a TT entry
// There is policy in here, because we need to decide whether to overwrite or not with different depths and eval types.
// TODO - tune
func updateTTEntry(entry *TTEntryT, eval EvalCp, bestMove dragon.Move, depthToGo int, evalType TTEvalT) {
	depthToGo8 := uint8(depthToGo)
	pEntry := &entry.parityHits[depthToGoParity(depthToGo)]

	// Try to update the lower-bound value
	if evalType == TTEvalExact || evalType == TTEvalLowerBound {
		// Replace if depth is greater or eval is more accurate
		if pEntry.lbEntry.depthToGo < depthToGo8 || (pEntry.lbEntry.depthToGo == depthToGo8 && pEntry.lbEntry.eval < eval) {
			pEntry.lbEntry.eval = eval
			pEntry.lbEntry.bestMove = bestMove
			pEntry.lbEntry.depthToGo = depthToGo8
			pEntry.lbEntry.evalType = evalType
		}
	}

	// Try to update the lower-bound value
	if evalType == TTEvalExact || evalType == TTEvalUpperBound {
		// Replace if depth is greater or eval is more accurate
		if pEntry.ubEntry.depthToGo < depthToGo8 || (pEntry.ubEntry.depthToGo == depthToGo8 && eval < pEntry.ubEntry.eval) {
			pEntry.ubEntry.eval = eval
			pEntry.ubEntry.bestMove = bestMove
			pEntry.ubEntry.depthToGo = depthToGo8
			pEntry.ubEntry.evalType = evalType
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
