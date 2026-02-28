package vm

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// Value — the universal value type on the VM stack
//
// A stack-based VM must be able to hold any kind of value on its operand stack.
// The simplest approach is a "tagged union" (also called a discriminated union
// or variant type): a single struct that carries both a type tag (ValueType)
// and enough storage for each possible kind of data.
//
// Tagged union design trade-offs:
//   - Every Value allocates space for all three fields (Int, Float, Str) even
//     though only one field is active at a time.  This wastes a few bytes per
//     value but is simple and avoids interface boxing / heap allocation.
//   - The Type field must be checked before reading Int, Float, or Str.
//     Forgetting the check is a programming error; the VM's strongly-typed
//     arithmetic helpers enforce this discipline.
//   - An alternative design uses Go interfaces (type Value interface{}), which
//     avoids wasted fields but causes heap allocation on every push.  The struct
//     approach keeps most values on the stack or in a pre-allocated slice.
// ---------------------------------------------------------------------------

// ValueType classifies the type of a Value.
// The iota constants make the zero value (ValInt / 0) mean "integer", which
// is a reasonable default — arithmetic operations produce integers by default.
type ValueType byte

const (
	ValInt    ValueType = iota // Integer — active field is Int (int64)
	ValFloat                  // Float   — active field is Float (float64)
	ValString                 // String  — active field is Str (string)
)

// Value is the tagged union that lives on the operand stack and in variables.
// Every runtime value — literal numbers, string results, intermediate
// expression results — is represented as a Value.  The Type field determines
// which of Int, Float, or Str holds the active data.  Only the field
// corresponding to the active type should be read; the other fields contain
// stale or zero data.
type Value struct {
	Type  ValueType // discriminator tag — always check this before reading data fields
	Int   int64     // active when Type == ValInt
	Float float64   // active when Type == ValFloat
	Str   string    // active when Type == ValString
}

// String returns a human-readable representation of the value (for debugging).
func (v Value) String() string {
	switch v.Type {
	case ValInt:
		return fmt.Sprintf("Int(%d)", v.Int)
	case ValFloat:
		return fmt.Sprintf("Float(%g)", v.Float)
	case ValString:
		return fmt.Sprintf("String(%q)", v.Str)
	default:
		return "Value(?)"
	}
}

// IntVal creates an integer Value.
func IntVal(n int64) Value {
	return Value{Type: ValInt, Int: n}
}

// FloatVal creates a floating-point Value.
func FloatVal(f float64) Value {
	return Value{Type: ValFloat, Float: f}
}

// StringVal creates a string Value.
func StringVal(s string) Value {
	return Value{Type: ValString, Str: s}
}

// TrueValue is the BASIC true constant (-1).
var TrueValue = Value{Type: ValInt, Int: -1}

// FalseValue is the BASIC false constant (0).
var FalseValue = Value{Type: ValInt, Int: 0}

// asFloat returns the numeric value of v as a float64.
// Strings return 0.
func (v Value) asFloat() float64 {
	switch v.Type {
	case ValInt:
		return float64(v.Int)
	case ValFloat:
		return v.Float
	default:
		return 0
	}
}

// asInt returns the numeric value of v as an int64.
// Strings return 0.
func (v Value) asInt() int64 {
	switch v.Type {
	case ValInt:
		return v.Int
	case ValFloat:
		return int64(v.Float)
	default:
		return 0
	}
}

// isTruthy returns true if the value is considered "true" in BASIC
// (non-zero for numbers, non-empty for strings).
func (v Value) isTruthy() bool {
	switch v.Type {
	case ValInt:
		return v.Int != 0
	case ValFloat:
		return v.Float != 0
	case ValString:
		return len(v.Str) > 0
	default:
		return false
	}
}

// asString returns a string representation suitable for PRINT.
func (v Value) asString() string {
	switch v.Type {
	case ValInt:
		if v.Int >= 0 {
			return fmt.Sprintf(" %d ", v.Int)
		}
		return fmt.Sprintf("%d ", v.Int)
	case ValFloat:
		if v.Float >= 0 {
			return fmt.Sprintf(" %g ", v.Float)
		}
		return fmt.Sprintf("%g ", v.Float)
	case ValString:
		return v.Str
	default:
		return ""
	}
}

// ---------------------------------------------------------------------------
// Instruction and Chunk
// ---------------------------------------------------------------------------

// Instruction is a single bytecode instruction with an optional operand.
// Every instruction has exactly one 32-bit operand field.  Its meaning is
// opcode-dependent:
//   - For OpPush/OpLoad/OpStore: index into Chunk.Constants (the constant pool)
//   - For OpJmp/OpJmpTrue/OpJmpFalse/OpCall/OpGosub: target instruction index
//   - For OpBuiltin: BuiltinID constant identifying which function to call
//   - For OpDimArray: number of dimensions
//   - For OpLine: the BASIC source line number being marked
//   - For operand-free instructions (OpAdd, OpPop, …): the field is 0 and ignored
type Instruction struct {
	Op      Opcode
	Operand int32 // meaning depends on Op; see opcode comments above
}

// String returns a human-readable disassembly of the instruction.
func (inst Instruction) String() string {
	return fmt.Sprintf("%-12s %d", inst.Op.String(), inst.Operand)
}

// Chunk holds the compiled bytecode for a single BASIC program.
//
// A Chunk is the unit of compiled code passed to the VM.  It contains three
// parallel arrays that must stay in sync:
//
//   Code      — the instruction stream; each element is one VM operation.
//               The instruction pointer (vm.ip) is an index into this slice.
//
//   Constants — the constant pool; a flat list of all Values that were known
//               at compile time (number literals, string literals, variable
//               name strings).  Instructions reference constants by index so
//               the Values themselves are stored only once even if referenced
//               many times.
//
//   Lines     — one source-line number per instruction (same length as Code).
//               This lets runtimeError() report "error at line N" by looking
//               up Lines[ip-1].  The debug info costs very little space and
//               makes error messages enormously more useful.
type Chunk struct {
	Code      []Instruction // the instruction stream
	Constants []Value       // constant pool
	Lines     []int         // source line number for each instruction (for error messages)
}

// AddConstant appends a value to the constant pool and returns its index.
func (c *Chunk) AddConstant(v Value) int32 {
	idx := int32(len(c.Constants))
	c.Constants = append(c.Constants, v)
	return idx
}

// Emit appends an instruction and its corresponding line number.
func (c *Chunk) Emit(op Opcode, operand int32, line int) int {
	idx := len(c.Code)
	c.Code = append(c.Code, Instruction{Op: op, Operand: operand})
	c.Lines = append(c.Lines, line)
	return idx
}

// Disassemble returns a multi-line string showing all instructions.
func (c *Chunk) Disassemble() string {
	var b strings.Builder
	for i, inst := range c.Code {
		line := 0
		if i < len(c.Lines) {
			line = c.Lines[i]
		}
		fmt.Fprintf(&b, "%04d [L%d] %s\n", i, line, inst.String())
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// CallFrame
// ---------------------------------------------------------------------------

// CallFrame stores the return context for GOSUB/RETURN and SUB/FUNCTION calls.
type CallFrame struct {
	ReturnAddr int // instruction pointer to resume at after return
	LocalBase  int // base index in the locals slice for this frame
}

// ---------------------------------------------------------------------------
// Array storage
// ---------------------------------------------------------------------------

// arrayValue stores a BASIC array with its dimensions and flat data.
type arrayValue struct {
	dims []int   // size of each dimension
	data []Value // flat storage, row-major order
}
