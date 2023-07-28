// SPDX-License-Identifier: MIT

// The package is copyrighted by Pierre L'Ecuyer and the
// University of Montreal.
// Go translation Copyright 2023 University of Illinois Board of Trustees.
// See LICENSE.md for details.

// RngStreams is an object-oriented random-number package with many long
// streams and substreams, based on the MRG32k3a RNG from reference [1]
// below and proposed in [2].
//
// It has implementations in C, C++, Go, Java, R, OpenCL, and some other
// languages.
//
// The package is copyrighted by Pierre L'Ecuyer and the University of
// Montreal.  It can be used freely for any purpose.
//
// e-mail:  lecuyer@iro.umontreal.ca
// http://www.iro.umontreal.ca/~lecuyer/
//
// If you use it for your research, please cite the following relevant
// publications in which MRG32k3a and the package with multiple streams
// were proposed:
//
// [1] P. L'Ecuyer, “Good Parameter Sets for Combined
// Multiple Recursive Random Number Generators”,
// Operations Research, 47, 1 (1999), 159--164.  See
// https://www-labs.iro.umontreal.ca/~lecuyer/myftp/papers/opres-combmrg2-1999.pdf
//
// [2] P. L'Ecuyer, R. Simard, E. J. Chen, and W. D. Kelton, “An
// Objected-Oriented Random-Number Package with Many Long Streams and
// Substreams”, Operations Research, 50, 6 (2002), 1073--1075 See
// https://www-labs.iro.umontreal.ca/~lecuyer/myftp/papers/streams00.pdf
//
// This Go translation is copyright 2023 The Board of Trustees of the
// University of Illinois. All rights reserved.
package rngstream

import (
	"fmt"
	"strconv"
	"strings"
)

type RngStream struct {
	cg, bg, ig [6]float64
	anti       bool
	incPrec    bool
	name       string
}

const norm float64 = 2.328306549295727688e-10
const m1 float64 = 4294967087
const m2 float64 = 4294944443
const a12 float64 = 1403580
const a13n float64 = 810728
const a21 float64 = 527612
const a23n float64 = 1370589

const two17 float64 = 131072
const two53 float64 = 9007199254740992
const fact float64 = 5.9604644775390625e-8 /* 1 / 2^24 */

// Default initial seed of the package. Will be updated to become
// the seed of the next created stream. */
var nextSeedLow = [3]float64{12345, 12345, 12345}
var nextSeedHigh = [3]float64{12345, 12345, 12345}

func SetRngStreamMasterSeed(seed float64) {
	nextSeedLow[0] = seed
	nextSeedLow[1] = seed + 1
	nextSeedLow[2] = seed + 2
	nextSeedHigh[0] = seed + 3
	nextSeedHigh[1] = seed + 4
	nextSeedHigh[2] = seed + 5
}

// The following are the transition matrices of the two MRG components
// (in matrix form), raised to the powers -1, 1, 2^76, and 2^127, resp.
var (
	// Inverse of a1p0
	invA1 = [3][3]float64{
		{184888585, 0, 1945170933},
		{1, 0, 0},
		{0, 1, 0}}

	// Inverse of a2p0
	invA2 = [3][3]float64{ //
		{0, 360363334, 4225571728},
		{1, 0, 0},
		{0, 1, 0}}

	// First MRG component raised to the power 1.
	a1p0 = [3][3]float64{
		{0, 1, 0},
		{0, 0, 1},
		{-810728, 1403580, 0}}

	// Second MRG component raised to the power 1.
	a2p0 = [3][3]float64{
		{0, 1, 0},
		{0, 0, 1},
		{-1370589, 0, 527612}}

	// First MRG component raised to the power 2^76
	a1p76 = [3][3]float64{
		{82758667, 1871391091, 4127413238},
		{3672831523, 69195019, 1871391091},
		{3672091415, 3528743235, 69195019}}

	// Second MRG component raised to the power 2^76
	a2p76 = [3][3]float64{
		{1511326704, 3759209742, 1610795712},
		{4292754251, 1511326704, 3889917532},
		{3859662829, 4292754251, 3708466080}}

	// First MRG component raised to the power 2^127
	a1p127 = [3][3]float64{
		{2427906178, 3580155704, 949770784},
		{226153695, 1230515664, 3580155704},
		{1988835001, 986791581, 1230515664}}

	// Second MRG component raised to the power 2^127
	a2p127 = [3][3]float64{
		{1464411153, 277697599, 1610723613},
		{32183930, 1464411153, 1022607788},
		{2824425944, 32183930, 2093834863}}
)

// Compute (a*s + c) % m. m must be < 2^35.  Works also for s, c < 0
func multModM(a, s, c, m float64) float64 {
	var v float64
	var a1 int64

	v = a*s + c

	if (v >= two53) || (v <= -two53) {
		a1 := int64(a / two17)
		a -= (float64(a1) * two17)
		v = float64(a1) * s
		a1 = int64(v / m)
		v -= float64(a1) * m
		v = v*two17 + a*s + c
	}
	a1 = int64(v / m)

	v = v - float64(a1)*m
	if v < 0 {
		return v + m
	} else {
		return v
	}
}

// Returns v = A*s % m.  Assumes that -m < s[i] < m.
// Works even if v = s.
func matVecModM(A *[3][3]float64, s *[3]float64, v *[3]float64, m float64) {
	var x [3]float64
	for i := 0; i < 3; i++ {
		x[i] = multModM((*A)[i][0], (*s)[0], 0, m)
		x[i] = multModM((*A)[i][1], (*s)[1], x[i], m)
		x[i] = multModM((*A)[i][2], (*s)[2], x[i], m)
	}

	for i := 0; i < 3; i++ {
		(*v)[i] = x[i]
	}
}

/* Returns C = A*B % m. Work even if A = C or B = C or A = B = C. */
func matMatModM(A *[3][3]float64, B *[3][3]float64, C *[3][3]float64, m float64) {
	/* Returns C = A*B % m. Work even if A = C or B = C or A = B = C. */
	var V = [3]float64{0, 0, 0}

	var W = [3][3]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			V[j] = (*B)[j][i]
		}
		matVecModM(A, &V, &V, m)
		for j := 0; j < 3; j++ {
			W[j][i] = V[j]
		}
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			(*C)[i][j] = W[i][j]
		}
	}
}

/* Compute matrix B = (A^(2^e) % m);  works even if A = B */
func matTwoPowModM(A *[3][3]float64, B *[3][3]float64, m float64, e int64) {
	/* initialize: B = A */
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			(*B)[i][j] = (*A)[i][j]
		}
	}

	/* Compute B = A^{2^e} */
	for i := 0; int64(i) < e; i++ {
		matMatModM(B, B, B, m)
	}
}

// Compute matrix B = A^n % m ;  works even if A = B
func matPowModM(A *[3][3]float64, B *[3][3]float64, m float64, n int64) {

	var W = [3][3]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}

	/* initialize: W = A; B = I */
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			W[i][j] = (*A)[i][j]
			(*B)[i][j] = 0
		}
	}

	for j := 0; j < 3; j++ {
		(*B)[j][j] = 1
	}

	/* Compute B = A^n % m using the binary decomposition of n */

	for n > 0 {
		if n%2 != 0 {
			matMatModM(&W, B, B, m)
		}
		matMatModM(&W, &W, &W, m)
		n /= 2
	}
}

func (g *RngStream) u01() float64 {

	var p1, p2 float64
	var u float64

	/* Component 1 */
	p1 = a12*(*g).cg[1] - a13n*(*g).cg[0]
	k := int64(p1 / m1)
	p1 -= float64(k) * m1

	if p1 < 0 {
		p1 += m1
	}

	(*g).cg[0] = (*g).cg[1]
	(*g).cg[1] = (*g).cg[2]
	(*g).cg[2] = p1

	/* Component 2 */
	p2 = a21*(*g).cg[5] - a23n*(*g).cg[3]
	k = int64(p2 / m2)
	p2 -= float64(k) * m2

	if p2 < 0 {
		p2 += m2
	}

	(*g).cg[3] = (*g).cg[4]
	(*g).cg[4] = (*g).cg[5]
	(*g).cg[5] = p2

	/* Combination */
	if p1 > p2 {
		u = float64((p1 - p2)) * norm
	} else {
		u = float64((p1 - p2 + m1)) * norm
	}

	if g.anti {
		u = 1.0 - u
	}
	return u
}

func (g *RngStream) u01d() float64 {
	var u float64 = g.u01()
	if !g.anti {
		u += g.u01() * fact
		if u < 1.0 {
			return u
		} else {
			return u - 1.0
		}
	} else {
		/* Don't forget that u01() returns 1 - u in the antithetic case */
		u += (g.u01() - 1.0) * fact
		if u < 0.0 {
			return u + 1.0
		} else {
			return u
		}
	}
}

/*
Check that the seeds are legitimate values. Returns 0 if legal seeds,

	-1 otherwise
*/
func checkSeed(seed [6]uint64) bool {

	for i := 0; i < 3; i++ {
		if float64(seed[i]) >= m1 {
			fmt.Println("****************************************")
			fmt.Println("ERROR: Seed is not set")
			fmt.Println("****************************************")
			return false
		}
	}

	for i := 3; i < 6; i++ {
		if float64(seed[i]) >= m2 {
			fmt.Println("****************************************")
			fmt.Println("ERROR: Seed is not set")
			fmt.Println("****************************************")
			return false
		}
	}

	if seed[0] == 0 && seed[1] == 0 && seed[2] == 0 {
		fmt.Println("****************************************")
		fmt.Println("ERROR: First three seeds are zero")
		fmt.Println("****************************************")
		return false
	}

	if seed[3] == 0 && seed[4] == 0 && seed[5] == 0 {
		fmt.Println("****************************************")
		fmt.Println("ERROR: Last three seeds are zero")
		fmt.Println("****************************************")
		return false
	}
	return true
}

// New creates a new stream with (optional) descriptor `name`. It initializes
// its seed Ig, and sets Bg and Cg to Ig. It also sets its `anti` and `incPrec`
// switches to false. The seed Ig is equal to the initial seed of the
// package if this is the first stream created; otherwise it is Z steps
// ahead of the seed of the most recently created stream.
func New(name string) *RngStream {
	g := new(RngStream)

	if len(name) > 0 {
		g.name = name
	} else {
		g.name = ""
	}
	g.anti = false
	g.incPrec = false

	for i := 0; i < 3; i++ {
		g.bg[i] = nextSeedLow[i]
		g.cg[i] = nextSeedLow[i]
		g.ig[i] = nextSeedLow[i]
	}

	for i := 3; i < 6; i++ {
		g.bg[i] = nextSeedHigh[i-3]
		g.cg[i] = nextSeedHigh[i-3]
		g.ig[i] = nextSeedHigh[i-3]
	}

	matVecModM(&a1p127, &nextSeedLow, &nextSeedLow, m1)
	matVecModM(&a2p127, &nextSeedHigh, &nextSeedHigh, m2)
	return g
}

// ResetStartStream Reinitializes the stream to its initial state:
// Cg and Bg are set to Ig.
func (g *RngStream) ResetStartStream() {
	for i := 0; i < 6; i++ {
		g.cg[i] = g.ig[i]
		g.bg[i] = g.ig[i]
	}
}

// ResetNextSubstream reinitializes the stream to the beginning of its next
// substream: Ng is computed, and Cg and Bg are set to Ng.
func (g *RngStream) ResetNextSubstream() {

	modBgLow := [3]float64{g.bg[0], g.bg[1], g.bg[2]}
	matVecModM(&a1p76, &modBgLow, &modBgLow, m1)

	modBgHigh := [3]float64{g.bg[3], g.bg[4], g.bg[5]}
	matVecModM(&a2p76, &modBgHigh, &modBgHigh, m2)

	for i := 0; i < 3; i++ {
		g.bg[i] = modBgLow[i]
		g.bg[i+3] = modBgHigh[i]
	}
	for i := 0; i < 6; i++ {
		g.cg[i] = g.bg[i]
	}
}

// ResetStartSubstream reinitializes the stream to the beginning
// of its current substream: Cg is set to Bg.
func (g *RngStream) ResetStartSubstream() {
	for i := 0; i < 6; i++ {
		g.cg[i] = g.bg[i]
	}
}

// SetPackageSeed sets the initial seed s0 of the package to the six
// integers in the vector seed. The first 3 integers in the seed must
// all be less than m1 = 4294967087, and not all 0; and the last 3
// integers must all be less than m2 = 4294944443, and not all 0.
// If this method is not called, the default
// initial seed is (12345, 12345, 12345, 12345, 12345, 12345). Returns
// false for invalid seeds, and true otherwise.
func SetPackageSeed(seed [6]uint64) bool {
	if !checkSeed(seed) { // note inversion from C version
		return false /* FAILURE */
	}
	for i := 0; i < 3; i++ {
		nextSeedLow[i] = float64(seed[i])
	}

	for i := 0; i < 3; i++ {
		nextSeedHigh[i] = float64(seed[i+3])
	}
	return true /* SUCCESS */
}

// SetSeed sets the initial seed Ig of the stream to the vector
// seed. The vector seed should contain valid seed values as described in
// SetPackageSeed. The state of the stream is then reset to this initial
// seed. The states and seeds of the other streams are not modified. As a
// result, after calling this method, the initial seeds of the streams are
// no longer spaced Z values apart. We discourage the use of this method;
// proper use of the Reset* methods is preferable. Returns false for invalid
// seeds, and true otherwise.
func (g *RngStream) SetSeed(seed [6]uint64) bool {
	if !checkSeed(seed) {
		return false /* FAILURE */
	}

	for i := 0; i < 6; i++ {
		g.cg[i] = float64(seed[i])
		g.bg[i] = float64(seed[i])
		g.ig[i] = float64(seed[i])
	}
	return true /* SUCCESS */
}

// AdvanceState advances the state by n steps (see below for the meaning
// of n), without modifying the states of other streams or the values of
// Bg and Ig in the current object. If e > 0, then n = 2e + c; if e < 0,
// then n = −2−e + c; and if e = 0, then n = c. Note: c is allowed to
// take negative values.  We discourage the use of this method.
func (g *RngStream) AdvanceState(e, c int64) {

	var B1 = [3][3]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
	var C1 = [3][3]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
	var B2 = [3][3]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}
	var C2 = [3][3]float64{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}

	if e > 0 {
		matTwoPowModM(&a1p0, &B1, m1, e)
		matTwoPowModM(&a2p0, &B2, m2, e)
	} else if e < 0 {
		matTwoPowModM(&invA1, &B1, m1, -e)
		matTwoPowModM(&invA2, &B2, m2, -e)
	}

	if c >= 0 {
		matPowModM(&a1p0, &C1, m1, c)
		matPowModM(&a2p0, &C2, m2, c)
	} else {
		matPowModM(&invA1, &C1, m1, -c)
		matPowModM(&invA2, &C2, m2, -c)
	}

	if e != 0 {
		matMatModM(&B1, &C1, &C1, m1)
		matMatModM(&B2, &C2, &C2, m2)
	}

	var gcgLow = [3]float64{g.cg[0], g.cg[1], g.cg[2]}
	var gcgHigh = [3]float64{g.cg[3], g.cg[4], g.cg[5]}

	matVecModM(&C1, &gcgLow, &gcgLow, m1)
	matVecModM(&C2, &gcgHigh, &gcgHigh, m2)
	for i := 0; i < 3; i++ {
		g.cg[i] = gcgLow[i]
		g.cg[i+3] = gcgHigh[i]
	}
}

// GetState returns in seed[0..5] the current state Cg of this stream. This is
// convenient if we want to 14 save the state for subsequent use.
func (g *RngStream) GetState(seed []uint64) {
	for i := 0; i < 6; i++ {
		seed[i] = uint64(g.cg[i])
	}
}

// WriteState writes (to standard output) the current state Cg of this stream.
func (g *RngStream) WriteState() {
	if g == nil {
		return
	}
	fmt.Println(g.RngStreamStateString())
}

func (g *RngStream) RngStreamStateString() string {

	state_str := ""
	if len(g.name) > 0 {
		state_str += (" " + g.name)
	}
	state_str += (":\n  Cg = {")
	vec_str := make([]string, 6)
	for i := 0; i < 6; i++ {
		vec_str[i] = strconv.FormatUint(uint64(g.cg[i]), 10)
	}
	state_str += strings.Join(vec_str, ",")
	state_str += " }\n"
	return state_str
}

// WriteStateFull writes (to standard output) the value of all the
// internal variables of this stream: name, anti, incPrec, Ig, Bg, Cg.
func (g *RngStream) WriteStateFull() {
	fmt.Println(g.RngStreamFullStateString())
}

func (g *RngStream) RngStreamFullStateString() string {
	if g == nil {
		return ""
	}
	state_str := ""
	//state_str := "The RngStream"
	if len(g.name) > 0 {
		state_str += g.name
	}
	state_str += ":\n  Anti = "
	if g.anti {
		state_str += "true\n"
	} else {
		state_str += "false\n"
	}
	state_str += "    IncPrec = "

	if g.incPrec {
		state_str += "true\n"
	} else {
		state_str += "false\n"
	}
	state_str += "    Ig = { "
	vec_str := make([]string, 6)
	for i := 0; i < 6; i++ {
		vec_str[i] = strconv.FormatUint(uint64(g.ig[i]), 10)
	}
	state_str += strings.Join(vec_str, ",")
	state_str += " }\n  Bg = { "

	for i := 0; i < 6; i++ {
		vec_str[i] = strconv.FormatUint(uint64(g.bg[i]), 10)
	}
	state_str += strings.Join(vec_str, ",")
	state_str += " }\n  Cg = { "
	for i := 0; i < 6; i++ {
		vec_str[i] = strconv.FormatUint(uint64(g.bg[i]), 10)
	}
	state_str += strings.Join(vec_str, ",")
	state_str += "}\n"
	//fmt.Println(state_str)
	return state_str
}

// SetIncreasedPrecis writes to the internal incPrec variable.  After calling
// this method with incp = true, each call to the generator (direct or
// indirect) for this stream will return a uniform random number with
// more bits of resolution (53 bits if machine follows IEEE 754 standard)
// instead of 32 bits, and will advance the state of the stream by 2 steps
// instead of 1. More precisely, if s is a stream of the class RngStream,
// in the nonantithetic case, the instruction “u = s.RandU01()” will be
// equivalent to “u = (s.RandU01() + s.RandU01() * fact) % 1.0” where
// the constant fact is equal to 2−24. This also applies when calling
// RandU01 indirectly (e.g., via RandInt, etc.). By default, or if this
// method is called again with incp = false, each call to RandU01 for this
// stream advances the state by 1 step and returns a number with 32 bits
// of resolution.
func (g *RngStream) SetIncreasedPrecis(incp bool) {
	g.incPrec = incp
}

// SetAntitetic write the `anti` internal variable. If a = true, the stream
// will start generating antithetic variates, i.e., 1 − U instead of U, until
// this method is called again with a = false.
func (g *RngStream) SetAntithetic(a bool) {
	g.anti = a
}

// RandU01 normally returns a (pseudo)random number from the uniform
// distribution over the interval (0, 1), after advancing the state by one
// step. The returned number has 32 bits of precision in the sense that it is
// always a multiple of 1/(2^32 −208). However, if IncreasedPrecis(true)
// has been called for this stream, the state is advanced by two steps and
// the returned number has 53 bits of precision.
func (g *RngStream) RandU01() float64 {
	if g.incPrec {
		return g.u01d()
	} else {
		return g.u01()
	}
}

// RandInt returns a (pseudo)random number from the discrete uniform
// distribution over the integers {i, i + 1,...,j} Makes one call to RandU01.
func (g *RngStream) RandInt(i int, j int) int {
	diff := float64(j - i)
	return i + int((diff+1.0)*g.RandU01())
}
