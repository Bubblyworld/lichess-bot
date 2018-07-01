// Move history for checking 3-fold repetition

package engine

// Map: zobrist -> count
type HistoryTableT map[uint64]int

// Add a position and return the resulting count for this position
func (ht HistoryTableT) Add(zobrist uint64) int {
	count := ht[zobrist]
	count++
	ht[zobrist] = count
	return count
}

// Add a position and return the resulting position count for this position
// Removes entries with count zero so the history table doesn't explode in size.
func (ht HistoryTableT) Remove(zobrist uint64) int {
	count := ht[zobrist]
	count--
	if count > 0 {
		ht[zobrist] = count
	} else {
		delete(ht, zobrist)
	}
	return count
}
