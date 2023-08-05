// Copyright 2023 University of Illinois Board of Trustees.
// All rights reserved.

package rngstream_test

import (
	"fmt"
	"github.com/iti/rngstream"
)

var initialSeed = []uint64{12345, 12345, 12345, 12345, 12345, 12345}

func ExampleNew() {
	// Reset seed to make test deterministic
	rngstream.SetPackageSeed(initialSeed)

	// Create a random number generator
	g := rngstream.New("g")

	// Fetch some pseudo-random numbers
	f1 := g.RandU01()
	f2 := g.RandU01()

	fmt.Printf("%.5f %.5f\n", f1, f2)
	// Output: 0.12701 0.31853
}

func ExampleRngStream_RandU01() {
	// Reset seed to make test deterministic
	rngstream.SetPackageSeed(initialSeed)

	// Create a random number generator
	g := rngstream.New("g")

	// Fetch some pseudo-random numbers
	f1 := g.RandU01()
	f2 := g.RandU01()

	fmt.Printf("%.5f %.5f\n", f1, f2)
	// Output: 0.12701 0.31853
}

func ExampleNew_double() {
	// Reset seed to make test deterministic
	rngstream.SetPackageSeed(initialSeed)

	// Each call to New gets a distinct seed
	g1 := rngstream.New("g1")
	g2 := rngstream.New("g2")

	// Fetch some pseudo-random numbers
	f1_1 := g1.RandU01()
	f1_2 := g1.RandU01()
	f2_1 := g2.RandU01()
	f2_2 := g2.RandU01()

	fmt.Printf("%.5f %.5f %.5f %.5f\n", f1_1, f1_2, f2_1, f2_2)
	// Output: 0.12701 0.31853 0.75958 0.97831
}

func ExampleSetPackageSeed_equal() {
	// Construct two generators with identical seeds
	rngstream.SetPackageSeed(initialSeed)
	g1 := rngstream.New("g1")

	rngstream.SetPackageSeed(initialSeed)
	g2 := rngstream.New("g2")

	b1 := g1.RandU01() == g2.RandU01()
	b2 := g1.RandU01() == g2.RandU01()

	fmt.Printf("%v %v\n", b1, b2)
	// Output: true true
}

func ExampleSetPackageSeed_notequal() {
	// Construct two generators, but allow the seed to advance for second
	rngstream.SetPackageSeed(initialSeed)
	g1 := rngstream.New("g1")
	g2 := rngstream.New("g2")

	b1 := g1.RandU01() == g2.RandU01()
	b2 := g1.RandU01() == g2.RandU01()

	fmt.Printf("%v %v\n", b1, b2)
	// Output: false false
}

func ExampleSetRngStreamMasterSeed_equal() {
	// Construct two generators with identical seeds
	rngstream.SetRngStreamMasterSeed(5555)
	g1 := rngstream.New("g1")

	rngstream.SetRngStreamMasterSeed(5555)
	g2 := rngstream.New("g2")

	b1 := g1.RandU01() == g2.RandU01()
	b2 := g1.RandU01() == g2.RandU01()

	fmt.Printf("%v %v\n", b1, b2)
	// Output: true true
}

func ExampleSetRngStreamMasterSeed_notequal() {
	// Construct two generators, but allow the seed to advance for second
	rngstream.SetRngStreamMasterSeed(5555)
	g1 := rngstream.New("g1")
	g2 := rngstream.New("g2")

	b1 := g1.RandU01() == g2.RandU01()
	b2 := g1.RandU01() == g2.RandU01()

	fmt.Printf("%v %v\n", b1, b2)
	// Output: false false
}

func ExampleRngStream() {
	var g *rngstream.RngStream

	// Reset seed to make test deterministic
	rngstream.SetPackageSeed(initialSeed)

	g = rngstream.New("g")
	b := g.RandU01() == g.RandU01()
	fmt.Printf("%v\n", b)
	// Output: false
}

func ExampleRngStream_AdvanceState_equal() {
	// Construct two generators with identical seeds
	rngstream.SetRngStreamMasterSeed(5555)
	g1 := rngstream.New("g1")

	rngstream.SetRngStreamMasterSeed(5555)
	g2 := rngstream.New("g2")

	// Consume 10 random numbers
	for i := 0; i < 10; i++ {
		g1.RandU01()
	}

	// Consume 10 random numbers
	g2.AdvanceState(0, 10)

	b := g1.RandU01() == g2.RandU01()
	fmt.Printf("%v\n", b)
	// Output: true
}

func ExampleRngStream_GetState() {
	// Construct generator with known state
	rngstream.SetRngStreamMasterSeed(5555)
	g1 := rngstream.New("g1")

	// Consume 10 random numbers
	g1.AdvanceState(0, 10)

	// Grab the current state & get a random number.
	state := g1.GetState()
	r1 := g1.RandU01()

	// Consume 10 more random numbers
	g1.AdvanceState(0, 10)

	// Restore the state & get a random number.
	g1.SetSeed(state)
	r2 := g1.RandU01()

	b := r1 == r2
	fmt.Printf("%v\n", b)
	// Output: true
}

func ExampleRngStream_SetSeed() {
	// Construct generator with known state
	rngstream.SetRngStreamMasterSeed(5555)
	g1 := rngstream.New("g1")

	// Consume 10 random numbers
	g1.AdvanceState(0, 10)

	// Grab the current state & get a random number.
	state := g1.GetState()
	r1 := g1.RandU01()

	// Consume 10 more random numbers
	g1.AdvanceState(0, 10)

	// Restore the state & get a random number.
	g1.SetSeed(state)
	r2 := g1.RandU01()

	b := r1 == r2
	fmt.Printf("%v\n", b)
	// Output: true
}

func ExampleRngStream_RandInt() {
	// Construct generator with known state
	rngstream.SetRngStreamMasterSeed(5555)
	g1 := rngstream.New("g1")

	// Generate some numbers
	fmt.Printf("%v %v %v %v\n",
		g1.RandInt(1, 3),
		g1.RandInt(1, 3),
		g1.RandInt(1, 3),
		g1.RandInt(1, 3))
	// Output: 3 3 1 2
}

func ExampleRngStream_WriteState() {
	// Reset seed to make test deterministic
	rngstream.SetPackageSeed(initialSeed)

	// Construct a generator
	g := rngstream.New("g")

	// Initial state value
	g.WriteState()

	// Consume 10 random numbers
	for i := 0; i < 10; i++ {
		g.RandU01()
	}

	// Current state value
	g.WriteState()

	// Output:
	// g:
	//   Cg = {12345,12345,12345,12345,12345,12345 }
	//
	//  g:
	//   Cg = {2989318136,3378525425,1773647758,1462200156,2794459678,2822254363 }
}

func ExampleRngStream_WriteStateFull() {
	// Reset seed to make test deterministic
	rngstream.SetPackageSeed(initialSeed)

	// Construct a generator
	g := rngstream.New("g")

	// Initial state value
	g.WriteStateFull()

	// Output:
	// g:
	//   Anti = false
	//     IncPrec = false
	//     Ig = { 12345,12345,12345,12345,12345,12345 }
	//   Bg = { 12345,12345,12345,12345,12345,12345 }
	//   Cg = { 12345,12345,12345,12345,12345,12345}

}
