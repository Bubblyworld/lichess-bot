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

	entry.zobrist = zobrist
	
	entry.eval = eval
	entry.bestMove = bestMove
	entry.qDepthToGo = uint8(qDepthToGo)
	entry.evalType = evalType
	entry.isQuiesced = isQuiesced
	

	index := qttIndex(qtt, zobrist)
	qtt[index] = entry
}

// Update a QTT entry
// The entry MUST be a QTT hit - we're just updating the entry.
// There is policy in here, because we need to decide whether to overwrite or not when we see a new depth.
// From web sources best policy is to always choose exact eval over lower-bound regardless of depth; otherwise choose greater depth.
// For QTT we have the additional special case for evals that are fully quiesced in which case lower depth is better(!)
// TODO - no idea what best policy is for replacing lb with ub and vice-versa.
func updateQttEntry(entry *QSearchTTEntryT, eval EvalCp, bestMove dragon.Move, qDepthToGo int, evalType TTEvalT, isQuiesced bool) {
	qDepthToGo8 := uint8(qDepthToGo)
	updateIsBetter := false
	switch entry.evalType {
	case TTEvalExact:
		// Replace with exact evals of greater depth, or
		//   quiesced evals of lower depth
		updateIsBetter = evalType == TTEvalExact && (isQuiesced || qDepthToGo8 > entry.qDepthToGo)
	case TTEvalLowerBound:
		// Always replace with exact; otherwise with greater depth or higher eval
		// Never replace with upper bound(?)
		if evalType == TTEvalExact {
			updateIsBetter = true
		} else if evalType == TTEvalLowerBound {
			updateIsBetter = qDepthToGo8 > entry.qDepthToGo || eval > entry.eval
		}
	case TTEvalUpperBound:
		// Always replace with exact or lower bound(?); otherwise with greater depth or lower eval
		if evalType != TTEvalUpperBound {
			updateIsBetter = true
		} else {
			updateIsBetter = qDepthToGo8 > entry.qDepthToGo || eval < entry.eval
		}
	}

	if updateIsBetter {
		entry.eval = eval
		entry.bestMove = bestMove
		entry.qDepthToGo = uint8(qDepthToGo)
		entry.evalType = evalType
		entry.isQuiesced = isQuiesced
	}
}

// Return QTT hit and isExactHit, or nil
func probeQtt(qtt []QSearchTTEntryT, zobrist uint64, qDepthToGo int) *QSearchTTEntryT {

	index := qttIndex(qtt, zobrist) 
	var entry *QSearchTTEntryT = &qtt[index]
	
	if isQttHit(entry, zobrist) {
		// update stats
		entry.nHits++
		return entry
	}
	return nil
}

