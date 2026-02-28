// math.go — Mathematical built-in functions for the BASIC runtime.
//
// # Compiler Design Note: Wrapping a Host-Language Library
//
// Turbo BASIC provides a rich set of mathematical functions as part of the
// language itself (SIN, COS, LOG, SQR, …). When we transpile BASIC to Go we
// need those functions to exist somewhere in Go. The simplest strategy is to
// wrap Go's own math package.
//
// Type signatures here accept float64 — BASIC's double-precision type (#).
// The compiler inserts explicit type casts in the generated call sites when the
// BASIC source uses a narrower type (!, %, &), so this package never needs to
// handle them directly. This is a common transpiler technique: normalise to the
// widest type in the runtime to minimise the number of overloads needed.
//
// Each wrapping function also enforces the domain rules that BASIC mandates
// (e.g. SQR of a negative number is an error, not a NaN) and returns a Go
// error so the generated code can propagate it to the user.

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
// BASIC: ABS(x) — always non-negative regardless of sign.
func Abs(x float64) float64 {
	return math.Abs(x)
}

// Sgn returns the sign of x: -1 if x < 0, 0 if x == 0, 1 if x > 0.
// BASIC: SGN(x) — useful for normalising direction vectors in game code.
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
//
// Note that BASIC's INT floors toward negative infinity, not toward zero.
// This differs from most language truncation operators. The function is named
// IntFloor rather than Int to avoid shadowing Go's built-in int type.
func IntFloor(x float64) float64 {
	return math.Floor(x)
}

// Fix truncates the fractional part toward zero (BASIC's FIX).
// FIX(-2.5) = -2, FIX(2.5) = 2
//
// FIX and INT differ only for negative numbers. This distinction matters when
// implementing index calculations or range-checking in generated BASIC programs.
func Fix(x float64) float64 {
	return math.Trunc(x)
}

// Ceil returns the smallest integer >= x. Turbo BASIC extension.
// BASIC: CEIL(x) — the ceiling function, opposite of INT (floor).
func Ceil(x float64) float64 {
	return math.Ceil(x)
}

// Sqr returns the square root. Returns error for negative input.
// BASIC: SQR(x) — errors on negative numbers rather than returning NaN,
// matching Turbo BASIC's "Illegal function call" behaviour.
func Sqr(x float64) (float64, error) {
	if x < 0 {
		return 0, fmt.Errorf("%s: square root of negative number", errIllegalFunctionCall)
	}
	return math.Sqrt(x), nil
}

// Exp returns e^x (e raised to the power x).
// BASIC: EXP(x) — the natural exponential function, inverse of LOG.
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

// Log returns the natural logarithm (base e). Returns error for x <= 0.
// BASIC: LOG(x) — used for exponential growth/decay calculations.
// The domain restriction (x > 0) is a mathematical requirement enforced here
// so the generated program gets a meaningful error instead of -Inf or NaN.
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

// Sin returns the sine of x (argument in radians).
// BASIC: SIN(x) — like all BASIC trig functions, expects radians, not degrees.
func Sin(x float64) float64 {
	return math.Sin(x)
}

// Cos returns the cosine of x (argument in radians).
// BASIC: COS(x) — often paired with SIN for circular/polar coordinate work.
func Cos(x float64) float64 {
	return math.Cos(x)
}

// Tan returns the tangent of x (argument in radians).
// BASIC: TAN(x) — undefined at odd multiples of π/2 (returns ±Inf there).
func Tan(x float64) float64 {
	return math.Tan(x)
}

// Atn returns the arctangent of x (result in radians, range -π/2 to +π/2).
// BASIC: ATN(x) — Turbo BASIC only provides ATN; other inverse trig functions
// (ASIN, ACOS) must be derived from ATN using trigonometric identities.
func Atn(x float64) float64 {
	return math.Atan(x)
}

// Cint converts a float64 to int16 with banker's rounding (round half to even).
// Returns error on overflow.
// BASIC: CINT(x) — converts to integer (%) type; overflow triggers error 6.
//
// Compiler design note: BASIC has four numeric types (%, &, !, #). The
// CINTx/CLNG/CSNG/CDBL family are the explicit type-cast functions that appear
// in the source. The code generator emits calls to these when the programmer
// writes CINT(expr), and also inserts them implicitly during type-coercion in
// assignments (e.g. assigning a float to an integer variable).
func Cint(x float64) (int16, error) {
	rounded := math.RoundToEven(x)
	if rounded < -32768 || rounded > 32767 {
		return 0, fmt.Errorf("%s: overflow in CINT (value %g out of int16 range)", errIllegalFunctionCall, x)
	}
	return int16(rounded), nil
}

// Clng converts a float64 to int32 with banker's rounding. Returns error on overflow.
// BASIC: CLNG(x) — converts to long integer (&) type; range -2,147,483,648 to 2,147,483,647.
func Clng(x float64) (int32, error) {
	rounded := math.RoundToEven(x)
	if rounded < -2147483648 || rounded > 2147483647 {
		return 0, fmt.Errorf("%s: overflow in CLNG (value %g out of int32 range)", errIllegalFunctionCall, x)
	}
	return int32(rounded), nil
}

// Csng converts a float64 to float32 (single precision).
// BASIC: CSNG(x) — converts to single-precision float (!) type; loses precision.
func Csng(x float64) float32 {
	return float32(x)
}

// Cdbl converts to float64 (identity for an already-float64 value).
// BASIC: CDBL(x) — converts to double-precision float (#) type.
// Although this is a no-op at the Go level, having an explicit function keeps
// the code generator simple: it always emits rt.Cdbl(expr) for CDBL() calls
// without needing a special case for "already the right type."
func Cdbl(x float64) float64 {
	return x
}

// ---------------------------------------------------------------------------
// Random number generator (BASIC's RND / RANDOMIZE)
//
// Compiler design note: stateful runtime objects
//
// Most runtime functions are pure (given the same inputs they return the same
// output). RND is different — it maintains hidden state (the seed and last
// result) that persists across calls. The generated Go program creates a single
// *RNG at startup and passes it to every rt.Rnd() call site.
//
// This is a general pattern: whenever the source language has global mutable
// state (random number generator, open file table, screen mode, error handler),
// the runtime models it as a struct with methods and the code generator emits
// a global variable declaration in the program's main() initialisation block.
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

// Rnd implements BASIC's RND function with its three distinct behaviours:
//   - RND or RND(positive): advance the sequence and return the next number in [0, 1).
//   - RND(0): return the last generated number without advancing the sequence.
//     This lets a program replay the same value, e.g. for animated retries.
//   - RND(negative): reseed the generator with the absolute value of the argument
//     (encoded as raw IEEE 754 bits for reproducibility), then return the first
//     number from the new sequence. This gives deterministic results from code.
//
// The caller passes arg=1 when the BASIC source writes plain RND (no argument),
// because the code generator normalises zero-argument calls to RND(1).
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
