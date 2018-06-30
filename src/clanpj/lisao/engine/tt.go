// Transposition table for Main Search

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Members ordered by descending size for better packing
type TTEntryT struct {
	zobrist uint64 // Zobrist hash from dragontoothmg
	               // Could just store hi bits cos the hash index encodes the low bits implicitly which would ring the struct size to < 16 bytes
	nHits uint32
	eval EvalCp
	bestMove dragon.Move
	depthToGo int8
	isLowerBound bool // is this a cut-off value (since we use straight AB, we don't ever generate upper bounds)
}

func ttIndex(tt []TTEntryT, zobrist uint64) int {
	// Note: assumes tt size is a power of 2!!!
	return int(zobrist) & (len(tt)-1)
}

func isTTHit(entry *TTEntryT, zobrist uint64) bool {
	return entry.zobrist == zobrist
}

// Initialise a TT entry
func writeTTEntry(tt []TTEntryT, zobrist uint64, eval EvalCp, bestMove dragon.Move, depthToGo int, isLowerBound bool) {
	var entry TTEntryT // use a full struct overwrite to obliterate old data

	entry.zobrist = zobrist
	
	entry.eval = eval
	entry.bestMove = bestMove
	entry.depthToGo = int8(depthToGo)
	entry.isLowerBound = isLowerBound
	

	index := ttIndex(tt, zobrist)
	tt[index] = entry
}

// Update a TT entry
// The entry MUST be a TT hit - we're just updating the entry.
// There is policy in here, because we need to decide whether to overwrite or not when we see a new depth.
// From web sources best policy is to always choose exact eval over lower-bound regardless of depth; otherwise choose greater depth.
func updateTTEntry(entry *TTEntryT, eval EvalCp, bestMove dragon.Move, depthToGo int, isLowerBound bool) {
	// Prefer a quiesced result, even at lower depth,because it applies exactly to greater depths;
	//   otherwise pick the greater depth, or pick an exact result over a lower-bound
	updateIsBetter := false
	if !isLowerBound && entry.isLowerBound {
		// Choose exact value even if it's lower depth.
		// This is better according to web sources, since otherwise you can end up with a hash table
		//   full of bad lower-bounds that don't generate any cuts.
		updateIsBetter = true
	} else if depthToGo > int(entry.depthToGo) { // TODO maybe not if the deeper value is a lower bound and the entry is exact?
		// Greater depths cut off higher
		updateIsBetter = true
	} else if depthToGo == int(entry.depthToGo) {
		// Pick the more accurate result
		if eval > entry.eval {
			updateIsBetter = true
		}
	}

	if updateIsBetter {
		entry.eval = eval
		entry.bestMove = bestMove
		entry.depthToGo = int8(depthToGo)
		entry.isLowerBound = isLowerBound
	}
}

// Return TT hit and isExactHit, or nil
func probeTT(tt []TTEntryT, zobrist uint64, depthToGo int) (*TTEntryT, bool) {

	index := ttIndex(tt, zobrist) 
	var entry *TTEntryT = &tt[index]
	
	if isTTHit(entry, zobrist) {
		// update stats
		entry.nHits++

		// It's an exact hit if we're at the same depth.
		// Most engines are happy to use cached results from previous deeper searches,
		//   but we require precise depth match so that we retain exact behaviour parity with no TT.
		isExactHit := depthToGo == int(entry.depthToGo)

		return entry, isExactHit
	}
	return nil, false
}
