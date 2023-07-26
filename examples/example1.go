/*  Program to test the random number streams file:    RngStream.c   */
/* Build: `go build example1` */
package main

import (
	_ "fmt"
	"github.com/illinoisrobert/rngstream"
)

const (
	NS = 1000000
)

func main() {
	/* Create 3 parallel streams */
	g1 := rngstream.RngStream_CreateStream("Poisson")
	g2 := rngstream.RngStream_CreateStream("Cantor")
	g3 := rngstream.RngStream_CreateStream("Laplace")

	/* Generate 35 random integers in [5, 10] with stream g1 */
	for i := 0; i < 35; i++ {
		_ = (*g1).RandInt(5, 10)
	}

	/* Generate 100 random reals in (0, 1) with stream g3 */
	for i := 0; i < 100; i++ {
		_ = (*g3).RngStream_RandU01()
	}

	/* Restart stream g3 in its initial state and generate the same 100
	   random reals as above */
	rngstream.RngStream_ResetStartStream(g3)
	for i := 0; i < 100; i++ {
		_ = (*g3).RngStream_RandU01()
	}

	/* Send stream g3 to its next substream and generate 5
	   random reals in (0, 1) with double precision */
	rngstream.RngStream_ResetNextSubstream(g3)
	rngstream.RngStream_IncreasedPrecis(g3, true)
	for i := 0; i < 5; i++ {
		_ = (*g3).RngStream_RandU01()
	}

	/* Generate 100000 antithetic random reals in (0, 1) with stream g2 */
	rngstream.RngStream_SetAntithetic(g2, true)
	for i := 0; i < 100000; i++ {
		_ = (*g2).RngStream_RandU01()
	}

	/* Create NS = 1000000 parallel streams */
	gar := make([]rngstream.RngStream, NS)
	for i := 0; i < NS; i++ {
		gar[i] = rngstream.RngStream_CreateStream("")
	}

	/* Generate 1000 random real in (0, 1) with stream 55555 in gar */
	for i := 0; i < 1000; i++ {
		_ = (*gar[55554]).RngStream_RandU01()
	}

}
