// Extension of the 'killer-move' heuristic that maintains a set of N previously useful moves for each search depth (depth-from-root)
// I'm not sure that this is a standard approach, but it is motivated by some similar approaches in recent literature ('ADS' - adaptive data structures)

// TODO (RPJ) experiment with using depthToGo (and possibly search window) as additional move ordering clues.

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)

const NKillersPerDepth = 4

const MoveNotFound = -1

type KillerMoveTableT [MaxDepth][NKillersPerDepth]dragon.Move

// Install a new killer move
func (kt *KillerMoveTableT) addKillerMove(move dragon.Move, depthFromRoot int) {
	if move == NoMove {
		return
	}
	
	depthKillers := &kt[depthFromRoot]

	moveIndex := 0
	for ; moveIndex < NKillersPerDepth; moveIndex++ {
		if depthKillers[moveIndex] == move {
			break
		}
	}

	// Shift up to make space for the new move at the front
	for i := moveIndex; 0 < i; i-- {
		if i < NKillersPerDepth {
			depthKillers[i] = depthKillers[i-1]
		}
	}

	// Install the new move as the new best killer
	depthKillers[0] = move
}

// Return the killers for the given depth from most deadly to least deadly
func (kt *KillerMoveTableT) killersForDepth(depthFromRoot int) *[NKillersPerDepth]dragon.Move {
	return &kt[depthFromRoot]
}

// Return the index of the given move in the killers list, or MoveNotFound
func (kt *KillerMoveTableT) killerMoveIndex(move dragon.Move, depthFromRoot int) int {
	depthKillers := &kt[depthFromRoot]

	moveIndex := MoveNotFound
	for i := 0; i < NKillersPerDepth; i++ {
		if depthKillers[i] == move {
			moveIndex = i
			break
		}
	}

	return moveIndex
}
