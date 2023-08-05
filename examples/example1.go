// The package is copyrighted by Pierre L'Ecuyer and the
// University of Montreal.
// Go translation Copyright 2023 University of Illinois Board of Trustees.
// See LICENSE.md for details.
// SPDX-License-Identifier: MIT

// Example code.
package main

import (
	_ "fmt"
	"github.com/iti/rngstream"
)

const (
	nStreams = 1000000
)

func main() {
	/* Create 3 parallel streams */
	g1 := rngstream.New("Poisson")
	g2 := rngstream.New("Cantor")
	g3 := rngstream.New("Laplace")

	/* Generate 35 random integers in [5, 10] with stream g1 */
	for i := 0; i < 35; i++ {
		_ = g1.RandInt(5, 10)
	}

	/* Generate 100 random reals in (0, 1) with stream g3 */
	for i := 0; i < 100; i++ {
		_ = g3.RandU01()
	}

	/* Restart stream g3 in its initial state and generate the same 100
	   random reals as above */
	g3.ResetStartStream()
	for i := 0; i < 100; i++ {
		_ = g3.RandU01()
	}

	/* Send stream g3 to its next substream and generate 5
	   random reals in (0, 1) with double precision */
	g3.ResetNextSubstream()
	g3.SetIncreasedPrecis(true)
	for i := 0; i < 5; i++ {
		_ = (*g3).RandU01()
	}

	/* Generate 100000 antithetic random reals in (0, 1) with stream g2 */
	g2.SetAntithetic(true)
	for i := 0; i < 100000; i++ {
		_ = (*g2).RandU01()
	}

	/* Create nStreams = 1000000 parallel streams */
	gar := make([]*rngstream.RngStream, nStreams)
	for i := 0; i < nStreams; i++ {
		gar[i] = rngstream.New("")
	}

	/* Generate 1000 random real in (0, 1) with stream 55555 in gar */
	for i := 0; i < 1000; i++ {
		_ = (*gar[55554]).RandU01()
	}
}
