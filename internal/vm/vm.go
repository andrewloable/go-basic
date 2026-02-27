package vm

import (
	"fmt"
	"math"
	"strings"

	"github.com/loabletech/go-basic/internal/runtime"
)

// ---------------------------------------------------------------------------
// Value — the universal value type on the VM stack
// ---------------------------------------------------------------------------

// ValueType classifies the type of a Value.
type ValueType byte

const (
	ValInt    ValueType = iota // Integer (stored in Int field as int64)
	ValFloat                  // Floating-point (stored in Float field as float64)
	ValString                 // String (stored in Str field)
)

// Value is the tagged union that lives on the operand stack and in variables.
type Value struct {
	Type  ValueType
	Int   int64
	Float float64
	Str   string
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
type Instruction struct {
	Op      Opcode
	Operand int32 // index into constant pool, variable index, jump target, etc.
}

// String returns a human-readable disassembly of the instruction.
func (inst Instruction) String() string {
	return fmt.Sprintf("%-12s %d", inst.Op.String(), inst.Operand)
}

// Chunk holds the compiled bytecode for a program.
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

// ---------------------------------------------------------------------------
// VM
// ---------------------------------------------------------------------------

// VM is the virtual machine that executes bytecode.
type VM struct {
	chunk     *Chunk
	ip        int            // instruction pointer (index into chunk.Code)
	stack     []Value        // operand stack
	sp        int            // stack pointer (index of next free slot)
	globals   map[string]Value
	arrays    map[string]*arrayValue
	callStack []CallFrame
	dataPool  []Value // DATA values
	dataPtr   int     // current READ position
	running   bool
	rng       *runtime.RNG // random number generator
	output    *strings.Builder // captured output (nil = write to stdout)
}

// NewVM creates a VM ready to execute the given chunk.
func NewVM(chunk *Chunk) *VM {
	return &VM{
		chunk:   chunk,
		ip:      0,
		stack:   make([]Value, 0, 256),
		sp:      0,
		globals: make(map[string]Value),
		arrays:  make(map[string]*arrayValue),
		running: false,
		rng:     runtime.NewRNG(),
	}
}

// SetOutput redirects all PRINT output to the given builder instead of stdout.
// This is useful for testing.
func (vm *VM) SetOutput(b *strings.Builder) {
	vm.output = b
}

// SetDataPool sets the DATA pool that READ will consume.
func (vm *VM) SetDataPool(data []Value) {
	vm.dataPool = data
	vm.dataPtr = 0
}

// ---------------------------------------------------------------------------
// Stack operations
// ---------------------------------------------------------------------------

func (vm *VM) push(v Value) {
	if vm.sp >= len(vm.stack) {
		vm.stack = append(vm.stack, v)
	} else {
		vm.stack[vm.sp] = v
	}
	vm.sp++
}

func (vm *VM) pop() Value {
	if vm.sp <= 0 {
		// This should never happen in a correctly compiled program.
		panic("vm: stack underflow")
	}
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) peek() Value {
	if vm.sp <= 0 {
		panic("vm: stack underflow on peek")
	}
	return vm.stack[vm.sp-1]
}

// ---------------------------------------------------------------------------
// Runtime error helper
// ---------------------------------------------------------------------------

func (vm *VM) runtimeError(format string, args ...interface{}) error {
	line := 0
	// ip has already been incremented past the current instruction,
	// so use ip-1 for the line number of the faulting instruction.
	idx := vm.ip - 1
	if idx >= 0 && idx < len(vm.chunk.Lines) {
		line = vm.chunk.Lines[idx]
	}
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("runtime error at line %d: %s", line, msg)
}

// ---------------------------------------------------------------------------
// Output helper
// ---------------------------------------------------------------------------

func (vm *VM) printStr(s string) {
	if vm.output != nil {
		vm.output.WriteString(s)
	} else {
		fmt.Print(s)
	}
}

// ---------------------------------------------------------------------------
// Run — main execution loop
// ---------------------------------------------------------------------------

// Run executes the bytecode in the chunk until OpHalt or an error.
func (vm *VM) Run() error {
	vm.running = true
	vm.ip = 0

	for vm.running && vm.ip < len(vm.chunk.Code) {
		inst := vm.chunk.Code[vm.ip]
		vm.ip++

		switch inst.Op {

		// ---------------------------------------------------------------
		// Stack operations
		// ---------------------------------------------------------------
		case OpPush:
			idx := inst.Operand
			if idx < 0 || int(idx) >= len(vm.chunk.Constants) {
				return vm.runtimeError("invalid constant index %d", idx)
			}
			vm.push(vm.chunk.Constants[idx])

		case OpPop:
			vm.pop()

		case OpDup:
			vm.push(vm.peek())

		// ---------------------------------------------------------------
		// Variable operations
		// ---------------------------------------------------------------
		case OpLoad:
			name := vm.chunk.Constants[inst.Operand].Str
			v, ok := vm.globals[name]
			if !ok {
				// Uninitialized variable — return zero value.
				v = Value{Type: ValInt, Int: 0}
			}
			vm.push(v)

		case OpStore:
			name := vm.chunk.Constants[inst.Operand].Str
			vm.globals[name] = vm.pop()

		case OpLoadArray:
			name := vm.chunk.Constants[inst.Operand].Str
			arr, ok := vm.arrays[name]
			if !ok {
				return vm.runtimeError("array %q not dimensioned", name)
			}
			idx, err := vm.computeArrayIndex(arr)
			if err != nil {
				return err
			}
			vm.push(arr.data[idx])

		case OpStoreArray:
			name := vm.chunk.Constants[inst.Operand].Str
			arr, ok := vm.arrays[name]
			if !ok {
				return vm.runtimeError("array %q not dimensioned", name)
			}
			val := vm.pop()
			idx, err := vm.computeArrayIndex(arr)
			if err != nil {
				return err
			}
			arr.data[idx] = val

		// ---------------------------------------------------------------
		// Arithmetic
		// ---------------------------------------------------------------
		case OpAdd:
			b := vm.pop()
			a := vm.pop()
			if a.Type == ValString && b.Type == ValString {
				// String addition is concatenation.
				vm.push(StringVal(a.Str + b.Str))
			} else if a.Type == ValInt && b.Type == ValInt {
				vm.push(IntVal(a.Int + b.Int))
			} else {
				vm.push(FloatVal(a.asFloat() + b.asFloat()))
			}

		case OpSub:
			b := vm.pop()
			a := vm.pop()
			if a.Type == ValInt && b.Type == ValInt {
				vm.push(IntVal(a.Int - b.Int))
			} else {
				vm.push(FloatVal(a.asFloat() - b.asFloat()))
			}

		case OpMul:
			b := vm.pop()
			a := vm.pop()
			if a.Type == ValInt && b.Type == ValInt {
				vm.push(IntVal(a.Int * b.Int))
			} else {
				vm.push(FloatVal(a.asFloat() * b.asFloat()))
			}

		case OpDiv:
			b := vm.pop()
			a := vm.pop()
			bf := b.asFloat()
			if bf == 0 {
				return vm.runtimeError("division by zero")
			}
			vm.push(FloatVal(a.asFloat() / bf))

		case OpIDiv:
			b := vm.pop()
			a := vm.pop()
			bi := b.asInt()
			if bi == 0 {
				return vm.runtimeError("division by zero")
			}
			vm.push(IntVal(a.asInt() / bi))

		case OpMod:
			b := vm.pop()
			a := vm.pop()
			bi := b.asInt()
			if bi == 0 {
				return vm.runtimeError("division by zero (MOD)")
			}
			vm.push(IntVal(a.asInt() % bi))

		case OpPow:
			b := vm.pop()
			a := vm.pop()
			vm.push(FloatVal(math.Pow(a.asFloat(), b.asFloat())))

		case OpNeg:
			a := vm.pop()
			if a.Type == ValInt {
				vm.push(IntVal(-a.Int))
			} else {
				vm.push(FloatVal(-a.asFloat()))
			}

		// ---------------------------------------------------------------
		// Comparison — push -1 (true) or 0 (false)
		// ---------------------------------------------------------------
		case OpEq:
			b := vm.pop()
			a := vm.pop()
			vm.push(boolVal(vm.compareValues(a, b) == 0))

		case OpNe:
			b := vm.pop()
			a := vm.pop()
			vm.push(boolVal(vm.compareValues(a, b) != 0))

		case OpLt:
			b := vm.pop()
			a := vm.pop()
			vm.push(boolVal(vm.compareValues(a, b) < 0))

		case OpGt:
			b := vm.pop()
			a := vm.pop()
			vm.push(boolVal(vm.compareValues(a, b) > 0))

		case OpLe:
			b := vm.pop()
			a := vm.pop()
			vm.push(boolVal(vm.compareValues(a, b) <= 0))

		case OpGe:
			b := vm.pop()
			a := vm.pop()
			vm.push(boolVal(vm.compareValues(a, b) >= 0))

		// ---------------------------------------------------------------
		// Logical (operate on integers)
		// ---------------------------------------------------------------
		case OpAnd:
			b := vm.pop()
			a := vm.pop()
			vm.push(IntVal(a.asInt() & b.asInt()))

		case OpOr:
			b := vm.pop()
			a := vm.pop()
			vm.push(IntVal(a.asInt() | b.asInt()))

		case OpXor:
			b := vm.pop()
			a := vm.pop()
			vm.push(IntVal(a.asInt() ^ b.asInt()))

		case OpNot:
			a := vm.pop()
			vm.push(IntVal(^a.asInt()))

		case OpEqv:
			b := vm.pop()
			a := vm.pop()
			vm.push(IntVal(^(a.asInt() ^ b.asInt())))

		case OpImp:
			b := vm.pop()
			a := vm.pop()
			vm.push(IntVal((^a.asInt()) | b.asInt()))

		// ---------------------------------------------------------------
		// String operations
		// ---------------------------------------------------------------
		case OpConcat:
			b := vm.pop()
			a := vm.pop()
			vm.push(StringVal(a.Str + b.Str))

		// ---------------------------------------------------------------
		// Control flow
		// ---------------------------------------------------------------
		case OpJmp:
			vm.ip = int(inst.Operand)

		case OpJmpTrue:
			v := vm.pop()
			if v.isTruthy() {
				vm.ip = int(inst.Operand)
			}

		case OpJmpFalse:
			v := vm.pop()
			if !v.isTruthy() {
				vm.ip = int(inst.Operand)
			}

		case OpCall:
			vm.callStack = append(vm.callStack, CallFrame{
				ReturnAddr: vm.ip,
				LocalBase:  0,
			})
			vm.ip = int(inst.Operand)

		case OpRet:
			if len(vm.callStack) == 0 {
				return vm.runtimeError("RETURN without CALL")
			}
			frame := vm.callStack[len(vm.callStack)-1]
			vm.callStack = vm.callStack[:len(vm.callStack)-1]
			vm.ip = frame.ReturnAddr

		case OpGosub:
			vm.callStack = append(vm.callStack, CallFrame{
				ReturnAddr: vm.ip,
				LocalBase:  0,
			})
			vm.ip = int(inst.Operand)

		case OpReturn:
			if len(vm.callStack) == 0 {
				return vm.runtimeError("RETURN without GOSUB")
			}
			frame := vm.callStack[len(vm.callStack)-1]
			vm.callStack = vm.callStack[:len(vm.callStack)-1]
			vm.ip = frame.ReturnAddr

		// ---------------------------------------------------------------
		// I/O
		// ---------------------------------------------------------------
		case OpPrint:
			v := vm.pop()
			vm.printStr(v.asString())

		case OpPrintNewline:
			vm.printStr("\n")

		case OpPrintTab:
			vm.printStr("\t")

		case OpPrintSemicolon:
			// No-op spacer — prevents newline in PRINT but doesn't emit output.

		case OpInput:
			prompt := ""
			if inst.Operand >= 0 && int(inst.Operand) < len(vm.chunk.Constants) {
				prompt = vm.chunk.Constants[inst.Operand].Str
			}
			line := runtime.InputPrompt(prompt)
			vm.push(StringVal(line))

		// ---------------------------------------------------------------
		// Type conversion
		// ---------------------------------------------------------------
		case OpToInt:
			v := vm.pop()
			vm.push(IntVal(v.asInt()))

		case OpToLong:
			v := vm.pop()
			vm.push(IntVal(v.asInt()))

		case OpToSingle:
			v := vm.pop()
			vm.push(FloatVal(float64(float32(v.asFloat()))))

		case OpToDouble:
			v := vm.pop()
			vm.push(FloatVal(v.asFloat()))

		case OpToString:
			v := vm.pop()
			vm.push(StringVal(v.asString()))

		// ---------------------------------------------------------------
		// Built-in functions
		// ---------------------------------------------------------------
		case OpBuiltin:
			if err := vm.execBuiltin(BuiltinID(inst.Operand)); err != nil {
				return err
			}

		// ---------------------------------------------------------------
		// Array
		// ---------------------------------------------------------------
		case OpDimArray:
			numDims := int(inst.Operand)
			nameVal := vm.pop()
			name := nameVal.Str

			dims := make([]int, numDims)
			totalSize := 1
			for i := numDims - 1; i >= 0; i-- {
				sizeVal := vm.pop()
				size := int(sizeVal.asInt()) + 1 // BASIC arrays are 0..N, so N+1 elements
				dims[i] = size
				totalSize *= size
			}

			vm.arrays[name] = &arrayValue{
				dims: dims,
				data: make([]Value, totalSize),
			}

		// ---------------------------------------------------------------
		// Data
		// ---------------------------------------------------------------
		case OpRead:
			if vm.dataPtr >= len(vm.dataPool) {
				return vm.runtimeError("out of DATA")
			}
			val := vm.dataPool[vm.dataPtr]
			vm.dataPtr++
			name := vm.chunk.Constants[inst.Operand].Str
			vm.globals[name] = val

		case OpRestore:
			target := int(inst.Operand)
			if target < 0 {
				vm.dataPtr = 0
			} else {
				vm.dataPtr = target
			}

		// ---------------------------------------------------------------
		// Misc
		// ---------------------------------------------------------------
		case OpHalt:
			vm.running = false

		case OpNop:
			// do nothing

		case OpLine:
			// Line marker — used only for debugging/error reporting. No action.

		default:
			return vm.runtimeError("unknown opcode %d", inst.Op)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Comparison helper
// ---------------------------------------------------------------------------

// compareValues returns <0, 0, or >0 analogous to strcmp semantics.
func (vm *VM) compareValues(a, b Value) int {
	// String comparison.
	if a.Type == ValString && b.Type == ValString {
		return strings.Compare(a.Str, b.Str)
	}
	// Numeric comparison.
	af := a.asFloat()
	bf := b.asFloat()
	if af < bf {
		return -1
	}
	if af > bf {
		return 1
	}
	return 0
}

// boolVal converts a Go bool to the BASIC true/false Value.
func boolVal(b bool) Value {
	if b {
		return TrueValue
	}
	return FalseValue
}

// ---------------------------------------------------------------------------
// Array index computation
// ---------------------------------------------------------------------------

func (vm *VM) computeArrayIndex(arr *arrayValue) (int, error) {
	numDims := len(arr.dims)
	indices := make([]int, numDims)
	for i := numDims - 1; i >= 0; i-- {
		indices[i] = int(vm.pop().asInt())
	}

	// Compute flat index (row-major order).
	flat := 0
	multiplier := 1
	for i := numDims - 1; i >= 0; i-- {
		idx := indices[i]
		if idx < 0 || idx >= arr.dims[i] {
			return 0, vm.runtimeError("array index out of bounds: dimension %d, index %d, size %d",
				i, idx, arr.dims[i])
		}
		flat += idx * multiplier
		multiplier *= arr.dims[i]
	}
	return flat, nil
}

// ---------------------------------------------------------------------------
// Built-in function dispatch
// ---------------------------------------------------------------------------

func (vm *VM) execBuiltin(id BuiltinID) error {
	switch id {

	// --- Math (1-argument, numeric) ---
	case BuiltinAbs:
		v := vm.pop()
		vm.push(FloatVal(runtime.Abs(v.asFloat())))

	case BuiltinSgn:
		v := vm.pop()
		vm.push(IntVal(int64(runtime.Sgn(v.asFloat()))))

	case BuiltinInt:
		v := vm.pop()
		vm.push(FloatVal(runtime.IntFloor(v.asFloat())))

	case BuiltinFix:
		v := vm.pop()
		vm.push(FloatVal(runtime.Fix(v.asFloat())))

	case BuiltinCeil:
		v := vm.pop()
		vm.push(FloatVal(runtime.Ceil(v.asFloat())))

	case BuiltinSqr:
		v := vm.pop()
		result, err := runtime.Sqr(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	case BuiltinExp:
		v := vm.pop()
		vm.push(FloatVal(runtime.Exp(v.asFloat())))

	case BuiltinExp2:
		v := vm.pop()
		vm.push(FloatVal(runtime.Exp2(v.asFloat())))

	case BuiltinExp10:
		v := vm.pop()
		vm.push(FloatVal(runtime.Exp10(v.asFloat())))

	case BuiltinLog:
		v := vm.pop()
		result, err := runtime.Log(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	case BuiltinLog2:
		v := vm.pop()
		result, err := runtime.Log2(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	case BuiltinLog10:
		v := vm.pop()
		result, err := runtime.Log10(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	case BuiltinSin:
		v := vm.pop()
		vm.push(FloatVal(runtime.Sin(v.asFloat())))

	case BuiltinCos:
		v := vm.pop()
		vm.push(FloatVal(runtime.Cos(v.asFloat())))

	case BuiltinTan:
		v := vm.pop()
		vm.push(FloatVal(runtime.Tan(v.asFloat())))

	case BuiltinAtn:
		v := vm.pop()
		vm.push(FloatVal(runtime.Atn(v.asFloat())))

	case BuiltinCint:
		v := vm.pop()
		result, err := runtime.Cint(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinClng:
		v := vm.pop()
		result, err := runtime.Clng(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinCsng:
		v := vm.pop()
		vm.push(FloatVal(float64(runtime.Csng(v.asFloat()))))

	case BuiltinCdbl:
		v := vm.pop()
		vm.push(FloatVal(runtime.Cdbl(v.asFloat())))

	case BuiltinRnd:
		v := vm.pop()
		vm.push(FloatVal(vm.rng.Rnd(v.asFloat())))

	// --- String functions ---
	case BuiltinLeft:
		n := vm.pop()
		s := vm.pop()
		vm.push(StringVal(runtime.Left(s.Str, int(n.asInt()))))

	case BuiltinRight:
		n := vm.pop()
		s := vm.pop()
		vm.push(StringVal(runtime.Right(s.Str, int(n.asInt()))))

	case BuiltinMid:
		length := vm.pop()
		start := vm.pop()
		s := vm.pop()
		vm.push(StringVal(runtime.Mid(s.Str, int(start.asInt()), int(length.asInt()))))

	case BuiltinLen:
		s := vm.pop()
		vm.push(IntVal(int64(runtime.Len(s.Str))))

	case BuiltinAsc:
		s := vm.pop()
		result, err := runtime.Asc(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinChr:
		n := vm.pop()
		result, err := runtime.Chr(int(n.asInt()))
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(StringVal(result))

	case BuiltinStr:
		n := vm.pop()
		vm.push(StringVal(runtime.Str(n.asFloat())))

	case BuiltinVal:
		s := vm.pop()
		vm.push(FloatVal(runtime.Val(s.Str)))

	case BuiltinInstr:
		find := vm.pop()
		s := vm.pop()
		start := vm.pop()
		vm.push(IntVal(int64(runtime.Instr(int(start.asInt()), s.Str, find.Str))))

	case BuiltinUCase:
		s := vm.pop()
		vm.push(StringVal(runtime.UCase(s.Str)))

	case BuiltinLCase:
		s := vm.pop()
		vm.push(StringVal(runtime.LCase(s.Str)))

	case BuiltinLTrim:
		s := vm.pop()
		vm.push(StringVal(runtime.LTrim(s.Str)))

	case BuiltinRTrim:
		s := vm.pop()
		vm.push(StringVal(runtime.RTrim(s.Str)))

	case BuiltinTrim:
		s := vm.pop()
		vm.push(StringVal(runtime.Trim(s.Str)))

	case BuiltinSpace:
		n := vm.pop()
		vm.push(StringVal(runtime.Space(int(n.asInt()))))

	case BuiltinString:
		char := vm.pop()
		n := vm.pop()
		vm.push(StringVal(runtime.StringRepeat(int(n.asInt()), byte(char.asInt()))))

	case BuiltinHex:
		n := vm.pop()
		vm.push(StringVal(runtime.Hex(int(n.asInt()))))

	case BuiltinOct:
		n := vm.pop()
		vm.push(StringVal(runtime.Oct(int(n.asInt()))))

	case BuiltinBin:
		n := vm.pop()
		vm.push(StringVal(runtime.Bin(int(n.asInt()))))

	// --- Binary conversion functions ---
	case BuiltinMki:
		n := vm.pop()
		vm.push(StringVal(runtime.Mki(int16(n.asInt()))))

	case BuiltinMkl:
		n := vm.pop()
		vm.push(StringVal(runtime.Mkl(int32(n.asInt()))))

	case BuiltinMks:
		n := vm.pop()
		vm.push(StringVal(runtime.Mks(float32(n.asFloat()))))

	case BuiltinMkd:
		n := vm.pop()
		vm.push(StringVal(runtime.Mkd(n.asFloat())))

	case BuiltinCvi:
		s := vm.pop()
		result, err := runtime.Cvi(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinCvl:
		s := vm.pop()
		result, err := runtime.Cvl(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinCvs:
		s := vm.pop()
		result, err := runtime.Cvs(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(float64(result)))

	case BuiltinCvd:
		s := vm.pop()
		result, err := runtime.Cvd(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	// --- I/O helpers ---
	case BuiltinTab:
		n := vm.pop()
		vm.push(StringVal(runtime.Spc(int(n.asInt()))))

	case BuiltinSpc:
		n := vm.pop()
		vm.push(StringVal(runtime.Spc(int(n.asInt()))))

	default:
		return vm.runtimeError("unknown built-in function ID %d", id)
	}

	return nil
}
