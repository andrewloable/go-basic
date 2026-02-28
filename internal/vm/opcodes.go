package vm

import "fmt"

// ---------------------------------------------------------------------------
// Opcode — bytecode instruction set for the Turbo BASIC VM
//
// What is an opcode?
// An opcode (operation code) is a small integer that identifies a single
// operation the virtual machine knows how to perform — things like "push a
// value", "add two numbers", or "jump to another instruction".  A running
// program is represented as a flat array of these instructions (the bytecode),
// and the VM executes them one by one.
//
// Why bytecode instead of just running the AST?
// Walking the AST at runtime is simple to implement but slow: every expression
// evaluation requires traversing tree nodes, type-switching on interfaces, and
// allocating memory.  Bytecode is a compact linear array; the VM loops over it
// with a single switch statement and minimal allocations.  This makes bytecode
// 5-20x faster to interpret than a tree-walk interpreter.
//
// How this VM's instruction set compares to a real CPU
// A real CPU (x86, ARM) has hundreds of instructions and uses hardware
// registers to pass values between them.  This VM is stack-based: operands
// are pushed onto an explicit in-memory stack, and results land back on top.
// There are no named registers to manage, which makes the compiler much
// simpler — it never needs to decide which register to use.
//
// Why have a bytecode VM alongside the Go transpiler?
// The Go transpiler (internal/codegen) turns BASIC source directly into Go
// source files that are then compiled by the Go toolchain.  This produces the
// fastest possible output but requires the Go compiler to be present.  The
// bytecode VM is self-contained: it can execute BASIC programs immediately
// after parsing, with no external tools.  This makes it ideal for quick
// scripting, testing, and interactive use.
//
// Instruction format (see Instruction in vm_types.go):
//   Op      Opcode  — which operation to perform (1 byte)
//   Operand int32   — a parameter whose meaning depends on the opcode
//   (Line numbers are stored separately in Chunk.Lines for error reporting.)
// ---------------------------------------------------------------------------

// Opcode represents a single bytecode instruction.
type Opcode byte

const (
	// ---------------------------------------------------------------------------
	// Stack operations
	//
	// The operand stack is the VM's scratchpad.  Every computation pushes its
	// result onto the stack; every instruction reads its inputs by popping them
	// off.  Think of it as a pile of sticky notes: you always work with the
	// topmost note (TOS = Top Of Stack) and NOS = Next On Stack.
	// ---------------------------------------------------------------------------

	// OpPush pushes a constant value onto the stack.
	// Operand: index into Chunk.Constants (the constant pool).
	// Stack effect: [] → [value]
	OpPush Opcode = iota

	// OpPop discards the top value on the stack.
	// Used when an expression is evaluated for side effects only (e.g., a
	// function call whose return value is not used).
	// Stack effect: [v] → []
	OpPop

	// OpDup duplicates the top of stack without consuming it.
	// Useful when the same value is needed twice in sequence.
	// Stack effect: [v] → [v, v]
	OpDup

	// ---------------------------------------------------------------------------
	// Variable operations
	//
	// Variables are stored in a flat map (vm.globals) keyed by name.  At compile
	// time each variable is given a "slot index" — the index of its name string
	// in the constant pool.  OpLoad/OpStore use this index to look up the name
	// and then read/write the variable.
	// ---------------------------------------------------------------------------

	// OpLoad reads a variable and pushes its current value onto the stack.
	// Operand: constant pool index of the variable's name string.
	// Stack effect: [] → [value]
	OpLoad

	// OpStore pops the top of stack and writes it into a variable.
	// Operand: constant pool index of the variable's name string.
	// Stack effect: [value] → []
	OpStore

	// OpLoadArray reads one element from a BASIC array.
	// The index expressions for each dimension must already be on the stack
	// (pushed left-to-right); this instruction pops them all and pushes the
	// element value.
	// Operand: constant pool index of the array variable's name.
	// Stack effect: [i1, i2, …, iN] → [element]
	OpLoadArray

	// OpStoreArray writes a value into one element of a BASIC array.
	// The stack must contain the value to store followed by all dimension
	// indices (value pushed last, so it is at TOS).
	// Operand: constant pool index of the array variable's name.
	// Stack effect: [i1, i2, …, iN, value] → []
	OpStoreArray

	// ---------------------------------------------------------------------------
	// Arithmetic
	//
	// All binary arithmetic instructions pop two values (NOS op TOS) and push
	// the result.  Unary instructions (OpNeg) pop one value and push the result.
	// The compiler emits the left operand first, the right operand second, so
	// when the operator executes, NOS holds the left and TOS holds the right —
	// which matters for non-commutative operators like subtraction.
	// ---------------------------------------------------------------------------

	// OpAdd adds the top two stack values.  For strings, performs concatenation.
	// Stack effect: [a, b] → [a+b]
	OpAdd

	// OpSub subtracts TOS from NOS (result = NOS - TOS).
	// Stack effect: [a, b] → [a-b]
	OpSub

	// OpMul multiplies the top two stack values.
	// Stack effect: [a, b] → [a*b]
	OpMul

	// OpDiv performs floating-point division (NOS / TOS).
	// Returns a runtime error on division by zero.
	// Stack effect: [a, b] → [a/b]
	OpDiv

	// OpIDiv performs integer division using BASIC's backslash operator (NOS \ TOS).
	// Both operands are truncated to integers before dividing.
	// Stack effect: [a, b] → [int(a) \ int(b)]
	OpIDiv

	// OpMod computes the integer modulo (NOS MOD TOS).
	// Stack effect: [a, b] → [a MOD b]
	OpMod

	// OpPow raises NOS to the power TOS (NOS ^ TOS).
	// Stack effect: [a, b] → [a^b]
	OpPow

	// OpNeg negates the top of stack (unary minus).
	// Stack effect: [a] → [-a]
	OpNeg

	// ---------------------------------------------------------------------------
	// Comparison operators
	//
	// BASIC uses -1 for TRUE and 0 for FALSE (unlike C's 1/0).  All comparison
	// opcodes pop two values and push either IntVal(-1) or IntVal(0).  This
	// convention lets BASIC programs use boolean results in arithmetic expressions.
	// ---------------------------------------------------------------------------

	// OpEq pushes -1 if NOS == TOS, else 0.  Stack effect: [a, b] → [-1 or 0]
	OpEq

	// OpNe pushes -1 if NOS != TOS, else 0.  Stack effect: [a, b] → [-1 or 0]
	OpNe

	// OpLt pushes -1 if NOS < TOS, else 0.  Stack effect: [a, b] → [-1 or 0]
	OpLt

	// OpGt pushes -1 if NOS > TOS, else 0.  Stack effect: [a, b] → [-1 or 0]
	OpGt

	// OpLe pushes -1 if NOS <= TOS, else 0.  Stack effect: [a, b] → [-1 or 0]
	OpLe

	// OpGe pushes -1 if NOS >= TOS, else 0.  Stack effect: [a, b] → [-1 or 0]
	OpGe

	// ---------------------------------------------------------------------------
	// Logical operators
	//
	// BASIC's logical operators work bitwise on integer operands, mirroring the
	// behaviour of Turbo BASIC / QBasic.  On float operands the values are first
	// truncated to int64 before the bitwise operation.
	// ---------------------------------------------------------------------------

	// OpAnd computes bitwise AND of the top two stack values.
	// Stack effect: [a, b] → [a AND b]
	OpAnd

	// OpOr computes bitwise OR of the top two stack values.
	// Stack effect: [a, b] → [a OR b]
	OpOr

	// OpXor computes bitwise XOR of the top two stack values.
	// Stack effect: [a, b] → [a XOR b]
	OpXor

	// OpNot computes bitwise NOT of the top stack value (unary).
	// Stack effect: [a] → [NOT a]
	OpNot

	// OpEqv computes logical equivalence: NOT (a XOR b).
	// Result is -1 when both operands have the same truth value.
	// Stack effect: [a, b] → [a EQV b]
	OpEqv

	// OpImp computes logical implication: (NOT a) OR b.
	// Result is 0 only when a is true and b is false.
	// Stack effect: [a, b] → [a IMP b]
	OpImp

	// ---------------------------------------------------------------------------
	// String operations
	// ---------------------------------------------------------------------------

	// OpConcat concatenates the top two string values (NOS & TOS).
	// Numeric values are converted to strings before concatenation.
	// Stack effect: [a$, b$] → [a$+b$]
	OpConcat

	// ---------------------------------------------------------------------------
	// Control flow
	//
	// Control flow in bytecode is implemented with jump instructions.  Unlike
	// source-level IF or FOR statements, the VM only knows "go to instruction N".
	// The compiler is responsible for converting structured control flow into the
	// correct combination of conditional and unconditional jumps.
	//
	// All jump operands are absolute instruction indices into Chunk.Code.
	// ---------------------------------------------------------------------------

	// OpJmp performs an unconditional jump to the instruction at operand index.
	// Operand: target instruction index.
	// Stack effect: none (does not touch the stack).
	OpJmp

	// OpJmpTrue pops the top of stack and jumps if it is truthy (non-zero).
	// Operand: target instruction index.
	// Stack effect: [cond] → []
	OpJmpTrue

	// OpJmpFalse pops the top of stack and jumps if it is falsy (zero / empty).
	// This is the workhorse of IF conditions and loop exit tests.
	// Operand: target instruction index.
	// Stack effect: [cond] → []
	OpJmpFalse

	// OpCall pushes the current ip as a return address and jumps to a
	// SUB/FUNCTION body.  The callee uses OpRet to restore the ip.
	// Operand: target instruction index.
	// Stack effect: args remain on stack; call frame is pushed internally.
	OpCall

	// OpRet returns from a SUB/FUNCTION call by restoring the saved ip.
	// Stack effect: none (return value, if any, remains on stack).
	OpRet

	// OpGosub implements BASIC's GOSUB: pushes the return address onto the
	// call stack, then jumps to the subroutine.
	// Operand: target instruction index.
	// Stack effect: none.
	OpGosub

	// OpReturn implements BASIC's RETURN: pops the call stack and jumps back
	// to the instruction after the matching GOSUB.
	// Stack effect: none.
	OpReturn

	// ---------------------------------------------------------------------------
	// I/O
	// ---------------------------------------------------------------------------

	// OpPrint pops the top of stack and writes its string representation to output.
	// Stack effect: [value] → []
	OpPrint

	// OpPrintNewline writes a newline character to output.
	// Stack effect: none.
	OpPrintNewline

	// OpPrintTab advances the output cursor to the next 14-character print zone,
	// implementing BASIC's comma separator in PRINT statements.
	// Stack effect: none.
	OpPrintTab

	// OpPrintSemicolon is a no-op spacer emitted for PRINT's semicolon separator.
	// It signals "no newline yet" without moving the cursor.
	// Stack effect: none.
	OpPrintSemicolon

	// OpInput reads a line from stdin, displays an optional prompt, and stores
	// the result in a variable.
	// Operand: constant pool index of the prompt string (may be empty).
	// Stack effect: none (result stored directly into variable by the handler).
	OpInput

	// ---------------------------------------------------------------------------
	// Type conversion
	//
	// BASIC allows implicit narrowing and widening conversions between its numeric
	// types.  These opcodes perform explicit conversions when the type suffix of
	// a variable demands a particular representation.
	// ---------------------------------------------------------------------------

	// OpToInt converts the top of stack to a 16-bit integer (int16 range).
	// Stack effect: [v] → [int16(v)]
	OpToInt

	// OpToLong converts the top of stack to a 32-bit integer (int32 range).
	// Stack effect: [v] → [int32(v)]
	OpToLong

	// OpToSingle converts the top of stack to a 32-bit float.
	// Stack effect: [v] → [float32(v)]
	OpToSingle

	// OpToDouble converts the top of stack to a 64-bit float.
	// Stack effect: [v] → [float64(v)]
	OpToDouble

	// OpToString converts the top of stack to its string representation.
	// Stack effect: [v] → [str(v)]
	OpToString

	// ---------------------------------------------------------------------------
	// Built-in functions
	//
	// Rather than giving every built-in function (ABS, SIN, LEFT$, …) its own
	// opcode, they all share a single OpBuiltin instruction.  The operand
	// encodes which function to call as a BuiltinID integer.  This keeps the
	// opcode set small while still dispatching in O(1) via a switch statement.
	// ---------------------------------------------------------------------------

	// OpBuiltin calls a built-in runtime function identified by the operand.
	// Operand: BuiltinID constant (see below).
	// Stack effect: varies by function — the function pops its arguments and
	// pushes one result.
	OpBuiltin

	// ---------------------------------------------------------------------------
	// Array
	// ---------------------------------------------------------------------------

	// OpDimArray allocates a BASIC array with the given dimensions.
	// The dimension sizes must already be on the stack (pushed left-to-right).
	// Operand: number of dimensions.
	// Stack effect: [size1, size2, …, sizeN, nameIdx] → []
	OpDimArray

	// ---------------------------------------------------------------------------
	// Data
	//
	// BASIC's DATA/READ/RESTORE mechanism is implemented as a separate pool of
	// values collected during compilation (pass 1).  At runtime, READ simply
	// advances a pointer into that pool.
	// ---------------------------------------------------------------------------

	// OpRead copies the next value from the DATA pool into a variable.
	// Operand: constant pool index of the target variable's name.
	// Stack effect: none (writes directly to globals map).
	OpRead

	// OpRestore resets the DATA pool read pointer.
	// Operand: target data index, or -1 to reset to the beginning.
	// Stack effect: none.
	OpRestore

	// ---------------------------------------------------------------------------
	// Miscellaneous
	// ---------------------------------------------------------------------------

	// OpHalt stops the VM's execution loop and returns control to the caller.
	// Every program ends with an implicit OpHalt.
	// Stack effect: none.
	OpHalt

	// OpNop (no operation) does nothing and advances ip by 1.
	// Used as a placeholder when an AST node has no bytecode equivalent yet.
	// Stack effect: none.
	OpNop

	// OpLine records the current BASIC source line number for error messages.
	// The VM does not execute any logic for this instruction — it is purely a
	// debugging aid that lets runtimeError() report accurate line numbers.
	// Operand: BASIC source line number.
	// Stack effect: none.
	OpLine

	// OpPoke implements BASIC's POKE address, value statement.
	// Real memory-mapped I/O is unavailable in Go, so this pops both operands
	// for side-effect evaluation and then discards them.
	// Stack effect: [address, value] → []
	OpPoke

	// OpClear resets all global variables to their zero values (0 for numbers,
	// "" for strings), implementing BASIC's CLEAR statement.
	// Stack effect: none.
	OpClear
)

// opNames maps each opcode to its human-readable name.
var opNames = map[Opcode]string{
	OpPush:           "PUSH",
	OpPop:            "POP",
	OpDup:            "DUP",
	OpLoad:           "LOAD",
	OpStore:          "STORE",
	OpLoadArray:      "LOAD_ARRAY",
	OpStoreArray:     "STORE_ARRAY",
	OpAdd:            "ADD",
	OpSub:            "SUB",
	OpMul:            "MUL",
	OpDiv:            "DIV",
	OpIDiv:           "IDIV",
	OpMod:            "MOD",
	OpPow:            "POW",
	OpNeg:            "NEG",
	OpEq:             "EQ",
	OpNe:             "NE",
	OpLt:             "LT",
	OpGt:             "GT",
	OpLe:             "LE",
	OpGe:             "GE",
	OpAnd:            "AND",
	OpOr:             "OR",
	OpXor:            "XOR",
	OpNot:            "NOT",
	OpEqv:            "EQV",
	OpImp:            "IMP",
	OpConcat:         "CONCAT",
	OpJmp:            "JMP",
	OpJmpTrue:        "JMP_TRUE",
	OpJmpFalse:       "JMP_FALSE",
	OpCall:           "CALL",
	OpRet:            "RET",
	OpGosub:          "GOSUB",
	OpReturn:         "RETURN",
	OpPrint:          "PRINT",
	OpPrintNewline:   "PRINT_NL",
	OpPrintTab:       "PRINT_TAB",
	OpPrintSemicolon: "PRINT_SEMI",
	OpInput:          "INPUT",
	OpToInt:          "TO_INT",
	OpToLong:         "TO_LONG",
	OpToSingle:       "TO_SINGLE",
	OpToDouble:       "TO_DOUBLE",
	OpToString:       "TO_STRING",
	OpBuiltin:        "BUILTIN",
	OpDimArray:       "DIM_ARRAY",
	OpRead:           "READ",
	OpRestore:        "RESTORE",
	OpHalt:           "HALT",
	OpNop:            "NOP",
	OpLine:           "LINE",
	OpPoke:           "POKE",
	OpClear:          "CLEAR",
}

// String returns the human-readable name of the opcode.
func (op Opcode) String() string {
	if name, ok := opNames[op]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", int(op))
}

// ---------------------------------------------------------------------------
// BuiltinID — identifies built-in runtime functions
//
// Why number the built-ins rather than calling them by name at runtime?
// When the VM executes OpBuiltin it is in a tight inner loop.  A hash-map
// lookup on a string name on every function call would be noticeably slower
// than a switch on a small integer.  By encoding the function identity as a
// compile-time constant (BuiltinID) and using a Go switch in the VM, the
// dispatch is reduced to a single branch with no allocations.
// ---------------------------------------------------------------------------

// BuiltinID identifies a built-in function for the OpBuiltin instruction.
type BuiltinID byte

const (
	// ---------------------------------------------------------------------------
	// Math functions — numeric input, numeric output
	// ---------------------------------------------------------------------------

	BuiltinAbs   BuiltinID = iota // ABS(x) — absolute value; removes negative sign
	BuiltinSgn                    // SGN(x) — sign of x: returns -1, 0, or 1
	BuiltinInt                    // INT(x) — floor: largest integer <= x
	BuiltinFix                    // FIX(x) — truncate: removes fractional part (toward zero)
	BuiltinCeil                   // CEIL(x) — ceiling: smallest integer >= x (Turbo BASIC extension)
	BuiltinSqr                    // SQR(x) — square root of x
	BuiltinExp                    // EXP(x) — e raised to the power x
	BuiltinExp2                   // EXP2(x) — 2 raised to the power x (Turbo BASIC extension)
	BuiltinExp10                  // EXP10(x) — 10 raised to the power x (Turbo BASIC extension)
	BuiltinLog                    // LOG(x) — natural logarithm (base e)
	BuiltinLog2                   // LOG2(x) — logarithm base 2 (Turbo BASIC extension)
	BuiltinLog10                  // LOG10(x) — logarithm base 10 (Turbo BASIC extension)
	BuiltinSin                    // SIN(x) — sine of x (radians)
	BuiltinCos                    // COS(x) — cosine of x (radians)
	BuiltinTan                    // TAN(x) — tangent of x (radians)
	BuiltinAtn                    // ATN(x) — arctangent of x; returns angle in radians
	BuiltinCint                   // CINT(x) — convert to int16, rounding to nearest even
	BuiltinClng                   // CLNG(x) — convert to int32, rounding to nearest even
	BuiltinCsng                   // CSNG(x) — convert to single-precision float32
	BuiltinCdbl                   // CDBL(x) — convert to double-precision float64
	BuiltinRnd                    // RND(x) — pseudo-random float in [0, 1); seed if x<0

	// ---------------------------------------------------------------------------
	// String functions — string manipulation and inspection
	// ---------------------------------------------------------------------------

	BuiltinLeft   // LEFT$(s, n) — first n characters of s
	BuiltinRight  // RIGHT$(s, n) — last n characters of s
	BuiltinMid    // MID$(s, start, length) — substring of s starting at 1-based position
	BuiltinLen    // LEN(s) — number of characters in s
	BuiltinAsc    // ASC(s) — ASCII code of the first character of s
	BuiltinChr    // CHR$(n) — character whose ASCII code is n
	BuiltinStr    // STR$(n) — string representation of numeric value n
	BuiltinVal    // VAL(s) — numeric value parsed from string s
	BuiltinInstr  // INSTR(start, s, find) — position of find$ within s$, starting at start
	BuiltinUCase  // UCASE$(s) — convert s to upper case
	BuiltinLCase  // LCASE$(s) — convert s to lower case
	BuiltinLTrim  // LTRIM$(s) — remove leading spaces from s
	BuiltinRTrim  // RTRIM$(s) — remove trailing spaces from s
	BuiltinTrim   // TRIM$(s) — remove both leading and trailing spaces (Turbo BASIC extension)
	BuiltinSpace  // SPACE$(n) — string of n space characters
	BuiltinString // STRING$(n, char) — string of n repetitions of char
	BuiltinHex    // HEX$(n) — hexadecimal string representation of integer n
	BuiltinOct    // OCT$(n) — octal string representation of integer n
	BuiltinBin    // BIN$(n) — binary string representation of integer n (Turbo BASIC extension)

	// ---------------------------------------------------------------------------
	// Binary conversion functions (string <-> binary encoding)
	//
	// These functions pack and unpack numeric values into raw byte strings,
	// historically used for writing typed data to binary files in Turbo BASIC.
	// ---------------------------------------------------------------------------

	BuiltinMki // MKI$(n) — pack int16 n into a 2-byte string
	BuiltinMkl // MKL$(n) — pack int32 n into a 4-byte string
	BuiltinMks // MKS$(n) — pack float32 n into a 4-byte string
	BuiltinMkd // MKD$(n) — pack float64 n into an 8-byte string
	BuiltinCvi // CVI(s) — unpack 2-byte string s into int16
	BuiltinCvl // CVL(s) — unpack 4-byte string s into int32
	BuiltinCvs // CVS(s) — unpack 4-byte string s into float32
	BuiltinCvd // CVD(s) — unpack 8-byte string s into float64

	// ---------------------------------------------------------------------------
	// I/O formatting functions
	// ---------------------------------------------------------------------------

	BuiltinTab // TAB(n) — move print cursor to column n; used inside PRINT
	BuiltinSpc // SPC(n) — insert n spaces into PRINT output
)

// builtinNames maps each BuiltinID to its human-readable name.
var builtinNames = map[BuiltinID]string{
	BuiltinAbs:    "ABS",
	BuiltinSgn:    "SGN",
	BuiltinInt:    "INT",
	BuiltinFix:    "FIX",
	BuiltinCeil:   "CEIL",
	BuiltinSqr:    "SQR",
	BuiltinExp:    "EXP",
	BuiltinExp2:   "EXP2",
	BuiltinExp10:  "EXP10",
	BuiltinLog:    "LOG",
	BuiltinLog2:   "LOG2",
	BuiltinLog10:  "LOG10",
	BuiltinSin:    "SIN",
	BuiltinCos:    "COS",
	BuiltinTan:    "TAN",
	BuiltinAtn:    "ATN",
	BuiltinCint:   "CINT",
	BuiltinClng:   "CLNG",
	BuiltinCsng:   "CSNG",
	BuiltinCdbl:   "CDBL",
	BuiltinRnd:    "RND",
	BuiltinLeft:   "LEFT$",
	BuiltinRight:  "RIGHT$",
	BuiltinMid:    "MID$",
	BuiltinLen:    "LEN",
	BuiltinAsc:    "ASC",
	BuiltinChr:    "CHR$",
	BuiltinStr:    "STR$",
	BuiltinVal:    "VAL",
	BuiltinInstr:  "INSTR",
	BuiltinUCase:  "UCASE$",
	BuiltinLCase:  "LCASE$",
	BuiltinLTrim:  "LTRIM$",
	BuiltinRTrim:  "RTRIM$",
	BuiltinTrim:   "TRIM$",
	BuiltinSpace:  "SPACE$",
	BuiltinString: "STRING$",
	BuiltinHex:    "HEX$",
	BuiltinOct:    "OCT$",
	BuiltinBin:    "BIN$",
	BuiltinMki:    "MKI$",
	BuiltinMkl:    "MKL$",
	BuiltinMks:    "MKS$",
	BuiltinMkd:    "MKD$",
	BuiltinCvi:    "CVI",
	BuiltinCvl:    "CVL",
	BuiltinCvs:    "CVS",
	BuiltinCvd:    "CVD",
	BuiltinTab:    "TAB",
	BuiltinSpc:    "SPC",
}

// String returns the human-readable name of the built-in function.
func (id BuiltinID) String() string {
	if name, ok := builtinNames[id]; ok {
		return name
	}
	return fmt.Sprintf("BUILTIN_UNKNOWN(%d)", int(id))
}
