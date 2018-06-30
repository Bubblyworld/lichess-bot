// Transposition table for Quiescence Search

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Members ordered by descending size for better packing
type QSearchTTEntryT struct {
	zobrist uint64 // Zobrist hash from dragontoothmg
	               // Could just store hi bits cos the hash index encodes the low bits implicitly which would ring the struct size to < 16 bytes
	nHits uint32
	eval EvalCp
	bestMove dragon.Move
	qDepthToGo int8
	isLowerBound bool // is this a cut-off value (since we use straight AB, we don't ever generate upper bounds)
	isQuiesced bool // true iff this applies to all greater depths because there are no noisy leafs
}

func qttIndex(qtt []QSearchTTEntryT, zobrist uint64) int {
	// Note: assumes qtt size is a power of 2!!!
	return int(zobrist) & (len(qtt)-1)
}

func isHit(entry *QSearchTTEntryT, zobrist uint64) bool {
	return entry.zobrist == zobrist
}

// Initialise a QTT entry
func writeQttEntry(qtt []QSearchTTEntryT, zobrist uint64, eval EvalCp, bestMove dragon.Move, qDepthToGo int, isLowerBound bool, isQuiesced bool) {
	var entry QSearchTTEntryT // use a full struct overwrite to obliterate old data

	entry.zobrist = zobrist
	
	entry.eval = eval
	entry.bestMove = bestMove
	entry.qDepthToGo = int8(qDepthToGo)
	entry.isLowerBound = isLowerBound
	entry.isQuiesced = isQuiesced
	

	index := qttIndex(qtt, zobrist)
	qtt[index] = entry
}

// Update a QTT entry
// The entry MUST be a QTT hit - we're just updating the entry.
// There is policy in here, because we need to decide whether to overwrite or not when we see a new depth.
// From web sources best policy is to always choose exact eval over lower-bound regardless of depth; otherwise choose greater depth.
func updateQttEntry(entry *QSearchTTEntryT, eval EvalCp, bestMove dragon.Move, qDepthToGo int, isLowerBound bool, isQuiesced bool) {
	// Prefer a quiesced result, even at lower depth,because it applies exactly to greater depths;
	//   otherwise pick the greater depth, or pick an exact result over a lower-bound
	updateIsBetter := false
	if isQuiesced && qDepthToGo <= int(entry.qDepthToGo) {
		// Quiesced results apply to all greater depths
		updateIsBetter = true
	} else if !isLowerBound && entry.isLowerBound {
		// Choose exact value even if it's lower depth.
		// This is better according to web sources, since otherwise you can end up with a hash table
		//   full of bad lower-bounds that don't generate any cuts.
		updateIsBetter = true
	} else if qDepthToGo > int(entry.qDepthToGo) {
		// Greater depths cut off higher
		updateIsBetter = true
	} else if qDepthToGo == int(entry.qDepthToGo) {
		// Pick the more accurate result
		if eval > entry.eval {
			updateIsBetter = true
		}
	}

	if updateIsBetter {
		entry.eval = eval
		entry.bestMove = bestMove
		entry.qDepthToGo = int8(qDepthToGo)
		entry.isLowerBound = isLowerBound
		entry.isQuiesced = isQuiesced
	}
}

// Return QTT hit and isExactHit, or nil
func probeQtt(qtt []QSearchTTEntryT, zobrist uint64, qDepthToGo int) (*QSearchTTEntryT, bool) {

	index := qttIndex(qtt, zobrist) 
	var entry *QSearchTTEntryT = &qtt[index]
	
	if isHit(entry, zobrist) {
		// update stats
		entry.nHits++

		// It's an exact hit if we're at the same depth.
		// Most engines are happy to use cached results from previous deeper searches,
		//   but we require precise depth match so that we retain exact behaviour parity with no QTT.
		isExactHit := qDepthToGo == int(entry.qDepthToGo)
		//    ... or if this is a fully quiesced result
		isExactHit = isExactHit || qDepthToGo > int(entry.qDepthToGo) && entry.isQuiesced

		return entry, isExactHit
	}
	return nil, false
}
