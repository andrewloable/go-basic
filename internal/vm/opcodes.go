package vm

import "fmt"

// ---------------------------------------------------------------------------
// Opcode — bytecode instruction set for the Turbo BASIC VM
// ---------------------------------------------------------------------------

// Opcode represents a single bytecode instruction.
type Opcode byte

const (
	// Stack operations
	OpPush Opcode = iota // Push constant onto stack (operand = constant pool index)
	OpPop                // Pop and discard top of stack
	OpDup                // Duplicate top of stack

	// Variable operations
	OpLoad       // Load variable value (operand = variable name index in constant pool)
	OpStore      // Store into variable (operand = variable name index in constant pool)
	OpLoadArray  // Load array element (operand = variable name index; indices on stack)
	OpStoreArray // Store into array element (operand = variable name index; indices+value on stack)

	// Arithmetic
	OpAdd  // Add top two values
	OpSub  // Subtract (NOS - TOS)
	OpMul  // Multiply
	OpDiv  // Floating-point division
	OpIDiv // Integer division (\)
	OpMod  // Modulo
	OpPow  // Exponentiation (NOS ^ TOS)
	OpNeg  // Unary negate

	// Comparison — push -1 (true) or 0 (false) per BASIC convention
	OpEq // Equal
	OpNe // Not equal
	OpLt // Less than
	OpGt // Greater than
	OpLe // Less than or equal
	OpGe // Greater than or equal

	// Logical (bitwise on integers, boolean on floats)
	OpAnd // AND
	OpOr  // OR
	OpXor // XOR
	OpNot // NOT (unary)
	OpEqv // EQV (equivalence: NOT (a XOR b))
	OpImp // IMP (implication: (NOT a) OR b)

	// String operations
	OpConcat // String concatenation

	// Control flow
	OpJmp      // Unconditional jump (operand = target instruction index)
	OpJmpTrue  // Jump if top of stack is true (non-zero); pops the value
	OpJmpFalse // Jump if top of stack is false (zero); pops the value
	OpCall     // Call subroutine/function (operand = target instruction index)
	OpRet      // Return from SUB/FUNCTION
	OpGosub    // GOSUB — push return address and jump (operand = target)
	OpReturn   // RETURN from GOSUB

	// I/O
	OpPrint          // Print top of stack as string
	OpPrintNewline   // Print newline
	OpPrintTab       // Print tab (advance to next 14-char zone)
	OpPrintSemicolon // Suppress trailing space (no-op spacer for PRINT layout)
	OpInput          // Input from stdin (operand = constant pool index of prompt string)

	// Type conversion
	OpToInt    // Convert TOS to integer
	OpToLong   // Convert TOS to long integer
	OpToSingle // Convert TOS to single-precision float
	OpToDouble // Convert TOS to double-precision float
	OpToString // Convert TOS to string

	// Built-in functions
	OpBuiltin // Call built-in function by BuiltinID (operand = BuiltinID)

	// Array
	OpDimArray // DIM array (operand = number of dimensions; sizes on stack, then name index)

	// Data
	OpRead    // READ from DATA pool into variable (operand = variable name index)
	OpRestore // RESTORE data pointer (operand = target data index, or -1 for beginning)

	// Misc
	OpHalt // End program
	OpNop  // No operation
	OpLine // Line number marker for debugging (operand = source line number)
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
// ---------------------------------------------------------------------------

// BuiltinID identifies a built-in function for the OpBuiltin instruction.
type BuiltinID byte

const (
	// Math functions
	BuiltinAbs   BuiltinID = iota // ABS(x)
	BuiltinSgn                    // SGN(x)
	BuiltinInt                    // INT(x) — floor
	BuiltinFix                    // FIX(x) — truncate toward zero
	BuiltinCeil                   // CEIL(x) — Turbo BASIC extension
	BuiltinSqr                   // SQR(x) — square root
	BuiltinExp                   // EXP(x) — e^x
	BuiltinExp2                  // EXP2(x) — Turbo BASIC extension
	BuiltinExp10                 // EXP10(x) — Turbo BASIC extension
	BuiltinLog                   // LOG(x) — natural log
	BuiltinLog2                  // LOG2(x) — Turbo BASIC extension
	BuiltinLog10                 // LOG10(x) — Turbo BASIC extension
	BuiltinSin                   // SIN(x)
	BuiltinCos                   // COS(x)
	BuiltinTan                   // TAN(x)
	BuiltinAtn                   // ATN(x) — arctangent
	BuiltinCint                  // CINT(x) — convert to int16
	BuiltinClng                  // CLNG(x) — convert to int32
	BuiltinCsng                  // CSNG(x) — convert to float32
	BuiltinCdbl                  // CDBL(x) — convert to float64
	BuiltinRnd                   // RND(x) — random number

	// String functions
	BuiltinLeft   // LEFT$(s, n)
	BuiltinRight  // RIGHT$(s, n)
	BuiltinMid    // MID$(s, start, length)
	BuiltinLen    // LEN(s)
	BuiltinAsc    // ASC(s)
	BuiltinChr    // CHR$(n)
	BuiltinStr    // STR$(n)
	BuiltinVal    // VAL(s)
	BuiltinInstr  // INSTR(start, s, find)
	BuiltinUCase  // UCASE$(s)
	BuiltinLCase  // LCASE$(s)
	BuiltinLTrim  // LTRIM$(s)
	BuiltinRTrim  // RTRIM$(s)
	BuiltinTrim   // TRIM$(s) — Turbo BASIC extension
	BuiltinSpace  // SPACE$(n)
	BuiltinString // STRING$(n, char)
	BuiltinHex   // HEX$(n)
	BuiltinOct   // OCT$(n)
	BuiltinBin   // BIN$(n) — Turbo BASIC extension

	// Conversion functions (string <-> binary)
	BuiltinMki // MKI$(n) — int16 to 2-byte string
	BuiltinMkl // MKL$(n) — int32 to 4-byte string
	BuiltinMks // MKS$(n) — float32 to 4-byte string
	BuiltinMkd // MKD$(n) — float64 to 8-byte string
	BuiltinCvi // CVI(s) — 2-byte string to int16
	BuiltinCvl // CVL(s) — 4-byte string to int32
	BuiltinCvs // CVS(s) — 4-byte string to float32
	BuiltinCvd // CVD(s) — 8-byte string to float64

	// I/O functions
	BuiltinTab // TAB(n) — cursor position
	BuiltinSpc // SPC(n) — spaces
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
