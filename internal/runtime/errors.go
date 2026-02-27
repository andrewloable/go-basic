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

// BasicError represents a Turbo BASIC runtime error.
type BasicError struct {
	Code    int    // ERR - error code
	Message string // human-readable description
	Line    int    // ERL - line where error occurred
}

// Error implements the error interface.
func (e *BasicError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("Error %d: %s in line %d", e.Code, e.Message, e.Line)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// ErrorState holds global error handling state for ON ERROR / RESUME.
type ErrorState struct {
	HandlerActive bool         // ON ERROR GOTO is set
	HandlerLabel  string       // target label
	LastError     *BasicError  // most recent error
	LastERL       int          // line of last error (ERL)
	LastERR       int          // code of last error (ERR)
}

// NewErrorState creates a fresh error state.
func NewErrorState() *ErrorState {
	return &ErrorState{}
}

// TriggerError creates and stores a BasicError. If a handler is active, it
// returns the error for the calling code to jump to the handler. Otherwise,
// it panics with the error (unhandled BASIC runtime error).
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
