// Transposition table for Quiescence Search

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

type QTTBoundEntryT struct {
	eval      EvalCp
	bestMove  dragon.Move
	qDepthToGo uint8
	depthToGo uint8
 	evalType  TTEvalT
	isQuiesced bool // true iff this applies to all greater depths because there are no noisy leafs
}

// Members ordered by descending size for better packing
type QSearchTTEntryT struct {
	zobrist uint64 // Zobrist hash from dragontoothmg
	               // Could just store hi bits cos the hash index encodes the low bits implicitly which would bring the struct size to < 16 bytes
	lbEntry QTTBoundEntryT
	ubEntry QTTBoundEntryT
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
		// initialise a new entry
		entry.zobrist = zobrist
	
		if evalType == TTEvalExact || evalType == TTEvalLowerBound {
			entry.lbEntry.eval = eval
			entry.lbEntry.bestMove = bestMove
			entry.lbEntry.qDepthToGo = uint8(qDepthToGo)
			entry.lbEntry.evalType = evalType
			entry.lbEntry.isQuiesced = isQuiesced
		}

		if evalType == TTEvalExact || evalType == TTEvalUpperBound {
			entry.ubEntry.eval = eval
			entry.ubEntry.bestMove = bestMove
			entry.ubEntry.qDepthToGo = uint8(qDepthToGo)
			entry.ubEntry.evalType = evalType
			entry.ubEntry.isQuiesced = isQuiesced
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
	// Try to update the lower-bound value
	if evalType == TTEvalExact || evalType == TTEvalLowerBound {
		// Replace if depth is greater or eval is more accurate (taking isQuiesced into account)
		if (isQuiesced && (!entry.lbEntry.isQuiesced || uint8(qDepthToGo) < entry.lbEntry.qDepthToGo)) ||
			entry.lbEntry.qDepthToGo < uint8(qDepthToGo) ||
			(entry.lbEntry.qDepthToGo == uint8(qDepthToGo) && entry.lbEntry.eval < eval) {
			
			entry.lbEntry.eval = eval
			entry.lbEntry.bestMove = bestMove
			entry.lbEntry.qDepthToGo = uint8(qDepthToGo)
			entry.lbEntry.evalType = evalType
			entry.lbEntry.isQuiesced = isQuiesced
		}
	}

	// Try to update the lower-bound value
	if evalType == TTEvalExact || evalType == TTEvalUpperBound {
		// Replace if depth is greater or eval is more accurate (taking isQuiesced into account)
		if (isQuiesced && (!entry.lbEntry.isQuiesced || uint8(qDepthToGo) < entry.lbEntry.qDepthToGo)) ||
			entry.lbEntry.qDepthToGo < uint8(qDepthToGo) ||
			(entry.lbEntry.qDepthToGo == uint8(qDepthToGo) && entry.lbEntry.eval < eval) {

			entry.ubEntry.eval = eval
			entry.ubEntry.bestMove = bestMove
			entry.ubEntry.qDepthToGo = uint8(qDepthToGo)
			entry.ubEntry.evalType = evalType
			entry.ubEntry.isQuiesced = isQuiesced
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

