// Transposition table implementation

package engine

import (
	// "errors"
	// "fmt"
	// "time"

	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Members ordered by descending size for better packing
type QSearchTTEntryT struct {
	zobrist uint64 // Zobrist hash from dragontoothmg
	nHits uint32
	nBadDepthHits uint32 // hash hits where we're at a different depth
	nSadDepthHits uint32 // hash hits where we're at a different depth but we've seen this depth before
	depthHitBits uint16  // Bit set for each depth that we've seen - used to calculate whether this is a 'sad' depth miss
	eval EvalCp
	bestMove dragon.Move
	qDepthToGo int8
	isLowerBound bool // is this a cut-off value (since we use straight AB, we don't ever generate upper bounds)
	isWhiteToMove bool // sanity on the position
	isQuiesced bool // true iff this applies to all greater depths because there are no noisy moves
}

const WhiteToMoveHashMix = 100000001693 // it's big, it's prime, no idea if it's a good choice for a hash mixin

func qttIndex(qtt []QSearchTTEntryT, zobrist uint64, isWhiteToMove bool) int {
	hash := zobrist
	if isWhiteToMove {
		// Mixin with addition seems likely to be better than xor
		hash += WhiteToMoveHashMix
	}
	
	// Just do mod - let's see how it performs (~40+ cycles on modern x86)
	return int(hash % uint64(len(qtt)))
}

func isHit(entry *QSearchTTEntryT, zobrist uint64, isWhiteToMove bool) bool {
	return entry.zobrist == zobrist && entry.isWhiteToMove == isWhiteToMove
}

// Initialise a QTT entry
// There is policy in here, because we need to decide whether to overwrite or not when we see a new depth (currently pick greater depth).
func writeQttEntry(qtt []QSearchTTEntryT, zobrist uint64, isWhiteToMove bool, eval EvalCp, bestMove dragon.Move, qDepthToGo int, isLowerBound bool, isQuiesced bool) {
	var entry QSearchTTEntryT // use a full struct overwrite to obliterate old data

	entry.zobrist = zobrist
	entry.isWhiteToMove = isWhiteToMove
	
	entry.depthHitBits |= (uint16(1) << uint(qDepthToGo))
	
	entry.eval = eval
	entry.bestMove = bestMove
	entry.qDepthToGo = int8(qDepthToGo)
	entry.isLowerBound = isLowerBound
	entry.isQuiesced = isQuiesced
	

	index := qttIndex(qtt, zobrist, isWhiteToMove)
	qtt[index] = entry
}

// Update a QTT entry
// The entry MUST be a QTT hit - we're just updating the entry.
// There is policy in here, because we need to decide whether to overwrite or not when we see a new depth (currently pick greater depth).
func updateQttEntry(entry *QSearchTTEntryT, eval EvalCp, bestMove dragon.Move, qDepthToGo int, isLowerBound bool, isQuiesced bool) {
	// Note that we've seen this depth even if we don't use its results
	entry.depthHitBits |= (uint16(1) << uint(qDepthToGo))

	// Prefer a quiesced result, even at lower depth,because it applies exactly to greater depths;
	//   otherwise pick the greater depth, or pick an exact result over a lower-bound
	updateIsBetter := false
	if isQuiesced && qDepthToGo <= int(entry.qDepthToGo) {
		// Quiesced results apply to all greater depths
		updateIsBetter = true
	} else if qDepthToGo > int(entry.qDepthToGo) {
		// Greater depths cut off higher
		updateIsBetter = true
	} else if qDepthToGo == int(entry.qDepthToGo) {
		// Pick the more accurate result
		if entry.isLowerBound && !isLowerBound {
			updateIsBetter = true
		} else if eval > entry.eval {
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
func probeQtt(qtt []QSearchTTEntryT, zobrist uint64, isWhiteToMove bool, qDepthToGo int) (*QSearchTTEntryT, bool) {

	index := qttIndex(qtt, zobrist, isWhiteToMove) 
	var entry *QSearchTTEntryT = &qtt[index]
	
	if isHit(entry, zobrist, isWhiteToMove) {
		// update stats
		entry.nHits++

		// It's an exact hit if we're at the same depth.
		// Most engines are happy to use cached results from previous deeper searches,
		//   but we require precise depth match so that we retain exact behaviour parity with no QTT.
		isExactHit := qDepthToGo == int(entry.qDepthToGo)
		//    ... or if this is a fully quiesced result
		isExactHit = isExactHit || qDepthToGo > int(entry.qDepthToGo) && entry.isQuiesced

		if !isExactHit {
			entry.nBadDepthHits++
			// Have we seen this depth before (and overwritten it)?
			if entry.depthHitBits & (uint16(1) << uint(qDepthToGo)) != 0 {
				entry.nSadDepthHits++
			}
		}
		
		return entry, isExactHit
	}
	return nil, false
}
