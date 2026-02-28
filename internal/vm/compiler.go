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
	"math"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// ---------------------------------------------------------------------------
// Compiler -- translates a Turbo BASIC AST into VM bytecode (Chunk).
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
}

// loopInfo tracks a loop context for EXIT/break handling.
type loopInfo struct {
	loopType   string // "FOR", "DO", "WHILE"
	loopStart  int    // instruction index of loop start
	breakJumps []int  // instruction indices needing patching for break
}

// labelPatch records a forward reference to a label.
type labelPatch struct {
	instrIndex int    // index in chunk.Code to patch
	labelName  string // the label being referenced
}

// subPatch records a forward reference to a SUB/FUNCTION.
type subPatch struct {
	instrIndex int    // index in chunk.Code to patch
	name       string // SUB/FUNCTION name
}

// NewCompiler creates a new bytecode compiler.  If table is nil, the compiler
// will still work but without access to semantic information.
func NewCompiler(table *semantic.SymbolTable) *Compiler {
	return &Compiler{
		chunk:      &Chunk{},
		table:      table,
		varIndex:   make(map[string]int32),
		labelAddrs: make(map[string]int),
		subAddrs:   make(map[string]int),
	}
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Compile takes a parsed BASIC program and returns the compiled bytecode.
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
// creating it if necessary.  This ensures the same variable always references
// the same constant pool entry.
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

// ---------------------------------------------------------------------------
// Statement compilation
//
// compileStatement is the heart of the AST visitor pattern for bytecode
// emission.  The compiler walks every node in the AST and, for each concrete
// statement type, calls the appropriate compile* helper.  Each helper is
// responsible for emitting the exact sequence of opcodes that implements that
// statement's semantics on the stack-based VM.
//
// The visitor approach keeps the compiler open for extension: adding support
// for a new statement type means adding one case here and writing a new
// compile* method — no existing code needs to change.
// ---------------------------------------------------------------------------

func (c *Compiler) compileStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	line := stmt.Pos().Line

	switch s := stmt.(type) {
	case *ast.PrintStatement:
		c.compilePrint(s)
	case *ast.LetStatement:
		c.compileLet(s)
	case *ast.IfStatement:
		c.compileIf(s)
	case *ast.ForStatement:
		c.compileFor(s)
	case *ast.WhileStatement:
		c.compileWhile(s)
	case *ast.DoLoopStatement:
		c.compileDoLoop(s)
	case *ast.SelectCaseStatement:
		c.compileSelectCase(s)
	case *ast.GotoStatement:
		c.compileGoto(s)
	case *ast.GosubStatement:
		c.compileGosub(s)
	case *ast.ReturnStatement:
		c.emit(OpReturn, 0, line)
	case *ast.EndStatement:
		c.emit(OpHalt, 0, line)
	case *ast.StopStatement:
		c.emit(OpHalt, 0, line)
	case *ast.SystemStatement:
		c.emit(OpHalt, 0, line)
	case *ast.LabelStatement:
		c.labelAddrs[strings.ToUpper(s.Name)] = c.currentAddr()
	case *ast.LineNumberStatement:
		label := fmt.Sprintf("%d", s.Number)
		c.labelAddrs[strings.ToUpper(label)] = c.currentAddr()
		c.emit(OpLine, int32(s.Number), line)
	case *ast.RemStatement:
		// Comments produce no bytecode.
	case *ast.DataStatement:
		// DATA is collected in pass 1; no bytecode emitted.
	case *ast.ReadStatement:
		c.compileRead(s)
	case *ast.RestoreStatement:
		c.compileRestore(s)
	case *ast.DimStatement:
		c.compileDim(s)
	case *ast.ArrayAssignment:
		c.compileArrayAssignment(s)
	case *ast.SubDeclaration:
		c.compileSubDecl(s)
	case *ast.FunctionDeclaration:
		c.compileFunctionDecl(s)
	case *ast.DefFnDeclaration:
		c.compileDefFn(s)
	case *ast.ExitStatement:
		c.compileExit(s)
	case *ast.SwapStatement:
		c.compileSwap(s)
	case *ast.IncrStatement:
		c.compileIncr(s)
	case *ast.DecrStatement:
		c.compileDecr(s)
	case *ast.RedimStatement:
		for _, d := range s.Declarations {
			c.compileDimDecl(d, line)
		}
	case *ast.DefTypeStatement,
		*ast.OptionBaseStatement,
		*ast.ScopeStatement,
		*ast.ClearStatement,
		*ast.TypeBlockStatement,
		*ast.FnAssignStatement,
		*ast.OnComputedGotoStatement,
		*ast.OnComputedGosubStatement,
		*ast.ConstStatement,
		*ast.PokeStatement,
		*ast.FieldAssignStatement:
		// Emit NOP for statement types not yet supported in VM.
		c.emit(OpNop, 0, line)
	default:
		// Emit NOP for unimplemented statement types so the instruction
		// stream keeps correct position information.
		c.emit(OpNop, 0, line)
	}
}

// ---------------------------------------------------------------------------
// PRINT
//
// compilePrint demonstrates the general pattern for statement compilation:
//   1. Recursively compile each sub-expression — this leaves a value on the
//      stack for each expression in the PRINT list.
//   2. Emit the OpPrint opcode to pop and display that value.
//   3. Emit formatting opcodes (OpPrintTab, OpPrintSemicolon) based on the
//      separator character between items.
//   4. Emit OpPrintNewline at the end unless a trailing separator suppresses it.
//
// No explicit control flow is needed here because PRINT is a straight-line
// statement; every instruction executes in sequence.
// ---------------------------------------------------------------------------

func (c *Compiler) compilePrint(s *ast.PrintStatement) {
	line := s.Pos().Line

	if len(s.Expressions) == 0 {
		// Bare PRINT emits just a newline.
		c.emit(OpPrintNewline, 0, line)
		return
	}

	for i, expr := range s.Expressions {
		c.compileExpression(expr)
		c.emit(OpPrint, 0, line)

		if i < len(s.Separators) {
			switch s.Separators[i] {
			case ",":
				c.emit(OpPrintTab, 0, line)
			case ";":
				c.emit(OpPrintSemicolon, 0, line)
			}
		}
	}

	if !s.HasTrailingSep {
		c.emit(OpPrintNewline, 0, line)
	}
}

// ---------------------------------------------------------------------------
// LET (assignment)
// ---------------------------------------------------------------------------

func (c *Compiler) compileLet(s *ast.LetStatement) {
	line := s.Pos().Line
	c.compileExpression(s.Value)
	name := s.Name.Name + s.Name.TypeSuffix
	idx := c.getVarIndex(name)
	c.emit(OpStore, idx, line)
}

// ---------------------------------------------------------------------------
// IF / ELSEIF / ELSE
//
// Conditional control flow is implemented with two jump opcodes:
//   - OpJmpFalse <addr>  — pops the stack; jumps if the value is falsy.
//   - OpJmp      <addr>  — unconditional jump (used to skip the ELSE branch).
//
// The compilation strategy for IF/ELSEIF/ELSE follows this pattern:
//
//   [compile condition]
//   OpJmpFalse → falseJump  (placeholder; patched after THEN block is emitted)
//   [compile THEN body]
//   OpJmp      → endJump    (placeholder; patched after all branches are emitted)
//   <falseJump lands here>
//   [compile ELSEIF / ELSE bodies, each with their own falseJump / endJump]
//   <all endJumps land here>
//
// Because the THEN block's length is unknown when we first emit OpJmpFalse,
// we use backpatching: emit the jump with operand 0 and record the index so
// we can fill in the real target once we know where the next section starts.
// ---------------------------------------------------------------------------

func (c *Compiler) compileIf(s *ast.IfStatement) {
	line := s.Pos().Line

	// Compile the condition.
	c.compileExpression(s.Condition)

	// Jump past the THEN block when false.
	falseJump := c.emitJump(OpJmpFalse, line)

	// Compile the THEN block.
	for _, stmt := range s.ThenBlock {
		c.compileStatement(stmt)
	}

	if len(s.ElseIfClauses) == 0 && len(s.ElseBlock) == 0 {
		c.patchJump(falseJump)
		return
	}

	// At end of THEN block, jump past all ELSEIF/ELSE blocks.
	endJumps := []int{c.emitJump(OpJmp, line)}
	c.patchJump(falseJump)

	// Compile ELSEIF clauses.
	for _, clause := range s.ElseIfClauses {
		clauseLine := clause.BasePos.Line
		c.compileExpression(clause.Condition)
		nextFalse := c.emitJump(OpJmpFalse, clauseLine)

		for _, stmt := range clause.Body {
			c.compileStatement(stmt)
		}

		endJumps = append(endJumps, c.emitJump(OpJmp, clauseLine))
		c.patchJump(nextFalse)
	}

	// Compile ELSE block.
	for _, stmt := range s.ElseBlock {
		c.compileStatement(stmt)
	}

	// Patch all end-jumps to the current position.
	for _, j := range endJumps {
		c.patchJump(j)
	}
}

// ---------------------------------------------------------------------------
// FOR / NEXT
//
// FOR loops are the most complex control-flow construct because the step value
// can be positive or negative, which determines the loop-exit direction.  The
// emitted bytecode structure is:
//
//   [init: counter = start; hidden __step_var = step]
//   <loopStart:>
//   [push step; push 0; OpGe]      ← is step >= 0?
//   OpJmpFalse → negJump
//   [push counter; push end; OpGt] ← positive step: exit if counter > end
//   OpJmpTrue  → exitPosJump
//   OpJmp      → bodyJump          ← fall through to body
//   <negJump lands here>
//   [push counter; push end; OpLt] ← negative step: exit if counter < end
//   OpJmpTrue  → exitNegJump
//   <bodyJump lands here>
//   [body statements]
//   [counter = counter + step]
//   OpJmp → loopStart              ← back-edge (known address, no patch needed)
//   <exitPosJump / exitNegJump / break jumps all land here>
//
// EXIT FOR works by emitting an OpJmp with a placeholder operand and storing
// the instruction's index in loopInfo.breakJumps.  All break jumps are
// backpatched to loopEnd after the NEXT is processed.
// ---------------------------------------------------------------------------

func (c *Compiler) compileFor(s *ast.ForStatement) {
	line := s.Pos().Line
	varName := s.Counter.Name + s.Counter.TypeSuffix
	varIdx := c.getVarIndex(varName)

	// counter = start
	c.compileExpression(s.Start)
	c.emit(OpStore, varIdx, line)

	// Store step in a hidden variable so the condition check can use it.
	stepVarIdx := c.getVarIndex("__step_" + varName)
	if s.Step != nil {
		c.compileExpression(s.Step)
	} else {
		c.emit(OpPush, c.addConstant(IntVal(1)), line)
	}
	c.emit(OpStore, stepVarIdx, line)

	// Loop start.
	loopStart := c.currentAddr()
	c.pushLoop("FOR", loopStart)

	// Condition check using the step sign.
	//   if step >= 0: exit when counter > end
	//   if step <  0: exit when counter < end
	zeroConst := c.addConstant(IntVal(0))

	c.emit(OpLoad, stepVarIdx, line)
	c.emit(OpPush, zeroConst, line)
	c.emit(OpGe, 0, line) // step >= 0 ?
	negJump := c.emitJump(OpJmpFalse, line)

	// Positive step: counter > end => exit
	c.emit(OpLoad, varIdx, line)
	c.compileExpression(s.End)
	c.emit(OpGt, 0, line)
	exitPosJump := c.emitJump(OpJmpTrue, line)
	bodyJump := c.emitJump(OpJmp, line) // jump to body

	// Negative step: counter < end => exit
	c.patchJump(negJump)
	c.emit(OpLoad, varIdx, line)
	c.compileExpression(s.End)
	c.emit(OpLt, 0, line)
	exitNegJump := c.emitJump(OpJmpTrue, line)

	// Body.
	c.patchJump(bodyJump)
	for _, stmt := range s.Body {
		c.compileStatement(stmt)
	}

	// Increment: counter = counter + step
	c.emit(OpLoad, varIdx, line)
	c.emit(OpLoad, stepVarIdx, line)
	c.emit(OpAdd, 0, line)
	c.emit(OpStore, varIdx, line)

	// Jump back to condition check.
	c.emit(OpJmp, int32(loopStart), line)

	// Patch exit jumps.
	loopEnd := c.currentAddr()
	c.chunk.Code[exitPosJump].Operand = int32(loopEnd)
	c.chunk.Code[exitNegJump].Operand = int32(loopEnd)

	// Patch break jumps.
	info := c.popLoop()
	for _, bj := range info.breakJumps {
		c.chunk.Code[bj].Operand = int32(loopEnd)
	}
}

// ---------------------------------------------------------------------------
// WHILE / WEND
//
// WHILE is the simplest loop: emit the condition test at the top, jump out on
// false, run the body, then jump unconditionally back to the condition test.
// The back-edge target (loopStart) is already known when we emit OpJmp, so no
// backpatching is required for that jump.  Only the exit jump (OpJmpFalse) and
// any EXIT WHILE jumps need patching, which happens after the body is emitted.
//
//   <loopStart:>
//   [compile condition]
//   OpJmpFalse → exitJump   (patched to loopEnd)
//   [body statements]
//   OpJmp → loopStart       (direct back-edge, no patch needed)
//   <exitJump / break jumps land here>
// ---------------------------------------------------------------------------

func (c *Compiler) compileWhile(s *ast.WhileStatement) {
	line := s.Pos().Line

	loopStart := c.currentAddr()
	c.pushLoop("WHILE", loopStart)

	c.compileExpression(s.Condition)
	exitJump := c.emitJump(OpJmpFalse, line)

	for _, stmt := range s.Body {
		c.compileStatement(stmt)
	}

	c.emit(OpJmp, int32(loopStart), line)
	c.patchJump(exitJump)

	info := c.popLoop()
	loopEnd := c.currentAddr()
	for _, bj := range info.breakJumps {
		c.chunk.Code[bj].Operand = int32(loopEnd)
	}
}

// ---------------------------------------------------------------------------
// DO / LOOP
//
// DO loops support four variants that share the same compilation skeleton:
//   - DO WHILE <cond> … LOOP  (test at top, exit when false)
//   - DO UNTIL <cond> … LOOP  (test at top, exit when true)
//   - DO … LOOP WHILE <cond>  (test at bottom, exit when false)
//   - DO … LOOP UNTIL <cond>  (test at bottom, exit when true)
//   - DO … LOOP               (infinite; requires EXIT DO to escape)
//
// For top-tested loops a forward exit jump is emitted before the body and
// must be backpatched once the loop end is known.  For bottom-tested loops the
// back-edge conditional jump is emitted after the body; since loopStart is
// already recorded its address is written directly into the operand without
// needing a patch list.
// ---------------------------------------------------------------------------

func (c *Compiler) compileDoLoop(s *ast.DoLoopStatement) {
	line := s.Pos().Line

	loopStart := c.currentAddr()
	c.pushLoop("DO", loopStart)

	exitJump := -1

	if s.TestAtTop && s.Condition != nil {
		c.compileExpression(s.Condition)
		if s.IsUntil {
			exitJump = c.emitJump(OpJmpTrue, line)
		} else {
			exitJump = c.emitJump(OpJmpFalse, line)
		}
	}

	for _, stmt := range s.Body {
		c.compileStatement(stmt)
	}

	if !s.TestAtTop && s.Condition != nil {
		c.compileExpression(s.Condition)
		if s.IsUntil {
			// LOOP UNTIL: continue looping when condition is false.
			jmp := c.emitJump(OpJmpFalse, line)
			c.chunk.Code[jmp].Operand = int32(loopStart)
		} else {
			// LOOP WHILE: continue looping when condition is true.
			jmp := c.emitJump(OpJmpTrue, line)
			c.chunk.Code[jmp].Operand = int32(loopStart)
		}
	} else if s.Condition == nil {
		// Infinite DO ... LOOP
		c.emit(OpJmp, int32(loopStart), line)
	} else {
		// TestAtTop: jump back to start.
		c.emit(OpJmp, int32(loopStart), line)
	}

	if exitJump >= 0 {
		c.patchJump(exitJump)
	}

	info := c.popLoop()
	loopEnd := c.currentAddr()
	for _, bj := range info.breakJumps {
		c.chunk.Code[bj].Operand = int32(loopEnd)
	}
}

// ---------------------------------------------------------------------------
// SELECT CASE
// ---------------------------------------------------------------------------

func (c *Compiler) compileSelectCase(s *ast.SelectCaseStatement) {
	line := s.Pos().Line

	// Evaluate the test expression and store in a temp variable.
	c.compileExpression(s.TestExpr)
	tempIdx := c.getVarIndex("__select_temp")
	c.emit(OpStore, tempIdx, line)

	var endJumps []int

	for _, cas := range s.Cases {
		caseLine := cas.BasePos.Line
		var trueJumps []int

		for _, cv := range cas.Values {
			if cv.IsRange {
				// CASE val1 TO val2
				c.emit(OpLoad, tempIdx, caseLine)
				c.compileExpression(cv.Value)
				c.emit(OpGe, 0, caseLine)
				lowerFail := c.emitJump(OpJmpFalse, caseLine)

				c.emit(OpLoad, tempIdx, caseLine)
				c.compileExpression(cv.EndValue)
				c.emit(OpLe, 0, caseLine)
				trueJumps = append(trueJumps, c.emitJump(OpJmpTrue, caseLine))

				c.patchJump(lowerFail)
			} else if cv.IsComparison {
				// CASE IS op val
				c.emit(OpLoad, tempIdx, caseLine)
				c.compileExpression(cv.Value)
				c.emitComparisonOp(cv.Comparison, caseLine)
				trueJumps = append(trueJumps, c.emitJump(OpJmpTrue, caseLine))
			} else {
				// CASE val
				c.emit(OpLoad, tempIdx, caseLine)
				c.compileExpression(cv.Value)
				c.emit(OpEq, 0, caseLine)
				trueJumps = append(trueJumps, c.emitJump(OpJmpTrue, caseLine))
			}
		}

		// None matched -- jump to next case.
		nextCaseJump := c.emitJump(OpJmp, caseLine)

		// Patch true jumps to the body.
		for _, tj := range trueJumps {
			c.patchJump(tj)
		}

		for _, stmt := range cas.Body {
			c.compileStatement(stmt)
		}

		endJumps = append(endJumps, c.emitJump(OpJmp, caseLine))
		c.patchJump(nextCaseJump)
	}

	// CASE ELSE.
	for _, stmt := range s.ElseBlock {
		c.compileStatement(stmt)
	}

	for _, ej := range endJumps {
		c.patchJump(ej)
	}
}

func (c *Compiler) emitComparisonOp(op string, line int) {
	switch strings.TrimSpace(op) {
	case "=":
		c.emit(OpEq, 0, line)
	case "<>", "><":
		c.emit(OpNe, 0, line)
	case "<":
		c.emit(OpLt, 0, line)
	case ">":
		c.emit(OpGt, 0, line)
	case "<=", "=<":
		c.emit(OpLe, 0, line)
	case ">=", "=>":
		c.emit(OpGe, 0, line)
	default:
		c.emit(OpEq, 0, line)
	}
}

// ---------------------------------------------------------------------------
// GOTO / GOSUB
// ---------------------------------------------------------------------------

func (c *Compiler) compileGoto(s *ast.GotoStatement) {
	line := s.Pos().Line
	target := strings.ToUpper(s.Target)

	if addr, ok := c.labelAddrs[target]; ok {
		c.emit(OpJmp, int32(addr), line)
	} else {
		idx := c.emitJump(OpJmp, line)
		c.labelPatches = append(c.labelPatches, labelPatch{
			instrIndex: idx,
			labelName:  s.Target,
		})
	}
}

func (c *Compiler) compileGosub(s *ast.GosubStatement) {
	line := s.Pos().Line
	target := strings.ToUpper(s.Target)

	if addr, ok := c.labelAddrs[target]; ok {
		c.emit(OpGosub, int32(addr), line)
	} else {
		idx := c.emitJump(OpGosub, line)
		c.labelPatches = append(c.labelPatches, labelPatch{
			instrIndex: idx,
			labelName:  s.Target,
		})
	}
}

// ---------------------------------------------------------------------------
// READ / RESTORE
// ---------------------------------------------------------------------------

func (c *Compiler) compileRead(s *ast.ReadStatement) {
	line := s.Pos().Line
	for _, v := range s.Variables {
		switch target := v.(type) {
		case *ast.Identifier:
			name := target.Name + target.TypeSuffix
			idx := c.getVarIndex(name)
			c.emit(OpRead, idx, line)
		default:
			c.addError("READ target must be a variable at line %d", line)
		}
	}
}

func (c *Compiler) compileRestore(s *ast.RestoreStatement) {
	line := s.Pos().Line
	c.emit(OpRestore, -1, line)
}

// ---------------------------------------------------------------------------
// DIM
// ---------------------------------------------------------------------------

func (c *Compiler) compileDim(s *ast.DimStatement) {
	line := s.Pos().Line
	for _, d := range s.Declarations {
		c.compileDimDecl(d, line)
	}
}

func (c *Compiler) compileDimDecl(d ast.DimDecl, line int) {
	if len(d.Dimensions) == 0 {
		return
	}

	for _, dim := range d.Dimensions {
		if dim.Upper != nil {
			c.compileExpression(dim.Upper)
		} else {
			c.emit(OpPush, c.addConstant(IntVal(10)), line)
		}
	}

	name := d.Name + d.TypeSuffix
	idx := c.getVarIndex(name)
	c.emit(OpPush, idx, line)
	c.emit(OpDimArray, int32(len(d.Dimensions)), line)
}

// ---------------------------------------------------------------------------
// Array assignment
// ---------------------------------------------------------------------------

func (c *Compiler) compileArrayAssignment(s *ast.ArrayAssignment) {
	line := s.Pos().Line
	name := s.Array.Name + s.Array.TypeSuffix
	nameIdx := c.getVarIndex(name)

	for _, idx := range s.Array.Indices {
		c.compileExpression(idx)
	}

	c.compileExpression(s.Value)
	c.emit(OpStoreArray, nameIdx, line)
}

// ---------------------------------------------------------------------------
// SUB / FUNCTION / DEF FN declarations
// ---------------------------------------------------------------------------

func (c *Compiler) compileSubDecl(s *ast.SubDeclaration) {
	if s.IsForward {
		return
	}
	line := s.Pos().Line

	skipJump := c.emitJump(OpJmp, line)
	c.subAddrs[strings.ToUpper(s.Name)] = c.currentAddr()

	for _, stmt := range s.Body {
		c.compileStatement(stmt)
	}
	c.emit(OpRet, 0, line)

	c.patchJump(skipJump)
}

func (c *Compiler) compileFunctionDecl(s *ast.FunctionDeclaration) {
	if s.IsForward {
		return
	}
	line := s.Pos().Line

	skipJump := c.emitJump(OpJmp, line)
	c.subAddrs[strings.ToUpper(s.Name)] = c.currentAddr()

	for _, stmt := range s.Body {
		c.compileStatement(stmt)
	}

	funcVarIdx := c.getVarIndex(s.Name)
	c.emit(OpLoad, funcVarIdx, line)
	c.emit(OpRet, 0, line)

	c.patchJump(skipJump)
}

func (c *Compiler) compileDefFn(s *ast.DefFnDeclaration) {
	line := s.Pos().Line

	skipJump := c.emitJump(OpJmp, line)
	c.subAddrs[strings.ToUpper(s.Name)] = c.currentAddr()

	if s.SingleLineExpr != nil {
		c.compileExpression(s.SingleLineExpr)
		fnVarIdx := c.getVarIndex(s.Name)
		c.emit(OpStore, fnVarIdx, line)
	} else {
		for _, stmt := range s.Body {
			c.compileStatement(stmt)
		}
	}

	fnVarIdx := c.getVarIndex(s.Name)
	c.emit(OpLoad, fnVarIdx, line)
	c.emit(OpRet, 0, line)

	c.patchJump(skipJump)
}

// ---------------------------------------------------------------------------
// EXIT
// ---------------------------------------------------------------------------

func (c *Compiler) compileExit(s *ast.ExitStatement) {
	line := s.Pos().Line

	switch strings.ToUpper(s.ExitType) {
	case "FOR":
		if info := c.findLoop("FOR"); info != nil {
			idx := c.emitJump(OpJmp, line)
			info.breakJumps = append(info.breakJumps, idx)
		} else {
			c.addError("EXIT FOR without FOR at line %d", line)
		}
	case "DO":
		if info := c.findLoop("DO"); info != nil {
			idx := c.emitJump(OpJmp, line)
			info.breakJumps = append(info.breakJumps, idx)
		} else {
			c.addError("EXIT DO without DO at line %d", line)
		}
	case "WHILE":
		if info := c.findLoop("WHILE"); info != nil {
			idx := c.emitJump(OpJmp, line)
			info.breakJumps = append(info.breakJumps, idx)
		} else {
			c.addError("EXIT WHILE without WHILE at line %d", line)
		}
	case "SUB", "FUNCTION", "DEF":
		c.emit(OpRet, 0, line)
	}
}

// ---------------------------------------------------------------------------
// SWAP
// ---------------------------------------------------------------------------

func (c *Compiler) compileSwap(s *ast.SwapStatement) {
	line := s.Pos().Line

	id1, ok1 := s.Var1.(*ast.Identifier)
	id2, ok2 := s.Var2.(*ast.Identifier)
	if !ok1 || !ok2 {
		c.addError("SWAP requires two variables at line %d", line)
		return
	}

	name1 := id1.Name + id1.TypeSuffix
	name2 := id2.Name + id2.TypeSuffix
	idx1 := c.getVarIndex(name1)
	idx2 := c.getVarIndex(name2)

	// temp = var1; var1 = var2; var2 = temp
	c.emit(OpLoad, idx1, line)
	c.emit(OpLoad, idx2, line)
	c.emit(OpStore, idx1, line)
	c.emit(OpStore, idx2, line)
}

// ---------------------------------------------------------------------------
// INCR / DECR
// ---------------------------------------------------------------------------

func (c *Compiler) compileIncr(s *ast.IncrStatement) {
	line := s.Pos().Line
	id, ok := s.Variable.(*ast.Identifier)
	if !ok {
		c.addError("INCR requires a variable at line %d", line)
		return
	}
	name := id.Name + id.TypeSuffix
	idx := c.getVarIndex(name)

	c.emit(OpLoad, idx, line)
	if s.Amount != nil {
		c.compileExpression(s.Amount)
	} else {
		c.emit(OpPush, c.addConstant(IntVal(1)), line)
	}
	c.emit(OpAdd, 0, line)
	c.emit(OpStore, idx, line)
}

func (c *Compiler) compileDecr(s *ast.DecrStatement) {
	line := s.Pos().Line
	id, ok := s.Variable.(*ast.Identifier)
	if !ok {
		c.addError("DECR requires a variable at line %d", line)
		return
	}
	name := id.Name + id.TypeSuffix
	idx := c.getVarIndex(name)

	c.emit(OpLoad, idx, line)
	if s.Amount != nil {
		c.compileExpression(s.Amount)
	} else {
		c.emit(OpPush, c.addConstant(IntVal(1)), line)
	}
	c.emit(OpSub, 0, line)
	c.emit(OpStore, idx, line)
}

// ---------------------------------------------------------------------------
// Expression compilation
//
// Every expression in a stack-based VM is compiled into a sequence of
// "push / operate" instructions that leaves exactly one value on the stack:
//
//   Literal      → OpPush <constant-pool-index>
//   Variable     → OpLoad <variable-index>
//   a + b        → [compile a] [compile b] OpAdd
//   f(x)         → [compile x] OpBuiltin <id>   (or OpCall for user functions)
//   (a + b) * c  → [compile a] [compile b] OpAdd [compile c] OpMul
//
// Because expressions are compiled recursively and each sub-expression leaves
// its result on the stack, the operands for any operator are always sitting
// on top of the stack in the correct order when the operator's opcode is
// reached.  This is why operand stacks and recursive descent compile so
// naturally together.
// ---------------------------------------------------------------------------

func (c *Compiler) compileExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.NumberLiteral:
		c.compileNumberLiteral(e)
	case *ast.StringLiteral:
		idx := c.addConstant(StringVal(e.Value))
		c.emit(OpPush, idx, e.Pos().Line)
	case *ast.Identifier:
		name := e.Name + e.TypeSuffix
		idx := c.getVarIndex(name)
		c.emit(OpLoad, idx, e.Pos().Line)
	case *ast.BinaryExpr:
		c.compileBinaryExpr(e)
	case *ast.UnaryExpr:
		c.compileUnaryExpr(e)
	case *ast.GroupExpr:
		c.compileExpression(e.Inner)
	case *ast.FunctionCall:
		c.compileFunctionCall(e)
	case *ast.ArrayAccess:
		c.compileArrayAccess(e)
	case *ast.FnCallExpression:
		// DEF FN call — emit a NOP placeholder (not yet supported in VM)
		c.emit(OpNop, 0, e.Pos().Line)
		idx := c.addConstant(FloatVal(0))
		c.emit(OpPush, idx, e.Pos().Line)
	case *ast.FieldAccessExpression:
		// TYPE field access — emit a NOP placeholder (not yet supported in VM)
		c.emit(OpNop, 0, e.Pos().Line)
		idx := c.addConstant(FloatVal(0))
		c.emit(OpPush, idx, e.Pos().Line)
	default:
		c.addError("unsupported expression type %T", expr)
	}
}

func (c *Compiler) compileNumberLiteral(n *ast.NumberLiteral) {
	var v Value
	if n.NumType == ast.NumInt || n.NumType == ast.NumLong {
		v = IntVal(int64(n.Value))
	} else {
		v = FloatVal(n.Value)
	}
	idx := c.addConstant(v)
	c.emit(OpPush, idx, n.Pos().Line)
}

// compileBinaryExpr compiles a binary (two-operand) expression.
//
// Operator dispatch: after both operands are on the stack the operator string
// is mapped to a single opcode.  Every arithmetic, relational, and logical
// operator in BASIC has a dedicated opcode (OpAdd, OpEq, OpAnd, …) so the VM
// can implement each operation in a tight native switch with no string
// comparisons at runtime — all the string-to-opcode translation happens here,
// once, at compile time.
//
// The compilation order — left then right — ensures that when the VM pops the
// two operands the left-hand value was pushed first, which matters for
// non-commutative operators like subtraction and division.
func (c *Compiler) compileBinaryExpr(e *ast.BinaryExpr) {
	line := e.Pos().Line

	c.compileExpression(e.Left)
	c.compileExpression(e.Right)

	switch strings.ToUpper(e.Operator) {
	case "+":
		c.emit(OpAdd, 0, line)
	case "-":
		c.emit(OpSub, 0, line)
	case "*":
		c.emit(OpMul, 0, line)
	case "/":
		c.emit(OpDiv, 0, line)
	case "\\":
		c.emit(OpIDiv, 0, line)
	case "MOD":
		c.emit(OpMod, 0, line)
	case "^":
		c.emit(OpPow, 0, line)
	case "=":
		c.emit(OpEq, 0, line)
	case "<>", "><":
		c.emit(OpNe, 0, line)
	case "<":
		c.emit(OpLt, 0, line)
	case ">":
		c.emit(OpGt, 0, line)
	case "<=", "=<":
		c.emit(OpLe, 0, line)
	case ">=", "=>":
		c.emit(OpGe, 0, line)
	case "AND":
		c.emit(OpAnd, 0, line)
	case "OR":
		c.emit(OpOr, 0, line)
	case "XOR":
		c.emit(OpXor, 0, line)
	case "EQV":
		c.emit(OpEqv, 0, line)
	case "IMP":
		c.emit(OpImp, 0, line)
	default:
		c.addError("unsupported binary operator %q at line %d", e.Operator, line)
	}
}

func (c *Compiler) compileUnaryExpr(e *ast.UnaryExpr) {
	line := e.Pos().Line

	// Optimization: fold constant negation.
	if e.Operator == "-" {
		if num, ok := e.Operand.(*ast.NumberLiteral); ok {
			var v Value
			if num.NumType == ast.NumInt || num.NumType == ast.NumLong {
				v = IntVal(-int64(num.Value))
			} else {
				v = FloatVal(-num.Value)
			}
			idx := c.addConstant(v)
			c.emit(OpPush, idx, line)
			return
		}
	}

	c.compileExpression(e.Operand)

	switch strings.ToUpper(e.Operator) {
	case "-":
		c.emit(OpNeg, 0, line)
	case "NOT":
		c.emit(OpNot, 0, line)
	default:
		c.addError("unsupported unary operator %q at line %d", e.Operator, line)
	}
}

// ---------------------------------------------------------------------------
// Function calls (built-in and user-defined)
// ---------------------------------------------------------------------------

// builtinMap maps upper-cased function names to BuiltinID.
var builtinMap = map[string]BuiltinID{
	"ABS":     BuiltinAbs,
	"SGN":     BuiltinSgn,
	"INT":     BuiltinInt,
	"FIX":     BuiltinFix,
	"CEIL":    BuiltinCeil,
	"SQR":     BuiltinSqr,
	"EXP":     BuiltinExp,
	"EXP2":    BuiltinExp2,
	"EXP10":   BuiltinExp10,
	"LOG":     BuiltinLog,
	"LOG2":    BuiltinLog2,
	"LOG10":   BuiltinLog10,
	"SIN":     BuiltinSin,
	"COS":     BuiltinCos,
	"TAN":     BuiltinTan,
	"ATN":     BuiltinAtn,
	"CINT":    BuiltinCint,
	"CLNG":    BuiltinClng,
	"CSNG":    BuiltinCsng,
	"CDBL":    BuiltinCdbl,
	"RND":     BuiltinRnd,
	"LEFT$":   BuiltinLeft,
	"RIGHT$":  BuiltinRight,
	"MID$":    BuiltinMid,
	"LEN":     BuiltinLen,
	"ASC":     BuiltinAsc,
	"CHR$":    BuiltinChr,
	"STR$":    BuiltinStr,
	"VAL":     BuiltinVal,
	"INSTR":   BuiltinInstr,
	"UCASE$":  BuiltinUCase,
	"LCASE$":  BuiltinLCase,
	"LTRIM$":  BuiltinLTrim,
	"RTRIM$":  BuiltinRTrim,
	"TRIM$":   BuiltinTrim,
	"SPACE$":  BuiltinSpace,
	"STRING$": BuiltinString,
	"HEX$":    BuiltinHex,
	"OCT$":    BuiltinOct,
	"BIN$":    BuiltinBin,
	"MKI$":    BuiltinMki,
	"MKL$":    BuiltinMkl,
	"MKS$":    BuiltinMks,
	"MKD$":    BuiltinMkd,
	"CVI":     BuiltinCvi,
	"CVL":     BuiltinCvl,
	"CVS":     BuiltinCvs,
	"CVD":     BuiltinCvd,
	"TAB":     BuiltinTab,
	"SPC":     BuiltinSpc,
}

// compileFunctionCall compiles a function-call expression.
//
// Built-in functions (ABS, SQR, LEFT$, …) are resolved at compile time via
// builtinMap.  Each built-in is assigned a numeric BuiltinID.  The call
// compiles to:
//   [compile each argument in order]   ← args land on stack left-to-right
//   OpBuiltin <BuiltinID>              ← VM pops args, runs native Go code,
//                                         pushes the result
//
// Using a numeric ID instead of a string name means the VM's dispatch loop
// only needs an integer switch — no hash lookups at runtime.
//
// User-defined functions (SUB/FUNCTION) compile to OpCall <addr>, where
// <addr> is the instruction index of the function body.  If the function has
// not been declared yet (forward reference) the address is recorded as a
// subPatch and backpatched after all declarations are compiled.
func (c *Compiler) compileFunctionCall(fc *ast.FunctionCall) {
	line := fc.Pos().Line
	name := strings.ToUpper(fc.Name)

	// Check if this is a built-in function.
	if id, ok := builtinMap[name]; ok {
		// INSTR can take 2 or 3 args: INSTR(s$, find$) or INSTR(start%, s$, find$).
		// The VM always pops 3, so push a default start=1 when only 2 args given.
		if id == BuiltinInstr && len(fc.Args) == 2 {
			c.emit(OpPush, c.addConstant(IntVal(1)), line)
		}
		for _, arg := range fc.Args {
			c.compileExpression(arg)
		}
		c.emit(OpBuiltin, int32(id), line)
		return
	}

	// User-defined FUNCTION or SUB call.
	for _, arg := range fc.Args {
		c.compileExpression(arg)
	}

	if addr, ok := c.subAddrs[name]; ok {
		c.emit(OpCall, int32(addr), line)
	} else {
		idx := c.emitJump(OpCall, line)
		c.subPatches = append(c.subPatches, subPatch{
			instrIndex: idx,
			name:       fc.Name,
		})
	}
}

func (c *Compiler) compileArrayAccess(aa *ast.ArrayAccess) {
	line := aa.Pos().Line
	name := aa.Name + aa.TypeSuffix
	nameIdx := c.getVarIndex(name)

	for _, idx := range aa.Indices {
		c.compileExpression(idx)
	}

	c.emit(OpLoadArray, nameIdx, line)
}

// ---------------------------------------------------------------------------
// Utility
// ---------------------------------------------------------------------------

// isWholeNumber reports whether f has no fractional part.
func isWholeNumber(f float64) bool {
	return f == math.Trunc(f) && !math.IsInf(f, 0) && !math.IsNaN(f)
}
