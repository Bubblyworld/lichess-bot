// Transposition table for Quiescence Search

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Members ordered by descending size for better packing
type QSearchTTEntryT struct {
	// Could just store hi bits cos the hash index encodes the low bits implicitly.
	zobrist uint64 // Zobrist hash from dragontoothmg
	bestMove  dragon.Move
	lbEval EvalCp
	ubEval EvalCp
	lbQDepthToGoPlus1 uint8 // 0 means value is not valid
	ubQDepthToGoPlus1 uint8 // 0 means value is not valid
	lbIsQuiesced bool // true iff this applies to all greater depths because there are no noisy leafs
	ubIsQuiesced bool // true iff this applies to all greater depths because there are no noisy leafs
}

func qttIndex(qtt []QSearchTTEntryT, zobrist uint64) int {
	// Note: assumes qtt size is a power of 2!!!
	return int(zobrist) & (len(qtt)-1)
}

func isQttHit(entry *QSearchTTEntryT, zobrist uint64) bool {
	return entry.zobrist == zobrist
}

// Initialise a QTT entry
func writeQttEntry(qtt []QSearchTTEntryT, zobrist uint64, eval EvalCp, bestMove dragon.Move, qDepthToGo int, evalType TTEvalT, isQuiesced bool) {
	var entry QSearchTTEntryT // use a full struct overwrite to obliterate old data

	// Do we already have an entry for the hash?
	oldQttEntry, isHit := probeQtt(qtt, zobrist)

	if isHit {
		entry = oldQttEntry
		updateQttEntry(&entry, eval, bestMove, qDepthToGo, evalType, isQuiesced)
	} else {
		qDepthToGoPlus1 := uint8(qDepthToGo+1)
		
		// initialise a new entry
		entry.zobrist = zobrist
	
		if evalType == TTEvalExact || evalType == TTEvalLowerBound {
			entry.bestMove = bestMove
			entry.lbEval = eval
			entry.lbQDepthToGoPlus1 = qDepthToGoPlus1
			entry.lbIsQuiesced = isQuiesced
		}

		if evalType == TTEvalExact || evalType == TTEvalUpperBound {
			entry.bestMove = bestMove
			entry.ubEval = eval
			entry.ubQDepthToGoPlus1 = qDepthToGoPlus1
			entry.ubIsQuiesced = isQuiesced
		}
	}

	index := qttIndex(qtt, zobrist)
	qtt[index] = entry
}

// Update a QTT entry
// The entry MUST be a QTT hit - we're just updating the entry.
// There is policy in here, because we need to decide whether to overwrite or not with different depths and eval types.
// TODO - tune
func updateQttEntry(entry *QSearchTTEntryT, eval EvalCp, bestMove dragon.Move, qDepthToGo int, evalType TTEvalT, isQuiesced bool) {
	qDepthToGoPlus1 := uint8(qDepthToGo+1)
	
	// Try to update the lower-bound value
	if evalType == TTEvalExact || evalType == TTEvalLowerBound {
		// Replace if depth is greater or eval is more accurate (taking isQuiesced into account)
		if (isQuiesced && (!entry.lbIsQuiesced || qDepthToGoPlus1 < entry.lbQDepthToGoPlus1)) ||
			entry.lbQDepthToGoPlus1 < qDepthToGoPlus1 ||
			(entry.lbQDepthToGoPlus1 == qDepthToGoPlus1 && entry.lbEval < eval) {
			
			entry.bestMove = bestMove
			entry.lbEval = eval
			entry.lbQDepthToGoPlus1 = qDepthToGoPlus1
			entry.lbIsQuiesced = isQuiesced
		}
	}

	// Try to update the lower-bound value
	if evalType == TTEvalExact || evalType == TTEvalUpperBound {
		// Replace if depth is greater or eval is more accurate (taking isQuiesced into account)
		if (isQuiesced && (!entry.ubIsQuiesced || qDepthToGoPlus1 < entry.ubQDepthToGoPlus1)) ||
			entry.ubQDepthToGoPlus1 < qDepthToGoPlus1 ||
			(entry.ubQDepthToGoPlus1 == qDepthToGoPlus1 &&  eval < entry.ubEval) {

			entry.bestMove = bestMove
			entry.ubEval = eval
			entry.ubQDepthToGoPlus1 = qDepthToGoPlus1
			entry.ubIsQuiesced = isQuiesced
		}
	}
}

// Return a copy of the TT entry, and whether it is a hit
// We copy to avoid entry overwrite shenanigans
func probeQtt(qtt []QSearchTTEntryT, zobrist uint64) (QSearchTTEntryT, bool) {
	index := qttIndex(qtt, zobrist) 
	var entry QSearchTTEntryT = qtt[index]

	return entry, isQttHit(&entry, zobrist)
}

