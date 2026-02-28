package vm

import (
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

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
	name := strings.ToUpper(s.Name)
	if s.IsForward {
		// DECLARE SUB: register as a known function so compileArrayAccess can
		// distinguish a SUB call from an array element access.
		c.declaredFuncs[name] = true
		return
	}
	line := s.Pos().Line

	c.declaredFuncs[name] = true
	skipJump := c.emitJump(OpJmp, line)
	c.subAddrs[name] = c.currentAddr()

	// Bind formal parameters: caller pushes args L-to-R, so TOS holds the
	// last parameter.  Pop each arg R-to-L and store into its parameter var.
	for i := len(s.Params) - 1; i >= 0; i-- {
		paramName := strings.ToUpper(s.Params[i].Name) // Name already includes type suffix (e.g. "msg$" → "MSG$")
		c.emit(OpStore, c.getVarIndex(paramName), line)
	}

	for _, stmt := range s.Body {
		c.compileStatement(stmt)
	}

	// SUBs have no meaningful return value.  Push a dummy zero so that the
	// calling LetStatement (CALL foo → LetStatement) can OpStore without
	// underflowing the stack.
	c.emit(OpPush, c.addConstant(IntVal(0)), line)
	c.emit(OpRet, 0, line)

	c.patchJump(skipJump)
}

func (c *Compiler) compileFunctionDecl(s *ast.FunctionDeclaration) {
	name := strings.ToUpper(s.Name)
	if s.IsForward {
		// DECLARE FUNCTION: register as a known function so compileArrayAccess
		// recognises Factorial(n) as a call rather than an array element access.
		c.declaredFuncs[name] = true
		return
	}
	line := s.Pos().Line

	c.declaredFuncs[name] = true
	skipJump := c.emitJump(OpJmp, line)
	c.subAddrs[name] = c.currentAddr()

	// Bind formal parameters: caller pushes args L-to-R, TOS = last param.
	// Pop each arg R-to-L into its parameter variable.
	for i := len(s.Params) - 1; i >= 0; i-- {
		paramName := strings.ToUpper(s.Params[i].Name) // Name already includes type suffix
		c.emit(OpStore, c.getVarIndex(paramName), line)
	}

	for _, stmt := range s.Body {
		c.compileStatement(stmt)
	}

	// Load the FUNCTION's return variable (set inside the body via FnAssign or
	// by assigning to the function name) and leave it on the stack for the caller.
	funcVarIdx := c.getVarIndex(s.Name)
	c.emit(OpLoad, funcVarIdx, line)
	c.emit(OpRet, 0, line)

	c.patchJump(skipJump)
}

func (c *Compiler) compileDefFn(s *ast.DefFnDeclaration) {
	// Store the declaration for inline expansion at FnCallExpression sites.
	// No bytecode is emitted here; the body is inlined when called.
	c.defFnMap[strings.ToUpper(s.Name)] = s
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

	// Fast path: both are simple scalar variables.
	// Stack trick: push both, then store in reverse order — no temp variable needed.
	if ok1 && ok2 {
		name1 := id1.Name + id1.TypeSuffix
		name2 := id2.Name + id2.TypeSuffix
		idx1 := c.getVarIndex(name1)
		idx2 := c.getVarIndex(name2)
		// stack after pushes: [val1, val2]
		// OpStore idx1 pops val2 → var1 = val2; stack: [val1]
		// OpStore idx2 pops val1 → var2 = val1; stack: []
		c.emit(OpLoad, idx1, line)
		c.emit(OpLoad, idx2, line)
		c.emit(OpStore, idx1, line)
		c.emit(OpStore, idx2, line)
		return
	}

	// General path: at least one operand is an array element.
	// Strategy:
	//   Step 1 — temp = Var1   (load Var1, store to hidden __swaptmp)
	//   Step 2 — Var1 = Var2   (push Var1 destination indices, load Var2, StoreArray/Store)
	//   Step 3 — Var2 = temp   (push Var2 destination indices, load __swaptmp, StoreArray/Store)
	//
	// For OpStoreArray the stack layout must be [idx1, idx2, …, value] — indices first,
	// value at TOS.  We achieve this by pushing the destination indices before loading
	// the source value (OpLoadArray pops the source indices and leaves the value at TOS).
	aa1, isArr1 := s.Var1.(*ast.ArrayAccess)
	aa2, isArr2 := s.Var2.(*ast.ArrayAccess)
	if (!ok1 && !isArr1) || (!ok2 && !isArr2) {
		c.addError("SWAP requires two variables at line %d", line)
		return
	}

	tempIdx := c.getVarIndex("__swaptmp")

	// Step 1: temp = Var1
	if ok1 {
		c.emit(OpLoad, c.getVarIndex(id1.Name+id1.TypeSuffix), line)
	} else {
		for _, idx := range aa1.Indices {
			c.compileExpression(idx)
		}
		c.emit(OpLoadArray, c.getVarIndex(aa1.Name+aa1.TypeSuffix), line)
	}
	c.emit(OpStore, tempIdx, line)

	// Step 2: Var1 = Var2
	// Push destination indices for Var1 (if array), then load Var2 value, then store.
	if isArr1 {
		nameIdx1 := c.getVarIndex(aa1.Name + aa1.TypeSuffix)
		for _, idx := range aa1.Indices {
			c.compileExpression(idx) // destination indices
		}
		// Now load Var2 value on top of destination indices.
		if ok2 {
			c.emit(OpLoad, c.getVarIndex(id2.Name+id2.TypeSuffix), line)
		} else {
			for _, idx := range aa2.Indices {
				c.compileExpression(idx)
			}
			c.emit(OpLoadArray, c.getVarIndex(aa2.Name+aa2.TypeSuffix), line)
		}
		c.emit(OpStoreArray, nameIdx1, line)
	} else {
		// Var1 is a simple identifier.
		if ok2 {
			c.emit(OpLoad, c.getVarIndex(id2.Name+id2.TypeSuffix), line)
		} else {
			for _, idx := range aa2.Indices {
				c.compileExpression(idx)
			}
			c.emit(OpLoadArray, c.getVarIndex(aa2.Name+aa2.TypeSuffix), line)
		}
		c.emit(OpStore, c.getVarIndex(id1.Name+id1.TypeSuffix), line)
	}

	// Step 3: Var2 = temp
	if isArr2 {
		nameIdx2 := c.getVarIndex(aa2.Name + aa2.TypeSuffix)
		for _, idx := range aa2.Indices {
			c.compileExpression(idx)
		}
		c.emit(OpLoad, tempIdx, line)
		c.emit(OpStoreArray, nameIdx2, line)
	} else {
		c.emit(OpLoad, tempIdx, line)
		c.emit(OpStore, c.getVarIndex(id2.Name+id2.TypeSuffix), line)
	}
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

// compileOnComputedBranch emits a conditional jump chain for ON expr GOTO/GOSUB.
//
// Stack layout during the chain:
//   - Before each iteration: [index]  (the original index value)
//   - Dup → [index, dup_index]
//   - Push(i) → [index, dup_index, literal_i]
//   - OpEq → consumes dup_index and literal_i, pushes bool → [index, bool]
//   - OpJmpFalse(skip) → POPS bool.  False → jump to skip, stack: [index]
//                                     True  → fall through, stack: [index]
//   - Equal path (fall through): branch to target. Stack [index] is left for
//     GOTO (never reached) or GOSUB (returned to, [index] still there).
//   - skip: next iteration starts with [index] on stack.
// After all iterations: OpPop discards [index].
func (c *Compiler) compileOnComputedBranch(expr ast.Expression, targets []string, branchOp Opcode, line int) {
	if len(targets) == 0 {
		return // no labels — nothing to do
	}

	// Push and truncate the index to an integer.
	c.compileExpression(expr)
	c.emit(OpToInt, 0, line)

	for i, label := range targets {
		idx := i + 1 // ON GOTO is 1-based
		target := strings.ToUpper(label)

		// Dup index so we can compare; OpEq will consume the dup copy.
		c.emit(OpDup, 0, line)
		// Push the 1-based position to compare against.
		valIdx := c.addConstant(IntVal(int64(idx)))
		c.emit(OpPush, valIdx, line)
		// OpEq: pops dup_index and literal_i, pushes bool. Stack: [index, bool]
		c.emit(OpEq, 0, line)
		// OpJmpFalse: POPS the bool. If false, jump past the branch.
		// If true, fall through to the branch. Stack after jump/fall: [index]
		skipIdx := c.emitJump(OpJmpFalse, line)
		// Equal path — the index is still on stack but we do not pop it here:
		// - For GOTO: we jump away; stack cleanup is not needed (program jumps).
		// - For GOSUB: we call the sub and return. The [index] remains on stack
		//   so the loop can continue after return.
		// However, for GOTO the leftover [index] is harmless (program branches away).
		// For GOSUB we intentionally leave [index] for the loop to use after return.
		if addr, ok := c.labelAddrs[target]; ok {
			c.emit(branchOp, int32(addr), line)
		} else {
			patchIdx := c.emitJump(branchOp, line)
			c.labelPatches = append(c.labelPatches, labelPatch{
				instrIndex: patchIdx,
				labelName:  label,
			})
		}
		// Patch the skip jump to land here (beginning of next iteration).
		c.patchJump(skipIdx)
	}
	// No match (or GOSUB returned): pop the index and fall through.
	c.emit(OpPop, 0, line)
}
