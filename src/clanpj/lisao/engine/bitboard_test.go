package engine

import (
	"testing"
)

func testResult(t *testing.T, s string, val uint64, expected uint64) {
	if val != expected {
		t.Errorf(s, val, expected)
	}
}

func TestBitboard(t *testing.T) {
	// Lose H file when going E
	testResult(t, "E(0x8000008080800000) is 0x%016x expected 0x%016x\n", E(0x8000008080800000), 0)
	// Lose A file when going W
	testResult(t, "W(0x0100000101010000) is 0x%016x expected 0x%016x\n", W(0x0100000101010000), 0)
	// Non-H files move E
	testResult(t, "E(0x8040201008040201) is 0x%016x expected 0x%016x\n", E(0x8040201008040201), 0x0080402010080402)
	// Non-A files move W
	testResult(t, "W(0x8040201008040201) is 0x%016x expected 0x%016x\n", W(0x8040201008040201), 0x4020100804020100)

	testResult(t, "N(0x8040201008040201) is 0x%016x expected 0x%016x\n", N(0x8040201008040201), 0x4020100804020100)
	testResult(t, "S(0x8040201008040201) is 0x%016x expected 0x%016x\n", S(0x8040201008040201), 0x0080402010080402)
	testResult(t, "N(0x0804020180402010) is 0x%016x expected 0x%016x\n", N(0x0804020180402010), 0x0402018040201000)
	testResult(t, "S(0x0804020180402010) is 0x%016x expected 0x%016x\n", S(0x0804020180402010), 0x0008040201804020)

	testResult(t, "WPawnScope(0x0180000000000100) is 0x%016x expected 0x%016x\n", WPawnScope(0x0180000000000100), 0xc303030303030000)
	testResult(t, "WPawnScope(0x8140000000000200) is 0x%016x expected 0x%016x\n", WPawnScope(0x8140000000000200), 0xe707070707070000)
	
	testResult(t, "BPawnScope(0x0001000000008001) is 0x%016x expected 0x%016x\n", BPawnScope(0x0001000000008001), 0x00000303030303c3)
	testResult(t, "BPawnScope(0x0002000000004081) is 0x%016x expected 0x%016x\n", BPawnScope(0x0002000000004081), 0x00000707070707e7)
}
