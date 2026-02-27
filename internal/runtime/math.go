package runtime

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// errIllegalFunctionCall is the canonical BASIC error for invalid arguments.
const errIllegalFunctionCall = "Illegal function call"

// Abs returns the absolute value of x.
func Abs(x float64) float64 {
	return math.Abs(x)
}

// Sgn returns the sign of x: -1 if x < 0, 0 if x == 0, 1 if x > 0.
func Sgn(x float64) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// IntFloor returns the largest integer less than or equal to x (BASIC's INT).
// INT(-2.5) = -3, INT(2.5) = 2
func IntFloor(x float64) float64 {
	return math.Floor(x)
}

// Fix truncates the fractional part toward zero (BASIC's FIX).
// FIX(-2.5) = -2, FIX(2.5) = 2
func Fix(x float64) float64 {
	return math.Trunc(x)
}

// Ceil returns the smallest integer >= x. Turbo BASIC extension.
func Ceil(x float64) float64 {
	return math.Ceil(x)
}

// Sqr returns the square root. Returns error for negative input.
func Sqr(x float64) (float64, error) {
	if x < 0 {
		return 0, fmt.Errorf("%s: square root of negative number", errIllegalFunctionCall)
	}
	return math.Sqrt(x), nil
}

// Exp returns e^x.
func Exp(x float64) float64 {
	return math.Exp(x)
}

// Exp2 returns 2^x. Turbo BASIC extension.
func Exp2(x float64) float64 {
	return math.Exp2(x)
}

// Exp10 returns 10^x. Turbo BASIC extension.
func Exp10(x float64) float64 {
	return math.Pow(10, x)
}

// Log returns the natural logarithm. Returns error for x <= 0.
func Log(x float64) (float64, error) {
	if x <= 0 {
		return 0, fmt.Errorf("%s: logarithm of non-positive number", errIllegalFunctionCall)
	}
	return math.Log(x), nil
}

// Log2 returns the base-2 logarithm. Turbo BASIC extension.
// Returns error for x <= 0.
func Log2(x float64) (float64, error) {
	if x <= 0 {
		return 0, fmt.Errorf("%s: logarithm of non-positive number", errIllegalFunctionCall)
	}
	return math.Log2(x), nil
}

// Log10 returns the base-10 logarithm. Turbo BASIC extension.
// Returns error for x <= 0.
func Log10(x float64) (float64, error) {
	if x <= 0 {
		return 0, fmt.Errorf("%s: logarithm of non-positive number", errIllegalFunctionCall)
	}
	return math.Log10(x), nil
}

// Sin returns the sine (argument in radians).
func Sin(x float64) float64 {
	return math.Sin(x)
}

// Cos returns the cosine (argument in radians).
func Cos(x float64) float64 {
	return math.Cos(x)
}

// Tan returns the tangent (argument in radians).
func Tan(x float64) float64 {
	return math.Tan(x)
}

// Atn returns the arctangent (result in radians).
func Atn(x float64) float64 {
	return math.Atan(x)
}

// Cint converts to int16 with rounding. Returns error on overflow.
// BASIC's CINT rounds and must fit in -32768 to 32767.
func Cint(x float64) (int16, error) {
	rounded := math.RoundToEven(x)
	if rounded < -32768 || rounded > 32767 {
		return 0, fmt.Errorf("%s: overflow in CINT (value %g out of int16 range)", errIllegalFunctionCall, x)
	}
	return int16(rounded), nil
}

// Clng converts to int32 with rounding. Returns error on overflow.
func Clng(x float64) (int32, error) {
	rounded := math.RoundToEven(x)
	if rounded < -2147483648 || rounded > 2147483647 {
		return 0, fmt.Errorf("%s: overflow in CLNG (value %g out of int32 range)", errIllegalFunctionCall, x)
	}
	return int32(rounded), nil
}

// Csng converts to float32.
func Csng(x float64) float32 {
	return float32(x)
}

// Cdbl converts to float64 (identity, but included for completeness).
func Cdbl(x float64) float64 {
	return x
}

// ---------------------------------------------------------------------------
// Random number generator (BASIC's RND / RANDOMIZE)
// ---------------------------------------------------------------------------

// RNG holds the state for BASIC's RND function.
type RNG struct {
	src  *rand.Rand
	last float64
}

// NewRNG creates a new random number generator with a time-based seed.
func NewRNG() *RNG {
	seed := time.Now().UnixNano()
	r := &RNG{
		src: rand.New(rand.NewSource(seed)),
	}
	// Generate an initial value so that RND(0) has something to return.
	r.last = r.src.Float64()
	return r
}

// Rnd implements BASIC's RND function:
//   - RND or RND(positive): return next random number 0 <= n < 1
//   - RND(0): return the last generated number
//   - RND(negative): reseed with the given value, return first number from new sequence
func (r *RNG) Rnd(arg float64) float64 {
	switch {
	case arg < 0:
		// Reseed with the absolute value of the argument cast to int64.
		// Using the raw bits ensures different negative values give different sequences.
		r.src = rand.New(rand.NewSource(int64(math.Float64bits(arg))))
		r.last = r.src.Float64()
	case arg == 0:
		// Return the last generated number; do not advance.
		return r.last
	default:
		// arg > 0 (or called without argument, which the caller passes as 1).
		r.last = r.src.Float64()
	}
	return r.last
}

// Randomize reseeds the RNG. If seed is 0, use time-based seed.
func (r *RNG) Randomize(seed float64) {
	var s int64
	if seed == 0 {
		s = time.Now().UnixNano()
	} else {
		s = int64(seed)
	}
	r.src = rand.New(rand.NewSource(s))
	// Generate an initial value so RND(0) returns a valid number after reseed.
	r.last = r.src.Float64()
}
