// 0x88 board representation facilities

package engine

import (
	dragon "github.com/Bubblyworld/dragontoothmg"
)


func init() {
	initKnightDiffs()
	initBishopDiffs()
	initRookDiffs()
}

func initKnightDiffs() {
	const nnwDiff = 0x33-0x12
	const nneDiff = 0x33-0x14
	const wnwDiff = 0x33-0x21
	const eneDiff = 0x33-0x25
	const sseDiff = -nnwDiff
	const sswDiff = -nneDiff
	const eseDiff = -wnwDiff
	const wswDiff = -eneDiff

	knightDirDist := DirDistT{KnightDir, 1}

	x88DiffToDirDist[diffIdx(nnwDiff)] = knightDirDist
	x88DiffToDirDist[diffIdx(nneDiff)] = knightDirDist
	x88DiffToDirDist[diffIdx(wnwDiff)] = knightDirDist
	x88DiffToDirDist[diffIdx(eneDiff)] = knightDirDist
	x88DiffToDirDist[diffIdx(sseDiff)] = knightDirDist
	x88DiffToDirDist[diffIdx(sswDiff)] = knightDirDist
	x88DiffToDirDist[diffIdx(eseDiff)] = knightDirDist
	x88DiffToDirDist[diffIdx(wswDiff)] = knightDirDist
}

func init7Diffs(diff int, dir DirT) {
	ndiffs := diff
	for dist := uint8(1); dist <= uint8(7); dist++ {
		x88DiffToDirDist[diffIdx(ndiffs)] = DirDistT{dir, dist}
		ndiffs += diff
	}
}

func initBishopDiffs() {
	const nwDiff = 0x11-0x02
	const neDiff = 0x11-0x00
	const seDiff = -nwDiff
	const swDiff = -neDiff

	init7Diffs(nwDiff, NWDir)
	init7Diffs(neDiff, NEDir)
	init7Diffs(seDiff, SEDir)
	init7Diffs(swDiff, SWDir)
}

func initRookDiffs() {
	const nDiff = 0x10-0x00
	const sDiff = 0x00-0x10
	const wDiff = 0x00-0x01
	const eDiff = 0x01-0x00

	init7Diffs(nDiff, NDir)
	init7Diffs(sDiff, SDir)
	init7Diffs(wDiff, WDir)
	init7Diffs(eDiff, EDir)
}

// From square index (0-64) to 0x88 index.
var squareTo0x88 = [64]int {
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
	0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
	0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
	0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
	0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57,
	0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
	0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77}

type DirT uint8

const (
	InvalidDir DirT = iota
	// Rook dirs
	NDir
	SDir
	WDir
	EDir
	// Bishop dirs
	NWDir
	NEDir
	SWDir
	SEDir
	// Knight dir (just need one cos it's not a slider)
	KnightDir

	NDirs
)

// Opposite direction of a given direction
var oppositeDir = [NDirs]DirT {
	InvalidDir,
	// Rook dirs
	SDir, //NDir
	NDir, //SDir
	EDir, //WDir
	WDir, //EDir
	// Bishop dirs
	SEDir, //NWDir
	SWDir, //NEDir
	NEDir, //SWDir
	NWDir, //SEDir
	// Knight dir (not valid here)
	KnightDir}

type DirFlagT int

func dirFlag(dir DirT) DirFlagT {
	return DirFlagT(1 << dir)
}

const knightDirFlags = DirFlagT(1 << KnightDir)
const rookDirFlags = DirFlagT(1 << NDir) | DirFlagT(1 << SDir) | DirFlagT(1 << WDir) | DirFlagT(1 << EDir)
const bishopDirFlags = DirFlagT(1 << NWDir) | DirFlagT(1 << NEDir) | DirFlagT(1 << SWDir) | DirFlagT(1 << SEDir)
const queenDirFlags = rookDirFlags | bishopDirFlags

var PieceToDirFlags = [dragon.NoPieces]DirFlagT {
	DirFlagT(0), // Nothing
	DirFlagT(0), // Pawn - needs special handling
	knightDirFlags, // Knight
	bishopDirFlags, // Bishop
	rookDirFlags, // Rook
	queenDirFlags, // Queen
	queenDirFlags} // King - but distance 1 only (not used anywhere?)

const whitePawnPushFlags = DirFlagT(1 << NDir)
const whitePawnCaptureFlags = DirFlagT(1 << NWDir) | DirFlagT(1 << NEDir)
const blackPawnPushFlags = DirFlagT(1 << SDir)
const blackPawnCaptureFlags = DirFlagT(1 << SWDir) | DirFlagT(1 << SEDir)

type DirDistT struct {
	dir DirT
	dist uint8
}

const min0x88Diff = 0x00 - 0x77

var x88DiffToDirDist [0x77+0x77+1]DirDistT

const maxPathDist = uint8(7)

// Bitboard base bit for each path direction
var pathBasePos = [NDirs]uint8 {
	0,
	// Rook dirs
	0, // NDir - a1-based
	63, // SDir - h8-based
	63, // WDir - h8-based
	0, // EDir - a1-based
	// Bishop dirs
	7, // NWDir - a8-based
	0, // NEDir a1-based
	63, // SWDir h8-based
	56, // SEDir h1-based
	// Knight dir (not valid here)
	0}

// 
// These always exclude the starting square.
var dirDistPathBits = [NDirs][maxPathDist+1]uint64 {
	//InvalidDir
	{ 0, 0, 0, 0, 0, 0, 0, 0},
	
	// Rook dirs
	//NDir - a1-based
	{ 0, 0x0000000000000100, 0x0000000000010100, 0x0000000001010100, 0x0000000101010100, 0x0000010101010100, 0x0001010101010100, 0x0101010101010100},
	//SDir - h8-based
	{ 0, 0x0080000000000000, 0x0080800000000000, 0x0080808000000000, 0x0080808080000000, 0x0080808080800000, 0x0080808080808000, 0x0080808080808080},
	//WDir - h8-based
	{ 0, 0x4000000000000000, 0x6000000000000000, 0x7000000000000000, 0x7800000000000000, 0x7c00000000000000, 0x7e00000000000000, 0x7f00000000000000},
	//EDir - a1-based
	{ 0, 0x0000000000000002, 0x0000000000000006, 0x000000000000000e, 0x000000000000001e, 0x000000000000003e, 0x000000000000007e, 0x00000000000000fe},
	
	// Bishop dirs
	//NWDir - a8-based
	{ 0, 0x0000000000004000, 0x0000000000204000, 0x0000000010204000, 0x0000000810204000, 0x0000040810204000, 0x0002040810204000, 0x0102040810204000},
	//NEDir - a1-based
	{ 0, 0x0000000000000200, 0x0000000000040200, 0x0000000008040200, 0x0000001008040200, 0x0000201008040200, 0x0040201008040200, 0x8040201008040200},
	//SWDir - h8-based
	{ 0, 0x0040000000000000, 0x0040200000000000, 0x0040201000000000, 0x0040201008000000, 0x0040201008040000, 0x0040201008040200, 0x0040201008040201},
	//SEDir - h1-based
	{ 0, 0x0002000000000000, 0x0002040000000000, 0x0002040800000000, 0x0002040810000000, 0x0002040810200000, 0x0002040810204000, 0x0002040810204080},
	
	// Knight dir (not valid here)
	//KnightDir
	{ 0, 0, 0, 0, 0, 0, 0, 0}}

func x88Diff(from uint8, to uint8) int {
	return squareTo0x88[to] - squareTo0x88[from]
}

func diffIdx(diff int) int {
	return diff - min0x88Diff
}

// Direction and distance (knight moves are 1) between squares.
func dirDist(from uint8, to uint8) DirDistT {
	return x88DiffToDirDist[diffIdx(x88Diff(from, to))]
}

// Excludes from and to squares
func sliderPathMiddleBits(from uint8, dir DirT, dist uint8, debug bool) uint64 {
	basePath := dirDistPathBits[dir][dist-1]
	base := pathBasePos[dir]

	if base < from {
		return basePath << (from-base)
	} else {
		return basePath >> (base-from)
	}
}

// Blockers is typically all pieces of both sides
func isSliderPathBlocked(from uint8, dir DirT, dist uint8, blockers uint64, debug bool) bool {
	return (sliderPathMiddleBits(from, dir, dist, debug) & blockers) != 0
}
