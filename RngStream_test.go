package rngstream

import (
	"fmt"
	"math"
	"testing"
)

// Compute (a*s + c) % m. m must be < 2^35.  Works also for s, c < 0
// func multModM(a, s, c, m int64) int64 {
func TestMultModM(t *testing.T) {
	SetPackageSeed([6]uint64{12345, 12345, 12345, 12345, 12345, 12345})
	got := multModM(3, 5, 11, 7)
	want := float64((3*5 + 11) % 7)
	fmt.Printf("TestMultModM: got %v; expected %v\n", got, want)
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}

func Test1(t *testing.T) {
	SetPackageSeed([6]uint64{12345, 12345, 12345, 12345, 12345, 12345})
	var sum float64 = 0.0
	var sum3 float64 = 0.0
	var sumi int = 0

	germe := [6]uint64{1, 1, 1, 1, 1, 1}

	g1 := New("g1")
	g2 := New("g2")
	g3 := New("g3")

	sum = g2.RandU01() + g3.RandU01()

	g1.AdvanceState(5, 3)
	sum += g1.RandU01()

	g1.ResetStartStream()
	for i := 0; i < 35; i++ {
		g1.AdvanceState(0, 1)
	}
	sum += g1.RandU01()

	g1.ResetStartStream()
	sumi = 0
	for i := 0; i < 35; i++ {
		sumi += g1.RandInt(1, 10)
	}
	sum += float64(sumi) / 100.0

	sum3 = 0.0
	for i := 0; i < 100; i++ {
		sum3 += g3.RandU01()
	}
	sum += sum3 / 10.0

	g3.ResetStartStream()
	for i := 1; i <= 5; i++ {
		sum += g3.RandU01()
	}

	for i := 0; i < 4; i++ {
		g3.ResetNextSubstream()
	}
	for i := 0; i < 5; i++ {
		sum += g3.RandU01()
	}

	g3.ResetStartSubstream()
	for i := 0; i < 5; i++ {
		sum += g3.RandU01()
	}

	g2.ResetNextSubstream()
	sum3 = 0.0
	for i := 1; i <= 100000; i++ {
		sum3 += g2.RandU01()
	}
	sum += sum3 / 10000.0

	g3.SetAntithetic(true)
	sum3 = 0.0
	for i := 1; i <= 100000; i++ {
		sum3 += g3.RandU01()
	}
	sum += sum3 / 10000.0

	SetPackageSeed(germe)
	gar := make([]*RngStream, 4)
	gar[0] = New("Poisson")
	gar[1] = New("Laplace")
	gar[2] = New("Galois")
	gar[3] = New("Cantor")

	for i := 0; i < 4; i++ {
		sum += gar[i].RandU01()
	}

	gar[2].AdvanceState(-127, 0)
	sum += gar[2].RandU01()

	gar[2].ResetNextSubstream()
	gar[2].SetIncreasedPrecis(true)
	sum3 = 0.0
	for i := 0; i < 100000; i++ {
		sum3 += gar[2].RandU01()
	}
	sum += sum3 / 10000.0

	gar[2].SetAntithetic(true)
	sum3 = 0.0
	for i := 0; i < 100000; i++ {
		sum3 += gar[2].RandU01()
	}
	sum += sum3 / 10000.0
	gar[2].SetAntithetic(false)

	gar[2].SetIncreasedPrecis(false)

	for i := 0; i < 4; i++ {
		sum += gar[i].RandU01()
	}

	fmt.Printf("-----------------------------------------------------\n")
	fmt.Printf("This test program should print the number   39.697547 \n\n")
	fmt.Printf("Actual test result = %.6f", sum)

	if math.Abs(sum-39.697547) < 1e-6 {
		fmt.Printf("\t ... ok\n\n")
	} else {
		t.Error("no match")
	}

}
