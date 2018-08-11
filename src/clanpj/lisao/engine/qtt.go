// Transposition table for Quiescence Search

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

// Members ordered by descending size for better packing
type QSearchTTEntryT struct {
	zobrist uint64 // Zobrist hash from dragontoothmg
	               // Could just store hi bits cos the hash index encodes the low bits implicitly which would bring the struct size to < 16 bytes
	nHits uint32
	eval EvalCp
	bestMove dragon.Move
	qDepthToGo uint8
	evalType TTEvalT // is this a cut-off value (since we use straight AB, we don't ever generate upper bounds)
	isQuiesced bool // true iff this applies to all greater depths because there are no noisy leafs
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
	
		entry.eval = eval
		entry.bestMove = bestMove
		entry.qDepthToGo = uint8(qDepthToGo)
		entry.evalType = evalType
		entry.isQuiesced = isQuiesced
	}

	index := qttIndex(qtt, zobrist)
	qtt[index] = entry
}

// Replacement policy that deeper is always better.
// return true iff the new eval should replace the qtt entry
func qevalIsBetter(entry *QSearchTTEntryT, eval EvalCp, qDepthToGo8 uint8, evalType TTEvalT, isQuiesced bool) bool {
	// Prefer quiesced evals of lower depths because they apply exactly to greater depths
	if qDepthToGo8 > entry.qDepthToGo || (isQuiesced && qDepthToGo8 <= entry.qDepthToGo)  {
		return true
	} else if qDepthToGo8 == entry.qDepthToGo {
		// Replace same depth only if the value is more accurate
		switch entry.evalType {
		case TTEvalLowerBound:
			// Always replace with exact; otherwise with higher eval
			// Never replace with upper bound(?)
			if evalType == TTEvalExact || eval > entry.eval {
				return true
			}
		case TTEvalUpperBound:
			// Always replace with exact or lower bound(?); otherwise with lower eval
			if evalType != TTEvalUpperBound || eval < entry.eval {
				return true
			}
		}
	}
	return false
}

// Replacement policy that chooses exact eval over lower-bound regardless of depth; otherwise choose greater depth.
// There was a source on the web shat suggested that this is better tha n straight depth-is-better because the latter
//   ends up with poor but deep bounds values that are no help near the leaves.
// return true iff the new eval should replace the qtt entry
func qevalIsBetter2(entry *QSearchTTEntryT, eval EvalCp, qDepthToGo8 uint8, evalType TTEvalT, isQuiesced bool) bool {
	switch entry.evalType {
	case TTEvalExact:
		// Replace with exact evals of greater depth, or quiesced evals of lower depth.
		return evalType == TTEvalExact && (isQuiesced || qDepthToGo8 > entry.qDepthToGo)
	case TTEvalLowerBound:
		// Always replace with exact; otherwise with greater depth or higher eval
		// Never replace with upper bound(?)
		if evalType == TTEvalExact {
			return true
		} else if evalType == TTEvalLowerBound {
			return qDepthToGo8 > entry.qDepthToGo || eval > entry.eval
		}
	case TTEvalUpperBound:
		// Always replace with exact or lower bound(?); otherwise with greater depth or lower eval
		if evalType != TTEvalUpperBound {
			return true
		} else {
			return qDepthToGo8 > entry.qDepthToGo || eval < entry.eval
		}
	}
	return false // unreachable
}

// Update a QTT entry
// The entry MUST be a QTT hit - we're just updating the entry.
// There is policy in here, because we need to decide whether to overwrite or not with different depths and eval types.
// TODO - tune
func updateQttEntry(entry *QSearchTTEntryT, eval EvalCp, bestMove dragon.Move, qDepthToGo int, evalType TTEvalT, isQuiesced bool) {
	qDepthToGo8 := uint8(qDepthToGo)

	if qevalIsBetter(entry, eval, qDepthToGo8, evalType, isQuiesced) {
		entry.eval = eval
		entry.bestMove = bestMove
		entry.qDepthToGo = uint8(qDepthToGo)
		entry.evalType = evalType
		entry.isQuiesced = isQuiesced
	}
}

// Return a copy of the TT entry, and whether it is a hit
// We copy to avoid entry overwrite shenanigans
func probeQtt(qtt []QSearchTTEntryT, zobrist uint64) (QSearchTTEntryT, bool) {
	index := qttIndex(qtt, zobrist) 
	var entry QSearchTTEntryT = qtt[index]

	return entry, isQttHit(&entry, zobrist)
}

