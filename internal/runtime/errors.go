// errors.go — Structured error handling for the BASIC runtime.
//
// # Compiler Design Note: BASIC's Error Handling Model
//
// Turbo BASIC's error handling is fundamentally different from Go's:
//
//	ON ERROR GOTO label  — install a global error handler; all subsequent
//	                       runtime errors jump to 'label' instead of aborting.
//	ERR                  — integer variable holding the last error code.
//	ERL                  — integer variable holding the line number of the error.
//	RESUME               — after handling, jump back to the statement that failed.
//	RESUME NEXT          — after handling, continue at the statement after the error.
//	RESUME label         — after handling, jump to a specific label.
//	ERROR n              — manually raise error n (like Go's panic).
//
// This is a non-local control flow mechanism similar to setjmp/longjmp in C or
// try/catch in Java — but BASIC's version is global (only one handler at a time)
// and is not scoped to a call stack frame.
//
// # Implementation Strategy
//
// Because Go does not have a setjmp equivalent, the generated code implements
// BASIC error handling using a combination of:
//
//  1. ErrorState — a struct that holds the handler label, last error, ERR, and ERL.
//  2. TriggerError — called at every potential error site; if a handler is active
//     it returns the error (and generated code checks the return value to jump to
//     the handler label via a Go switch on a state variable); otherwise it panics.
//  3. defer + recover — in the generated program's main function, a deferred
//     function catches panics from unhandled BASIC errors and prints them.
//
// This is an example of the N-way control flow problem in transpilers: the source
// language has control flow primitives (GOTO, ON ERROR, RESUME) that do not map
// cleanly to the target language. The solution is to encode the control flow
// state explicitly in a runtime variable and translate each jump to a Go switch
// or for-loop that checks that variable.
//
// # Error Code Table
//
// BASIC errors are identified by numeric codes (ERR = 5 means "Illegal function
// call"). The table below maps codes to human-readable messages. These codes are
// standardised across Microsoft BASIC dialects.

package runtime

import "fmt"

// basicErrorMessages maps standard BASIC error codes to their messages.
var basicErrorMessages = map[int]string{
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
	24: "Device timeout",
	25: "Device fault",
	27: "Out of paper",
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

// BasicError represents a Turbo BASIC runtime error with its numeric code,
// human-readable message, and the source line number where it occurred.
//
// This struct satisfies Go's error interface so it can be used in standard
// Go error-handling patterns (errors.Is, fmt.Errorf wrapping, etc.) while
// also carrying the BASIC-specific fields that ERR and ERL expose to programs.
type BasicError struct {
	Code    int    // ERR - error code visible to BASIC programs via the ERR variable
	Message string // human-readable description of the error
	Line    int    // ERL - the BASIC source line number where the error occurred
}

// Error implements the error interface.
func (e *BasicError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("Error %d: %s in line %d", e.Code, e.Message, e.Line)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// ErrorState holds the mutable global state that BASIC's error handling model
// requires. There is exactly one ErrorState per generated program, created at
// startup and passed (via pointer) to every generated function that can raise
// an error.
//
// Fields mirror the BASIC programmer's view:
//   - HandlerActive / HandlerLabel — set by "ON ERROR GOTO label".
//   - LastError / LastERR / LastERL — readable via the ERR and ERL variables.
//   - ResumeLabel — set by the RESUME statement so the handler can jump back.
type ErrorState struct {
	HandlerActive bool        // true when ON ERROR GOTO has been executed
	HandlerLabel  string      // the label to jump to on error
	LastError     *BasicError // the most recently triggered error (nil if none)
	LastERL       int         // line number of the last error (BASIC's ERL)
	LastERR       int         // error code of the last error (BASIC's ERR)
	ResumeLabel   string      // label to jump back to for a bare RESUME statement
}

// NewErrorState creates a fresh error state.
func NewErrorState() *ErrorState {
	return &ErrorState{}
}

// TriggerError creates a BasicError, records it in the ErrorState, and either
// returns it (when a handler is active) or panics (when no handler is set).
//
// Generated code calls this at every potential error site:
//
//	if err := _es.TriggerError(5, 42); err != nil {
//	    _gotoLabel = "errorHandler"
//	    goto _dispatch
//	}
//
// The nil-vs-non-nil return value is the bridge between BASIC's ON ERROR model
// and Go's structured control flow. When no handler is active, panicking lets
// Go's own stack unwinding and a top-level recover() print a clean error message.
func (es *ErrorState) TriggerError(code int, line int) *BasicError {
	msg := ErrorMessage(code)
	err := &BasicError{
		Code:    code,
		Message: msg,
		Line:    line,
	}
	es.LastError = err
	es.LastERR = code
	es.LastERL = line

	if es.HandlerActive {
		return err
	}
	// No handler active: unhandled error, panic.
	panic(err)
}

// SetHandler sets the ON ERROR GOTO target. Pass "" to disable (ON ERROR GOTO 0).
func (es *ErrorState) SetHandler(label string) {
	if label == "" {
		es.HandlerActive = false
		es.HandlerLabel = ""
	} else {
		es.HandlerActive = true
		es.HandlerLabel = label
	}
}

// Err returns the last error code (BASIC's ERR).
func (es *ErrorState) Err() int {
	return es.LastERR
}

// Erl returns the line of the last error (BASIC's ERL).
func (es *ErrorState) Erl() int {
	return es.LastERL
}

// ClearError clears the error state (after RESUME).
func (es *ErrorState) ClearError() {
	es.LastError = nil
	es.LastERR = 0
	es.LastERL = 0
}

// ErrorMessage returns the BASIC error message for a code.
// Returns "Unknown error" for unrecognized codes.
func ErrorMessage(code int) string {
	if msg, ok := basicErrorMessages[code]; ok {
		return msg
	}
	return fmt.Sprintf("Unknown error %d", code)
}
