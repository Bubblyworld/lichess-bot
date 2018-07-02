// Transposition table for Main Search

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Members ordered by descending size for better packing
type TTEntryT struct {
	zobrist uint64 // Zobrist hash from dragontoothmg
	               // Could just store hi bits cos the hash index encodes the low bits implicitly which would bring the struct size to < 16 bytes
	nHits uint32
	eval EvalCp
	bestMove dragon.Move
	depthToGo uint8
	evalType TTEvalT
}

// The eval for a TT entry can be exact, a lower bound, or an upper bound
type TTEvalT uint8

const (
	TTEvalExact TTEvalT = iota
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

// Initialise a TT entry
func writeTTEntry(tt []TTEntryT, zobrist uint64, eval EvalCp, bestMove dragon.Move, depthToGo int, evalType TTEvalT) {
	var entry TTEntryT // use a full struct overwrite to obliterate old data

	entry.zobrist = zobrist
	
	entry.eval = eval
	entry.bestMove = bestMove
	entry.depthToGo = uint8(depthToGo)
	entry.evalType = evalType
	

	index := ttIndex(tt, zobrist)
	tt[index] = entry
}

// Update a TT entry
// The entry MUST be a TT hit - we're just updating the entry.
// There is policy in here, because we need to decide whether to overwrite or not when we see a new depth.
// From web sources best policy is to always choose exact eval over lower-bound regardless of depth; otherwise choose greater depth.
// TODO - no idea what best policy is for replacing lb with ub and vice-versa.
func updateTTEntry(entry *TTEntryT, eval EvalCp, bestMove dragon.Move, depthToGo int, evalType TTEvalT) {
	depthToGo8 := uint8(depthToGo)
	updateIsBetter := false
	switch entry.evalType {
	case TTEvalExact:
		// Only replace with exact evals of greater depth
		updateIsBetter = evalType == TTEvalExact && depthToGo8 > entry.depthToGo
	case TTEvalLowerBound:
		// Always replace with exact; otherwise with greater depth or higher eval
		// Never replace with upper bound(?)
		if evalType == TTEvalExact {
			updateIsBetter = true
		} else if evalType == TTEvalLowerBound {
			updateIsBetter =
				depthToGo8 > entry.depthToGo ||
				eval > entry.eval
		}
	case TTEvalUpperBound:
		// Always replace with exact or lower bound(?); otherwise with greater depth or lower eval
		if evalType != TTEvalUpperBound {
			updateIsBetter = true
		} else {
			updateIsBetter =
				depthToGo8 > entry.depthToGo ||
				eval < entry.eval
		}
	}

	if updateIsBetter {
		entry.eval = eval
		entry.bestMove = bestMove
		entry.depthToGo = depthToGo8
		entry.evalType = evalType
	}
}

// Return TT hit or nil
func probeTT(tt []TTEntryT, zobrist uint64, depthToGo int) *TTEntryT {

	index := ttIndex(tt, zobrist) 
	var entry *TTEntryT = &tt[index]
	
	if isTTHit(entry, zobrist) {
		// update stats
		entry.nHits++
		return entry
	}
	return nil
}
