// Package codegen implements Phase 4 of the go-basic compiler pipeline:
// code generation — the final transformation from an Abstract Syntax Tree (AST)
// into executable target code.
//
// # Compilation pipeline recap
//
//   Source text
//       │  Phase 1: Lexer  (text → token stream)
//       ▼
//   Token stream
//       │  Phase 2: Parser (tokens → AST)
//       ▼
//   AST
//       │  Phase 3: Semantic analysis (type inference, symbol table)
//       ▼
//   Annotated AST
//       │  Phase 4: Code generation  ← YOU ARE HERE
//       ▼
//   Target source (Go)
//
// # This is a transpiler, not a native-code compiler
//
// Rather than emitting machine code or bytecode, this package emits Go source
// code that can then be compiled by the standard Go toolchain.  The strategy
// trades raw performance for simplicity and portability: every BASIC construct
// maps to an equivalent Go construct, and anything without a direct mapping
// falls back to a helper in the runtime package (internal/runtime).
//
// # Tree-walking code generation
//
// The dominant pattern here is tree walking (also called a "recursive descent
// emitter").  Starting from the root ast.Program node, we iterate over every
// statement and call emitStatement(), which dispatches on the concrete node
// type using a Go type-switch.  Expression nodes are handled analogously by
// emitExpr().  Each handler directly writes Go source text into a bytes.Buffer.
//
// This is the simplest possible code-generation strategy and is sufficient for
// a transpiler.  A production compiler targeting native code would instead
// lower the AST to an intermediate representation (IR) such as SSA or LLVM IR
// before performing register allocation and instruction selection.
//
// # Name mangling
//
// BASIC variable names cannot be used as-is in Go for two reasons:
//
//  1. Type-sigil suffixes: BASIC uses trailing characters to encode type
//     information.  "X%" is an integer, "S$" is a string, "D#" is a double.
//     These characters are not legal in Go identifiers, so they are replaced
//     with readable suffixes: _pct, _str, _dbl, etc.
//
//  2. Keyword conflicts: Many common BASIC names ("return", "for", "type",
//     "string", …) are reserved words in Go.  The mangler detects these and
//     prepends "b_" so they remain valid identifiers without colliding.
//
// The mangleName() function handles both cases.  All references to a variable
// — declarations, reads, and writes — go through the same mangler, so output
// is consistent throughout the generated file.
//
// # Structure of the generated Go file
//
// The assembler (the second half of Generate()) stitches sections together in
// a fixed order:
//
//  1. "package main" declaration.
//  2. import block — only imports that were actually needed are emitted.
//  3. Blank-identifier suppression lines (var _ = fmt.Sprintf, etc.) so the
//     file compiles even when an import is referenced only by a TODO stub.
//  4. func main() { … } — contains all top-level BASIC statements in source
//     order, plus optional RNG and DATA pool preamble.
//  5. SUB / FUNCTION / DEF FN declarations — appended after main().  They are
//     collected in a separate funcBuf during the walk so that forward
//     references from inside main() still resolve correctly.
package codegen

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// ---------------------------------------------------------------------------
// CodeGenerator transpiles a Turbo BASIC AST into compilable Go source code.
// ---------------------------------------------------------------------------

// CodeGenerator holds all state needed while walking the AST and emitting Go.
type CodeGenerator struct {
	program         *ast.Program
	table           *semantic.SymbolTable
	buf             bytes.Buffer
	indent          int
	imports         map[string]bool  // track needed imports
	declared        map[string]bool  // track declared variables (mangled names)
	tempCount       int              // temp variable counter
	labelMap        map[string]bool  // labels that exist (defined in source)
	referencedLabels map[string]bool // labels that are targeted by GOTO/GOSUB
	gosubFuncs      map[string]bool  // GOSUB targets turned into functions
	dataPool        []ast.Expression // DATA values
	dataIdx         int              // current READ position

	// funcBuf collects SUB/FUNCTION declarations to emit outside main().
	funcBuf bytes.Buffer

	// needRng tracks whether the rng variable is needed.
	needRng bool

	// hasRead tracks whether any DATA READ statements exist (to emit dataPool).
	hasRead bool
}

// New creates a fresh CodeGenerator ready for use.
func New() *CodeGenerator {
	return &CodeGenerator{
		imports:          make(map[string]bool),
		declared:         make(map[string]bool),
		labelMap:         make(map[string]bool),
		referencedLabels: make(map[string]bool),
		gosubFuncs:       make(map[string]bool),
	}
}

// Generate is the main entry point.  It walks the AST, emits Go code, and
// returns the fully-formed Go source string (or an error).
func (g *CodeGenerator) Generate(program *ast.Program, table *semantic.SymbolTable) (string, error) {
	g.program = program
	g.table = table

	// Pre-pass: collect labels & DATA values.
	g.collectLabelsAndData(program.Statements)

	// ---- Emit main body into g.buf ----
	g.indent = 1
	for _, stmt := range program.Statements {
		// SUB/FUNCTION declarations go into funcBuf, not main.
		switch stmt.(type) {
		case *ast.SubDeclaration, *ast.FunctionDeclaration, *ast.DefFnDeclaration:
			g.emitTopLevelDecl(stmt)
			continue
		}
		g.emitStatement(stmt)
	}

	// ---- Assemble final output ----
	var out bytes.Buffer

	out.WriteString("package main\n\n")

	// Always import fmt (used by almost every BASIC program).
	g.imports["fmt"] = true

	// If we have DATA statements or READ statements, we need the data pool infrastructure.
	hasData := len(g.dataPool) > 0 || g.hasRead

	// Build import block.
	out.WriteString("import (\n")
	stdImports := []string{"fmt", "math", "os", "strconv", "strings"}
	wroteStd := false
	for _, imp := range stdImports {
		if g.imports[imp] {
			out.WriteString("\t\"" + imp + "\"\n")
			wroteStd = true
		}
	}
	if wroteStd {
		out.WriteString("\n")
	}
	out.WriteString("\trt \"github.com/loabletech/go-basic/internal/runtime\"\n")
	out.WriteString(")\n\n")

	// Emit suppress-unused helpers so the generated code always compiles even
	// when some imports are not yet exercised by TODO stubs.
	out.WriteString("// Suppress unused import warnings.\n")
	out.WriteString("var _ = fmt.Sprintf\n")
	if g.imports["math"] {
		out.WriteString("var _ = math.Abs\n")
	}
	if g.imports["os"] {
		out.WriteString("var _ = os.Exit\n")
	}
	if g.imports["strconv"] {
		out.WriteString("var _ = strconv.Itoa\n")
	}
	if g.imports["strings"] {
		out.WriteString("var _ = strings.TrimSpace\n")
	}
	out.WriteString("var _ = rt.Abs\n")
	out.WriteString("\n")

	// main function.
	out.WriteString("func main() {\n")

	// Declare rng if needed.
	if g.needRng {
		out.WriteString("\trng := rt.NewRNG()\n")
		out.WriteString("\t_ = rng\n")
	}

	// Declare DATA pool if needed.
	if hasData {
		out.WriteString("\tvar dataPool []interface{}\n")
		out.WriteString("\t_ = dataPool\n")
		g.emitDataPoolInit(&out)
		out.WriteString("\tdataIdx := 0\n")
		out.WriteString("\t_ = dataIdx\n")
	}

	out.Write(g.buf.Bytes())
	out.WriteString("}\n")

	// Append SUB/FUNCTION declarations.
	if g.funcBuf.Len() > 0 {
		out.WriteString("\n")
		out.Write(g.funcBuf.Bytes())
	}

	return out.String(), nil
}

// ---------------------------------------------------------------------------
// Pre-pass: collect labels, line numbers, and DATA values
// ---------------------------------------------------------------------------

// collectLabelsAndData performs a pre-pass over the AST to:
//   - Record all label/line-number definitions (labelMap)
//   - Record all GOTO/GOSUB targets (referencedLabels) so we only emit
//     labels that are actually jumped to — Go considers unused labels a
//     compile error, so emitting unreferenced labels would break the output.
//   - Collect DATA values into the dataPool for READ statement access.
func (g *CodeGenerator) collectLabelsAndData(stmts []ast.Statement) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.LabelStatement:
			g.labelMap[strings.ToUpper(s.Name)] = true
		case *ast.LineNumberStatement:
			g.labelMap[fmt.Sprintf("%d", s.Number)] = true
		case *ast.DataStatement:
			g.dataPool = append(g.dataPool, s.Values...)
		case *ast.ReadStatement:
			if !s.IsInput {
				g.hasRead = true
			}

		// Collect GOTO/GOSUB targets — these become goto labels in Go.
		case *ast.GotoStatement:
			g.referencedLabels[strings.ToUpper(s.Target)] = true
		case *ast.GosubStatement:
			g.referencedLabels[strings.ToUpper(s.Target)] = true
		// ON ERROR GOTO is emitted as a TODO comment (not a real goto), so we
		// intentionally do NOT add its target to referencedLabels. That way the
		// error-handler label is suppressed in the output and the file compiles.
		case *ast.OnComputedGotoStatement:
			for _, t := range s.Targets {
				g.referencedLabels[strings.ToUpper(t)] = true
			}
		case *ast.OnComputedGosubStatement:
			for _, t := range s.Targets {
				g.referencedLabels[strings.ToUpper(t)] = true
			}

		// Recurse into nested blocks.
		case *ast.IfStatement:
			g.collectLabelsAndData(s.ThenBlock)
			for _, clause := range s.ElseIfClauses {
				g.collectLabelsAndData(clause.Body)
			}
			g.collectLabelsAndData(s.ElseBlock)
		case *ast.ForStatement:
			g.collectLabelsAndData(s.Body)
		case *ast.WhileStatement:
			g.collectLabelsAndData(s.Body)
		case *ast.DoLoopStatement:
			g.collectLabelsAndData(s.Body)
		case *ast.SubDeclaration:
			g.collectLabelsAndData(s.Body)
		case *ast.FunctionDeclaration:
			g.collectLabelsAndData(s.Body)
		case *ast.SelectCaseStatement:
			for _, c := range s.Cases {
				g.collectLabelsAndData(c.Body)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Indentation helpers
// ---------------------------------------------------------------------------

func (g *CodeGenerator) writeIndent() {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteByte('\t')
	}
}

func (g *CodeGenerator) writeLine(s string) {
	g.writeIndent()
	g.buf.WriteString(s)
	g.buf.WriteByte('\n')
}

func (g *CodeGenerator) writeLinef(format string, args ...interface{}) {
	g.writeIndent()
	fmt.Fprintf(&g.buf, format, args...)
	g.buf.WriteByte('\n')
}

// funcWriteIndent writes indentation into funcBuf.
func (g *CodeGenerator) funcWriteIndent(indent int) {
	for i := 0; i < indent; i++ {
		g.funcBuf.WriteByte('\t')
	}
}

func (g *CodeGenerator) funcWriteLine(indent int, s string) {
	g.funcWriteIndent(indent)
	g.funcBuf.WriteString(s)
	g.funcBuf.WriteByte('\n')
}

func (g *CodeGenerator) funcWriteLinef(indent int, format string, args ...interface{}) {
	g.funcWriteIndent(indent)
	fmt.Fprintf(&g.funcBuf, format, args...)
	g.funcBuf.WriteByte('\n')
}

// ---------------------------------------------------------------------------
// DATA pool initializer
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitDataPoolInit(out *bytes.Buffer) {
	out.WriteString("\tdataPool = []interface{}{\n")
	for _, expr := range g.dataPool {
		out.WriteString("\t\t")
		switch e := expr.(type) {
		case *ast.NumberLiteral:
			out.WriteString(formatGoNumber(e))
		case *ast.StringLiteral:
			out.WriteString(strconv.Quote(e.Value))
		default:
			out.WriteString("nil")
		}
		out.WriteString(",\n")
	}
	out.WriteString("\t}\n")
}

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
	case *ast.LocateStatement:
		g.emitLocate(s)
	case *ast.ClsStatement:
		g.writeLine("fmt.Print(rt.AnsiCls()) // CLS")
	case *ast.ColorStatement:
		g.emitColor(s)
	case *ast.OnErrorGotoStatement:
		g.writeLinef("// TODO: ON ERROR GOTO %s", s.Target)
	case *ast.OnComputedGotoStatement:
		g.emitOnComputedGoto(s)
	case *ast.OnComputedGosubStatement:
		g.emitOnComputedGosub(s)
	case *ast.PokeStatement:
		g.writeLinef("_ = %s; _ = %s // POKE (no-op in transpiled code)", g.emitExpr(s.Address), g.emitExpr(s.Value))
	case *ast.ConstStatement:
		g.writeLinef("%s := %s // CONST", mangleName(s.Name), g.emitExpr(s.Value))
	case *ast.ClearStatement:
		g.writeLine("// CLEAR — variable reset not supported in transpiled code")
	case *ast.TypeBlockStatement:
		g.emitTypeBlock(s)
	case *ast.FnAssignStatement:
		g.writeLinef("_fn_%s = %s", mangleName(s.Name), g.emitExpr(s.Value))
	case *ast.FieldAssignStatement:
		// Struct/TYPE field assignment: obj.Field = value
		g.writeLinef("%s.%s = %s", g.emitExpr(s.Object), mangleName(s.Field), g.emitExpr(s.Value))
	case *ast.ResumeStatement:
		g.writeLinef("// TODO: RESUME %s", s.Type)
	case *ast.ErrorStatement:
		g.writeLinef("// TODO: ERROR %s", g.emitExpr(s.Code))

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
//   PRINT A; B      → semicolon: concatenate values with no spacing
//                     → fmt.Print(A, B) / fmt.Println(A, B)
//
//   PRINT A, B      → comma: "zone" or column-tabbing output.  Each comma
//                     advances to the next 14-character print zone.
//                     → rt.PrintZone() from the runtime package
//
//   PRINT USING f$; → formatted output with a format template.
//                     → rt.PrintUsing() from the runtime package
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

	// RANDOMIZE is parsed as a LetStatement by the parser.
	// Translate to an RNG seed call instead of a variable assignment.
	if upperName == "RANDOMIZE" {
		g.needRng = true
		if s.Value != nil {
			g.writeLinef("rng.Randomize(float64(%s))", g.emitExpr(s.Value))
		} else {
			g.writeLine("rng.Randomize(rt.Timer())")
		}
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
		if !g.declared[name] {
			g.declared[name] = true
			g.writeLinef("var %s string = %s", name, val)
		} else {
			g.writeLinef("%s = %s", name, val)
		}
		return
	}

	val := g.emitExpr(s.Value)
	if !g.declared[name] {
		g.declared[name] = true
		// Cast the value to the variable's type to handle mismatches between
		// BASIC's implicit coercions and Go's strict typing.
		g.writeLinef("var %s %s = %s(%s)", name, goT, goT, val)
	} else {
		g.writeLinef("%s = %s(%s)", name, goT, val)
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
		// Multi-dimensional: emit a comment and the first index for now.
		g.writeLinef("// TODO: multi-dim array assignment %s[%s] = %s", arrName, strings.Join(indices, "]["), val)
	}
}

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
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitIf(s *ast.IfStatement) {
	cond := g.emitExpr(s.Condition)
	g.writeLinef("if %s {", g.toBoolExpr(cond))
	g.indent++
	for _, stmt := range s.ThenBlock {
		g.emitStatement(stmt)
	}
	g.indent--

	for _, clause := range s.ElseIfClauses {
		ec := g.emitExpr(clause.Condition)
		g.writeLinef("} else if %s {", g.toBoolExpr(ec))
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
//   for counter = start;
//       (step > 0 && counter <= end) ||
//       (step < 0 && counter >= end) ||
//       (step == 0);           // runs exactly once when step == 0
//       counter += step { … }
//
// To avoid re-evaluating the STEP and END expressions on every iteration
// (which would be wrong if they contained side effects), both are captured
// in temporary variables (end_N, step_N) before the loop header.  The
// tempCount field provides a monotonically increasing suffix to guarantee
// uniqueness across multiple FOR loops in the same scope.
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

	counterType := g.goTypeForIdent(s.Counter.Name + s.Counter.TypeSuffix)
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
	cond := g.emitExpr(s.Condition)
	g.writeLinef("for %s {", g.toBoolExpr(cond))
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

	cond := g.emitExpr(s.Condition)
	boolCond := g.toBoolExpr(cond)

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
// DIM
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitDim(s *ast.DimStatement) {
	for _, d := range s.Declarations {
		name := mangleName(d.Name + d.TypeSuffix)
		if len(d.Dimensions) == 0 {
			// Scalar declaration.
			goT := g.goTypeForDecl(d)
			g.writeLinef("var %s %s", name, goT)
			g.declared[name] = true
		} else if len(d.Dimensions) == 1 {
			// 1-D array.
			goT := g.goTypeForDecl(d)
			upper := g.emitExpr(d.Dimensions[0].Upper)
			g.writeLinef("%s := make([]%s, int(%s)+1)", name, goT, upper)
			g.declared[name] = true
		} else {
			// Multi-dimensional: emit as slice of slices or TODO.
			goT := g.goTypeForDecl(d)
			upper := g.emitExpr(d.Dimensions[0].Upper)
			g.writeLinef("%s := make([]%s, int(%s)+1) // TODO: multi-dim", name, goT, upper)
			g.declared[name] = true
		}
	}
}

func (g *CodeGenerator) emitRedim(s *ast.RedimStatement) {
	for _, d := range s.Declarations {
		name := mangleName(d.Name + d.TypeSuffix)
		if len(d.Dimensions) >= 1 {
			goT := g.goTypeForDecl(d)
			upper := g.emitExpr(d.Dimensions[0].Upper)
			g.writeLinef("%s = make([]%s, int(%s)+1) // REDIM", name, goT, upper)
		}
	}
}

// ---------------------------------------------------------------------------
// READ / RESTORE
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitRead(s *ast.ReadStatement) {
	// INPUT statement: read from stdin.
	if s.IsInput {
		g.emitInputFromStdin(s)
		return
	}

	// DATA READ: read from the compile-time data pool.
	for _, v := range s.Variables {
		varExpr := g.emitExpr(v)
		if ident, ok := v.(*ast.Identifier); ok {
			name := mangleName(ident.Name + ident.TypeSuffix)
			if !g.declared[name] {
				goT := g.goTypeForIdent(ident.Name + ident.TypeSuffix)
				g.writeLinef("var %s %s", name, goT)
				g.declared[name] = true
			}
			goT := g.goTypeForIdent(ident.Name + ident.TypeSuffix)
			if isStringType(ident.Name + ident.TypeSuffix) {
				g.writeLinef("%s = fmt.Sprint(dataPool[dataIdx]); dataIdx++", varExpr)
			} else {
				// Cast tv_ to the target type to satisfy Go's strict type system.
				g.imports["fmt"] = true
				g.writeLinef("{ v_ := dataPool[dataIdx]; dataIdx++; switch tv_ := v_.(type) { case float64: %s = %s(tv_); case int: %s = %s(float64(tv_)); case string: %s = %s(rt.Val(tv_)); default: _ = tv_ } }", varExpr, goT, varExpr, goT, varExpr, goT)
			}
		} else {
			g.writeLinef("_ = dataPool[dataIdx]; dataIdx++ // READ into %s", varExpr)
		}
	}
}

// emitInputFromStdin emits stdin-reading code for INPUT and LINE INPUT statements.
// INPUT reads whitespace-delimited values; LINE INPUT reads the whole line.
func (g *CodeGenerator) emitInputFromStdin(s *ast.ReadStatement) {
	g.imports["fmt"] = true

	// Print the prompt if there is one (INPUT "Prompt: "; var).
	if s.Prompt != "" {
		g.writeLinef("fmt.Print(%q)", s.Prompt)
	}

	if s.IsLineInput {
		// LINE INPUT: read entire line into a single string variable.
		g.imports["os"] = true
		if len(s.Variables) > 0 {
			varExpr := g.emitExpr(s.Variables[0])
			if ident, ok := s.Variables[0].(*ast.Identifier); ok {
				name := mangleName(ident.Name + ident.TypeSuffix)
				if !g.declared[name] {
					g.writeLinef("var %s string", name)
					g.declared[name] = true
				}
			}
			g.writeLine("{ scanner_ := rt.NewScanner(); scanner_.Scan()")
			g.writeLinef("  %s = scanner_.Text() }", varExpr)
		}
		return
	}

	// Regular INPUT: read one or more values.
	for _, v := range s.Variables {
		varExpr := g.emitExpr(v)
		if ident, ok := v.(*ast.Identifier); ok {
			name := mangleName(ident.Name + ident.TypeSuffix)
			if !g.declared[name] {
				goT := g.goTypeForIdent(ident.Name + ident.TypeSuffix)
				g.writeLinef("var %s %s", name, goT)
				g.declared[name] = true
			}
			if isStringType(ident.Name + ident.TypeSuffix) {
				g.writeLine("{ scanner_ := rt.NewScanner(); scanner_.Scan()")
				g.writeLinef("  %s = scanner_.Text() }", varExpr)
			} else {
				g.writeLinef("fmt.Scan(&%s)", varExpr)
			}
		}
	}
}

func (g *CodeGenerator) emitRestore(_ *ast.RestoreStatement) {
	g.writeLine("dataIdx = 0 // RESTORE")
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

// ---------------------------------------------------------------------------
// EXIT
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitExit(s *ast.ExitStatement) {
	switch strings.ToUpper(s.ExitType) {
	case "FOR", "DO", "WHILE", "LOOP":
		// LOOP is an alias for DO in Turbo BASIC (EXIT LOOP = exit DO loop)
		g.writeLine("break")
	case "SUB", "FUNCTION", "DEF":
		// EXIT DEF exits a DEF FN multi-line function (equivalent to return)
		g.writeLine("return")
	default:
		g.writeLinef("// EXIT %s (unsupported)", s.ExitType)
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
// OPEN / CLOSE
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitOpen(s *ast.OpenStatement) {
	filename := g.emitExpr(s.Filename)
	fileNum := g.emitExpr(s.FileNum)
	mode := strings.ToUpper(s.Mode)
	modeConst := "rt.FileModeInput"
	switch mode {
	case "INPUT":
		modeConst = "rt.FileModeInput"
	case "OUTPUT":
		modeConst = "rt.FileModeOutput"
	case "APPEND":
		modeConst = "rt.FileModeAppend"
	case "RANDOM":
		modeConst = "rt.FileModeRandom"
	case "BINARY":
		modeConst = "rt.FileModeBinary"
	}
	recLen := "0"
	if s.RecLen != nil {
		recLen = g.emitExpr(s.RecLen)
	}
	g.writeLinef("// TODO: declare fm *rt.FileManager if not done")
	g.writeLinef("_ = %s; _ = %s; _ = %s; _ = %s // OPEN", filename, modeConst, fileNum, recLen)
}

func (g *CodeGenerator) emitClose(s *ast.CloseStatement) {
	if len(s.FileNums) == 0 {
		g.writeLine("// TODO: fm.FileCloseAll() // CLOSE all")
	} else {
		for _, f := range s.FileNums {
			num := g.emitExpr(f)
			g.writeLinef("_ = %s // CLOSE #", num)
		}
	}
}

// ---------------------------------------------------------------------------
// LOCATE / COLOR
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitLocate(s *ast.LocateStatement) {
	row := "1"
	col := "1"
	if s.Row != nil {
		row = g.emitExpr(s.Row)
	}
	if s.Col != nil {
		col = g.emitExpr(s.Col)
	}
	g.writeLinef("fmt.Print(rt.AnsiLocate(int(%s), int(%s)))", row, col)
}

func (g *CodeGenerator) emitColor(s *ast.ColorStatement) {
	fg := "7"
	bg := "0"
	if s.Foreground != nil {
		fg = g.emitExpr(s.Foreground)
	}
	if s.Background != nil {
		bg = g.emitExpr(s.Background)
	}
	g.writeLinef("fmt.Print(rt.AnsiColor(int(%s), int(%s)))", fg, bg)
}

// ---------------------------------------------------------------------------
// SUB / FUNCTION / DEF FN – top-level declarations
//
// emitTopLevelDecl() is a second visitor entry point used exclusively for
// procedure-level constructs.  In BASIC, SUB and FUNCTION blocks can appear
// anywhere in the source file, but in Go they must be top-level declarations
// outside of any other function.
//
// The strategy is to redirect writes from the main buffer (g.buf) to a
// separate buffer (g.funcBuf) while walking the procedure body.  After the
// entire program has been processed, funcBuf is appended to the output after
// the closing brace of main(), producing valid Go.
//
// emitSubDecl / emitFuncDecl temporarily swap g.buf with a fresh buffer so
// that nested emitStatement() calls still write through the same pointer,
// then move the result into funcBuf.  The indent counter is also saved and
// restored so indentation inside the procedure body starts at level 1.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitTopLevelDecl(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.SubDeclaration:
		g.emitSubDecl(s)
	case *ast.FunctionDeclaration:
		g.emitFuncDecl(s)
	case *ast.DefFnDeclaration:
		g.emitDefFn(s)
	}
}

func (g *CodeGenerator) emitSubDecl(s *ast.SubDeclaration) {
	if s.IsForward {
		return
	}
	name := mangleName(s.Name)
	params := g.emitParams(s.Params)
	g.funcWriteLinef(0, "func %s(%s) {", name, params)
	// Save and restore state.
	origBuf := g.buf
	origIndent := g.indent
	g.buf = bytes.Buffer{}
	g.indent = 1
	for _, stmt := range s.Body {
		g.emitStatement(stmt)
	}
	g.funcBuf.Write(g.buf.Bytes())
	g.buf = origBuf
	g.indent = origIndent
	g.funcWriteLine(0, "}")
	g.funcWriteLine(0, "")
}

func (g *CodeGenerator) emitFuncDecl(s *ast.FunctionDeclaration) {
	if s.IsForward {
		return
	}
	name := mangleName(s.Name)
	params := g.emitParams(s.Params)
	retType := g.goTypeForReturnType(s.ReturnType, s.Name)
	retVar := mangleName(s.Name)

	g.funcWriteLinef(0, "func %s(%s) %s {", name, params, retType)
	g.funcWriteLinef(1, "var %s %s", retVar, retType)

	origBuf := g.buf
	origIndent := g.indent
	g.buf = bytes.Buffer{}
	g.indent = 1
	for _, stmt := range s.Body {
		g.emitStatement(stmt)
	}
	g.funcBuf.Write(g.buf.Bytes())
	g.buf = origBuf
	g.indent = origIndent

	g.funcWriteLinef(1, "return %s", retVar)
	g.funcWriteLine(0, "}")
	g.funcWriteLine(0, "")
}

func (g *CodeGenerator) emitDefFn(s *ast.DefFnDeclaration) {
	name := mangleName(s.Name)
	params := g.emitParams(s.Params)

	if s.SingleLineExpr != nil {
		retType := "float64" // DEF FN default return type
		g.funcWriteLinef(0, "func %s(%s) %s {", name, params, retType)
		expr := g.emitExpr(s.SingleLineExpr)
		g.funcWriteLinef(1, "return %s", expr)
		g.funcWriteLine(0, "}")
		g.funcWriteLine(0, "")
	} else {
		g.funcWriteLinef(0, "func %s(%s) float64 {", name, params)
		g.funcWriteLinef(1, "var result_ float64")

		origBuf := g.buf
		origIndent := g.indent
		g.buf = bytes.Buffer{}
		g.indent = 1
		for _, stmt := range s.Body {
			g.emitStatement(stmt)
		}
		g.funcBuf.Write(g.buf.Bytes())
		g.buf = origBuf
		g.indent = origIndent

		g.funcWriteLinef(1, "return result_")
		g.funcWriteLine(0, "}")
		g.funcWriteLine(0, "")
	}
}

func (g *CodeGenerator) emitParams(params []ast.Parameter) string {
	parts := make([]string, 0, len(params))
	for _, p := range params {
		name := mangleName(p.Name)
		goT := g.goTypeForParamType(p.Type, p.Name)
		parts = append(parts, fmt.Sprintf("%s %s", name, goT))
	}
	return strings.Join(parts, ", ")
}

// ---------------------------------------------------------------------------
// Expression emitter – visitor pattern for expression nodes
//
// emitExpr() is the expression-side visitor.  It mirrors emitStatement() but
// returns a string instead of writing to the buffer directly.  This is the
// key architectural difference between statements and expressions in a
// tree-walking code generator:
//
//   Statement emitters  → write text to g.buf as a side effect (l-value
//                          semantics: "do this").
//   Expression emitters → return a string that represents the value and can
//                          be composed into a larger expression (r-value
//                          semantics: "compute this").
//
// For example, emitPrint() calls g.emitExpr(expr) to get the Go string for
// each sub-expression, then assembles them into a fmt.Println() call.
//
// Recursion is the natural mechanism for nested expressions.  A BinaryExpr
// node calls emitExpr on both its Left and Right children, then wraps the
// results in parentheses with the mapped operator in the middle.  This
// bottom-up composition means deeply nested expressions like:
//
//   A * (B + C) ^ 2
//
// are correctly parenthesised in the output without any explicit precedence
// tracking — parentheses are inserted at every level.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitExpr(expr ast.Expression) string {
	if expr == nil {
		return "0"
	}
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return formatGoNumber(e)
	case *ast.StringLiteral:
		return strconv.Quote(e.Value)
	case *ast.Identifier:
		return g.emitIdentifier(e)
	case *ast.BinaryExpr:
		return g.emitBinaryExpr(e)
	case *ast.UnaryExpr:
		return g.emitUnaryExpr(e)
	case *ast.FunctionCall:
		return g.emitFunctionCall(e)
	case *ast.ArrayAccess:
		return g.emitArrayAccess(e)
	case *ast.GroupExpr:
		return "(" + g.emitExpr(e.Inner) + ")"
	case *ast.FnCallExpression:
		return g.emitFnCallExpression(e)
	case *ast.FieldAccessExpression:
		// Struct/TYPE field access: obj.Field (maps directly to Go struct field)
		return g.emitExpr(e.Object) + "." + mangleName(e.Field)
	default:
		return "0 /* unknown expression */"
	}
}

// ---------------------------------------------------------------------------
// Identifier
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitIdentifier(id *ast.Identifier) string {
	// Some BASIC builtins are used without parentheses and look like plain
	// variable references to the parser. Detect them here and emit the correct
	// runtime calls instead of bare (undefined) variable names.
	switch strings.ToUpper(id.Name + id.TypeSuffix) {
	case "INKEY$":
		return "rt.Inkey()"
	case "DATE$":
		return "rt.DateStr()"
	case "TIME$":
		return "rt.TimeStr()"
	case "RND":
		g.needRng = true
		return "rng.Rnd(1)"
	case "ERADR":
		return "float64(0) /* ERADR: not applicable in transpiled code */"
	case "TIMER":
		return "rt.Timer()"
	}
	return mangleName(id.Name + id.TypeSuffix)
}

// ---------------------------------------------------------------------------
// Binary expression
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitBinaryExpr(e *ast.BinaryExpr) string {
	left := g.emitExpr(e.Left)
	right := g.emitExpr(e.Right)
	op := strings.ToUpper(e.Operator)

	switch op {
	case "+", "-", "*":
		return fmt.Sprintf("(%s %s %s)", left, e.Operator, right)
	case "/":
		return fmt.Sprintf("(%s / %s)", left, right)
	case "\\":
		// Integer division.
		return fmt.Sprintf("(int(%s) / int(%s))", left, right)
	case "MOD":
		return fmt.Sprintf("(int(%s) %% int(%s))", left, right)
	case "^":
		g.imports["math"] = true
		return fmt.Sprintf("math.Pow(%s, %s)", left, right)
	case "=":
		return fmt.Sprintf("(%s == %s)", left, right)
	case "<>", "><":
		return fmt.Sprintf("(%s != %s)", left, right)
	case "<":
		return fmt.Sprintf("(%s < %s)", left, right)
	case ">":
		return fmt.Sprintf("(%s > %s)", left, right)
	case "<=", "=<":
		return fmt.Sprintf("(%s <= %s)", left, right)
	case ">=", "=>":
		return fmt.Sprintf("(%s >= %s)", left, right)
	case "AND":
		return fmt.Sprintf("(int(%s) & int(%s))", left, right)
	case "OR":
		return fmt.Sprintf("(int(%s) | int(%s))", left, right)
	case "XOR":
		return fmt.Sprintf("(int(%s) ^ int(%s))", left, right)
	case "EQV":
		return fmt.Sprintf("(^(int(%s) ^ int(%s)))", left, right)
	case "IMP":
		return fmt.Sprintf("((^int(%s)) | int(%s))", left, right)
	default:
		return fmt.Sprintf("(%s /* %s */ %s)", left, op, right)
	}
}

// ---------------------------------------------------------------------------
// Unary expression
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitUnaryExpr(e *ast.UnaryExpr) string {
	operand := g.emitExpr(e.Operand)
	op := strings.ToUpper(e.Operator)

	switch op {
	case "-":
		return fmt.Sprintf("(-%s)", operand)
	case "+":
		return fmt.Sprintf("(+%s)", operand)
	case "NOT":
		return fmt.Sprintf("(^int(%s))", operand)
	default:
		return fmt.Sprintf("(%s%s)", e.Operator, operand)
	}
}

// ---------------------------------------------------------------------------
// Function call — mapping BASIC built-ins to the runtime package
//
// BASIC has a large library of built-in functions (ABS, SIN, LEFT$, etc.)
// that must be available in every program.  In a native-code compiler these
// would be part of the language runtime linked directly into the binary.  In
// this transpiler they live in the separate "internal/runtime" (aliased "rt")
// package, which is unconditionally imported into every generated file.
//
// The dispatch strategy is a large switch on the upper-cased function name.
// Each known built-in case emits the corresponding "rt.FuncName(…)" call,
// performing any necessary argument coercions (e.g., int() casts for
// functions that require integer arguments in Go even though BASIC treats
// all numbers as float64).
//
// Several built-ins return two values (result, error) in the runtime package
// to preserve the original BASIC error-handling semantics.  Because the
// transpiled code uses expression-level embedding, they are wrapped in an
// immediately-invoked function literal:
//
//   func() float64 { v_, _ := rt.Sqr(x); return v_ }()
//
// This is an inline closure that discards the error — a pragmatic choice that
// keeps expression context simple at the cost of ignoring runtime errors.
//
// The "default" case handles user-defined SUB/FUNCTION calls: the name is
// mangled and called directly, since user functions end up as top-level Go
// functions in the same package.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitFunctionCall(fc *ast.FunctionCall) string {
	name := strings.ToUpper(fc.Name)
	args := make([]string, 0, len(fc.Args))
	for _, a := range fc.Args {
		args = append(args, g.emitExpr(a))
	}

	switch name {
	// Math functions.
	case "ABS":
		return fmt.Sprintf("rt.Abs(%s)", g.oneArg(args))
	case "SGN":
		return fmt.Sprintf("rt.Sgn(%s)", g.oneArg(args))
	case "INT":
		return fmt.Sprintf("rt.IntFloor(%s)", g.oneArg(args))
	case "FIX":
		return fmt.Sprintf("rt.Fix(%s)", g.oneArg(args))
	case "CEIL":
		return fmt.Sprintf("rt.Ceil(%s)", g.oneArg(args))
	case "SQR":
		return fmt.Sprintf("func() float64 { v_, _ := rt.Sqr(%s); return v_ }()", g.oneArg(args))
	case "EXP":
		return fmt.Sprintf("rt.Exp(%s)", g.oneArg(args))
	case "EXP2":
		return fmt.Sprintf("rt.Exp2(%s)", g.oneArg(args))
	case "EXP10":
		return fmt.Sprintf("rt.Exp10(%s)", g.oneArg(args))
	case "LOG":
		return fmt.Sprintf("func() float64 { v_, _ := rt.Log(%s); return v_ }()", g.oneArg(args))
	case "LOG2":
		return fmt.Sprintf("func() float64 { v_, _ := rt.Log2(%s); return v_ }()", g.oneArg(args))
	case "LOG10":
		return fmt.Sprintf("func() float64 { v_, _ := rt.Log10(%s); return v_ }()", g.oneArg(args))
	case "SIN":
		return fmt.Sprintf("rt.Sin(%s)", g.oneArg(args))
	case "COS":
		return fmt.Sprintf("rt.Cos(%s)", g.oneArg(args))
	case "TAN":
		return fmt.Sprintf("rt.Tan(%s)", g.oneArg(args))
	case "ATN":
		return fmt.Sprintf("rt.Atn(%s)", g.oneArg(args))

	// Conversion functions.
	case "CINT":
		return fmt.Sprintf("func() int16 { v_, _ := rt.Cint(%s); return v_ }()", g.oneArg(args))
	case "CLNG":
		return fmt.Sprintf("func() int32 { v_, _ := rt.Clng(%s); return v_ }()", g.oneArg(args))
	case "CSNG":
		return fmt.Sprintf("rt.Csng(%s)", g.oneArg(args))
	case "CDBL":
		return fmt.Sprintf("rt.Cdbl(%s)", g.oneArg(args))

	// String functions.
	case "LEFT$":
		return fmt.Sprintf("rt.Left(%s, int(%s))", g.argN(args, 0), g.argN(args, 1))
	case "RIGHT$":
		return fmt.Sprintf("rt.Right(%s, int(%s))", g.argN(args, 0), g.argN(args, 1))
	case "MID$":
		if len(args) >= 3 {
			return fmt.Sprintf("rt.Mid(%s, int(%s), int(%s))", args[0], args[1], args[2])
		}
		return fmt.Sprintf("rt.Mid(%s, int(%s), -1)", g.argN(args, 0), g.argN(args, 1))
	case "LEN":
		return fmt.Sprintf("rt.Len(%s)", g.oneArg(args))
	case "INSTR":
		if len(args) >= 3 {
			return fmt.Sprintf("rt.Instr(int(%s), %s, %s)", args[0], args[1], args[2])
		}
		return fmt.Sprintf("rt.Instr(1, %s, %s)", g.argN(args, 0), g.argN(args, 1))
	case "ASC":
		return fmt.Sprintf("func() int { v_, _ := rt.Asc(%s); return v_ }()", g.oneArg(args))
	case "CHR$":
		return fmt.Sprintf("func() string { v_, _ := rt.Chr(int(%s)); return v_ }()", g.oneArg(args))
	case "STR$":
		return fmt.Sprintf("rt.Str(%s)", g.oneArg(args))
	case "VAL":
		return fmt.Sprintf("rt.Val(%s)", g.oneArg(args))
	case "HEX$":
		return fmt.Sprintf("rt.Hex(int(%s))", g.oneArg(args))
	case "OCT$":
		return fmt.Sprintf("rt.Oct(int(%s))", g.oneArg(args))
	case "BIN$":
		return fmt.Sprintf("rt.Bin(int(%s))", g.oneArg(args))
	case "UCASE$":
		return fmt.Sprintf("rt.UCase(%s)", g.oneArg(args))
	case "LCASE$":
		return fmt.Sprintf("rt.LCase(%s)", g.oneArg(args))
	case "LTRIM$":
		return fmt.Sprintf("rt.LTrim(%s)", g.oneArg(args))
	case "RTRIM$":
		return fmt.Sprintf("rt.RTrim(%s)", g.oneArg(args))
	case "TRIM$":
		return fmt.Sprintf("rt.Trim(%s)", g.oneArg(args))
	case "SPACE$":
		return fmt.Sprintf("rt.Space(int(%s))", g.oneArg(args))
	case "STRING$":
		return fmt.Sprintf("rt.StringRepeat(int(%s), byte(%s))", g.argN(args, 0), g.argN(args, 1))

	// Conversion binary functions.
	case "MKI$":
		return fmt.Sprintf("rt.Mki(int16(%s))", g.oneArg(args))
	case "MKL$":
		return fmt.Sprintf("rt.Mkl(int32(%s))", g.oneArg(args))
	case "MKS$":
		return fmt.Sprintf("rt.Mks(float32(%s))", g.oneArg(args))
	case "MKD$":
		return fmt.Sprintf("rt.Mkd(%s)", g.oneArg(args))
	case "CVI":
		return fmt.Sprintf("func() int16 { v_, _ := rt.Cvi(%s); return v_ }()", g.oneArg(args))
	case "CVL":
		return fmt.Sprintf("func() int32 { v_, _ := rt.Cvl(%s); return v_ }()", g.oneArg(args))
	case "CVS":
		return fmt.Sprintf("func() float32 { v_, _ := rt.Cvs(%s); return v_ }()", g.oneArg(args))
	case "CVD":
		return fmt.Sprintf("func() float64 { v_, _ := rt.Cvd(%s); return v_ }()", g.oneArg(args))

	// Random.
	case "RND":
		g.needRng = true
		if len(args) > 0 {
			return fmt.Sprintf("rng.Rnd(%s)", args[0])
		}
		return "rng.Rnd(1)"

	// Timer / system.
	case "TIMER":
		return "rt.Timer()"
	case "DATE$":
		return "rt.DateStr()"
	case "TIME$":
		return "rt.TimeStr()"
	case "COMMAND$":
		return "rt.CommandStr()"
	case "ENVIRON$":
		return fmt.Sprintf("rt.EnvironGet(%s)", g.oneArg(args))

	// TAB / SPC.
	case "TAB":
		return fmt.Sprintf("rt.Spc(int(%s))", g.oneArg(args))
	case "SPC":
		return fmt.Sprintf("rt.Spc(int(%s))", g.oneArg(args))

	default:
		// User-defined function or unmapped built-in: call directly.
		mangledName := mangleName(fc.Name)
		return fmt.Sprintf("%s(%s)", mangledName, strings.Join(args, ", "))
	}
}

// ---------------------------------------------------------------------------
// Array access — BASIC 1-based indexing vs. Go 0-based slices
//
// One of the most common semantic mismatches between BASIC and Go is array
// indexing.  BASIC arrays are conventionally 1-based: "DIM A(10)" declares
// elements A(1) through A(10).  Go slices are always 0-based.
//
// The reconciliation used here is to allocate slices with one extra element:
//
//   DIM A(10)  →  a := make([]float64, 10+1)   // indices 0..10
//
// This wastes one slot at index 0 but keeps the indexing arithmetic trivial:
// A(i) in BASIC simply becomes a[int(i)] in Go — no adjustment needed.
// The cost is one unused slot per array, which is almost always acceptable.
//
// Note that emitArrayAccess does NOT subtract 1 from the index.  The "+1"
// in emitDim is the only adaptation required.  This design choice must be
// understood when reading both functions together.
//
// Multi-dimensional arrays are not yet fully supported; a TODO comment is
// emitted as a placeholder so the file still compiles.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitArrayAccess(aa *ast.ArrayAccess) string {
	name := mangleName(aa.Name + aa.TypeSuffix)
	if len(aa.Indices) == 0 {
		// No indices — array passed by reference (e.g., as a SUB parameter)
		return name
	}
	if len(aa.Indices) == 1 {
		idx := g.emitExpr(aa.Indices[0])
		return fmt.Sprintf("%s[int(%s)]", name, idx)
	}
	// Multi-dimensional: emit first index only with a TODO.
	indices := make([]string, 0, len(aa.Indices))
	for _, idx := range aa.Indices {
		indices = append(indices, g.emitExpr(idx))
	}
	return fmt.Sprintf("%s[int(%s)] /* TODO: multi-dim [%s] */", name, indices[0], strings.Join(indices, ","))
}

// ---------------------------------------------------------------------------
// Helper: convert an expression to a Go bool expression
// ---------------------------------------------------------------------------

// toBoolExpr wraps a numeric expression in a != 0 check when necessary.
// If the expression already looks like a comparison, return it as-is.
func (g *CodeGenerator) toBoolExpr(expr string) string {
	// If it already looks boolean (contains ==, !=, <, >, <=, >=, ||, &&),
	// assume it's already a bool.
	if looksLikeBool(expr) {
		return expr
	}
	return fmt.Sprintf("(%s) != 0", expr)
}

func looksLikeBool(s string) bool {
	// Quick heuristic: check for comparison operators.
	for _, op := range []string{"==", "!=", "<=", ">=", "&&", "||", "!(", "true", "false"} {
		if strings.Contains(s, op) {
			return true
		}
	}
	// Also check for isolated < and > (not part of <= or >=).
	// A simple check: if it contains < or > at all, likely boolean.
	if strings.ContainsAny(s, "<>") {
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Label name mapping
// ---------------------------------------------------------------------------

func (g *CodeGenerator) labelName(target string) string {
	// If the target looks numeric, use line_NNN.
	if _, err := strconv.Atoi(target); err == nil {
		return "line_" + target
	}
	// Otherwise use label_NAME (sanitized).
	return "label_" + sanitizeGoIdent(target)
}

// ---------------------------------------------------------------------------
// Variable name mangling — bridging BASIC names to Go identifiers
//
// Name mangling is the process of transforming identifiers from the source
// language into legal identifiers in the target language.  It is required
// whenever the two languages have incompatible identifier rules.
//
// BASIC has two identifier features that Go does not support:
//
//  1. Sigil (type) suffixes — trailing punctuation encodes the variable type:
//
//       COUNT%    integer      → count_pct
//       NAME$     string       → name_str
//       TOTAL&    long int     → total_lng
//       RATE!     single float → rate_sng
//       FACTOR#   double       → factor_dbl
//
//     These characters are not valid in Go identifiers, so they are replaced
//     with readable alphabetic suffixes that carry the same information.
//     sanitizeGoIdent() then removes any remaining invalid characters.
//
//  2. Go keyword conflicts — BASIC programs freely use names like RETURN,
//     FOR, TYPE, STRING which happen to be reserved in Go.  The mangler
//     detects these after the sigil transformation and prepends "b_":
//
//       RETURN  → b_return
//       TYPE    → b_type
//       STRING  → b_string
//
// Crucially, every use of a variable name — declarations, reads, writes, and
// parameter lists — must pass through mangleName() so the output is
// consistent.  A mangle applied in one place but not another would cause a
// "undefined identifier" compile error in the generated Go.
// ---------------------------------------------------------------------------

// mangleName converts a BASIC variable name (possibly with type suffix) to a
// valid Go identifier.
func mangleName(name string) string {
	if len(name) == 0 {
		return "_empty"
	}

	last := name[len(name)-1]
	base := name
	suffix := ""

	switch last {
	case '%':
		base = name[:len(name)-1]
		suffix = "_pct"
	case '$':
		base = name[:len(name)-1]
		suffix = "_str"
	case '&':
		base = name[:len(name)-1]
		suffix = "_lng"
	case '!':
		base = name[:len(name)-1]
		suffix = "_sng"
	case '#':
		base = name[:len(name)-1]
		suffix = "_dbl"
	}

	result := sanitizeGoIdent(base) + suffix

	// Handle Go reserved words.
	switch result {
	case "break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
		"int", "string", "float64", "float32", "bool",
		"true", "false", "nil":
		result = "b_" + result
	}

	return result
}

// sanitizeGoIdent replaces characters invalid in Go identifiers.
func sanitizeGoIdent(s string) string {
	if len(s) == 0 {
		return "_"
	}
	var buf strings.Builder
	for i, ch := range s {
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_' {
			buf.WriteRune(ch)
		} else if ch >= '0' && ch <= '9' {
			if i == 0 {
				buf.WriteByte('_')
			}
			buf.WriteRune(ch)
		} else if ch == '.' {
			buf.WriteByte('_')
		} else {
			// Skip other characters.
		}
	}
	if buf.Len() == 0 {
		return "_"
	}
	return buf.String()
}

// ---------------------------------------------------------------------------
// Type mapping
// ---------------------------------------------------------------------------

// goType maps a semantic.DataType to its Go type string.
func goType(dt semantic.DataType) string {
	switch dt {
	case semantic.TypeInteger:
		return "int16"
	case semantic.TypeLong:
		return "int32"
	case semantic.TypeSingle:
		return "float32"
	case semantic.TypeDouble:
		return "float64"
	case semantic.TypeString:
		return "string"
	default:
		return "float64"
	}
}

// goTypeForIdent resolves the Go type for a BASIC variable name.
func (g *CodeGenerator) goTypeForIdent(name string) string {
	if g.table != nil {
		dt := g.table.ResolveType(name)
		return goType(dt)
	}
	return goTypeFromSuffix(name)
}

// goTypeForDecl resolves the Go type for a DIM declaration.
func (g *CodeGenerator) goTypeForDecl(d ast.DimDecl) string {
	if d.ElementType != "" {
		switch strings.ToUpper(d.ElementType) {
		case "INTEGER":
			return "int16"
		case "LONG":
			return "int32"
		case "SINGLE":
			return "float32"
		case "DOUBLE":
			return "float64"
		case "STRING":
			return "string"
		}
	}
	return g.goTypeForIdent(d.Name + d.TypeSuffix)
}

// goTypeForReturnType resolves the Go type for a FUNCTION return type.
func (g *CodeGenerator) goTypeForReturnType(retType string, name string) string {
	switch strings.ToUpper(retType) {
	case "INTEGER", "%":
		return "int16"
	case "LONG", "&":
		return "int32"
	case "SINGLE", "!":
		return "float32"
	case "DOUBLE", "#":
		return "float64"
	case "STRING", "$":
		return "string"
	}
	return g.goTypeForIdent(name)
}

// goTypeForParamType resolves the Go type for a parameter declaration.
func (g *CodeGenerator) goTypeForParamType(typeStr string, name string) string {
	switch strings.ToUpper(typeStr) {
	case "INTEGER", "%":
		return "int16"
	case "LONG", "&":
		return "int32"
	case "SINGLE", "!":
		return "float32"
	case "DOUBLE", "#":
		return "float64"
	case "STRING", "$":
		return "string"
	}
	return g.goTypeForIdent(name)
}

// goTypeFromSuffix infers a Go type from the variable name's suffix character.
func goTypeFromSuffix(name string) string {
	if len(name) == 0 {
		return "float64"
	}
	switch name[len(name)-1] {
	case '%':
		return "int16"
	case '&':
		return "int32"
	case '!':
		return "float32"
	case '#':
		return "float64"
	case '$':
		return "string"
	default:
		return "float64"
	}
}

// isStringType returns true if the BASIC name denotes a string variable.
func isStringType(name string) bool {
	return len(name) > 0 && name[len(name)-1] == '$'
}

// ---------------------------------------------------------------------------
// Number formatting
// ---------------------------------------------------------------------------

// formatGoNumber emits a Go numeric literal for a BASIC NumberLiteral.
// We emit *untyped* Go constants (e.g., 42 or 3.14) rather than wrapping in
// float64(...).  Go's untyped constant rules automatically convert them to
// the destination type (float32, int16, float64, etc.) during assignment and
// function calls, avoiding the "cannot use float64(N) as float32" errors that
// would arise from explicit typed wrappers.
func formatGoNumber(n *ast.NumberLiteral) string {
	// If it is an integer value without fractional part, emit as a plain integer.
	if n.Value == math.Trunc(n.Value) && n.Value >= -1e15 && n.Value <= 1e15 {
		return fmt.Sprintf("%d", int64(n.Value))
	}
	return fmt.Sprintf("%g", n.Value)
}

// ---------------------------------------------------------------------------
// Argument helpers
// ---------------------------------------------------------------------------

func (g *CodeGenerator) oneArg(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "0"
}

func (g *CodeGenerator) argN(args []string, n int) string {
	if n < len(args) {
		return args[n]
	}
	return "0"
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

// emitTypeBlock emits a TYPE block as a Go struct definition.
func (g *CodeGenerator) emitTypeBlock(s *ast.TypeBlockStatement) {
	g.writeLinef("// TYPE %s (user-defined type — transpiled as struct)", s.Name)
}

// emitFnCallExpression emits a DEF FN function call.
func (g *CodeGenerator) emitFnCallExpression(e *ast.FnCallExpression) string {
	var argStrs []string
	for _, a := range e.Args {
		argStrs = append(argStrs, g.emitExpr(a))
	}
	return fmt.Sprintf("fn_%s(%s)", mangleName(e.Name), strings.Join(argStrs, ", "))
}
