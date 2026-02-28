package codegen

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// IF / THEN / ELSE — conditional control flow mapping
//
// BASIC's IF maps almost 1-to-1 onto Go's if/else if/else chain.  The
// structural difference that requires care is BASIC's numeric truth model:
// any non-zero numeric value is true (0 is false), whereas Go requires an
// explicit bool in a condition expression.
//
// The helper toBoolExpr() bridges this gap.  It inspects the emitted
// expression string: if it already contains a comparison operator it is
// already a Go bool; otherwise it appends "!= 0" to coerce the numeric value.
// This heuristic works well in practice because BASIC conditions are almost
// always either comparisons or the results of logical operators (AND/OR/NOT),
// which are themselves emitted as bitwise operations on integers.
//
// ELSEIF clauses are unrolled into the natural Go "} else if … {" pattern
// by iterating s.ElseIfClauses.  The recursive body walks emit nested
// statements at increased indentation, producing correctly formatted output.
//
// TUTORIAL — Type system differences between BASIC and Go
//
// BASIC uses a C-like numeric truth model: 0 is false, anything else is true.
// This means the following BASIC code is valid:
//
//   X = 5
//   IF X THEN PRINT "non-zero"    ' prints "non-zero"
//   IF X - 5 THEN PRINT "also"   ' does NOT print (X - 5 == 0)
//
// Go requires conditions to be explicitly boolean; "if 5 {}" is a compile error.
//
// When transpiling, we must bridge this semantic gap. The toBoolExpr() helper
// uses a text-level heuristic: if the emitted condition string already contains
// a comparison operator (==, !=, <, >, etc.), it is already boolean. Otherwise,
// it wraps the expression with "!= 0":
//
//   IF X > 3 → if (x > 3) {          (already boolean, no wrapping)
//   IF X     → if (x) != 0 {         (numeric, needs != 0)
//   IF X AND Y → if (int(x) & int(y)) != 0 {   (bitwise AND, numeric result)
//
// This heuristic is imperfect (it can be confused by complex strings) but
// handles the vast majority of real BASIC programs correctly. A more robust
// approach would use the type system to determine at AST level whether an
// expression is boolean or numeric — that is what a full type-checker would do.
//
// TUTORIAL — Single-line vs. block IF
//
// Turbo BASIC supports two syntactic forms of IF:
//
//   Single-line:  IF X > 0 THEN PRINT X ELSE PRINT 0
//   Block form:   IF X > 0 THEN
//                     PRINT X
//                 ELSE
//                     PRINT 0
//                 END IF
//
// The parser (Phase 2) normalises both forms into the same ast.IfStatement
// node with ThenBlock and ElseBlock slices. emitIf() therefore only needs to
// handle the block form — the single-line form is already transformed.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitIf(s *ast.IfStatement) {
	g.writeLinef("if %s {", g.toBoolExprFromNode(s.Condition))
	g.indent++
	for _, stmt := range s.ThenBlock {
		g.emitStatement(stmt)
	}
	g.indent--

	for _, clause := range s.ElseIfClauses {
		g.writeLinef("} else if %s {", g.toBoolExprFromNode(clause.Condition))
		g.indent++
		for _, stmt := range clause.Body {
			g.emitStatement(stmt)
		}
		g.indent--
	}

	if len(s.ElseBlock) > 0 {
		g.writeLine("} else {")
		g.indent++
		for _, stmt := range s.ElseBlock {
			g.emitStatement(stmt)
		}
		g.indent--
	}

	g.writeLine("}")
}

// ---------------------------------------------------------------------------
// FOR / NEXT — numeric loop control flow mapping
//
// BASIC's FOR loop is more subtle than it appears.  The STEP value may be
// positive, negative, or zero, and BASIC requires the loop body to execute
// zero times if the initial value is already past the end value in the given
// direction.  A simple Go "for i = start; i <= end; i += step" is wrong for
// negative steps.
//
// The emitted loop handles all three cases with a single compound condition:
//
//	for counter = start;
//	    (step > 0 && counter <= end) ||
//	    (step < 0 && counter >= end) ||
//	    (step == 0);           // runs exactly once when step == 0
//	    counter += step { … }
//
// To avoid re-evaluating the STEP and END expressions on every iteration
// (which would be wrong if they contained side effects), both are captured
// in temporary variables (end_N, step_N) before the loop header.  The
// tempCount field provides a monotonically increasing suffix to guarantee
// uniqueness across multiple FOR loops in the same scope.
//
// TUTORIAL — Source-language semantics vs. target-language idioms
//
// A key skill in compiler construction is understanding exactly what the
// source language guarantees and mapping those guarantees faithfully to the
// target language, even when the source construct has no direct target analog.
//
// BASIC's FOR loop guarantees:
//   1. The END and STEP expressions are evaluated exactly once before the loop
//      begins (not on every iteration).
//   2. The loop direction determines the termination condition: upward-counting
//      loops use <=, downward-counting loops use >=.
//   3. A STEP of zero is defined to run the body exactly once (some dialects
//      treat it as infinite; Turbo BASIC uses the once-semantics).
//
// Go's "for init; cond; post" loop evaluates the condition before each
// iteration, which matches BASIC. But Go has no built-in support for
// direction-aware stepping. The solution used here is a compound boolean
// condition that encodes all three cases simultaneously:
//
//   (step > 0 && counter <= end) ||
//   (step < 0 && counter >= end) ||
//   (step == 0)
//
// This is a common pattern when mapping a richer source construct to a more
// primitive target construct: encode the richer semantics as explicit logic in
// the target language rather than relying on language-level support.
//
// The tempCount field demonstrates another common technique: generating unique
// temporary variable names. Each FOR loop needs its own end_N / step_N pair,
// and a simple monotonically-increasing counter is the simplest way to
// guarantee uniqueness without a full name-allocation pass.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitFor(s *ast.ForStatement) {
	counter := mangleName(s.Counter.Name + s.Counter.TypeSuffix)
	startExpr := g.emitExpr(s.Start)
	endExpr := g.emitExpr(s.End)

	stepExpr := "1"
	if s.Step != nil {
		stepExpr = g.emitExpr(s.Step)
	}

	// Declare counter if first use.
	if !g.declared[counter] {
		g.declared[counter] = true
		goT := g.goTypeForIdent(s.Counter.Name + s.Counter.TypeSuffix)
		g.writeLinef("var %s %s", counter, goT)
	}

	// Use a helper temp for end and step so they are evaluated once.
	// Cast both to the counter's Go type to avoid mismatched-type errors: BASIC
	// is dynamically typed but Go requires homogeneous operands in comparisons.
	endVar := fmt.Sprintf("end_%d", g.tempCount)
	stepVar := fmt.Sprintf("step_%d", g.tempCount)
	g.tempCount++

	// Prefer the hoisted/declared type for the counter — it reflects
	// DIM AS declarations (e.g., DIM i AS SINGLE → float32) which override
	// the suffix-based inference from goTypeForIdent.
	counterType := g.goTypeForIdent(s.Counter.Name + s.Counter.TypeSuffix)
	if ht, ok := g.hoistedTypes[counter]; ok {
		counterType = ht
	}
	g.writeLinef("%s := %s(%s)", endVar, counterType, endExpr)
	g.writeLinef("%s := %s(%s)", stepVar, counterType, stepExpr)
	g.writeLinef("for %s = %s(%s); (%s > 0 && %s <= %s) || (%s < 0 && %s >= %s) || (%s == 0); %s += %s {",
		counter, counterType, startExpr,
		stepVar, counter, endVar,
		stepVar, counter, endVar,
		stepVar,
		counter, stepVar)
	g.indent++
	for _, stmt := range s.Body {
		g.emitStatement(stmt)
	}
	g.indent--
	g.writeLine("}")
}

// ---------------------------------------------------------------------------
// WHILE / WEND
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitWhile(s *ast.WhileStatement) {
	g.writeLinef("for %s {", g.toBoolExprFromNode(s.Condition))
	g.indent++
	for _, stmt := range s.Body {
		g.emitStatement(stmt)
	}
	g.indent--
	g.writeLine("}")
}

// ---------------------------------------------------------------------------
// DO / LOOP
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitDoLoop(s *ast.DoLoopStatement) {
	if s.Condition == nil {
		// Infinite loop: DO ... LOOP
		g.writeLine("for {")
		g.indent++
		for _, stmt := range s.Body {
			g.emitStatement(stmt)
		}
		g.indent--
		g.writeLine("}")
		return
	}

	boolCond := g.toBoolExprFromNode(s.Condition)

	if s.TestAtTop {
		if s.IsUntil {
			g.writeLinef("for !(%s) {", boolCond)
		} else {
			g.writeLinef("for %s {", boolCond)
		}
		g.indent++
		for _, stmt := range s.Body {
			g.emitStatement(stmt)
		}
		g.indent--
		g.writeLine("}")
	} else {
		// Test at bottom.
		g.writeLine("for {")
		g.indent++
		for _, stmt := range s.Body {
			g.emitStatement(stmt)
		}
		if s.IsUntil {
			g.writeLinef("if %s { break }", boolCond)
		} else {
			g.writeLinef("if !(%s) { break }", boolCond)
		}
		g.indent--
		g.writeLine("}")
	}
}

// ---------------------------------------------------------------------------
// SELECT CASE
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitSelectCase(s *ast.SelectCaseStatement) {
	testExpr := g.emitExpr(s.TestExpr)
	testVar := fmt.Sprintf("sel_%d", g.tempCount)
	g.tempCount++
	g.writeLinef("%s := %s", testVar, testExpr)

	// Emit as a Go "switch {}" (tagless switch on boolean conditions) rather
	// than an if/else chain.  This is important because GO requires "break"
	// statements to be inside a for/switch/select — and BASIC's EXIT FOR/
	// EXIT LOOP inside a CASE body should break out of the surrounding switch.
	g.writeLine("switch {")
	for _, c := range s.Cases {
		condParts := make([]string, 0, len(c.Values))
		for _, cv := range c.Values {
			if cv.IsRange {
				low := g.emitExpr(cv.Value)
				high := g.emitExpr(cv.EndValue)
				condParts = append(condParts, fmt.Sprintf("(%s >= %s && %s <= %s)", testVar, low, testVar, high))
			} else if cv.IsComparison {
				val := g.emitExpr(cv.Value)
				op := cv.Comparison
				condParts = append(condParts, fmt.Sprintf("(%s %s %s)", testVar, op, val))
			} else {
				val := g.emitExpr(cv.Value)
				condParts = append(condParts, fmt.Sprintf("(%s == %s)", testVar, val))
			}
		}
		cond := strings.Join(condParts, " || ")
		g.writeLinef("case %s:", cond)
		g.indent++
		for _, stmt := range c.Body {
			g.emitStatement(stmt)
		}
		g.indent--
	}

	if len(s.ElseBlock) > 0 {
		g.writeLine("default:")
		g.indent++
		for _, stmt := range s.ElseBlock {
			g.emitStatement(stmt)
		}
		g.indent--
	}

	if len(s.Cases) > 0 || len(s.ElseBlock) > 0 {
		g.writeLine("}")
	}
}

// ---------------------------------------------------------------------------
// GOTO / GOSUB / RETURN
//
// TUTORIAL — GOTO: a case of near-direct mapping
//
// GOTO is one of the simplest BASIC constructs to transpile because Go also
// has a goto statement. The mapping is almost 1-to-1:
//
//   BASIC:  GOTO 100           Go: goto line_100
//   BASIC:  GOTO ErrorHandler  Go: goto label_ErrorHandler
//
// The only work required is translating the BASIC label reference (which may
// be a bare line number like "100" or a symbolic name like "ErrorHandler")
// into a legal Go identifier. This is the job of labelName().
//
// However, Go's goto has an important restriction: it cannot jump forward
// over a variable declaration in the same scope. For example, this is a
// Go compile error:
//
//   goto skip
//   x := 5    // ← goto jumps over this declaration
//   skip:
//
// BASIC programs frequently do this because their variables are implicitly
// declared anywhere. The solution used here is variable hoisting: when any
// GOTO exists in the program, all main-level variables are declared with "var"
// at the top of main() (see collectMainVariables in codegen.go), before any
// goto target label, making the jump-over problem impossible.
//
// TUTORIAL — GOSUB: a case of semantic mismatch
//
// GOSUB is harder because it has no direct Go equivalent. BASIC's GOSUB
// saves the current program counter (return address) on an implicit call
// stack, jumps to the subroutine, and RETURN pops that address to continue
// execution after the original GOSUB.
//
// Go does not have this mechanism. A true simulation would require either:
//   - A runtime call stack (complex, heap-allocated)
//   - Converting each GOSUB target into a named function (requires knowing all
//     callers and all exit paths — a non-trivial analysis)
//
// This transpiler takes a pragmatic approximation: GOSUB is emitted as a
// plain goto. This is semantically incorrect for programs that GOSUB to the
// same subroutine from multiple call sites (the RETURN would only work for
// one of them), but it handles the common idiom where a subroutine is
// called from one place and terminates with RETURN. Programs relying on
// multi-caller GOSUB are a known limitation documented in the project README.
//
// This illustrates an important lesson: real-world transpilers often make
// deliberate semantic approximations for constructs that have no target-
// language equivalent. Documenting these limitations honestly is as important
// as the approximation itself.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitGoto(s *ast.GotoStatement) {
	label := g.labelName(s.Target)
	g.writeLinef("goto %s", label)
}

func (g *CodeGenerator) emitGosub(s *ast.GosubStatement) {
	// In Go we cannot do a true GOSUB/RETURN. Emit as a goto with a comment.
	label := g.labelName(s.Target)
	g.writeLinef("goto %s // GOSUB", label)
}

func (g *CodeGenerator) emitReturn(_ *ast.ReturnStatement) {
	g.writeLine("return // RETURN")
}

// ---------------------------------------------------------------------------
// EXIT
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitExit(s *ast.ExitStatement) {
	switch strings.ToUpper(s.ExitType) {
	case "FOR", "DO", "WHILE", "LOOP":
		// LOOP is an alias for DO in Turbo BASIC (EXIT LOOP = exit DO loop)
		g.writeLine("break")
	case "SUB", "FUNCTION":
		g.writeLine("return")
	case "DEF":
		// EXIT DEF exits a DEF FN multi-line function (equivalent to return)
		if g.inDefFn {
			g.writeLinef("return %s", zeroValueForType(g.defFnReturnType))
		} else {
			g.writeLine("return")
		}
	default:
		g.writeLinef("// EXIT %s (unsupported)", s.ExitType)
	}
}
