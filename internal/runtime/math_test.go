package runtime

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Abs
// ---------------------------------------------------------------------------

func TestAbs(t *testing.T) {
	tests := []struct {
		x, want float64
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-3.14, 3.14},
		{3.14, 3.14},
	}
	for _, tt := range tests {
		if got := Abs(tt.x); got != tt.want {
			t.Errorf("Abs(%g) = %g, want %g", tt.x, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Sgn
// ---------------------------------------------------------------------------

func TestSgn(t *testing.T) {
	tests := []struct {
		x    float64
		want int
	}{
		{5, 1},
		{-5, -1},
		{0, 0},
		{0.001, 1},
		{-0.001, -1},
	}
	for _, tt := range tests {
		if got := Sgn(tt.x); got != tt.want {
			t.Errorf("Sgn(%g) = %d, want %d", tt.x, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// IntFloor
// ---------------------------------------------------------------------------

func TestIntFloor(t *testing.T) {
	tests := []struct {
		x, want float64
	}{
		{2.5, 2},
		{-2.5, -3},
		{3.0, 3},
		{-3.0, -3},
		{0, 0},
		{0.9, 0},
		{-0.1, -1},
	}
	for _, tt := range tests {
		if got := IntFloor(tt.x); got != tt.want {
			t.Errorf("IntFloor(%g) = %g, want %g", tt.x, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Fix
// ---------------------------------------------------------------------------

func TestFix(t *testing.T) {
	tests := []struct {
		x, want float64
	}{
		{2.5, 2},
		{-2.5, -2},
		{3.0, 3},
		{-3.0, -3},
		{0, 0},
		{0.9, 0},
		{-0.9, 0},
	}
	for _, tt := range tests {
		if got := Fix(tt.x); got != tt.want {
			t.Errorf("Fix(%g) = %g, want %g", tt.x, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Ceil
// ---------------------------------------------------------------------------

func TestCeil(t *testing.T) {
	tests := []struct {
		x, want float64
	}{
		{2.1, 3},
		{-2.1, -2},
		{3.0, 3},
		{-3.0, -3},
		{0, 0},
		{0.1, 1},
		{-0.9, 0},
	}
	for _, tt := range tests {
		if got := Ceil(tt.x); got != tt.want {
			t.Errorf("Ceil(%g) = %g, want %g", tt.x, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Sqr
// ---------------------------------------------------------------------------

func TestSqr(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		got, err := Sqr(4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 2 {
			t.Errorf("Sqr(4) = %g, want 2", got)
		}
	})

	t.Run("zero", func(t *testing.T) {
		got, err := Sqr(0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Errorf("Sqr(0) = %g, want 0", got)
		}
	})

	t.Run("fractional", func(t *testing.T) {
		got, err := Sqr(2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if math.Abs(got-math.Sqrt(2)) > 1e-15 {
			t.Errorf("Sqr(2) = %g, want %g", got, math.Sqrt(2))
		}
	})

	t.Run("negative error", func(t *testing.T) {
		_, err := Sqr(-1)
		if err == nil {
			t.Fatal("expected error for negative input, got nil")
		}
	})
}

// ---------------------------------------------------------------------------
// Trig functions
// ---------------------------------------------------------------------------

func TestSin(t *testing.T) {
	if got := Sin(0); got != 0 {
		t.Errorf("Sin(0) = %g, want 0", got)
	}
	if got := Sin(math.Pi / 2); math.Abs(got-1) > 1e-15 {
		t.Errorf("Sin(Pi/2) = %g, want 1", got)
	}
}

func TestCos(t *testing.T) {
	if got := Cos(0); got != 1 {
		t.Errorf("Cos(0) = %g, want 1", got)
	}
	if got := Cos(math.Pi); math.Abs(got+1) > 1e-15 {
		t.Errorf("Cos(Pi) = %g, want -1", got)
	}
}

func TestTan(t *testing.T) {
	if got := Tan(0); got != 0 {
		t.Errorf("Tan(0) = %g, want 0", got)
	}
	if got := Tan(math.Pi / 4); math.Abs(got-1) > 1e-15 {
		t.Errorf("Tan(Pi/4) = %g, want 1", got)
	}
}

func TestAtn(t *testing.T) {
	if got := Atn(0); got != 0 {
		t.Errorf("Atn(0) = %g, want 0", got)
	}
	if got := Atn(1); math.Abs(got-math.Pi/4) > 1e-15 {
		t.Errorf("Atn(1) = %g, want %g", got, math.Pi/4)
	}
}

// ---------------------------------------------------------------------------
// Exp, Exp2, Exp10
// ---------------------------------------------------------------------------

func TestExp(t *testing.T) {
	if got := Exp(0); got != 1 {
		t.Errorf("Exp(0) = %g, want 1", got)
	}
	if got := Exp(1); math.Abs(got-math.E) > 1e-15 {
		t.Errorf("Exp(1) = %g, want %g", got, math.E)
	}
}

func TestExp2(t *testing.T) {
	if got := Exp2(0); got != 1 {
		t.Errorf("Exp2(0) = %g, want 1", got)
	}
	if got := Exp2(10); got != 1024 {
		t.Errorf("Exp2(10) = %g, want 1024", got)
	}
}

func TestExp10(t *testing.T) {
	if got := Exp10(0); got != 1 {
		t.Errorf("Exp10(0) = %g, want 1", got)
	}
	if got := Exp10(3); math.Abs(got-1000) > 1e-10 {
		t.Errorf("Exp10(3) = %g, want 1000", got)
	}
}

// ---------------------------------------------------------------------------
// Log, Log2, Log10
// ---------------------------------------------------------------------------

func TestLog(t *testing.T) {
	t.Run("e", func(t *testing.T) {
		got, err := Log(math.E)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if math.Abs(got-1) > 1e-15 {
			t.Errorf("Log(e) = %g, want 1", got)
		}
	})

	t.Run("one", func(t *testing.T) {
		got, err := Log(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Errorf("Log(1) = %g, want 0", got)
		}
	})

	t.Run("zero error", func(t *testing.T) {
		_, err := Log(0)
		if err == nil {
			t.Fatal("expected error for 0, got nil")
		}
	})

	t.Run("negative error", func(t *testing.T) {
		_, err := Log(-1)
		if err == nil {
			t.Fatal("expected error for negative, got nil")
		}
	})
}

func TestLog2(t *testing.T) {
	got, err := Log2(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3 {
		t.Errorf("Log2(8) = %g, want 3", got)
	}

	_, err = Log2(0)
	if err == nil {
		t.Fatal("expected error for 0, got nil")
	}
}

func TestLog10(t *testing.T) {
	got, err := Log10(1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(got-3) > 1e-15 {
		t.Errorf("Log10(1000) = %g, want 3", got)
	}

	_, err = Log10(-5)
	if err == nil {
		t.Fatal("expected error for negative, got nil")
	}
}

// ---------------------------------------------------------------------------
// Cint (banker's rounding)
// ---------------------------------------------------------------------------

func TestCint(t *testing.T) {
	tests := []struct {
		name    string
		x       float64
		want    int16
		wantErr bool
	}{
		{"round down", 2.3, 2, false},
		{"round up", 2.7, 3, false},
		{"banker round even 0.5", 0.5, 0, false},
		{"banker round even 1.5", 1.5, 2, false},
		{"banker round even 2.5", 2.5, 2, false},
		{"banker round even 3.5", 3.5, 4, false},
		{"negative round", -2.5, -2, false},
		{"negative round -3.5", -3.5, -4, false},
		{"zero", 0, 0, false},
		{"max int16", 32767, 32767, false},
		{"min int16", -32768, -32768, false},
		{"overflow positive", 32768, 0, true},
		{"overflow negative", -32769, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Cint(tt.x)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Cint(%g) = %d, want %d", tt.x, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Clng
// ---------------------------------------------------------------------------

func TestClng(t *testing.T) {
	tests := []struct {
		name    string
		x       float64
		want    int32
		wantErr bool
	}{
		{"round down", 2.3, 2, false},
		{"round up", 2.7, 3, false},
		{"banker round 0.5", 0.5, 0, false},
		{"banker round 1.5", 1.5, 2, false},
		{"zero", 0, 0, false},
		{"large", 100000.7, 100001, false},
		{"max int32", 2147483647, 2147483647, false},
		{"overflow", 2147483648, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Clng(tt.x)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Clng(%g) = %d, want %d", tt.x, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Csng, Cdbl
// ---------------------------------------------------------------------------

func TestCsng(t *testing.T) {
	got := Csng(3.141592653589793)
	expected := float32(3.141592653589793)
	if got != expected {
		t.Errorf("Csng(pi) = %g, want %g", got, expected)
	}
}

func TestCdbl(t *testing.T) {
	val := 3.141592653589793
	if got := Cdbl(val); got != val {
		t.Errorf("Cdbl(%g) = %g, want %g", val, got, val)
	}
}

// ---------------------------------------------------------------------------
// RNG (Rnd / Randomize)
// ---------------------------------------------------------------------------

func TestRndPositive(t *testing.T) {
	r := NewRNG()
	for i := 0; i < 100; i++ {
		v := r.Rnd(1)
		if v < 0 || v >= 1 {
			t.Fatalf("Rnd(1) returned %g, want [0,1)", v)
		}
	}
}

func TestRndZeroReturnsLast(t *testing.T) {
	r := NewRNG()
	v1 := r.Rnd(1)
	v2 := r.Rnd(0)
	if v1 != v2 {
		t.Errorf("Rnd(0) = %g, want same as last Rnd(1) = %g", v2, v1)
	}
	// Call Rnd(0) again -- should still be the same.
	v3 := r.Rnd(0)
	if v2 != v3 {
		t.Errorf("Rnd(0) changed: %g != %g", v2, v3)
	}
}

func TestRndNegativeReseeds(t *testing.T) {
	r1 := NewRNG()
	r2 := NewRNG()

	// Seed both with the same negative value.
	v1 := r1.Rnd(-42)
	v2 := r2.Rnd(-42)
	if v1 != v2 {
		t.Errorf("Rnd(-42) produced different values: %g vs %g", v1, v2)
	}

	// Subsequent values should also match.
	v1 = r1.Rnd(1)
	v2 = r2.Rnd(1)
	if v1 != v2 {
		t.Errorf("after reseed, Rnd(1) produced different values: %g vs %g", v1, v2)
	}
}

func TestRndDifferentNegativeSeeds(t *testing.T) {
	r := NewRNG()
	v1 := r.Rnd(-1)
	v2 := r.Rnd(-2)
	if v1 == v2 {
		t.Errorf("different negative seeds should produce different values")
	}
}

func TestRandomize(t *testing.T) {
	r1 := NewRNG()
	r2 := NewRNG()

	r1.Randomize(12345)
	r2.Randomize(12345)

	for i := 0; i < 10; i++ {
		v1 := r1.Rnd(1)
		v2 := r2.Rnd(1)
		if v1 != v2 {
			t.Errorf("after Randomize(12345), iteration %d: %g != %g", i, v1, v2)
		}
	}
}

func TestRandomizeZeroUsesTime(t *testing.T) {
	// We can only verify it does not panic and produces valid values.
	r := NewRNG()
	r.Randomize(0)
	v := r.Rnd(1)
	if v < 0 || v >= 1 {
		t.Errorf("after Randomize(0), Rnd(1) = %g out of range", v)
	}
}
