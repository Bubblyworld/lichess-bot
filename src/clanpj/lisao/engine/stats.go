package engine

import (
	"fmt"
)

type SearchStatsT struct {
	Nodes             uint64 // #nodes visited
	Mates             uint64 // #true terminal nodes
	NonLeafs          uint64 // #non-leaf nodes
	CutNodes          uint64 // #(beta-)cut nodes
	NullMoveCuts      uint64 // #nodes that cut due to null move heuristic
	FirstChildCuts    uint64 // #non-leaf nodes that (beta-)cut on the first child searched
	CutNodeChildren   uint64 // Total #children of cut nodes (in order to see how effective the move ordering is)
	ShallowCutNodes   uint64 // #(beta-)cut nodes less than MinIDMoveHintDepth from leaf
	ShallowNullMoveCuts uint64 // #shallow nodes that cut due to null move heuristic
	ShallowCutNodeChildren uint64 // Total #children of cut nodes less than MinIDMoveHintDepth from leaf
	DeepCutNodes      uint64 // #(beta-)cut nodes at least MinIDMoveHintDepth from leaf
	DeepNullMoveCuts  uint64 // #deep nodes that cut due to null move heuristic
	DeepCutNodeChildren uint64 // Total #children of cut nodes at least MinIDMoveHintDepth from leaf
	AllChildrenNodes  uint64 // #non-leaf nodes with no beta cut
	TTMoveCuts        uint64 // #nodes with tt move cut
	TTMoveCutsNotKDK  uint64 // #nodes with tt move cut that's not the killer or deep-killer move
	PosRepetitions    uint64 // #nodes with repeated position
	TTHits            uint64 // #nodes with successful TT probe
	TTDepthHits       uint64 // #nodes where TT hit was at the same depth
	TTDeeperHits      uint64 // #nodes where TT hit was deeper (and the same parity)
	TTBetaCuts        uint64 // #nodes with beta cutoff from TT hit
	TTAlphaCuts       uint64 // #nodes with alpha cutoff from TT hit
	TTTrueEvals       uint64 // #nodes with QQT hits that are the same depth and are not a lower bound
	QNodes            uint64 // #nodes visited in qsearch
	QMates            uint64 // #true terminal nodes in qsearch
	QNonLeafs         uint64 // #non-leaf qnodes
	QFirstChildCuts   uint64 // #non-leaf qnodes that (beta-)cut on the first child searched
	QAllChildrenNodes uint64 // #non-leaf qnodes with no beta cut
	QKillers          uint64 // #qnodes with killer move available
	QKillerCuts       uint64 // #qnodes with killer move cut
	QDeepKillers      uint64 // #qnodes with deep killer move available
	QDeepKillerCuts   uint64 // #qnodes with deep killer move cut
	QRampagePrunes    uint64 // #qnodes where we did queen rampage pruning
	QPats             uint64 // #qnodes with stand pat best
	QPatCuts          uint64 // #qnodes with stand pat cut
	QQuiesced         uint64 // #qnodes where we successfully quiesced
	QPrunes           uint64 // #qnodes where we reached full depth - i.e. likely failed to quiesce
	QttHits           uint64 // #qnodes with successful QTT probe
	QttDepthHits      uint64 // #qnodes where QTT hit was at the same depth
	QttBetaCuts       uint64 // #qnodes with beta cutoff from QTT hit
	QttAlphaCuts      uint64 // #qnodes with beta cutoff from QTT hit
	QttLateCuts       uint64 // #qnodes with beta cutoff from QTT hit
	QttTrueEvals      uint64 // #qnodes with QQT hits that are the same depth and are not a lower bound

	NonLeafsAt       [MaxDepth]uint64  // non-leafs by depth
	QNonLeafsAt      [MaxDepth]uint64 // q-search non-leafs by depth

	Killers           [MaxDepth][NKillersPerDepth]uint64 // #nodes with killer move available
	KillerCuts        [MaxDepth][NKillersPerDepth]uint64 // #nodes with killer move cut
}

func PerC(n uint64, N uint64) string {
	return fmt.Sprintf("%d [%.2f%%]", n, float64(n)/float64(N)*100)
}

func (s *SearchStatsT) Dump(finalDepth int) {
	// Reverse order from which it appears in the UCI driver
	fmt.Println("info string   "/*q-moves:", s.QMoves, "q-simple-moves:", PerC(s.QSimpleMoves, s.QMoves), "q-simple-captures:", PerC(s.QSimpleCaptures, s.QMoves), "q-nomoves:", s.QNoMoves, "q-1moves:", s.Q1Move, "q-movegens:", PerC(s.QMoveGens, s.QNonLeafs)*/, "q-mates:", PerC(s.QMates, s.QNonLeafs), "q-pat-cuts:", PerC(s.QPatCuts, s.QNonLeafs), "q-rampage-prunes:", PerC(s.QRampagePrunes, s.QNonLeafs), "q-killers:", PerC(s.QKillers, s.QNonLeafs), "q-killer-cuts:", PerC(s.QKillerCuts, s.QNonLeafs), "q-deep-killers:", PerC(s.QDeepKillers, s.QNonLeafs), "q-deep-killer-cuts:", PerC(s.QDeepKillerCuts, s.QNonLeafs))
	// if UseEarlyMoveHint {
	// 	fmt.Println("info string   mv-all:", s.MVAll, "mv-non-king:", PerC(s.MVNonKing, s.MVAll), "mv-ours:", PerC(s.MVOurPiece, s.MVAll), "mv-pawn:", PerC(s.MVPawn, s.MVAll), "mv-pawn-push:", PerC(s.MVPawnPush, s.MVPawn), "mv-pp-ok:", PerC(s.MVPawnPushOk, s.MVPawnPush), "mv-pawnok:", PerC(s.MVPawnOk, s.MVPawn), "mv-nonpawn:", PerC(s.MVNonPawn, s.MVAll), "mv-nonpawn-ok:", PerC(s.MVNonPawnOk, s.MVNonPawn), "mv-disc0:", PerC(s.MVDisc0, s.MVAll), "mv-disc1:", PerC(s.MVDisc1, s.MVAll), "mv-disc2:", PerC(s.MVDisc2, s.MVAll), "mv-disc-no:", PerC(s.MVDiscMaybe, s.MVAll))
	// }
	if UseQSearchTT {
		fmt.Println("info string   qtt-hits:", PerC(s.QttHits, s.QNonLeafs), "qtt-depth-hits:", PerC(s.QttDepthHits, s.QNonLeafs), "qtt-beta-cuts:", PerC(s.QttBetaCuts, s.QNonLeafs), "qtt-alpha-cuts:", PerC(s.QttAlphaCuts, s.QNonLeafs), "qtt-late-cuts:", PerC(s.QttLateCuts, s.QNonLeafs), "qtt-true-evals:", PerC(s.QttTrueEvals, s.QNonLeafs))
	}
	fmt.Print("info string    q-non-leafs by depth:")
	for i := 0; i < MaxQDepthStats && i < QSearchDepth; i++ {
		fmt.Printf(" %d: %s", i, PerC(s.QNonLeafsAt[i], s.QNonLeafs))
	}
	fmt.Println()
	fmt.Println("info string q-nodes:", s.QNodes, "q-non-leafs:", s.QNonLeafs, "q-all-nodes:", PerC(s.QAllChildrenNodes, s.QNonLeafs), "q-1st-child-cuts:", PerC(s.QFirstChildCuts, s.QNonLeafs), "q-pats:", PerC(s.QPats, s.QNonLeafs), "q-quiesced:", PerC(s.QQuiesced, s.QNonLeafs), "q-prunes:", PerC(s.QPrunes, s.QNonLeafs))
	fmt.Println()
	//fmt.Println("info string   null-cuts:", PerC(s.NullMoveCuts, s.NonLeafs), "mates:", PerC(s.Mates, s.NonLeafs), "killers:", PerC(s.Killers, s.NonLeafs), "killer-cuts:", PerC(s.KillerCuts, s.NonLeafs), "deep-killers:", PerC(s.DeepKillers, s.NonLeafs), "deep-killer-cuts:", PerC(s.DeepKillerCuts, s.NonLeafs))
	fmt.Println("            !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("info string   cuts:", PerC(s.CutNodes, s.NonLeafs), "null-cuts:", PerC(s.NullMoveCuts, s.CutNodes), "first-child-cuts:", PerC(s.FirstChildCuts, s.CutNodes), "cut-kids", PerC(s.CutNodeChildren, s.CutNodes-s.NullMoveCuts), "shallow-cut-kids", PerC(s.ShallowCutNodeChildren, s.ShallowCutNodes-s.ShallowNullMoveCuts), "deep-cut-kids", PerC(s.DeepCutNodeChildren, s.DeepCutNodes-s.DeepNullMoveCuts), "tt-move-cuts:", PerC(s.TTMoveCuts, s.CutNodes-s.NullMoveCuts))
	//, "killer-cuts:", PerC(s.KillerCuts, s.CutNodes-s.NullMoveCuts), "deep-killer-cuts:", PerC(s.DeepKillerCuts, s.CutNodes-s.NullMoveCuts), "tt-move-cuts-not-kdk:", PerC(s.TTMoveCutsNotKDK, s.CutNodes-s.NullMoveCuts), "killer-cuts-not-dk:", PerC(s.KillerCutsNotDK, s.CutNodes-s.NullMoveCuts), "deep-killer-cuts-not-k:", PerC(s.DeepKillerCutsNotK, s.CutNodes-s.NullMoveCuts))
	fmt.Println("            !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	if UseTT {
		fmt.Println("info string   tt-hits:", PerC(s.TTHits, s.NonLeafs), "tt-depth-hits:", PerC(s.TTDepthHits, s.NonLeafs), "tt-deeper-hits:", PerC(s.TTDeeperHits, s.NonLeafs), "tt-beta-cuts:", PerC(s.TTBetaCuts, s.NonLeafs), "tt-alpha-cuts:", PerC(s.TTAlphaCuts, s.NonLeafs), "tt-late-cuts:", PerC(s.TTMoveCuts, s.NonLeafs), "tt-true-evals:", PerC(s.TTTrueEvals, s.NonLeafs))
	}
	fmt.Print("info string    non-leafs by depth:")
	for i := 0; i < MaxDepthStats && i < finalDepth; i++ {
		fmt.Printf(" %d: %s", i, PerC(s.NonLeafsAt[i], s.NonLeafs))
	}
	fmt.Println()
	fmt.Println("info string nodes:", s.Nodes, "non-leafs:", s.NonLeafs, "all-nodes:", PerC(s.AllChildrenNodes, s.NonLeafs), "1st-child-cuts:", PerC(s.FirstChildCuts, s.NonLeafs), "pos-repetitions:", PerC(s.PosRepetitions, s.Nodes))
}
