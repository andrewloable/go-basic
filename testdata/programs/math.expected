package main

import (
	"fmt"
	"math"
	"os"

	rt "github.com/loabletech/go-basic/internal/runtime"
)

// Suppress unused import warnings.
var _ = fmt.Sprintf
var _ = math.Abs
var _ = os.Exit
var _ = rt.Abs

func main() {
	// Test math functions
	fmt.Println("ABS(-5):   ", rt.Abs((-float64(5))))
	fmt.Println("SGN(-3):   ", rt.Sgn((-float64(3))))
	fmt.Println("SGN(0):    ", rt.Sgn(float64(0)))
	fmt.Println("SGN(7):    ", rt.Sgn(float64(7)))
	fmt.Println("INT(3.7):  ", rt.IntFloor(3.7))
	fmt.Println("INT(-3.7): ", rt.IntFloor((-3.7)))
	fmt.Println("FIX(3.7):  ", rt.Fix(3.7))
	fmt.Println("FIX(-3.7): ", rt.Fix((-3.7)))
	fmt.Println("SQR(16):   ", func() float64 { v_, _ := rt.Sqr(float64(16)); return v_ }())
	fmt.Println("SIN(0):    ", rt.Sin(float64(0)))
	fmt.Println("COS(0):    ", rt.Cos(float64(0)))
	fmt.Println("LOG(1):    ", func() float64 { v_, _ := rt.Log(float64(1)); return v_ }())
	fmt.Println("EXP(0):    ", rt.Exp(float64(0)))
	fmt.Println("2^10:      ", math.Pow(float64(2), float64(10)))
	os.Exit(0)
}
