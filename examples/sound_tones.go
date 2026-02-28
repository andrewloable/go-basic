package main

import (
	"fmt"

	rt "github.com/loabletech/go-basic/internal/runtime"
)

// Suppress unused import warnings.
var _ = fmt.Sprintf
var _ = rt.Abs

func main() {
	// Hoisted variable declarations (avoids goto-over-declaration errors).
	var freq float32
	_ = freq

	//  SOUND Statement - Tone Generation
	//  Page: 357 (Chapter 5 - SOUND statement)
	end_0 := float32(2000)
	step_0 := float32(100)
	for freq = float32(100); (step_0 > 0 && freq <= end_0) || (step_0 < 0 && freq >= end_0) || (step_0 == 0); freq += step_0 {
		rt.Sound(float64(freq), float64(2))
	}
}
