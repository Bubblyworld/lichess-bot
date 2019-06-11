package engine

import (
	"fmt"
)

type SearchStatsT struct {
	Nodes             uint64 // #nodes visited
	AfterTTNodes      uint64 // #nodes surviving TT cut/exact
	AfterNullNodes    uint64 // #nodes surviving null-heuristic cut
	AfterChildLoopNodes uint64 // #nodes surviving 
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
	AllChildrenNodes2 uint64 // #non-leaf nodes with no beta cut (check)
	TTMoveCuts        uint64 // #nodes with tt move cut
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
	QCutNodes         uint64 // #(beta-)cut qnodes
	QCutNodeChildren  uint64 // Total #children of cut qnodes (in order to see how effective the move ordering is)
	QFirstChildCuts   uint64 // #non-leaf qnodes that (beta-)cut on the first child searched
	QAllChildrenNodes uint64 // #non-leaf qnodes with no beta cut
	QRampagePrunes    uint64 // #qnodes where we did queen rampage pruning
	QPats             uint64 // #qnodes with stand pat best
	QPatCuts          uint64 // #qnodes with stand pat cut
	QQuiesced         uint64 // #qnodes where we successfully quiesced
	QPrunes           uint64 // #qnodes where we reached full depth - i.e. likely failed to quiesce
	QttHits           uint64 // #qnodes with successful QTT probe
	QttMoveCuts       uint64 // #qnodes with tt move cut
	QttDepthHits      uint64 // #qnodes where QTT hit was at the same depth
	QttBetaCuts       uint64 // #qnodes with beta cutoff from QTT hit
	QttAlphaCuts      uint64 // #qnodes with beta cutoff from QTT hit
	QttLateCuts       uint64 // #qnodes with beta cutoff from QTT hit
	QttTrueEvals      uint64 // #qnodes with QQT hits that are the same depth and are not a lower bound

	NShallowBestMoveCalcs uint64 // #actual calculations of shallow best move
	NNoMoveShallowBestMove uint64 // #calculations of shallow best move that return NoMove
	NShallowBestMoves uint64 // #nodes with shallow-best-moves
	NShallowBestMovesBest uint64 // #nodes with best move == shallow-best-move
	NShallowBestMoveCuts uint64 // #nodes with shallow-best-move best and a cut
	NShallowBestMoveOtherCuts uint64 // #nodes with shallow-best-move and another move cuts

	NTTMoves uint64 // #nodes with tt-moves
	NTTMovesBest uint64 // #nodes with best move == tt-move
	NTTMoveCuts uint64 // #nodes with tt-move best and a cut
	NTTMoveOtherCuts uint64 // #nodes with tt-moveother move cuts

	AfterNullNodesByD [MaxDepth]uint64 // #nodes surviving null-heuristic cut
	NTTMovesByD [MaxDepth]uint64 // #nodes with tt-moves
	NTTMovesBestByD [MaxDepth]uint64 // #nodes with best move == tt-move
	NTTMoveCutsByD [MaxDepth]uint64 // #nodes with tt-move best and a cut
	NTTMoveOtherCutsByD [MaxDepth]uint64 // #nodes with tt-moveother move cuts

	AfterNullNodesByD1 [MaxDepth]uint64 // #nodes surviving null-heuristic cut
	NTTMovesByD1 [MaxDepth]uint64 // #nodes with tt-moves from depthToGo-1
	NTTMovesBestByD1 [MaxDepth]uint64 // #nodes with best move == tt-move from depthToGo-1
	NTTMoveCutsByD1 [MaxDepth]uint64 // #nodes with tt-move best and a cut from depthToGo-1
	NTTMoveOtherCutsByD1 [MaxDepth]uint64 // #nodes with tt-moveother move cuts from depthToGo-1
	
	NonLeafsAt       [MaxDepth]uint64  // non-leafs by depth
	QNonLeafsAt      [MaxDepth]uint64 // q-search non-leafs by depth

	Killers           [MaxDepth][NKillersPerDepth]uint64 // #nodes with killer move available
	KillerCuts        [MaxDepth][NKillersPerDepth]uint64 // #nodes with killer move cut

	QKillers          [MaxDepth][NKillersPerDepth]uint64 // #qnodes with killer move available
	QKillerCuts       [MaxDepth][NKillersPerDepth]uint64 // #qnodes with killer move cut
}

func PerC(n uint64, N uint64) string {
	return fmt.Sprintf("%d [%.2f%%]", n, float64(n)/float64(N)*100)
}

func (s *SearchStatsT) DumpOld(finalDepth int) {
	// Reverse order from which it appears in the UCI driver
	fmt.Println("info string   "/*q-moves:", s.QMoves, "q-simple-moves:", PerC(s.QSimpleMoves, s.QMoves), "q-simple-captures:", PerC(s.QSimpleCaptures, s.QMoves), "q-nomoves:", s.QNoMoves, "q-1moves:", s.Q1Move, "q-movegens:", PerC(s.QMoveGens, s.QNonLeafs)*/, "q-mates:", PerC(s.QMates, s.QNonLeafs), "q-pat-cuts:", PerC(s.QPatCuts, s.QNonLeafs), "q-rampage-prunes:", PerC(s.QRampagePrunes, s.QNonLeafs)/*, "q-killers:", PerC(s.QKillers, s.QNonLeafs), "q-killer-cuts:", PerC(s.QKillerCuts, s.QNonLeafs), "q-deep-killers:", PerC(s.QDeepKillers, s.QNonLeafs), "q-deep-killer-cuts:", PerC(s.QDeepKillerCuts, s.QNonLeafs)*/)
	// if UseEarlyMoveHint {
	// 	fmt.Println("info string   mv-all:", s.MVAll, "mv-non-king:", PerC(s.MVNonKing, s.MVAll), "mv-ours:", PerC(s.MVOurPiece, s.MVAll), "mv-pawn:", PerC(s.MVPawn, s.MVAll), "mv-pawn-push:", PerC(s.MVPawnPush, s.MVPawn), "mv-pp-ok:", PerC(s.MVPawnPushOk, s.MVPawnPush), "mv-pawnok:", PerC(s.MVPawnOk, s.MVPawn), "mv-nonpawn:", PerC(s.MVNonPawn, s.MVAll), "mv-nonpawn-ok:", PerC(s.MVNonPawnOk, s.MVNonPawn), "mv-disc0:", PerC(s.MVDisc0, s.MVAll), "mv-disc1:", PerC(s.MVDisc1, s.MVAll), "mv-disc2:", PerC(s.MVDisc2, s.MVAll), "mv-disc-no:", PerC(s.MVDiscMaybe, s.MVAll))
	// }
	fmt.Println("info string   qcuts:", PerC(s.QCutNodes, s.QNonLeafs), "qpat-cuts:", PerC(s.QPatCuts, s.CutNodes), "qfirst-child-cuts:", PerC(s.QFirstChildCuts, s.CutNodes), "cut-kids", PerC(s.QCutNodeChildren, s.QCutNodes-s.QPatCuts), "qtt-move-cuts:", PerC(s.QttMoveCuts, s.QCutNodes-s.QPatCuts))
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

func (s *SearchStatsT) Dump/*CutStats*/(finalDepth int) {
	fmt.Println()

	// First we hit the TT
	nodesLeft := s.Nodes
	fmt.Println("info string nodes:", nodesLeft, "tt-beta-cuts:", PerC(s.TTBetaCuts, nodesLeft), "tt-alpha-cuts:", PerC(s.TTAlphaCuts, nodesLeft), "tt-true-evals:", PerC(s.TTTrueEvals, nodesLeft))

	sanityCheck := nodesLeft - (s.TTBetaCuts + s.TTAlphaCuts + s.TTTrueEvals)
	nodesLeft = s.AfterTTNodes

	fmt.Println("info string nodes-after-tt-per-all-nodes:", PerC(nodesLeft, s.Nodes))

	// Then we do leaf nodes (q-search)
	fmt.Println("info string non-leafs-per-all-nodes:", PerC(s.NonLeafs, s.Nodes), "non-leafs-per-after-tt:", PerC(s.NonLeafs, nodesLeft), "cut-nodes:", PerC(s.CutNodes, s.NonLeafs), "all-nodes:", PerC(s.AllChildrenNodes, s.NonLeafs), "sanity-check:", s.AllChildrenNodes2)

	nodesLeft = s.NonLeafs

	// Then we do null-move heuristic
	//fmt.Println("info string nullodes-after-tt:", PerC(nodesLeft, s.Nodes), "sanity-check:", sanityCheck, "null-cuts:", PerC(s.NullMoveCuts, nodesLeft))

	sanityCheck = nodesLeft - s.NullMoveCuts
	nodesLeft = s.AfterNullNodes

	fmt.Println("info string nodes-after-null-per-all-nodes:", PerC(nodesLeft, s.Nodes), "nodes-after-null-per-non-leafs:", PerC(nodesLeft, s.NonLeafs), "sanity-check:", sanityCheck)

	fmt.Println("info string all-children:", PerC(s.AllChildrenNodes, nodesLeft), "sanity-check:", s.AllChildrenNodes2, "cut-nodes:", PerC(s.CutNodes, nodesLeft))
	
	fmt.Println()

	cutsWithShallowBestMove := s.NShallowBestMoveCuts + s.NShallowBestMoveOtherCuts
	allNodesWithShallowBestMove := s.NShallowBestMoves - cutsWithShallowBestMove
	allNodesWithShallowBestMoveBest := s.NShallowBestMovesBest - s.NShallowBestMoveCuts
		
	fmt.Println()

	fmt.Println("info string shallow-best-move-calls:", s.AfterNullNodes, "shallow-best-move-calcs:", PerC(s.NShallowBestMoveCalcs, s.AfterNullNodes), "nomove-shallow-best-move-calcs:", PerC(s.NNoMoveShallowBestMove, s.NShallowBestMoveCalcs))

	fmt.Println("info string shallow-best-moves:", PerC(s.NShallowBestMoves, nodesLeft), "shallow-best-move-best:", PerC(s.NShallowBestMovesBest, s.NShallowBestMoves), "cuts-with-shallow-best-move:", cutsWithShallowBestMove, "shallow-best-move-cuts:", PerC(s.NShallowBestMoveCuts, cutsWithShallowBestMove), "all-nodes-with-shallow-best-move", allNodesWithShallowBestMove, "all-nodes-with-shallow-best-move-best-of-all:", PerC(allNodesWithShallowBestMoveBest, allNodesWithShallowBestMove))
	
	fmt.Println()
	
	cutsWithTTMove := s.NTTMoveCuts + s.NTTMoveOtherCuts
	allNodesWithTTMove := s.NTTMoves - cutsWithTTMove
	allNodesWithTTMoveBest := s.NTTMovesBest - s.NTTMoveCuts
		
	fmt.Println("info string tt-moves:", PerC(s.NTTMoves, nodesLeft), "tt-move-best:", PerC(s.NTTMovesBest, s.NTTMoves), "cuts-with-tt-move:", cutsWithTTMove, "tt-move-cuts:", PerC(s.NTTMoveCuts, cutsWithTTMove), "all-nodes-with-tt-move", allNodesWithTTMove, "all-nodes-with-tt-move-best-of-all:", PerC(allNodesWithTTMoveBest, allNodesWithTTMove))

	fmt.Println()
	
	for d := 0; d < 13; d++ {
		nodesLeftByD := s.AfterNullNodesByD[d]
		
		cutsWithTTMove = s.NTTMoveCutsByD[d] + s.NTTMoveOtherCutsByD[d]
		allNodesWithTTMove = s.NTTMovesByD[d] - cutsWithTTMove
		allNodesWithTTMoveBest = s.NTTMovesBestByD[d] - s.NTTMoveCutsByD[d]
		
		fmt.Println("info string depth", d, "tt-moves:", PerC(s.NTTMovesByD[d], nodesLeftByD), "tt-move-best:", PerC(s.NTTMovesBestByD[d], s.NTTMovesByD[d]), "cuts-with-tt-move:", cutsWithTTMove, "tt-move-cuts:", PerC(s.NTTMoveCutsByD[d], cutsWithTTMove), "all-nodes-with-tt-move", allNodesWithTTMove, "all-nodes-with-tt-move-best-of-all:", PerC(allNodesWithTTMoveBest, allNodesWithTTMove))
	}

	fmt.Println()
	
	for d := 0; d < 13; d++ {
		nodesLeftByD1 := s.AfterNullNodesByD1[d]
		
		cutsWithTTMove = s.NTTMoveCutsByD1[d] + s.NTTMoveOtherCutsByD1[d]
		allNodesWithTTMove = s.NTTMovesByD1[d] - cutsWithTTMove
		allNodesWithTTMoveBest = s.NTTMovesBestByD1[d] - s.NTTMoveCutsByD1[d]
		
		fmt.Println("info string tt-d1", d, "tt-moves:", PerC(s.NTTMovesByD1[d], nodesLeftByD1), "tt-move-best:", PerC(s.NTTMovesBestByD1[d], s.NTTMovesByD1[d]), "cuts-with-tt-move:", cutsWithTTMove, "tt-move-cuts:", PerC(s.NTTMoveCutsByD1[d], cutsWithTTMove), "all-nodes-with-tt-move", allNodesWithTTMove, "all-nodes-with-tt-move-best-of-all:", PerC(allNodesWithTTMoveBest, allNodesWithTTMove))
	}
	
	fmt.Println()
}
