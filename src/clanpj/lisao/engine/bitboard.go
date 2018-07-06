// Bitboard utilities
// Note bit 0 (low bit) is square A1, bit 63 (hi bit) is square H8

package engine

const Shift uint = 1

const A uint64 = 0x0101010101010101
const H uint64 = 0x8080808080808080

func N(bb uint64) uint64 { return bb << 8 }

func S(bb uint64) uint64 { return bb >> 8 }

func W(bb uint64) uint64 { return (bb & ^A) >> 1 }

func E(bb uint64) uint64 { return (bb & ^H) << 1 }

func NFill(bb uint64) uint64 {
	fill := bb
	fill = fill | (fill << 8)
	fill = fill | (fill << 16)
	fill = fill | (fill << 32)
	return fill
}

func SFill(bb uint64) uint64 {
	fill := bb
	fill = fill | (fill >> 8)
	fill = fill | (fill >> 16)
	fill = fill | (fill >> 32)
	return fill
}

func WPawnScope(wPawns uint64) uint64 {
	// forward
	n := N(wPawns)
	// take west
	nw := W(n)
	// take east
	ne := E(n)

	// Pawns' influence is all the squares forward from there
	return NFill(n | nw | ne)
}

func BPawnScope(bPawns uint64) uint64 {
	// forward
	s := S(bPawns)
	// take west
	sw := W(s)
	// take east
	se := E(s)

	// Pawns' influence is all the squares forward from there
	return SFill(s | sw | se)
}

// Pawn attacks and defenses
func WPawnAttacks(wPawns uint64) uint64 {
	// forward
	n := N(wPawns)
	// take west
	nw := W(n)
	// take east
	ne := E(n)

	// Take either way
	return nw | ne
}

// Pawn attacks and defenses
func BPawnAttacks(bPawns uint64) uint64 {
	// forward
	s := S(bPawns)
	// take west
	sw := W(s)
	// take east
	se := E(s)

	// Take either way
	return sw | se
}
