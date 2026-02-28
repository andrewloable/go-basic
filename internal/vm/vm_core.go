package vm

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/runtime"
)

// ---------------------------------------------------------------------------
// VM — virtual machine state
//
// The VM struct holds all runtime state for one program execution.  Each field
// corresponds to a conceptual part of the "virtual computer":
//
//   chunk     — the ROM: the compiled program that the VM reads from but never
//               modifies (analogous to the text segment in a real process).
//
//   ip        — the instruction pointer (program counter): the index of the
//               next instruction to fetch from chunk.Code.  The VM pre-
//               increments ip before dispatching, so after a fetch ip points
//               past the just-fetched instruction.
//
//   stack/sp  — the operand stack: a pre-allocated slice used as a simple
//               fixed-size stack.  sp is the "stack pointer" — the index of
//               the next free slot.  Values below sp are live; values at sp
//               and above are dead (leftover from earlier pops).  Growing the
//               slice via append is avoided in the hot path; the slice is only
//               grown when sp reaches the current capacity.
//
//   globals   — a flat key/value store for all BASIC variables.  The key is
//               the upper-cased variable name (e.g. "COUNT", "X$").  Every
//               variable in the program lives here — there is no stack frame
//               for locals in this simple VM.
//
//   arrays    — a separate store for BASIC array variables; each entry holds
//               the dimension sizes and a flat Value slice.
//
//   callStack — return addresses and local-base offsets for GOSUB/RETURN and
//               OpCall/OpRet (SUB/FUNCTION calls).
//
//   dataPool  — all literal DATA values collected by the compiler's pass 1.
//               READ instructions walk this slice sequentially via dataPtr.
//
//   rng       — seeded random number generator; shared across the whole run
//               so that RND() produces a reproducible sequence.
//
//   output    — when non-nil, PRINT writes here instead of stdout.  Used by
//               the test suite to capture output without forking a process.
// ---------------------------------------------------------------------------

// VM is the virtual machine that executes bytecode.
type VM struct {
	chunk     *Chunk
	ip        int              // instruction pointer (index into chunk.Code)
	stack     []Value          // operand stack
	sp        int              // stack pointer (index of next free slot)
	globals   map[string]Value
	arrays    map[string]*arrayValue
	callStack []CallFrame
	dataPool  []Value          // DATA values
	dataPtr   int              // current READ position
	running   bool
	rng       *runtime.RNG    // random number generator
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
//
// The operand stack is the heart of a stack-based VM.  Every instruction that
// produces a value pushes it here; every instruction that consumes a value
// pops it.  The stack grows upward: index 0 is the bottom, index sp-1 is the
// top (TOS).
//
// Implementation note:
// Rather than using Go's built-in append/slice-shrink for every push/pop, the
// VM maintains an explicit integer stack pointer (vm.sp).  The underlying
// slice is grown with append only when sp reaches len(vm.stack); otherwise
// the value is written directly at vm.stack[vm.sp].  This avoids the overhead
// of slice-header copies and capacity checks on every operation.
// ---------------------------------------------------------------------------

// push places v on top of the operand stack and increments sp.
// If the stack slice is full it is grown by append (amortised O(1)).
func (vm *VM) push(v Value) {
	if vm.sp >= len(vm.stack) {
		vm.stack = append(vm.stack, v)
	} else {
		vm.stack[vm.sp] = v
	}
	vm.sp++
}

// pop removes and returns the top value from the operand stack.
// Stack underflow panics because it indicates a bug in the compiler or VM —
// a correctly compiled program never pops more values than it has pushed.
func (vm *VM) pop() Value {
	if vm.sp <= 0 {
		// This should never happen in a correctly compiled program.
		panic("vm: stack underflow")
	}
	vm.sp--
	return vm.stack[vm.sp]
}

// peek returns the top value without consuming it.
// Used when an instruction needs to inspect the top without removing it
// (e.g., OpDup copies TOS by peeking and then pushing the same value again).
func (vm *VM) peek() Value {
	if vm.sp <= 0 {
		panic("vm: stack underflow on peek")
	}
	return vm.stack[vm.sp-1]
}

// ---------------------------------------------------------------------------
// Runtime error helper
// ---------------------------------------------------------------------------

// runtimeError builds an error message that includes the BASIC source line
// number of the faulting instruction.
//
// Why ip-1?
// The VM's fetch-decode-execute loop pre-increments ip before dispatching:
//
//   inst := vm.chunk.Code[vm.ip]
//   vm.ip++      ← ip now points PAST the current instruction
//   switch inst.Op { ... }
//
// So when an error handler calls runtimeError(), vm.ip already points to the
// NEXT instruction.  To report the correct source line for the instruction
// that actually faulted, the handler looks up Chunk.Lines[vm.ip - 1].
//
// This ip-1 adjustment is a standard idiom in bytecode VMs; it is worth
// noting explicitly because it surprises most students at first.
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
// Run — main execution loop (fetch-decode-execute)
//
// This is the core of the virtual machine.  Every interpreter or VM is built
// around some variant of the "fetch-decode-execute" cycle:
//
//   Fetch   — read the next instruction from memory (Chunk.Code[vm.ip])
//   Decode  — identify what operation it represents (the Op field)
//   Execute — carry out the operation (the switch cases below)
//
// In this VM the cycle is a simple for-loop with a Go switch statement.  Go's
// switch compiles to a jump table on many platforms, making dispatch fast.
//
// How the loop terminates:
//   1. OpHalt sets vm.running = false  → loop condition becomes false.
//   2. vm.ip reaches len(chunk.Code)   → the for condition fails naturally.
//   3. An instruction returns an error → Run() returns that error immediately.
//
// Each case in the switch calls a helper method (execAdd, execJmp, …) that
// is responsible for popping operands, performing the operation, and pushing
// the result.  Keeping each case small makes the loop easy to reason about
// and the helpers easy to test in isolation.
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
			if err := vm.execLoad(inst); err != nil {
				return err
			}

		case OpStore:
			vm.execStore(inst)

		case OpLoadArray:
			if err := vm.execLoadArray(inst); err != nil {
				return err
			}

		case OpStoreArray:
			if err := vm.execStoreArray(inst); err != nil {
				return err
			}

		// ---------------------------------------------------------------
		// Arithmetic
		// ---------------------------------------------------------------
		case OpAdd:
			vm.execAdd()

		case OpSub:
			vm.execSub()

		case OpMul:
			vm.execMul()

		case OpDiv:
			if err := vm.execDiv(); err != nil {
				return err
			}

		case OpIDiv:
			if err := vm.execIDiv(); err != nil {
				return err
			}

		case OpMod:
			if err := vm.execMod(); err != nil {
				return err
			}

		case OpPow:
			vm.execPow()

		case OpNeg:
			vm.execNeg()

		// ---------------------------------------------------------------
		// Comparison — push -1 (true) or 0 (false)
		// ---------------------------------------------------------------
		case OpEq:
			vm.execEq()

		case OpNe:
			vm.execNe()

		case OpLt:
			vm.execLt()

		case OpGt:
			vm.execGt()

		case OpLe:
			vm.execLe()

		case OpGe:
			vm.execGe()

		// ---------------------------------------------------------------
		// Logical (operate on integers)
		// ---------------------------------------------------------------
		case OpAnd:
			vm.execAnd()

		case OpOr:
			vm.execOr()

		case OpXor:
			vm.execXor()

		case OpNot:
			vm.execNot()

		case OpEqv:
			vm.execEqv()

		case OpImp:
			vm.execImp()

		// ---------------------------------------------------------------
		// String operations
		// ---------------------------------------------------------------
		case OpConcat:
			vm.execConcat()

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
			vm.execCall(inst)

		case OpRet:
			if err := vm.execRet(); err != nil {
				return err
			}

		case OpGosub:
			vm.execGosub(inst)

		case OpReturn:
			if err := vm.execReturn(); err != nil {
				return err
			}

		// ---------------------------------------------------------------
		// I/O
		// ---------------------------------------------------------------
		case OpPrint:
			vm.execPrint()

		case OpPrintNewline:
			vm.execPrintNewline()

		case OpPrintTab:
			vm.execPrintTab()

		case OpPrintSemicolon:
			vm.execPrintSemicolon()

		case OpInput:
			vm.execInput(inst)

		// ---------------------------------------------------------------
		// Type conversion
		// ---------------------------------------------------------------
		case OpToInt:
			vm.execToInt()

		case OpToLong:
			vm.execToLong()

		case OpToSingle:
			vm.execToSingle()

		case OpToDouble:
			vm.execToDouble()

		case OpToString:
			vm.execToString()

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
			vm.execDimArray(inst)

		// ---------------------------------------------------------------
		// Data
		// ---------------------------------------------------------------
		case OpRead:
			if err := vm.execRead(inst); err != nil {
				return err
			}

		case OpRestore:
			vm.execRestore(inst)

		// ---------------------------------------------------------------
		// Misc
		// ---------------------------------------------------------------
		case OpHalt:
			vm.running = false

		case OpNop:
			// do nothing

		case OpLine:
			// Line marker — used only for debugging/error reporting. No action.

		case OpPoke:
			// POKE address, value — memory-mapped I/O is not available in Go;
			// pop both operands (for side-effect evaluation) and discard.
			vm.pop() // value
			vm.pop() // address

		case OpClear:
			// CLEAR — reset all global variables to their zero values.
			for k := range vm.globals {
				v := vm.globals[k]
				switch v.Type {
				case ValString:
					vm.globals[k] = StringVal("")
				default:
					vm.globals[k] = FloatVal(0)
				}
			}

		default:
			return vm.runtimeError("unknown opcode %d", inst.Op)
		}
	}

	return nil
}
