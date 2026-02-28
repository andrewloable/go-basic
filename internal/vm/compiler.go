// Package vm contains a bytecode compiler and its accompanying virtual machine.
//
// # Bytecode Compilation vs. Transpilation
//
// A transpiler (like the one in internal/transpiler) converts an AST directly
// into source code for another high-level language (e.g., BASIC → Go).  A
// bytecode compiler instead lowers the AST into a compact sequence of
// numbered instructions — the bytecode — that a purpose-built virtual machine
// (VM) interprets at runtime.  Bytecode sits between source code and native
// machine code: it is easier to generate than native code yet much faster for
// a VM to dispatch than re-parsing source on every execution.
//
// # The Stack-Based VM Model
//
// This VM is stack-based.  Every instruction operates on an implicit operand
// stack rather than named registers.  For example, to evaluate "A + B" the VM
// pushes the value of A, pushes the value of B, then executes OpAdd, which
// pops both values, adds them, and pushes the result.  The final result of any
// expression is always left on top of the stack for the next instruction to
// consume.  This model is simple to implement and to compile for because the
// compiler never has to manage register allocation.
//
// # Opcodes
//
// An opcode is a small integer (see opcode.go) that uniquely identifies one VM
// operation — OpPush, OpAdd, OpJmpFalse, etc.  Each instruction in the
// bytecode stream is a (Opcode, Operand, Line) triple stored as a flat array
// in Chunk.Code.  The operand is a 32-bit integer whose meaning depends on the
// opcode: it may be a constant-pool index, a variable index, a jump target
// address, or a built-in function ID.
//
// # Labels and Backpatching
//
// When the compiler encounters a forward jump (e.g., the jump that skips the
// THEN block when an IF condition is false) it does not yet know the target
// address because the destination code has not been emitted yet.  The solution
// is to emit the jump instruction immediately with a placeholder operand of 0
// and record the instruction's index.  Once the destination is known (after
// emitting all intermediate code) the compiler "backpatches" the placeholder
// by writing the real address into the operand field.  See patchJump(),
// labelPatches, and subPatches for the two flavours of backpatching used here.
//
// # Symbol Table / Variable Indices
//
// Variables are not stored in a hash map at runtime.  Instead, each unique
// variable name is interned into the constant pool as a string value, and its
// index in that pool (an int32) becomes the variable's identifier throughout
// the bytecode.  OpStore <idx> writes the top-of-stack into the VM's variable
// array at position <idx>; OpLoad <idx> pushes the current value back.  The
// compiler's varIndex map (name → constant-pool index) is the compile-time
// symbol table that performs this mapping.
package vm

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// ---------------------------------------------------------------------------
// Compiler -- translates a Turbo BASIC AST into VM bytecode (Chunk).
//
// The Compiler is a single-pass tree-walker with one pre-pass for DATA.
// It visits every AST node, decides which bytecode instructions implement
// that node's semantics, and appends them to the Chunk being built.
//
// Compiler state overview:
//   chunk       — the bytecode buffer being built (instructions + constants)
//   table       — optional semantic symbol table from the analysis phase
//   varIndex    — compile-time symbol table: maps variable names to their
//                 constant pool slot so OpLoad/OpStore know which slot to use
//   errors      — accumulates non-fatal compilation errors; reported at end
//   loopStack   — stack of active loops; EXIT FOR/DO/WHILE push break-jump
//                 indices here so patchJump() can fix them when the loop ends
//   labelAddrs  — maps BASIC line-number/label names to instruction indices;
//                 GOTO/GOSUB look up their target here
//   labelPatches — forward-reference table: when a GOTO is compiled before
//                  its target label, the jump instruction is recorded here and
//                  patched at the end of compilation
//   dataPool    — literal values from all DATA statements, collected in pass 1;
//                 READ instructions index into this pool at runtime
//   subAddrs    — maps SUB/FUNCTION names to their first instruction index
//   subPatches  — forward-reference table for function calls (same idea as
//                 labelPatches but for user-defined SUBs/FUNCTIONs)
//   constMap    — BASIC CONST values resolved at compile time; no variable
//                 slot is allocated; the value is inlined as a constant push
//   defFnMap    — DEF FN declarations stored for inline expansion (no OpCall
//                 overhead; the body is emitted directly at each call site)
//   typeDefMap  — TYPE block definitions stored for field-access mangling
//   defTypeMap  — per-letter default type suffix from DEFINT/DEFSNG/etc.
// ---------------------------------------------------------------------------

// Compiler walks the AST and emits bytecode instructions into a Chunk.
type Compiler struct {
	chunk     *Chunk
	table     *semantic.SymbolTable
	varIndex  map[string]int32 // variable name -> constant pool index
	errors    []string
	loopStack []loopInfo // for EXIT FOR/DO/WHILE

	// Label handling for GOTO/GOSUB
	labelAddrs   map[string]int // label name -> instruction index
	labelPatches []labelPatch   // forward references needing patching

	// DATA pool collected during compilation
	dataPool []Value

	// SUB/FUNCTION declarations: name -> instruction index of body start
	subAddrs   map[string]int
	subPatches []subPatch

	// constMap holds compile-time constants from CONST statements.
	// Values are pushed directly onto the stack (no variable slot needed).
	constMap map[string]Value

	// defFnMap holds DEF FN declarations for inline expansion at call sites.
	// The body is inlined instead of emitting a separate subroutine.
	defFnMap map[string]*ast.DefFnDeclaration

	// typeDefMap holds TYPE block field definitions for compile-time validation.
	// Key is the upper-cased type name; value is the list of fields.
	typeDefMap map[string][]ast.TypeField

	// defTypeMap maps each letter A-Z to its default BASIC type suffix.
	// Set by DEFINT/DEFLNG/DEFSNG/DEFDBL/DEFSTR statements.
	// Index 0 = 'A', index 25 = 'Z'.  Empty string means use default (float).
	defTypeMap [26]string

	// declaredFuncs records names that are known to be user-defined SUBs or
	// FUNCTIONs — populated from DECLARE SUB/FUNCTION (forward declarations)
	// and from the actual body declarations.  Used by compileArrayAccess to
	// distinguish a user-function call like Factorial(n) (parsed by the parser
	// as *ast.ArrayAccess because it is an unknown identifier) from a genuine
	// array element access.
	declaredFuncs map[string]bool
}

// loopInfo tracks a loop context for EXIT/break handling.
// The loopStack (a slice of loopInfo) implements a "loop nesting stack".
// When the compiler begins a FOR/WHILE/DO loop it pushes one loopInfo; when
// the loop ends it pops the info and patches all break jumps.  EXIT FOR inside
// a nested loop only needs to find the innermost FOR entry, which findLoop()
// does by scanning from the top.
type loopInfo struct {
	loopType   string // "FOR", "DO", "WHILE"
	loopStart  int    // instruction index of loop start (back-edge target)
	breakJumps []int  // instruction indices needing patching for EXIT FOR/DO/WHILE
}

// labelPatch records a forward reference to a label.
// When GOTO/GOSUB is compiled before the target label is seen, the compiler
// emits a jump with operand 0 and stores the jump's instruction index here.
// At the end of compilation all label patches are resolved using labelAddrs.
type labelPatch struct {
	instrIndex int    // index in chunk.Code to patch
	labelName  string // the label being referenced
}

// subPatch records a forward reference to a SUB/FUNCTION.
// Identical mechanism to labelPatch but used for user-defined functions called
// before their declaration appears in the source.
type subPatch struct {
	instrIndex int    // index in chunk.Code to patch
	name       string // SUB/FUNCTION name
}

// NewCompiler creates a new bytecode compiler.  If table is nil, the compiler
// will still work but without access to semantic information.
func NewCompiler(table *semantic.SymbolTable) *Compiler {
	return &Compiler{
		chunk:         &Chunk{},
		table:         table,
		varIndex:      make(map[string]int32),
		labelAddrs:    make(map[string]int),
		subAddrs:      make(map[string]int),
		constMap:      make(map[string]Value),
		defFnMap:      make(map[string]*ast.DefFnDeclaration),
		typeDefMap:    make(map[string][]ast.TypeField),
		declaredFuncs: make(map[string]bool),
	}
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Compile takes a parsed BASIC program and returns the compiled bytecode.
//
// Two-pass compilation pipeline:
//
// Pass 1 — collectData():
//   BASIC DATA statements declare literal values that READ consumes at runtime.
//   The data values can appear anywhere in the source (even after a READ that
//   uses them), so a single pass cannot handle them inline.  Pass 1 scans the
//   entire statement list, extracts every DATA literal, and appends it to
//   c.dataPool in source order.  The VM receives this pool before execution
//   via VM.SetDataPool().
//
// Pass 2 — compileStatement() loop:
//   Each statement in the AST is compiled in order, emitting zero or more
//   Instruction values into c.chunk.Code.  Expression sub-trees are compiled
//   recursively (see compile_expressions.go).  Jump instructions that reference
//   not-yet-emitted code use placeholder operands of 0 and are recorded in
//   labelPatches / subPatches for the post-pass below.
//
// Post-pass — backpatch forward references:
//   After all bytecode is emitted, every recorded forward-jump placeholder is
//   resolved: the compiler looks up the target name in labelAddrs / subAddrs
//   (populated when the target label/SUB was compiled) and writes the real
//   instruction index into the placeholder operand.
func (c *Compiler) Compile(program *ast.Program) (*Chunk, error) {
	if program == nil {
		return nil, fmt.Errorf("compiler: nil program")
	}

	if len(program.Statements) == 0 {
		c.emit(OpHalt, 0, 0)
		return c.chunk, nil
	}

	// Pass 1: collect DATA values.
	c.collectData(program.Statements)

	// Pass 2: emit bytecode.
	for _, stmt := range program.Statements {
		c.compileStatement(stmt)
	}

	// Emit halt at end of program.
	c.emit(OpHalt, 0, 0)

	// Patch all forward label references.
	for _, lp := range c.labelPatches {
		addr, ok := c.labelAddrs[strings.ToUpper(lp.labelName)]
		if !ok {
			c.addError("undefined label %q", lp.labelName)
			continue
		}
		c.chunk.Code[lp.instrIndex].Operand = int32(addr)
	}

	// Patch all forward SUB/FUNCTION references.
	for _, sp := range c.subPatches {
		addr, ok := c.subAddrs[strings.ToUpper(sp.name)]
		if !ok {
			c.addError("undefined SUB/FUNCTION %q", sp.name)
			continue
		}
		c.chunk.Code[sp.instrIndex].Operand = int32(addr)
	}

	if len(c.errors) > 0 {
		return nil, fmt.Errorf("compilation errors:\n%s", strings.Join(c.errors, "\n"))
	}

	return c.chunk, nil
}

// DataPool returns the DATA values collected during compilation.
// The caller should pass this to VM.SetDataPool before running.
func (c *Compiler) DataPool() []Value {
	return c.dataPool
}

// ---------------------------------------------------------------------------
// Pass 1: collect DATA statements
// ---------------------------------------------------------------------------

func (c *Compiler) collectData(stmts []ast.Statement) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.DataStatement:
			for _, expr := range s.Values {
				switch v := expr.(type) {
				case *ast.NumberLiteral:
					if v.NumType == ast.NumInt || v.NumType == ast.NumLong {
						c.dataPool = append(c.dataPool, IntVal(int64(v.Value)))
					} else {
						c.dataPool = append(c.dataPool, FloatVal(v.Value))
					}
				case *ast.StringLiteral:
					c.dataPool = append(c.dataPool, StringVal(v.Value))
				case *ast.UnaryExpr:
					if v.Operator == "-" {
						if num, ok := v.Operand.(*ast.NumberLiteral); ok {
							if num.NumType == ast.NumInt || num.NumType == ast.NumLong {
								c.dataPool = append(c.dataPool, IntVal(-int64(num.Value)))
							} else {
								c.dataPool = append(c.dataPool, FloatVal(-num.Value))
							}
							continue
						}
					}
					c.dataPool = append(c.dataPool, IntVal(0))
				default:
					c.dataPool = append(c.dataPool, IntVal(0))
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Emit helpers
//
// These helpers form the low-level interface between the compile* methods and
// the Chunk (the bytecode buffer).  The most important pair is emitJump /
// patchJump, which together implement backpatching — the technique used
// whenever a jump instruction must be emitted before its target address is
// known.
//
// How backpatching works, step by step:
//
//  1. emitJump(OpJmpFalse, line) appends the instruction to Chunk.Code with
//     operand = 0 (a placeholder) and returns the index of that instruction
//     in the Code slice.  The caller stores this index.
//
//  2. The compiler continues emitting subsequent instructions (the branch body,
//     the else block, etc.).  The Code slice grows; the placeholder sits in
//     the middle with the wrong address.
//
//  3. Once the target address is known — i.e., the current length of
//     Chunk.Code — the caller invokes patchJump(savedIndex).  patchJump
//     writes len(Chunk.Code) into the saved instruction's Operand field,
//     replacing the placeholder with the real target address.
//
// This two-step emit-then-patch strategy is standard in single-pass bytecode
// compilers.  Whenever you see a variable named *Jump or *Addr holding the
// return value of emitJump, you know a backpatch is coming once the target
// location is emitted.
// ---------------------------------------------------------------------------

func (c *Compiler) emit(op Opcode, operand int32, line int) int {
	return c.chunk.Emit(op, operand, line)
}

func (c *Compiler) emitJump(op Opcode, line int) int {
	c.emit(op, 0, line)
	return len(c.chunk.Code) - 1
}

func (c *Compiler) patchJump(offset int) {
	c.chunk.Code[offset].Operand = int32(len(c.chunk.Code))
}

func (c *Compiler) currentAddr() int {
	return len(c.chunk.Code)
}

func (c *Compiler) addConstant(v Value) int32 {
	return c.chunk.AddConstant(v)
}

// getVarIndex returns the constant pool index for a variable name string,
// creating a new slot if the variable has not been seen before.
//
// Variable slot allocation:
// Every distinct variable name is interned once into the constant pool as a
// string Value.  The index of that string in the pool becomes the variable's
// permanent "slot index" throughout the entire bytecode.  At runtime, OpLoad
// and OpStore use this index to look up the variable name in the pool and then
// access vm.globals[name].
//
// The varIndex map is the compile-time symbol table: it maps each upper-cased
// variable name to its constant pool slot, ensuring that every reference to,
// say, "COUNT" always produces the same slot index regardless of where in the
// source the reference appears.  This avoids re-scanning the pool on every
// reference and guarantees O(1) variable lookup both at compile time and
// at runtime.
func (c *Compiler) getVarIndex(name string) int32 {
	key := strings.ToUpper(name)
	if idx, ok := c.varIndex[key]; ok {
		return idx
	}
	idx := c.addConstant(StringVal(key))
	c.varIndex[key] = idx
	return idx
}

func (c *Compiler) addError(format string, args ...interface{}) {
	c.errors = append(c.errors, fmt.Sprintf(format, args...))
}

// ---------------------------------------------------------------------------
// Loop stack helpers
// ---------------------------------------------------------------------------

func (c *Compiler) pushLoop(loopType string, loopStart int) {
	c.loopStack = append(c.loopStack, loopInfo{
		loopType:  loopType,
		loopStart: loopStart,
	})
}

func (c *Compiler) popLoop() loopInfo {
	if len(c.loopStack) == 0 {
		return loopInfo{}
	}
	info := c.loopStack[len(c.loopStack)-1]
	c.loopStack = c.loopStack[:len(c.loopStack)-1]
	return info
}

func (c *Compiler) findLoop(loopType string) *loopInfo {
	for i := len(c.loopStack) - 1; i >= 0; i-- {
		if c.loopStack[i].loopType == loopType {
			return &c.loopStack[i]
		}
	}
	return nil
}
