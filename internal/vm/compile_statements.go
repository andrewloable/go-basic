package vm

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Statement compilation — dispatch switch (the AST visitor)
//
// compileStatement is the central dispatch function of the compiler.  It
// implements the "visitor" pattern: it accepts any AST statement node and
// routes it to the correct compile* helper based on the node's concrete type.
//
// Why a type switch rather than a method on each AST node?
// The "visitor" pattern keeps compilation logic separate from the AST data
// structures.  The AST package (internal/ast) only describes what the source
// says; the VM package decides how to translate it to bytecode.  Adding a new
// back-end (e.g., a different target architecture) only requires writing a new
// visitor, not modifying the AST.
//
// Each compile* helper is responsible for:
//   1. Compiling any sub-expressions (which push values onto the stack).
//   2. Emitting the opcodes that implement the statement's semantics.
//   3. Ensuring the stack is balanced — a statement must not leave extra
//      values on the stack when it finishes (unlike an expression, which
//      always leaves exactly one result).
//
// Statements that produce no runtime behaviour (REM, DATA, TYPE blocks) emit
// no instructions at all.  Unrecognised statement types fall through to OpNop
// so the instruction stream stays contiguous and line-number tracking remains
// accurate.
// ---------------------------------------------------------------------------

// compileStatement dispatches an AST statement node to the appropriate
// compile* helper and emits the corresponding bytecode.
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
	case *ast.ConstStatement:
		// Evaluate the constant value at compile time if possible.
		v := c.evalConstExpr(s.Value)
		c.constMap[strings.ToUpper(s.Name)] = v
		// No opcodes emitted — constants are inlined at use sites.
	case *ast.PokeStatement:
		// POKE address, value — evaluate both expressions (for side effects),
		// then emit OpPoke which is a no-op at runtime (no real memory access).
		c.compileExpression(s.Address)
		c.compileExpression(s.Value)
		c.emit(OpPoke, 0, line)
	case *ast.ClearStatement:
		// CLEAR — reset all global variables to zero.
		c.emit(OpClear, 0, line)
	case *ast.OnComputedGotoStatement:
		c.compileOnComputedBranch(s.Expr, s.Targets, OpJmp, line)
	case *ast.OnComputedGosubStatement:
		c.compileOnComputedBranch(s.Expr, s.Targets, OpGosub, line)
	case *ast.FnAssignStatement:
		// FN name = expr inside a multi-line DEF FN body.
		// Store the expression result in the function's result variable.
		c.compileExpression(s.Value)
		fnVarIdx := c.getVarIndex(s.Name)
		c.emit(OpStore, fnVarIdx, line)
	case *ast.TypeBlockStatement:
		// Store field definitions for compile-time use (validation, DIM init).
		// No bytecode is emitted — TYPE blocks are declarations only.
		c.typeDefMap[strings.ToUpper(s.Name)] = s.Fields
	case *ast.FieldAssignStatement:
		c.compileFieldAssign(s)
	case *ast.DefTypeStatement:
		// Record default type per letter range (e.g. DEFINT A-Z).
		// The VM is dynamically typed, but we store this for future use.
		c.compileDefType(s)
	case *ast.ScopeStatement:
		// SHARED/STATIC/LOCAL/COMMON — the VM uses a flat global namespace,
		// so scope modifiers are accepted but require no bytecode.
	case *ast.OptionBaseStatement:
		// OPTION BASE — affects array indexing; not enforced in VM.
		// Accepted without error.
	default:
		// Emit NOP for unimplemented statement types so the instruction
		// stream keeps correct position information.
		c.emit(OpNop, 0, line)
	}
}

// ---------------------------------------------------------------------------
// PRINT — output statement with separator-driven formatting
//
// compilePrint demonstrates the general pattern for statement compilation:
//   1. Compile each sub-expression recursively — this leaves a value on the
//      stack for each item in the PRINT list.
//   2. Emit OpPrint to pop and display that value.
//   3. Emit a formatting opcode based on the separator character that follows
//      the item in the source:
//        ','  → OpPrintTab  (advance to the next 14-character print zone)
//        ';'  → OpPrintSemicolon (stay on the same line, no extra spacing)
//   4. After the last item, emit OpPrintNewline — unless the last separator
//      was a trailing ',' or ';', which suppresses the newline.
//
// This design maps BASIC's flexible PRINT syntax to a small, predictable
// sequence of opcodes.  No runtime parsing of the PRINT list is needed; the
// compiler has already extracted the separator pattern at compile time.
//
// Example: PRINT A; B, C
//   compileExpression(A)  → stack: [valA]
//   OpPrint               → prints valA, stack: []
//   OpPrintSemicolon      → no-op (stays on line)
//   compileExpression(B)  → stack: [valB]
//   OpPrint               → prints valB, stack: []
//   OpPrintTab            → advances to next tab zone
//   compileExpression(C)  → stack: [valC]
//   OpPrint               → prints valC, stack: []
//   OpPrintNewline        → ends the line
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
// IF / ELSEIF / ELSE — conditional jump patching
//
// Conditional control flow is implemented entirely with two jump opcodes:
//   OpJmpFalse <addr>  — pops the stack and jumps if the value is falsy (0).
//   OpJmp      <addr>  — unconditional jump (used to skip ELSE branches).
//
// The bytecode layout for IF cond THEN body [ELSEIF...] [ELSE...] END IF:
//
//   [compile condition]
//   OpJmpFalse → falseJump      ← emitted with placeholder operand 0;
//                                  patched after THEN block is emitted
//   [compile THEN body]
//   OpJmp      → endJump        ← skip ELSE/ELSEIF; placeholder, patched last
//   <falseJump target: here>
//   [compile ELSEIF condition]
//   OpJmpFalse → nextFalse      ← same pattern for each ELSEIF clause
//   [compile ELSEIF body]
//   OpJmp      → endJump        ← another end-jump collected in endJumps slice
//   <nextFalse target: here>
//   [compile ELSE body]
//   <all endJumps patched to point here>
//
// The "endJumps" slice collects the index of every OpJmp that must jump past
// all branches.  After the last ELSE body is compiled we know the final
// address, so all collected endJumps are patched in one loop.
//
// Why multiple endJumps?  Each ELSEIF/ELSE clause needs its own exit jump to
// skip the remaining clauses.  Collecting them in a slice and patching them all
// at the end is simpler than computing each address eagerly.
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
// FOR / NEXT — step-direction-aware loop with backpatching
//
// FOR loops are the most complex control-flow construct to compile because the
// step value (STEP clause) can be positive or negative, which determines
// whether the exit condition is "counter > end" or "counter < end".  Since
// the step is an arbitrary expression we cannot always determine its sign at
// compile time, so the emitted bytecode checks the step sign at runtime.
//
// Emitted bytecode structure:
//
//   [counter = start]
//   [__step_var = step]
//   <loopStart: instruction index recorded here>
//   OpLoad __step_var; OpPush 0; OpGe      ← is step >= 0?
//   OpJmpFalse → negJump                   ← if negative step, go to negBranch
//   OpLoad counter; [compile end]; OpGt    ← positive branch: exit if counter > end
//   OpJmpTrue → exitPosJump                ← placeholder
//   OpJmp → bodyJump                       ← placeholder: jump past neg branch
//   <negJump patched here>
//   OpLoad counter; [compile end]; OpLt    ← negative branch: exit if counter < end
//   OpJmpTrue → exitNegJump                ← placeholder
//   <bodyJump patched here>
//   [body statements]
//   OpLoad counter; OpLoad __step_var; OpAdd; OpStore counter  ← counter += step
//   OpJmp → loopStart                      ← back-edge (no patch needed; address known)
//   <loopEnd: exitPosJump, exitNegJump, and all breakJumps patched here>
//
// EXIT FOR:
// Compiles to an OpJmp with placeholder operand 0.  The instruction index is
// appended to loopInfo.breakJumps.  After the loop body is fully emitted,
// compileFor() pops the loopInfo and patches all break jumps to loopEnd.
//
// Back-edge vs forward jump:
// The jump back to loopStart uses OpJmp with a known address (loopStart was
// recorded before any body code was emitted), so no backpatching is needed
// for the back-edge — only the exit and break jumps require it.
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
//	<loopStart:>
//	[compile condition]
//	OpJmpFalse → exitJump   (patched to loopEnd)
//	[body statements]
//	OpJmp → loopStart       (direct back-edge, no patch needed)
//	<exitJump / break jumps land here>
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

// ---------------------------------------------------------------------------
// DEFTYPE (DEFINT / DEFLNG / DEFSNG / DEFDBL / DEFSTR)
// ---------------------------------------------------------------------------

// compileDefType records the default type suffix for each letter in the ranges
// declared by a DEFINT/DEFLNG/DEFSNG/DEFDBL/DEFSTR statement.
// The VM is dynamically typed so this mostly serves documentation purposes,
// but the mapping is stored in defTypeMap for use by future tooling.
func (c *Compiler) compileDefType(s *ast.DefTypeStatement) {
	var suffix string
	switch strings.ToUpper(s.Type) {
	case "DEFINT":
		suffix = "%"
	case "DEFLNG":
		suffix = "&"
	case "DEFSNG":
		suffix = "!"
	case "DEFDBL":
		suffix = "#"
	case "DEFSTR":
		suffix = "$"
	default:
		suffix = "!"
	}
	for _, r := range s.LetterRanges {
		for ch := r.Start; ch <= r.End; ch++ {
			if ch >= 'A' && ch <= 'Z' {
				c.defTypeMap[ch-'A'] = suffix
			} else if ch >= 'a' && ch <= 'z' {
				c.defTypeMap[ch-'a'] = suffix
			}
		}
	}
}

// ---------------------------------------------------------------------------
// TYPE field assignment
// ---------------------------------------------------------------------------

// compileFieldAssign handles TYPE record field assignment: expr.field = value.
// Uses flat variable mangling: varname.field → varname_field as a distinct var.
func (c *Compiler) compileFieldAssign(s *ast.FieldAssignStatement) {
	line := s.Pos().Line
	id, ok := s.Object.(*ast.Identifier)
	if !ok {
		// Complex object (array element, etc.) — not yet supported; emit NOP.
		c.emit(OpNop, 0, line)
		return
	}
	// Mangle "varname.field" into a single flat variable "varname_field".
	flatName := (id.Name + id.TypeSuffix) + "_" + s.Field
	flatIdx := c.getVarIndex(flatName)
	c.compileExpression(s.Value)
	c.emit(OpStore, flatIdx, line)
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
