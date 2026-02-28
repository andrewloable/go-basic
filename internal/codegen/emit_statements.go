package codegen

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Statement emitter – visitor pattern for statement nodes
//
// emitStatement() is the central dispatch point for the statement half of the
// visitor.  It implements the Visitor pattern without a separate Visitor
// interface: instead of calling stmt.Accept(visitor), we perform a Go
// type-switch directly on the AST node.  The result is the same — each node
// type is handled by a dedicated method — but with less boilerplate.
//
// Every branch of the switch corresponds to one BASIC statement kind.  If a
// statement kind has non-trivial output it delegates to a dedicated emit*()
// method; trivial cases (END, BEEP, …) are inlined in the switch.
//
// Statements that have no meaningful Go equivalent (POKE, CLEAR, ON ERROR,
// unsupported file I/O) emit a comment or a no-op so the generated file still
// compiles.  The "default" branch at the bottom is a safety net that emits a
// TODO comment for any node type that was added to the parser but not yet
// handled here.
//
// TUTORIAL — The dispatch pattern
//
// In classical object-oriented compiler design, code generation uses the
// Visitor design pattern: each AST node type has an Accept(visitor) method,
// and the visitor implements a separate visit() method for every node type.
// This keeps the AST nodes and the code generator decoupled.
//
// Go's type-switch offers a simpler alternative that achieves the same
// dispatch without any interface plumbing. The pattern is:
//
//   switch s := stmt.(type) {
//   case *ast.PrintStatement:  g.emitPrint(s)
//   case *ast.ForStatement:    g.emitFor(s)
//   ...
//   }
//
// The compiler checks that all cases are handled (via the default fallthrough),
// and each case gives us a typed pointer to the concrete node — no reflection
// or unsafe casts needed. This is idiomatic Go and entirely equivalent to the
// classic visitor for a single-traversal code generator.
//
// When adding support for a new BASIC statement:
//  1. Add an AST node type in internal/ast (Phase 2 output).
//  2. Add a case in this switch pointing to a new emitXxx() method.
//  3. Implement emitXxx() to write the Go equivalent to g.buf.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.PrintStatement:
		g.emitPrint(s)
	case *ast.LetStatement:
		g.emitLet(s)
	case *ast.RandomizeStatement:
		g.emitRandomize(s)
	case *ast.ArrayAssignment:
		g.emitArrayAssignment(s)
	case *ast.IfStatement:
		g.emitIf(s)
	case *ast.ForStatement:
		g.emitFor(s)
	case *ast.WhileStatement:
		g.emitWhile(s)
	case *ast.DoLoopStatement:
		g.emitDoLoop(s)
	case *ast.SelectCaseStatement:
		g.emitSelectCase(s)
	case *ast.GotoStatement:
		g.emitGoto(s)
	case *ast.GosubStatement:
		g.emitGosub(s)
	case *ast.ReturnStatement:
		g.emitReturn(s)
	case *ast.DimStatement:
		g.emitDim(s)
	case *ast.RedimStatement:
		g.emitRedim(s)
	case *ast.DataStatement:
		// Already collected in pre-pass; emit a comment.
		g.writeLine("// DATA (values collected into dataPool)")
	case *ast.ReadStatement:
		g.emitRead(s)
	case *ast.RestoreStatement:
		g.emitRestore(s)
	case *ast.EndStatement:
		g.imports["os"] = true
		g.writeLine("os.Exit(0)")
	case *ast.StopStatement:
		g.imports["os"] = true
		g.writeLine("os.Exit(0) // STOP")
	case *ast.SystemStatement:
		g.imports["os"] = true
		g.writeLine("os.Exit(0) // SYSTEM")
	case *ast.LineNumberStatement:
		g.emitLineNumber(s)
	case *ast.LabelStatement:
		g.emitLabel(s)
	case *ast.RemStatement:
		g.emitRem(s)
	case *ast.ExitStatement:
		g.emitExit(s)
	case *ast.SwapStatement:
		g.emitSwap(s)
	case *ast.IncrStatement:
		g.emitIncr(s)
	case *ast.DecrStatement:
		g.emitDecr(s)
	case *ast.OptionBaseStatement:
		g.writeLinef("// OPTION BASE %d", s.Value)
	case *ast.DefTypeStatement:
		g.writeLinef("// %s (handled at compile time)", s.Type)
	case *ast.ScopeStatement:
		g.writeLinef("// %s %s", s.Modifier, strings.Join(s.Variables, ", "))
	case *ast.EraseStatement:
		g.emitErase(s)
	case *ast.BeepStatement:
		g.writeLine("fmt.Print(\"\\a\") // BEEP")
	case *ast.OpenStatement:
		g.emitOpen(s)
	case *ast.CloseStatement:
		g.emitClose(s)
	case *ast.FilePrintStatement:
		g.emitFilePrint(s)
	case *ast.FileInputStatement:
		g.emitFileInput(s)
	case *ast.FileWriteStatement:
		g.emitFileWrite(s)
	case *ast.FieldStatement:
		g.emitField(s)
	case *ast.LsetStatement:
		g.emitLset(s)
	case *ast.RsetStatement:
		g.emitRset(s)
	case *ast.PutStatement:
		g.emitPut(s)
	case *ast.GetStatement:
		g.emitGet(s)
	case *ast.SeekStatement:
		g.emitSeek(s)
	case *ast.LocateStatement:
		g.emitLocate(s)
	case *ast.ClsStatement:
		g.writeLine("fmt.Print(rt.AnsiCls()) // CLS")
	case *ast.ScreenStatement:
		g.writeLinef("rt.ScreenMode(int(%s))", g.emitExpr(s.Mode))
	case *ast.ColorStatement:
		g.emitColor(s)
	case *ast.CircleStmt:
		g.emitCircle(s)
	case *ast.LineStmt:
		g.emitLine(s)
	case *ast.PsetStatement:
		g.emitPset(s)
	case *ast.PaintStmt:
		g.emitPaint(s)
	case *ast.DrawStmt:
		g.writeLinef("rt.Draw(%s)", g.emitExpr(s.CommandString))
	case *ast.ViewStatement:
		g.emitView(s)
	case *ast.OnErrorGotoStatement:
		g.needErrState = true
		if s.Target == "0" || s.Target == "" {
			g.writeLine("errState.SetHandler(\"\") // ON ERROR GOTO 0 — disable error trapping")
			g.onErrorLabel = ""
		} else {
			lbl := g.labelName(s.Target)
			g.writeLinef("errState.SetHandler(\"%s\") // ON ERROR GOTO %s", lbl, s.Target)
			g.onErrorLabel = lbl
			// Emit a dead-code guard so Go's "label defined and not used" check is
			// satisfied even when the label is never the target of a syntactic goto
			// (e.g. when no ERROR n statement exists to emit one).  The guard is
			// harmless when a real goto already exists.
			key := strings.ToUpper(s.Target)
			if g.onErrorTargets[key] {
				g.writeLine("if false { goto " + lbl + " } // ON ERROR GOTO reachability guard")
				// Clear so emitLabel/emitLineNumber don't emit a second guard.
				delete(g.onErrorTargets, key)
			}
		}
	case *ast.OnComputedGotoStatement:
		g.emitOnComputedGoto(s)
	case *ast.OnComputedGosubStatement:
		g.emitOnComputedGosub(s)
	case *ast.PokeStatement:
		g.writeLinef("_ = %s; _ = %s // POKE (no-op in transpiled code)", g.emitExpr(s.Address), g.emitExpr(s.Value))
	case *ast.ConstStatement:
		name := mangleName(s.Name)
		val := g.emitExpr(s.Value)
		if g.constNames[name] {
			// Promoted to package level — use assignment form with type cast.
			goT := g.goTypeForIdent(s.Name)
			if goT == "string" {
				g.writeLinef("%s = %s // CONST", name, val)
			} else {
				g.writeLinef("%s = %s(%s) // CONST", name, goT, val)
			}
		} else {
			g.writeLinef("%s := %s // CONST", name, val)
		}
	case *ast.ClearStatement:
		// CLEAR resets all numeric variables to 0 and string variables to "".
		// Emit an assignment to the zero value for every hoisted variable.
		for _, hv := range g.hoistedVars {
			switch hv.typ {
			case "string":
				g.writeLinef("%s = \"\"", hv.name)
			default:
				g.writeLinef("%s = 0", hv.name)
			}
		}
		// Also reset any package-level (SHARED) variables.
		for _, pv := range g.packageVars {
			switch pv.typ {
			case "string":
				g.writeLinef("%s = \"\"", pv.name)
			default:
				g.writeLinef("%s = 0", pv.name)
			}
		}
		if len(g.hoistedVars) == 0 && len(g.packageVars) == 0 {
			g.writeLine("// CLEAR — no variables to reset")
		}
	case *ast.TypeBlockStatement:
		g.emitTypeBlock(s)
	case *ast.FnAssignStatement:
		if g.inDefFn {
			g.writeLinef("return %s(%s)", g.defFnReturnType, g.emitExpr(s.Value))
		} else {
			g.writeLinef("_fn_%s = %s", mangleName(s.Name), g.emitExpr(s.Value))
		}
	case *ast.FieldAssignStatement:
		// Struct/TYPE field assignment: obj.Field = value
		g.writeLinef("%s.%s = %s", g.emitExpr(s.Object), mangleName(s.Field), g.emitExpr(s.Value))
	case *ast.ResumeStatement:
		g.needErrState = true
		switch s.Type {
		case "NEXT":
			g.writeLine("errState.ClearError() // RESUME NEXT — clear error and continue")
		case "":
			// RESUME (bare): clear error and jump back to the label saved when the error
			// was caught. errState.ResumeLabel is set by the ON ERROR GOTO handler.
			g.writeLine("errState.ClearError()")
			g.writeLine("if errState.ResumeLabel != \"\" {")
			g.indent++
			g.writeLine("panic(\"RESUME: cannot jump to \" + errState.ResumeLabel + \" (runtime goto not supported)\")")
			g.indent--
			g.writeLine("}")
		default:
			// RESUME <label> — clear error and jump to label
			g.writeLine("errState.ClearError()")
			g.writeLinef("goto %s", g.labelName(s.Type))
		}
	case *ast.PlayStatement:
		g.writeLinef("rt.Play(%s)", g.emitExpr(s.CommandString))
	case *ast.SoundStatement:
		g.writeLinef("rt.Sound(float64(%s), float64(%s))", g.emitExpr(s.Frequency), g.emitExpr(s.Duration))
	case *ast.ErrorStatement:
		// ERROR n simulates a runtime error with code n.
		// TriggerError panics if no handler is active, or stores the error for
		// the ON ERROR GOTO handler to examine via ERR/ERL.
		g.needErrState = true
		if g.onErrorLabel != "" {
			g.writeLinef("if _berr := errState.TriggerError(int(%s), 0); _berr != nil { goto %s }",
				g.emitExpr(s.Code), g.onErrorLabel)
		} else {
			g.writeLinef("errState.TriggerError(int(%s), 0)", g.emitExpr(s.Code))
		}

	// Sub/Function/DefFn declarations are handled separately via emitTopLevelDecl.
	case *ast.SubDeclaration:
		// Should not reach here in normal flow.
	case *ast.FunctionDeclaration:
		// Should not reach here in normal flow.
	case *ast.DefFnDeclaration:
		// Should not reach here in normal flow.

	default:
		// Catch-all: emit a TODO comment so the output still compiles.
		g.writeLinef("// TODO: %s", stmt.TokenLiteral())
	}
}

// ---------------------------------------------------------------------------
// PRINT — control flow / output mapping from BASIC to Go
//
// BASIC's PRINT statement is richer than a plain fmt.Println() call because it
// supports three distinct output modes controlled by the separator characters
// between expressions:
//
//	PRINT A; B      → semicolon: concatenate values with no spacing
//	                  → fmt.Print(A, B) / fmt.Println(A, B)
//
//	PRINT A, B      → comma: "zone" or column-tabbing output.  Each comma
//	                  advances to the next 14-character print zone.
//	                  → rt.PrintZone() from the runtime package
//
//	PRINT USING f$; → formatted output with a format template.
//	                  → rt.PrintUsing() from the runtime package
//
// A trailing semicolon suppresses the newline (PRINT A;), mapping to
// fmt.Print() instead of fmt.Println().
//
// This illustrates a general challenge in transpilation: a single source
// keyword may need to be mapped to several different target constructs
// depending on its runtime-determined arguments.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitPrint(s *ast.PrintStatement) {
	// PRINT USING
	if s.Format != nil {
		g.emitPrintUsing(s)
		return
	}

	// No expressions: just println.
	if len(s.Expressions) == 0 {
		g.writeLine("fmt.Println()")
		return
	}

	// Check if we have comma separators (zone printing).
	hasComma := false
	for _, sep := range s.Separators {
		if sep == "," {
			hasComma = true
			break
		}
	}

	if hasComma {
		// Zone-based printing: use rt.PrintZone.
		g.writeLine("{")
		g.indent++
		g.writeLine("col_ := 0")
		g.writeLine("out_ := \"\"")
		for i, expr := range s.Expressions {
			val := g.emitExpr(expr)
			sep := ""
			if i < len(s.Separators) {
				sep = s.Separators[i]
			}
			if sep == "," {
				g.writeLinef("{ s_, c_ := rt.PrintZone(col_, fmt.Sprint(%s)); out_ += s_; col_ = c_ }", val)
			} else {
				g.writeLinef("{ v_ := fmt.Sprint(%s); out_ += v_; col_ += len(v_) }", val)
			}
		}
		if s.HasTrailingSep {
			g.writeLine("fmt.Print(out_)")
		} else {
			g.writeLine("fmt.Println(out_)")
		}
		g.indent--
		g.writeLine("}")
		return
	}

	// Semicolons or no separator: concatenate values.
	if len(s.Expressions) == 1 && !s.HasTrailingSep {
		g.writeLinef("fmt.Println(%s)", g.emitExpr(s.Expressions[0]))
		return
	}

	parts := make([]string, 0, len(s.Expressions))
	for _, expr := range s.Expressions {
		parts = append(parts, g.emitExpr(expr))
	}

	if s.HasTrailingSep {
		g.writeLinef("fmt.Print(%s)", strings.Join(parts, ", "))
	} else {
		g.writeLinef("fmt.Println(%s)", strings.Join(parts, ", "))
	}
}

func (g *CodeGenerator) emitPrintUsing(s *ast.PrintStatement) {
	formatExpr := g.emitExpr(s.Format)
	args := make([]string, 0, len(s.Expressions))
	for _, expr := range s.Expressions {
		args = append(args, g.emitExpr(expr))
	}
	g.writeLinef("fmt.Print(rt.PrintUsing(%s, []interface{}{%s}))", formatExpr, strings.Join(args, ", "))
	if !s.HasTrailingSep {
		g.writeLine("fmt.Println()")
	}
}

// ---------------------------------------------------------------------------
// LET (assignment)
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitLet(s *ast.LetStatement) {
	upperName := strings.ToUpper(s.Name.Name)

	// Inside a multi-line DEF FN, assignments to the function name become return statements.
	if g.inDefFn && strings.EqualFold(s.Name.Name+s.Name.TypeSuffix, g.defFnName) {
		g.writeLinef("return %s(%s)", g.defFnReturnType, g.emitExpr(s.Value))
		return
	}

	// CALL statements are parsed as LetStatement where LHS name == RHS FunctionCall name.
	// Detect this pattern and emit a plain function call (discarding the return value)
	// rather than an assignment, to avoid "declared and not used" compile errors.
	if fc, ok := s.Value.(*ast.FunctionCall); ok {
		if upperName == strings.ToUpper(fc.Name) && s.Name.TypeSuffix == "" {
			args := make([]string, 0, len(fc.Args))
			for _, a := range fc.Args {
				args = append(args, g.emitExpr(a))
			}
			// Apply coercion (including by-ref pointer wrapping) for SUB/FUNCTION calls.
			if g.program != nil && g.subFuncNames[strings.ToUpper(fc.Name)] {
				args = g.coerceCallArgs(fc, args)
			}
			g.writeLinef("%s(%s)", mangleName(fc.Name), strings.Join(args, ", "))
			return
		}
	}

	name := mangleName(s.Name.Name + s.Name.TypeSuffix)
	goT := g.goTypeForIdent(s.Name.Name + s.Name.TypeSuffix)

	// For string type, don't wrap with a type cast (Go doesn't allow string(expr)
	// on non-byte-slice values; string concatenation and fmt.Sprint handle coercion).
	if goT == "string" {
		val := g.emitExpr(s.Value)
		if g.paramsByRef[name] {
			g.writeLinef("*%s = %s", name, val)
			return
		}
		if !g.declared[name] {
			g.declared[name] = true
			g.writeLinef("var %s string = %s", name, val)
		} else {
			g.writeLinef("%s = %s", name, val)
		}
		return
	}

	val := g.emitExpr(s.Value)
	if g.paramsByRef[name] {
		g.writeLinef("*%s = %s(%s)", name, goT, val)
		return
	}
	if !g.declared[name] {
		g.declared[name] = true
		// Cast the value to the variable's type to handle mismatches between
		// BASIC's implicit coercions and Go's strict typing.
		g.writeLinef("var %s %s = %s(%s)", name, goT, goT, val)
	} else {
		// Use the hoisted/package type if available — it reflects the DIM/AS declaration
		// and may differ from the suffix-based inference returned by goTypeForIdent.
		castT := goT
		if ht, ok := g.hoistedTypes[name]; ok {
			castT = ht
		} else if g.sharedVars[name] {
			// For package-level variables, find the declared type.
			for _, pv := range g.packageVars {
				if pv.name == name {
					// Strip array prefix for assignment cast
					t := pv.typ
					for strings.HasPrefix(t, "[]") {
						t = t[2:]
					}
					castT = t
					break
				}
			}
		}
		g.writeLinef("%s = %s(%s)", name, castT, val)
	}
}

// ---------------------------------------------------------------------------
// RANDOMIZE
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitRandomize(s *ast.RandomizeStatement) {
	g.needRng = true
	if s.Seed != nil {
		seed := g.emitExpr(s.Seed)
		g.writeLinef("rng.Randomize(float64(%s))", seed)
	} else {
		g.writeLine("rng.Randomize(rt.Timer())")
	}
}

// ---------------------------------------------------------------------------
// Array assignment
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitArrayAssignment(s *ast.ArrayAssignment) {
	arrName := mangleName(s.Array.Name + s.Array.TypeSuffix)
	indices := make([]string, 0, len(s.Array.Indices))
	for _, idx := range s.Array.Indices {
		indices = append(indices, g.emitExpr(idx))
	}
	val := g.emitExpr(s.Value)

	// Cast value to the array element type to prevent Go type mismatches.
	elemType := g.goTypeForIdent(s.Array.Name + s.Array.TypeSuffix)
	if len(indices) == 1 {
		g.writeLinef("%s[int(%s)] = %s(%s)", arrName, indices[0], elemType, val)
	} else {
		// Multi-dimensional: emit chained bracket access for assignment.
		var accessBuf strings.Builder
		accessBuf.WriteString(arrName)
		for _, idx := range indices {
			accessBuf.WriteString(fmt.Sprintf("[int(%s)]", idx))
		}
		g.writeLinef("%s = %s(%s)", accessBuf.String(), elemType, val)
	}
}

// ---------------------------------------------------------------------------
// SWAP, INCR, DECR
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitSwap(s *ast.SwapStatement) {
	a := g.emitExpr(s.Var1)
	b := g.emitExpr(s.Var2)
	g.writeLinef("%s, %s = %s, %s", a, b, b, a)
}

func (g *CodeGenerator) emitIncr(s *ast.IncrStatement) {
	v := g.emitExpr(s.Variable)
	if s.Amount != nil {
		amt := g.emitExpr(s.Amount)
		g.writeLinef("%s += %s", v, amt)
	} else {
		g.writeLinef("%s++", v)
	}
}

func (g *CodeGenerator) emitDecr(s *ast.DecrStatement) {
	v := g.emitExpr(s.Variable)
	if s.Amount != nil {
		amt := g.emitExpr(s.Amount)
		g.writeLinef("%s -= %s", v, amt)
	} else {
		g.writeLinef("%s--", v)
	}
}

// ---------------------------------------------------------------------------
// ERASE
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitErase(s *ast.EraseStatement) {
	for _, name := range s.Names {
		mn := mangleName(name)
		g.writeLinef("%s = nil // ERASE", mn)
	}
}

// ---------------------------------------------------------------------------
// Labels and line numbers
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitLineNumber(s *ast.LineNumberStatement) {
	// Only emit the label if it is the target of a GOTO or GOSUB.
	// Go treats unused labels as compile errors, so emitting unreferenced
	// line-number labels (which BASIC programs often have for every line)
	// would make the output fail to compile.
	key := fmt.Sprintf("%d", s.Number)
	if !g.referencedLabels[key] {
		return
	}
	label := "line_" + key
	if g.indent > 0 {
		old := g.indent
		g.indent = 0
		g.writeLinef("%s:", label)
		g.indent = old
	} else {
		g.writeLinef("%s:", label)
	}
}

func (g *CodeGenerator) emitLabel(s *ast.LabelStatement) {
	// Same as above: only emit if referenced.
	key := strings.ToUpper(s.Name)
	if !g.referencedLabels[key] {
		return
	}
	label := g.labelName(s.Name)
	if g.indent > 0 {
		old := g.indent
		g.indent = 0
		g.writeLinef("%s:", label)
		g.indent = old
	} else {
		g.writeLinef("%s:", label)
	}
}

func (g *CodeGenerator) emitRem(s *ast.RemStatement) {
	g.writeLinef("// %s", s.Text)
}

// emitRestore emits the RESTORE statement (reset dataIdx to start of pool).
func (g *CodeGenerator) emitRestore(_ *ast.RestoreStatement) {
	g.writeLine("dataIdx = 0 // RESTORE")
}

// emitTypeBlock emits a TYPE block as a Go struct definition.
func (g *CodeGenerator) emitTypeBlock(s *ast.TypeBlockStatement) {
	// TYPE blocks are emitted as package-level struct definitions.
	// Since they must be at package level, write to funcBuf.
	// Use uppercase name to match DIM/param type resolution (parser stores types as uppercase).
	g.funcWriteLinef(0, "type %s struct {", mangleName(strings.ToUpper(s.Name)))
	for _, f := range s.Fields {
		goT := g.goTypeForParamType(f.TypeName, f.Name)
		g.funcWriteLinef(1, "%s %s", mangleName(f.Name), goT)
	}
	g.funcWriteLine(0, "}")
	g.funcWriteLine(0, "")
}

// emitOnComputedGoto emits an ON expr GOTO t1, t2, ... as a series of if/goto.
// The target label names are converted using labelName() to match the emitted labels.
func (g *CodeGenerator) emitOnComputedGoto(s *ast.OnComputedGotoStatement) {
	expr := g.emitExpr(s.Expr)
	g.writeLine("{")
	g.indent++
	g.writeLinef("_on_idx := int(%s)", expr)
	for i, t := range s.Targets {
		g.writeLinef("if _on_idx == %d { goto %s }", i+1, g.labelName(t))
	}
	g.indent--
	g.writeLine("}")
}

// emitOnComputedGosub emits ON expr GOSUB t1, t2, ... as a series of if/goto.
// True GOSUB (save return address, jump, return) can't be emulated with goto,
// so we use goto for now — the RETURN at the subroutine end exits the function.
func (g *CodeGenerator) emitOnComputedGosub(s *ast.OnComputedGosubStatement) {
	expr := g.emitExpr(s.Expr)
	g.writeLine("{")
	g.indent++
	g.writeLinef("_on_idx := int(%s)", expr)
	for i, t := range s.Targets {
		g.writeLinef("if _on_idx == %d { goto %s }", i+1, g.labelName(t))
	}
	g.indent--
	g.writeLine("}")
}
