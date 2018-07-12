// Transposition table for Main Search

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

type TTParityEntryT struct {
	eval EvalCp
	bestMove dragon.Move
	depthToGo uint8
	evalType TTEvalT
}	

// Members ordered by descending size for better packing
type TTEntryT struct {
	zobrist uint64 // Zobrist hash from dragontoothmg
	               // Could just store hi bits cos the hash index encodes the low bits implicitly which would bring the struct size to < 16 bytes
	parityHits [2]TTParityEntryT
}

// The eval for a TT entry can be exact, a lower bound, or an upper bound
type TTEvalT uint8

const (
	TTInvalid TTEvalT = iota   // must be the 0 item
	TTEvalExact
	TTEvalLowerBound           // from beta cut-off
	TTEvalUpperBound           // from alpha cut-off
)

func ttIndex(tt []TTEntryT, zobrist uint64) int {
	// Note: assumes tt size is a power of 2!!!
	return int(zobrist) & (len(tt)-1)
}

func isTTHit(entry *TTEntryT, zobrist uint64) bool {
	return entry.zobrist == zobrist
}

func depthToGoParity(depthToGo int) int { return depthToGo & 1 }

// Initialise a TT entry
func writeTTEntry(tt []TTEntryT, zobrist uint64, eval EvalCp, bestMove dragon.Move, depthToGo int, evalType TTEvalT) {
	var entry TTEntryT // use a full struct overwrite to obliterate old data

	entry.zobrist = zobrist

	pEntry := &entry.parityHits[depthToGoParity(depthToGo)]
	
	pEntry.eval = eval
	pEntry.bestMove = bestMove
	pEntry.depthToGo = uint8(depthToGo)
	pEntry.evalType = evalType
	

	index := ttIndex(tt, zobrist)
	tt[index] = entry
}

// Replacement policy that deeper is always better.
// Seems to work much better for end-games than the other more complicated policy (but possibly worse in start game)
// return true iff the new eval should replace the tt entry
func evalIsBetter(pEntry *TTParityEntryT, eval EvalCp, depthToGo8 uint8, evalType TTEvalT) bool {
	if depthToGo8 > pEntry.depthToGo {
		return true
	} else if depthToGo8 == pEntry.depthToGo {
		// Replace same depth only if the value is more accurate
		switch pEntry.evalType {
		case TTEvalLowerBound:
			// Always replace with exact; otherwise with higher eval
			// Never replace with upper bound(?)
			if evalType == TTEvalExact || eval > pEntry.eval {
				return true
			}
		case TTEvalUpperBound:
			// Always replace with exact or lower bound(?); otherwise with lower eval
			if evalType != TTEvalUpperBound || eval < pEntry.eval {
				return true
			}
		}
	}
	return false
}

// Replacement policy that chooses exact eval over lower-bound regardless of depth; otherwise choose greater depth.
// There was a source on the web shat suggested that this is better tha n straight depth-is-better because the latter
//   ends up with poor but deep bounds values that are no help near the leaves.
// return true iff the new eval should replace the tt entry
func evalIsBetter2(pEntry *TTParityEntryT, eval EvalCp, depthToGo8 uint8, evalType TTEvalT) bool {
	switch pEntry.evalType {
	case TTEvalExact:
		// Only replace with exact evals of greater depth
		return evalType == TTEvalExact && depthToGo8 > pEntry.depthToGo
	case TTEvalLowerBound:
		// Always replace with exact; otherwise with greater depth or higher eval
		// Never replace with upper bound(?)
		if evalType == TTEvalExact {
			return true
		} else if evalType == TTEvalLowerBound {
			return depthToGo8 > pEntry.depthToGo || eval > pEntry.eval
		}
	case TTEvalUpperBound:
		// Always replace with exact or lower bound(?); otherwise with greater depth or lower eval
		if evalType != TTEvalUpperBound {
			return true
		} else {
			return depthToGo8 > pEntry.depthToGo || eval < pEntry.eval
		}
	}
	return false // unreachable
}


// Update a TT entry
// There is policy in here, because we need to decide whether to overwrite or not with different depths and eval types.
// TODO - tune
func updateTTEntry(entry *TTEntryT, eval EvalCp, bestMove dragon.Move, depthToGo int, evalType TTEvalT) {
	depthToGo8 := uint8(depthToGo)
	pEntry := &entry.parityHits[depthToGoParity(depthToGo)]
	
	if evalIsBetter(pEntry, eval, depthToGo8, evalType) {
		pEntry.eval = eval
		pEntry.bestMove = bestMove
		pEntry.depthToGo = depthToGo8
		pEntry.evalType = evalType
	}
}

// Return TT hit or nil
func probeTT(tt []TTEntryT, zobrist uint64, depthToGo int) *TTEntryT {

	index := ttIndex(tt, zobrist) 
	var entry *TTEntryT = &tt[index]
	
	if isTTHit(entry, zobrist) {
		return entry
	}
	return nil
}
