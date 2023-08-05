/*  Programme pour tester le generateur   RngStreams.c  */

package rngstream

import (
	"fmt"
	"math"
	"testing"
)

func Test2(t *testing.T) {
	SetPackageSeed([]uint64{12345, 12345, 12345, 12345, 12345, 12345})
	var sum = 0.0
	var sum3 = 0.0
	var sumi = 0
	var i int

	var gar [4]*RngStream
	germe := []uint64{1, 1, 1, 1, 1, 1}

	g1 := New("g1")
	g2 := New("g2")
	g3 := New("g3")

	fmt.Printf("Initial states of g1, g2, and g3:\n\n")
	g1.WriteState()
	g2.WriteState()
	g3.WriteState()
	sum = g2.RandU01() + g3.RandU01()
	for i = 0; i < 12345; i++ {
		g2.RandU01()
	}

	g1.AdvanceState(5, 3)
	fmt.Printf("State of g1 after advancing by 2^5 + 3 = 35 steps:\n")
	g1.WriteState()
	fmt.Printf("RandU01 (g1) = %12.8f\n\n", g1.RandU01())

	g1.ResetStartStream()
	for i = 0; i < 35; i++ {
		g1.AdvanceState(0, 1)
	}
	fmt.Printf("State of g1 after reset and advancing 35 times by 1:\n")
	g1.WriteState()
	fmt.Printf("RandU01 (g1) = %12.8f\n\n", g1.RandU01())

	g1.ResetStartStream()
	for i = 0; i < 35; i++ {
		sumi += g1.RandInt(1, 10)
	}
	fmt.Printf("State of g1 after reset and 35 calls to RandInt (1, 10):\n")
	g1.WriteState()
	fmt.Printf("   sum of 35 integers in [1, 10] = %v\n\n", sumi)
	sum += float64(sumi) / 100.0
	fmt.Printf("RandU01 (g1) = %12.8f\n\n", g1.RandU01())

	sum3 = 0.0
	g1.ResetStartStream()
	g1.SetIncreasedPrecis(true)
	sumi = 0
	for i = 0; i < 17; i++ {
		fmt.Printf("State after reset, IncreasedPrecis(true), and %d calls to RandInt(1, 10):\n", i)
		g1.WriteState()
		sumi += g1.RandInt(1, 10)
	}
	fmt.Printf("State of g1 after reset, IncreasedPrecis (1) and 17 calls to RandInt (1, 10):\n")
	g1.WriteState()
	g1.SetIncreasedPrecis(false)
	g1.RandInt(1, 10)
	fmt.Printf("State of g1 after IncreasedPrecis (0) and 1 call to RandInt\n")
	g1.WriteState()
	sum3 = float64(sumi) / 10.0

	g1.ResetStartStream()
	g1.SetIncreasedPrecis(true)
	for i = 0; i < 17; i++ {
		sum3 += g1.RandU01()
	}
	fmt.Printf("State of g1 after reset, IncreasedPrecis (1) and 17 calls to RandU01:\n")
	g1.WriteState()
	g1.SetIncreasedPrecis(false)
	g1.RandU01()
	fmt.Printf("State of g1 after IncreasedPrecis (0) and 1 call to RandU01\n")
	g1.WriteState()
	sum += sum3 / 10.0

	sum3 = 0.0
	fmt.Printf("Sum of first 100 output values from stream g3:\n")
	for i = 0; i < 100; i++ {
		sum3 += g3.RandU01()
	}
	fmt.Printf("   sum = %12.6f\n\n", sum3)
	sum += sum3 / 10.0

	fmt.Printf("\nReset stream g3 to its initial seed.\n")
	g3.ResetStartSubstream()
	fmt.Printf("First 5 output values from stream g3:\n")
	for i = 1; i <= 5; i++ {
		fmt.Printf("%12.8f\n", g3.RandU01())
	}
	sum += g3.RandU01()

	fmt.Printf("\nReset stream g3 to the next Substream, 4 times.\n")
	for i = 1; i <= 4; i++ {
		g3.ResetNextSubstream()
	}
	fmt.Printf("First 5 output values from stream g3, fourth Substream:\n")
	for i = 1; i <= 5; i++ {
		fmt.Printf("%12.8f\n", g3.RandU01())
	}
	sum += g3.RandU01()

	fmt.Printf("\nReset stream g2 to the beginning of Substream.\n")
	g2.ResetStartSubstream()
	fmt.Printf(" Sum of 100000 values from stream g2 with double precision:   ")
	g2.SetIncreasedPrecis(true)
	sum3 = 0.0
	for i = 1; i <= 100000; i++ {
		sum3 += g2.RandU01()
	}
	fmt.Printf("%12.4f\n", sum3)
	sum += sum3 / 10000.0
	g2.SetIncreasedPrecis(false)

	g3.SetAntithetic(true)
	fmt.Printf(" Sum of 100000 antithetic output values from stream g3:   ")
	sum3 = 0.0
	for i = 1; i <= 100000; i++ {
		sum3 += g3.RandU01()
	}
	fmt.Printf("%12.4f\n", sum3)
	sum += sum3 / 10000.0

	fmt.Printf("\nSetPackageSeed to seed = { 1, 1, 1, 1, 1, 1 }\n")
	SetPackageSeed(germe)

	fmt.Printf("\nCreate an array of 4 named streams and write their full state\n")
	gar[0] = New("Poisson")
	gar[1] = New("Laplace")
	gar[2] = New("Galois")
	gar[3] = New("Cantor")

	for i = 0; i < 4; i++ {
		gar[i].WriteStateFull()
	}

	fmt.Printf("Jump stream Galois by 2^127 steps backward\n")
	gar[2].AdvanceState(-127, 0)
	gar[2].WriteState()
	gar[2].ResetNextSubstream()

	for i = 0; i < 4; i++ {
		sum += gar[i].RandU01()
	}

	fmt.Printf("-----------------------------------------------------\n")
	fmt.Printf("This test program should print the number  23.705324\n\n")
	fmt.Printf("Actual test result = %.6f", sum)

	if math.Abs(sum-23.705324) < 1e-6 {
		fmt.Printf("\t ... ok\n\n")
	} else {
		fmt.Printf("\t ... failed\n\n")
		t.Error("mismatch")
	}

}
