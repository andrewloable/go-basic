package runtime

import (
	"os"
	"regexp"
	"testing"
)

// ---------------------------------------------------------------------------
// Timer
// ---------------------------------------------------------------------------

func TestTimer(t *testing.T) {
	v := Timer()
	if v < 0 {
		t.Errorf("Timer() = %g, want >= 0", v)
	}
	// Seconds since midnight should be < 86400.
	if v >= 86400 {
		t.Errorf("Timer() = %g, want < 86400", v)
	}
}

// ---------------------------------------------------------------------------
// DateStr
// ---------------------------------------------------------------------------

func TestDateStr(t *testing.T) {
	got := DateStr()
	// Should match MM-DD-YYYY format.
	matched, err := regexp.MatchString(`^\d{2}-\d{2}-\d{4}$`, got)
	if err != nil {
		t.Fatalf("regex error: %v", err)
	}
	if !matched {
		t.Errorf("DateStr() = %q, want MM-DD-YYYY format", got)
	}
}

// ---------------------------------------------------------------------------
// TimeStr
// ---------------------------------------------------------------------------

func TestTimeStr(t *testing.T) {
	got := TimeStr()
	// Should match HH:MM:SS format.
	matched, err := regexp.MatchString(`^\d{2}:\d{2}:\d{2}$`, got)
	if err != nil {
		t.Fatalf("regex error: %v", err)
	}
	if !matched {
		t.Errorf("TimeStr() = %q, want HH:MM:SS format", got)
	}
}

// ---------------------------------------------------------------------------
// SwapInt
// ---------------------------------------------------------------------------

func TestSwapInt(t *testing.T) {
	a, b := 10, 20
	SwapInt(&a, &b)
	if a != 20 || b != 10 {
		t.Errorf("after SwapInt: a=%d, b=%d, want a=20, b=10", a, b)
	}
}

func TestSwapIntSameValue(t *testing.T) {
	a, b := 5, 5
	SwapInt(&a, &b)
	if a != 5 || b != 5 {
		t.Errorf("after SwapInt same: a=%d, b=%d, want a=5, b=5", a, b)
	}
}

func TestSwapIntNegative(t *testing.T) {
	a, b := -3, 7
	SwapInt(&a, &b)
	if a != 7 || b != -3 {
		t.Errorf("after SwapInt: a=%d, b=%d, want a=7, b=-3", a, b)
	}
}

// ---------------------------------------------------------------------------
// SwapFloat
// ---------------------------------------------------------------------------

func TestSwapFloat(t *testing.T) {
	a, b := 1.5, 2.5
	SwapFloat(&a, &b)
	if a != 2.5 || b != 1.5 {
		t.Errorf("after SwapFloat: a=%g, b=%g, want a=2.5, b=1.5", a, b)
	}
}

func TestSwapFloatZero(t *testing.T) {
	a, b := 0.0, 3.14
	SwapFloat(&a, &b)
	if a != 3.14 || b != 0 {
		t.Errorf("after SwapFloat: a=%g, b=%g, want a=3.14, b=0", a, b)
	}
}

// ---------------------------------------------------------------------------
// SwapString
// ---------------------------------------------------------------------------

func TestSwapString(t *testing.T) {
	a, b := "hello", "world"
	SwapString(&a, &b)
	if a != "world" || b != "hello" {
		t.Errorf("after SwapString: a=%q, b=%q, want a=\"world\", b=\"hello\"", a, b)
	}
}

func TestSwapStringEmpty(t *testing.T) {
	a, b := "", "something"
	SwapString(&a, &b)
	if a != "something" || b != "" {
		t.Errorf("after SwapString: a=%q, b=%q", a, b)
	}
}

// ---------------------------------------------------------------------------
// CommandStr
// ---------------------------------------------------------------------------

func TestCommandStr(t *testing.T) {
	// CommandStr returns os.Args[1:] joined by space.
	// In test mode os.Args[0] is the test binary. We cannot easily control
	// os.Args in tests, but we can at least verify it does not panic and
	// returns a string.
	_ = CommandStr()
}

// ---------------------------------------------------------------------------
// EnvironGet / EnvironSet
// ---------------------------------------------------------------------------

func TestEnvironGetSet(t *testing.T) {
	key := "GO_BASIC_TEST_ENV_VAR_12345"
	// Make sure it does not exist initially.
	os.Unsetenv(key)
	got := EnvironGet(key)
	if got != "" {
		t.Errorf("EnvironGet before set = %q, want empty", got)
	}

	err := EnvironSet(key, "test_value")
	if err != nil {
		t.Fatalf("EnvironSet: %v", err)
	}

	got = EnvironGet(key)
	if got != "test_value" {
		t.Errorf("EnvironGet after set = %q, want %q", got, "test_value")
	}

	// Clean up.
	os.Unsetenv(key)
}

func TestEnvironGetNonExistent(t *testing.T) {
	got := EnvironGet("THIS_VARIABLE_SHOULD_NOT_EXIST_XYZ")
	if got != "" {
		t.Errorf("EnvironGet for nonexistent = %q, want empty", got)
	}
}

func TestEnvironSetOverwrite(t *testing.T) {
	key := "GO_BASIC_TEST_OVERWRITE"
	os.Unsetenv(key)

	EnvironSet(key, "first")
	EnvironSet(key, "second")

	got := EnvironGet(key)
	if got != "second" {
		t.Errorf("EnvironGet after overwrite = %q, want %q", got, "second")
	}

	os.Unsetenv(key)
}

// ---------------------------------------------------------------------------
// Delay (just test it does not panic for zero/negative)
// ---------------------------------------------------------------------------

func TestDelayZero(t *testing.T) {
	Delay(0)
}

func TestDelayNegative(t *testing.T) {
	Delay(-1)
}
