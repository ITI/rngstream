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
// [1] P. L'Ecuyer, ``Good Parameter Sets for Combined
// Multiple Recursive Random Number Generators'',
// Operations Research, 47, 1 (1999), 159--164.  See
// https://www-labs.iro.umontreal.ca/~lecuyer/myftp/papers/opres-combmrg2-1999.pdf
// 
// [2] P. L'Ecuyer, R. Simard, E. J. Chen, and W. D. Kelton, ``An
// Objected-Oriented Random-Number Package with Many Long Streams and
// Substreams'', Operations Research, 50, 6 (2002), 1073--1075 See
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

type RngStream_InfoState struct {
    Cg, Bg, Ig [6]int64
    Anti bool
    IncPrec bool 
    name string}

type RngStream *RngStream_InfoState

/*---------------------------------------------------------------------*/
/* Private part.                                                       */
/*---------------------------------------------------------------------*/

const norm float64 = 2.328306549295727688e-10
const m1 int64 = 4294967087
const m2 int64 =  4294944443
const a12 int64 = 1403580
const a13n int64 = 810728
const a21 int64 =  527612
const a23n int64 = 1370589

const two17 int64 = 131072
const two53 int64 = 9007199254740992
const fact float64 = 5.9604644775390625e-8    /* 1 / 2^24 */


// /* Default initial seed of the package. Will be updated to become
//    the seed of the next created stream. */
var nextSeedLow = [3]int64{12345, 23456, 34567}
var nextSeedHigh = [3]int64{45678, 56789, 67890}

func SetRngStreamMasterSeed(seed int64) {
    nextSeedLow[0] = seed
    nextSeedLow[1] = seed+1
    nextSeedLow[2] = seed+2
    nextSeedHigh[0] = seed+3
    nextSeedHigh[1] = seed+4
    nextSeedHigh[2] = seed+5
}


/* The following are the transition matrices of the two MRG components */
/* (in matrix form), raised to the powers -1, 1, 2^76, and 2^127, resp.*/
var (
	// Inverse of A1p0
	InvA1 = [3][3]int64{
          { 184888585,   0,  1945170933 },
          {         1,   0,           0 },
          {         0,   1,           0 }}

// Inverse of A2p0
	 InvA2 = [3][3]int64{          //
          {      0,  360363334,  4225571728 },
          {      1,          0,           0 },
          {      0,          1,           0 }}

// First MRG component raised to the power 1.
	 A1p0 = [3][3]int64{
          {       0,        1,       0 },
          {       0,        0,       1 },
          { -810728,  1403580,       0 }}

// Second MRG component raised to the power 1.
	 A2p0 = [3][3]int64{
          {        0,        1,       0 },
          {        0,        0,       1 },
          { -1370589,        0,  527612 }}

// First MRG component raised to the power 2^76
	 A1p76 = [3][3]int64{
          {      82758667, 1871391091, 4127413238 }, 
          {    3672831523,   69195019, 1871391091 }, 
          {    3672091415, 3528743235,   69195019 }}

// Second MRG component raised to the power 2^76
	 A2p76 = [3][3]int64{
          {    1511326704, 3759209742, 1610795712 }, 
          {    4292754251, 1511326704, 3889917532 }, 
          {    3859662829, 4292754251, 3708466080 }}

// First MRG component raised to the power 2^127
	 A1p127 = [3][3]int64{
          {    2427906178, 3580155704,  949770784 }, 
          {     226153695, 1230515664, 3580155704 },
          {    1988835001,  986791581, 1230515664 }}

// Second MRG component raised to the power 2^127
	 A2p127 = [3][3]int64{
          {    1464411153,  277697599, 1610723613 },
          {      32183930, 1464411153, 1022607788 },
          {    2824425944,   32183930, 2093834863 }}
)


// Compute (a*s + c) % m. m must be < 2^35.  Works also for s, c < 0
func MultModM (a, s, c, m int64) int64 {
   var v int64
   var a1 int64 

   v = a * s + c

   if ((v >= two53) || (v <= -two53)) {
      a1 := int64(a / two17)
      a -= (a1 * two17)
      v = a1 * s
      a1 = int64(v / m)
      v -= a1 * m
      v = v * two17 + a * s + c
   }
   a1 = int64(v / m)

   tv := v - a1*m
   if ( tv < 0) {
      return v + m
   } else {
      return v
   }
}


// Returns v = A*s % m.  Assumes that -m < s[i] < m.
// Works even if v = s.
func MatVecModM( A *[3][3]int64, s *[3]int64, v *[3]int64, m int64 ) {
   var x [3]int64
   for i := 0; i < 3; i++ {
      x[i] = MultModM ((*A)[i][0], (*s)[0], 0, m)
      x[i] = MultModM ((*A)[i][1], (*s)[1], x[i], m)
      x[i] = MultModM ((*A)[i][2], (*s)[2], x[i], m)
   }

   for i := 0; i < 3; i++ {
      (*v)[i] = x[i]
  }
}


/* Returns C = A*B % m. Work even if A = C or B = C or A = B = C. */
func MatMatModM( A *[3][3]int64, B *[3][3]int64, C *[3][3]int64, m int64) { 
   /* Returns C = A*B % m. Work even if A = C or B = C or A = B = C. */
   var V = [3]int64{0,0,0}

   var W = [3][3]int64{{0,0,0},{0,0,0},{0,0,0}}

   for i := 0; i < 3; i++ {
     for j := 0; j < 3; j++ {
         V[j] = (*B)[j][i]
     }
     MatVecModM (A, &V, &V, m)
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
func MatTwoPowModM(A *[3][3]int64, B *[3][3]int64, m int64, e int) {
	/* initialize: B = A */
   for i := 0; i < 3; i++ {
           for j := 0; j < 3; j++ {
              (*B)[i][j] = (*A)[i][j]
           }
   }

   /* Compute B = A^{2^e} */
   for i := 0; i < e; i++ {
      MatMatModM (B, B, B, m)
  }
}

// Compute matrix B = A^n % m ;  works even if A = B 
func MatPowModM( A *[3][3]int64, B *[3][3]int64, m int64, n int) {

   var W = [3][3]int64{{0,0,0},{0,0,0},{0,0,0}}

   /* initialize: W = A; B = I */
   for i := 0; i < 3; i++ {
       for j := 0; j < 3; j++ {
         W[i][j] = (*A)[i][j];
         (*B)[i][j] = 0;
      }
   }

   for j := 0; j < 3; j++ {
      (*B)[j][j] = 1;
   }

   /* Compute B = A^n % m using the binary decomposition of n */

   for n > 0 {
      if (n % 2 != 0) {
         MatMatModM (&W, B, B, m)
     }
     MatMatModM (&W, &W, &W, m)
     n /= 2
   }
}

func (g *RngStream_InfoState) U01 () float64 {

   var p1, p2 int64
   var u float64

   /* Component 1 */
   p1 = a12 * (*g).Cg[1] - a13n * (*g).Cg[0]
   k := int64(p1 / m1)
   p1 -= k * m1

   if p1 < 0 {
      p1 += m1
   }

   (*g).Cg[0] = (*g).Cg[1]
   (*g).Cg[1] = (*g).Cg[2]
   (*g).Cg[2] = p1

   /* Component 2 */
   p2 = a21 * (*g).Cg[5] - a23n * (*g).Cg[3]
   k = int64( p2 / m2 )
   p2 -= k * m2

   if (p2 < 0) {
      p2 += m2
   }

   (*g).Cg[3] = (*g).Cg[4]
   (*g).Cg[4] = (*g).Cg[5]
   (*g).Cg[5] = p2

   /* Combination */
   if p1 > p2 {
       u = float64((p1 - p2)) * norm
   } else {
       u = float64((p1 - p2 + m1)) * norm
   }

   if g.Anti {
       u = 1.0 - u
    }
   return u
}

func (g *RngStream_InfoState) U01d () float64 {
   var u float64 = g.U01()
   if !g.Anti {
      u += g.U01() * fact;
      if u < 1.0 {
          return u
      } else {
          return u - 1.0
      }
   } else {
      /* Don't forget that U01() returns 1 - u in the antithetic case */
      u += (g.U01() - 1.0) * fact
      if u < 0.0 {
          return u+1.0
      } else {
          return u
      }
   }
}

/* Check that the seeds are legitimate values. Returns 0 if legal seeds,
   -1 otherwise */
func CheckSeed(seed [6]uint64) bool {

     for i := 0; i < 3; i++ {
     if (int64(seed[i]) >= m1) {
         fmt.Println("****************************************")
         fmt.Println("ERROR: Seed is not set")
         fmt.Println("****************************************")
         return false
     }
   }

   for i := 3; i < 6; i++ {
     if (int64(seed[i]) >= m2) {
         fmt.Println("****************************************")
         fmt.Println("ERROR: Seed is not set")
         fmt.Println("****************************************")
         return false
     }
   }

   if (seed[0] == 0 && seed[1] == 0 && seed[2] == 0) {
         fmt.Println("****************************************")
         fmt.Println("ERROR: First three seeds are zero")
         fmt.Println("****************************************")
         return false
   }

   if (seed[3] == 0 && seed[4] == 0 && seed[5] == 0) {
         fmt.Println("****************************************")
         fmt.Println("ERROR: Last three seeds are zero")
         fmt.Println("****************************************")
         return false
   }
   return true
}


func RngStream_CreateStream (name string) RngStream {
   g := new(RngStream_InfoState)

   if len(name) > 0 {
      g.name = name
   } else {
      g.name = "";
   }
   g.Anti = false
   g.IncPrec = false

   for i := 0; i < 3; i++ {
      g.Bg[i] = nextSeedLow[i]
      g.Cg[i] = nextSeedLow[i]
      g.Ig[i] = nextSeedLow[i]
   }

   for i := 3; i < 6; i++ {
      g.Bg[i] = nextSeedHigh[i-3]
      g.Cg[i] = nextSeedHigh[i-3]
      g.Ig[i] = nextSeedHigh[i-3]
   }

   MatVecModM (&A1p127, &nextSeedLow, &nextSeedLow, m1)
   MatVecModM (&A2p127, &nextSeedHigh, &nextSeedHigh, m2)
   return g
}

func RngStream_ResetStartStream (g RngStream ) {
    for i := 0; i < 6; i++ { 
      g.Cg[i] = g.Ig[i]
      g.Bg[i] = g.Ig[i]
    }
}

func RngStream_ResetNextSubstream (g RngStream) {

   modBgLow := [3]int64{g.Bg[0],g.Bg[1],g.Bg[2]}
   MatVecModM (&A1p76, &modBgLow, &modBgLow, m1)

   modBgHigh := [3]int64{g.Bg[3],g.Bg[4],g.Bg[5]}
   MatVecModM (&A2p76, &modBgHigh, &modBgHigh, m2)
  
   for i:= 0; i<3; i++ {
        g.Bg[i]   = modBgLow[i]
        g.Bg[i+3] = modBgHigh[i]
   }
   for i := 0; i < 6; i++ { 
      g.Cg[i] = g.Bg[i]
   }
}

func RngStream_ResetStartSubstream (g RngStream ) {
 for i := 0; i < 6; i++ {
      g.Cg[i] = g.Bg[i]
  }
}

func RngStream_SetPackageSeed (seed [6]uint64) bool {
  if !CheckSeed (seed) {           // note inversion from C version
      return false                    /* FAILURE */
  }
  for i:=0; i<3; i++ {
      nextSeedLow[i] = int64(seed[i])
  }

  for i:=0; i<3; i++ {
      nextSeedHigh[i] = int64(seed[i+3])
  }
  return true                       /* SUCCESS */
}

func RngStream_SetSeed (g RngStream, seed [6]uint64) bool {
  if (!CheckSeed (seed)) {
      return false;                    /* FAILURE */
  }

  for i := 0; i < 6; i++ {
    g.Cg[i] = int64(seed[i])
    g.Bg[i] = int64(seed[i])
    g.Ig[i] = int64(seed[i])
  } 
  return true                       /* SUCCESS */ 
}


func RngStream_AdvanceState (g RngStream, e, c int ) {

   var B1 = [3][3]int64{{0,0,0},{0,0,0},{0,0,0}}
   var C1 = [3][3]int64{{0,0,0},{0,0,0},{0,0,0}}
   var B2 = [3][3]int64{{0,0,0},{0,0,0},{0,0,0}}
   var C2 = [3][3]int64{{0,0,0},{0,0,0},{0,0,0}}

   if (e > 0) {
      MatTwoPowModM (&A1p0, &B1, m1, e)
      MatTwoPowModM (&A2p0, &B2, m2, e)
   } else if (e < 0) {
      MatTwoPowModM (&InvA1, &B1, m1, -e)
      MatTwoPowModM (&InvA2, &B2, m2, -e)
   }

   if (c >= 0) {
      MatPowModM (&A1p0, &C1, m1, c)
      MatPowModM (&A2p0, &C2, m2, c)
   } else {
      MatPowModM (&InvA1, &C1, m1, -c)
      MatPowModM (&InvA2, &C2, m2, -c)
   }

   if (e != 0) {
      MatMatModM (&B1, &C1, &C1, m1)
      MatMatModM (&B2, &C2, &C2, m2)
   }

   var gCgLow = [3]int64{g.Cg[0], g.Cg[1], g.Cg[2]}
   var gCgHigh = [3]int64{g.Cg[3], g.Cg[4], g.Cg[5]}

   MatVecModM (&C1, &gCgLow, &gCgLow, m1)
   MatVecModM (&C2, &gCgHigh, &gCgHigh, m2)
   for i:= 0; i<3; i++ {
       g.Cg[i] = gCgLow[i]
       g.Cg[i+3] = gCgHigh[i]
   }
}

func RngStream_GetState (g RngStream, seed []uint64) {
  for i := 0; i < 6; i++ {
      seed[i] = uint64(g.Cg[i])
  }
}

func RngStream_WriteState (g RngStream) {
    if (g == nil) {
          return
    }
    fmt.Println(RngStreamStateString(g))
}

func RngStreamStateString (g RngStream) string {

    state_str := ""
    if ( len(g.name)>0 ) {
      state_str += (" "+g.name)
    }
    state_str += (":\n  Cg = {")
    vec_str := make([]string,6)
    for i:=0; i<6; i++ {
        vec_str[i] = strconv.FormatUint(uint64(g.Cg[i]),10)
    }
    state_str += strings.Join(vec_str,",")
    state_str += " }\n"
    return state_str
}

func RngStream_WriteStateFull (g RngStream ) {
    fmt.Println(RngStreamFullStateString(g))
}

func RngStreamFullStateString(g RngStream) string {
    if (g == nil) { 
      return ""
   }
   state_str := ""
   //state_str := "The RngStream" 
   if (len(g.name) > 0) {
       state_str += g.name
   }
   state_str += ":\n  Anti = "
   if g.Anti {
       state_str += "true\n"
   } else {
       state_str += "false\n"
   }
   state_str += "    IncPrec = "

   if g.IncPrec {
       state_str += "true\n"
   } else {
       state_str += "false\n"
   }
   state_str += "    Ig = { "
   vec_str := make([]string,6) 
   for i := 0; i < 6; i++ {
      vec_str[i] = strconv.FormatUint(uint64(g.Ig[i]),10)
   }
   state_str += strings.Join(vec_str,",")
   state_str += " }\n  Bg = { "

   for i := 0; i < 6; i++ {
      vec_str[i] = strconv.FormatUint(uint64(g.Bg[i]),10)
   }
   state_str += strings.Join(vec_str,",")
   state_str += " }\n  Cg = { "
   for i := 0; i < 6; i++ {
      vec_str[i] = strconv.FormatUint(uint64(g.Bg[i]),10)
   }
   state_str += strings.Join(vec_str,",")
   state_str += "}\n"
   //fmt.Println(state_str)
   return state_str
}


func RngStream_IncreasedPrecis (g RngStream,  incp bool) {
   g.IncPrec = incp
}

func RngStream_SetAntithetic (g RngStream, a bool) {
   g.Anti = a
}

func (g *RngStream_InfoState) RngStream_RandU01 () float64 {
  if g.IncPrec {
      return g.U01d ()
  } else {
      return g.U01 ()
  }
}

func (g *RngStream_InfoState) RandInt (i int, j int) int {
   diff := float64(j-i) 
   return i + int( ((diff + 1.0)) * g.U01 () )
}
