package engine

type SearchStatsT struct {
	Nodes             uint64 // #nodes visited
	MoveGens uint64
	Moves uint64
	SimpleMoves uint64
	SimpleCaptures uint64
	Mates             uint64 // #true terminal nodes
	NonLeafs          uint64 // #non-leaf nodes
	FirstChildCuts    uint64 // #non-leaf nodes that (beta-)cut on the first child searched
	AllChildrenNodes  uint64 // #non-leaf nodes with no beta cut
	NullMoveCuts      uint64 // #nodes that cut due to null move heuristic
	ValidHintMoves    uint64 // #nodes with a known valid move before we do movegen - either a TT hit or a known valid killer move
	EarlyKillers      uint64 // #nodes with early killer
	ValidEarlyKillers uint64 // #nodes with killer used as early hint move
	HintMoveCuts      uint64 // #nodes with hint move cut (before movegen)
	Killers           uint64 // #nodes with killer move available
	KillerCuts        uint64 // #nodes with killer move cut
	DeepKillers       uint64 // #nodes with deep killer move available
	DeepKillerCuts    uint64 // #nodes with deep killer move cut
	PosRepetitions    uint64 // #nodes with repeated position
	TTHits            uint64 // #nodes with successful TT probe
	TTDepthHits       uint64 // #nodes where TT hit was at the same depth
	TTDeeperHits      uint64 // #nodes where TT hit was deeper (and the same parity)
	TTBetaCuts        uint64 // #nodes with beta cutoff from TT hit
	TTAlphaCuts       uint64 // #nodes with alpha cutoff from TT hit
	TTLateCuts        uint64 // #nodes with beta cutoff from TT hit
	TTTrueEvals       uint64 // #nodes with QQT hits that are the same depth and are not a lower bound
	QNodes            uint64 // #nodes visited in qsearch
	QMoveGens uint64
	QNoMoves uint64
	Q1Move uint64
	QMoves uint64
	QSimpleMoves uint64
	QSimpleCaptures uint64
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

	MVAll             uint64
	MVNonKing         uint64
	MVOurPiece        uint64
	MVPawn            uint64
	MVPawnPush        uint64
	MVPawnPushOk      uint64
	MVPawnOk          uint64
	MVNonPawn         uint64
	MVNonPawnOk       uint64
	MVDisc0 uint64
	MVDisc1 uint64
	MVDisc2 uint64
	MVDiscMaybe uint64

	NonLeafsAt       [MaxDepthStats]uint64  // non-leafs by depth
	FirstChildCutsAt [MaxDepthStats]uint64  // first-child cuts by depth
	QNonLeafsAt      [MaxQDepthStats]uint64 // q-search non-leafs by depth
}