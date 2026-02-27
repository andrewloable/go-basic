package runtime

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// BasicError
// ---------------------------------------------------------------------------

func TestBasicErrorWithLine(t *testing.T) {
	e := &BasicError{Code: 11, Message: "Division by zero", Line: 100}
	got := e.Error()
	want := "Error 11: Division by zero in line 100"
	if got != want {
		t.Errorf("BasicError.Error() = %q, want %q", got, want)
	}
}

func TestBasicErrorWithoutLine(t *testing.T) {
	e := &BasicError{Code: 5, Message: "Illegal function call", Line: 0}
	got := e.Error()
	want := "Error 5: Illegal function call"
	if got != want {
		t.Errorf("BasicError.Error() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// NewErrorState
// ---------------------------------------------------------------------------

func TestNewErrorState(t *testing.T) {
	es := NewErrorState()
	if es == nil {
		t.Fatal("NewErrorState() returned nil")
	}
	if es.HandlerActive {
		t.Error("new error state should not have active handler")
	}
	if es.LastError != nil {
		t.Error("new error state should have nil LastError")
	}
	if es.Err() != 0 {
		t.Errorf("new error state Err() = %d, want 0", es.Err())
	}
	if es.Erl() != 0 {
		t.Errorf("new error state Erl() = %d, want 0", es.Erl())
	}
}

// ---------------------------------------------------------------------------
// SetHandler
// ---------------------------------------------------------------------------

func TestSetHandler(t *testing.T) {
	es := NewErrorState()

	es.SetHandler("error_handler")
	if !es.HandlerActive {
		t.Error("expected handler to be active after SetHandler")
	}
	if es.HandlerLabel != "error_handler" {
		t.Errorf("HandlerLabel = %q, want %q", es.HandlerLabel, "error_handler")
	}

	// Disable handler.
	es.SetHandler("")
	if es.HandlerActive {
		t.Error("expected handler to be inactive after SetHandler(\"\")")
	}
	if es.HandlerLabel != "" {
		t.Errorf("HandlerLabel = %q, want empty", es.HandlerLabel)
	}
}

// ---------------------------------------------------------------------------
// TriggerError with handler active
// ---------------------------------------------------------------------------

func TestTriggerErrorWithHandler(t *testing.T) {
	es := NewErrorState()
	es.SetHandler("on_error")

	err := es.TriggerError(11, 50)
	if err == nil {
		t.Fatal("expected BasicError, got nil")
	}
	if err.Code != 11 {
		t.Errorf("error code = %d, want 11", err.Code)
	}
	if err.Line != 50 {
		t.Errorf("error line = %d, want 50", err.Line)
	}
	if es.Err() != 11 {
		t.Errorf("Err() = %d, want 11", es.Err())
	}
	if es.Erl() != 50 {
		t.Errorf("Erl() = %d, want 50", es.Erl())
	}
	if es.LastError != err {
		t.Error("LastError not set correctly")
	}
}

// ---------------------------------------------------------------------------
// TriggerError without handler (panics)
// ---------------------------------------------------------------------------

func TestTriggerErrorWithoutHandler(t *testing.T) {
	es := NewErrorState()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		be, ok := r.(*BasicError)
		if !ok {
			t.Fatalf("expected *BasicError in panic, got %T", r)
		}
		if be.Code != 6 {
			t.Errorf("panic error code = %d, want 6", be.Code)
		}
	}()

	es.TriggerError(6, 200)
}

// ---------------------------------------------------------------------------
// ClearError
// ---------------------------------------------------------------------------

func TestClearError(t *testing.T) {
	es := NewErrorState()
	es.SetHandler("handler")
	es.TriggerError(5, 10)

	if es.Err() == 0 {
		t.Fatal("expected non-zero Err() before clear")
	}

	es.ClearError()
	if es.Err() != 0 {
		t.Errorf("after ClearError, Err() = %d, want 0", es.Err())
	}
	if es.Erl() != 0 {
		t.Errorf("after ClearError, Erl() = %d, want 0", es.Erl())
	}
	if es.LastError != nil {
		t.Error("after ClearError, LastError should be nil")
	}
}

// ---------------------------------------------------------------------------
// ErrorMessage for known and unknown codes
// ---------------------------------------------------------------------------

func TestErrorMessage(t *testing.T) {
	knownCodes := map[int]string{
		1:  "NEXT without FOR",
		2:  "Syntax error",
		3:  "RETURN without GOSUB",
		4:  "Out of DATA",
		5:  "Illegal function call",
		6:  "Overflow",
		7:  "Out of memory",
		9:  "Subscript out of range",
		10: "Array already DIMensioned",
		11: "Division by zero",
		13: "Type mismatch",
		14: "Out of string space",
		20: "RESUME without error",
		52: "Bad file number",
		53: "File not found",
		54: "Bad file mode",
		55: "File already open",
		58: "File already exists",
		61: "Disk full",
		62: "Input past end of file",
		63: "Bad record number",
		64: "Bad filename",
		67: "Too many files",
		68: "Device unavailable",
		71: "Disk not ready",
		72: "Disk media error",
		76: "Path not found",
	}
	for code, wantMsg := range knownCodes {
		got := ErrorMessage(code)
		if got != wantMsg {
			t.Errorf("ErrorMessage(%d) = %q, want %q", code, got, wantMsg)
		}
	}
}

func TestErrorMessageUnknown(t *testing.T) {
	got := ErrorMessage(999)
	if !strings.Contains(got, "Unknown error") {
		t.Errorf("ErrorMessage(999) = %q, want to contain \"Unknown error\"", got)
	}
	if !strings.Contains(got, "999") {
		t.Errorf("ErrorMessage(999) = %q, want to contain \"999\"", got)
	}
}

// ---------------------------------------------------------------------------
// Err / Erl
// ---------------------------------------------------------------------------

func TestErrErl(t *testing.T) {
	es := NewErrorState()
	es.SetHandler("h")

	es.TriggerError(53, 300)
	if es.Err() != 53 {
		t.Errorf("Err() = %d, want 53", es.Err())
	}
	if es.Erl() != 300 {
		t.Errorf("Erl() = %d, want 300", es.Erl())
	}

	// Trigger another error; values update.
	es.TriggerError(11, 400)
	if es.Err() != 11 {
		t.Errorf("Err() = %d, want 11", es.Err())
	}
	if es.Erl() != 400 {
		t.Errorf("Erl() = %d, want 400", es.Erl())
	}
}
