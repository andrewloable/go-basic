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
	"strconv"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// ---------------------------------------------------------------------------
// CodeGenerator transpiles a Turbo BASIC AST into compilable Go source code.
// ---------------------------------------------------------------------------

// CodeGenerator holds all state needed while walking the AST and emitting Go.
//
// TUTORIAL — Generator state
//
// A code generator is essentially a stateful tree visitor. As it walks the AST
// it accumulates information that influences how later nodes should be emitted.
// Keeping this state in a struct (rather than global variables) makes the
// generator re-entrant and testable. Key fields explained:
//
//   buf         — The primary output buffer. Statement emitters write Go source
//                 text here incrementally. Think of it as a "virtual pen" that
//                 moves forward-only through the output file.
//
//   indent      — Current indentation level. Incremented when entering a block
//                 (if/for/while) and decremented on exit, so child statements
//                 are indented correctly without any caller needing to track it.
//
//   table       — The symbol table produced by Phase 3 (semantic analysis).
//                 The generator consults it to look up the type of every
//                 variable so it can emit the correct Go type cast. Without
//                 this, the generator would have to re-infer types from scratch.
//
//   imports     — Tracks which Go packages are actually used. A production Go
//                 file must not import packages it doesn't use (the compiler
//                 rejects it). This set lets Generate() emit only the imports
//                 that are genuinely needed.
//
//   declared    — Tracks which Go variable names have already had a "var"
//                 declaration emitted. BASIC programs declare variables
//                 implicitly on first use; the generator must emit a Go "var"
//                 only on the first occurrence to avoid "declared but not used"
//                 or "redeclared" errors.
//
//   tempCount   — Monotonically-increasing counter for generating unique temp
//                 names (end_0, step_0, end_1, step_1, …). Multiple FOR loops
//                 in the same scope each need their own temp variables.
//
//   labelMap    — Set of all labels/line-numbers defined in the source, built
//                 in the pre-pass. Go treats unused labels as compile errors, so
//                 this is cross-referenced with referencedLabels to decide which
//                 labels to actually emit.
//
//   referencedLabels — Set of labels that are the target of at least one GOTO,
//                 GOSUB, ON ERROR GOTO, etc. Only these labels are emitted in
//                 the generated Go; the rest are suppressed to avoid unused-label
//                 compile errors.
//
//   dataPool    — All DATA statement values collected in order. BASIC's READ
//                 statement consumes values from this pool sequentially, so the
//                 entire pool must be materialised before main() body is emitted.
//
//   funcBuf     — A secondary output buffer that collects SUB and FUNCTION
//                 declarations. These are appended after the closing brace of
//                 main(), producing the correct Go file structure where top-level
//                 functions appear at package level rather than nested inside main.
type CodeGenerator struct {
	program          *ast.Program
	table            *semantic.SymbolTable
	buf              bytes.Buffer
	indent           int
	imports          map[string]bool  // track needed imports
	declared         map[string]bool  // track declared variables (mangled names)
	tempCount        int              // temp variable counter
	labelMap         map[string]bool  // labels that exist (defined in source)
	referencedLabels map[string]bool  // labels that are targeted by GOTO/GOSUB
	gosubFuncs       map[string]bool  // GOSUB targets turned into functions
	dataPool         []ast.Expression // DATA values
	dataIdx          int              // current READ position

	// funcBuf collects SUB/FUNCTION declarations to emit outside main().
	funcBuf bytes.Buffer

	// needRng tracks whether the rng variable is needed.
	needRng bool

	// needErrState tracks whether the errState variable is needed (for ERR builtin).
	needErrState bool

	// hasRead tracks whether any DATA READ statements exist (to emit dataPool).
	hasRead bool

	// needFileManager tracks whether the fm *rt.FileManager variable is needed.
	needFileManager bool

	// inDefFn is true while emitting the body of a multi-line DEF FN.
	inDefFn bool
	// defFnReturnType is the Go return type of the DEF FN currently being emitted.
	defFnReturnType string
	// defFnName is the BASIC name of the DEF FN currently being emitted (e.g. "FNFactorial#").
	defFnName string

	// hoistedVars collects variable declarations to hoist to the top of main()
	// to avoid goto-over-declaration errors in Go. Each entry is {mangledName, goType}.
	hoistedVars []hoistedVar
	// hoistedSet tracks which mangled names have been added to hoistedVars.
	hoistedSet map[string]bool
	// hasGoto tracks whether the program contains any GOTO/GOSUB statements.
	hasGoto bool

	// sharedVars tracks mangled names of variables declared SHARED in any SUB/FUNCTION.
	// These variables are emitted as package-level var declarations.
	sharedVars map[string]bool
	// packageVars collects package-level variable declarations for SHARED variables.
	packageVars []hoistedVar
	// packageVarSet tracks which mangled names have been added to packageVars.
	packageVarSet map[string]bool
	// inSubOrFunc is true while emitting the body of a SUB or FUNCTION.
	inSubOrFunc bool

	// arrayDims tracks the number of dimensions for each array (by mangled name).
	// Used by emitArrayAccess and emitArrayAssignment to emit the correct
	// number of index brackets for multi-dimensional arrays.
	arrayDims map[string]int

	// onErrorLabel holds the most recently set ON ERROR GOTO target label
	// (mangled). Used by ERROR statement codegen to emit a goto.
	onErrorLabel string
}

// hoistedVar represents a variable declaration to be hoisted to the top of main().
type hoistedVar struct {
	name string // mangled Go name
	typ  string // Go type (e.g. "float32", "string")
}

// New creates a fresh CodeGenerator ready for use.
func New() *CodeGenerator {
	return &CodeGenerator{
		imports:          make(map[string]bool),
		declared:         make(map[string]bool),
		labelMap:         make(map[string]bool),
		referencedLabels: make(map[string]bool),
		gosubFuncs:       make(map[string]bool),
		hoistedSet:       make(map[string]bool),
		sharedVars:       make(map[string]bool),
		packageVarSet:    make(map[string]bool),
		arrayDims:        make(map[string]int),
	}
}

// Generate is the main entry point.  It walks the AST, emits Go code, and
// returns the fully-formed Go source string (or an error).
//
// TUTORIAL — Top-level code emission strategy
//
// Generate() orchestrates three distinct phases:
//
//  1. Pre-passes (information gathering)
//     Before any Go code is written, the generator scans the entire AST to
//     build maps it will need during emission:
//       - collectLabelsAndData: finds all label definitions, GOTO targets, and
//         DATA values. This is necessary because GOTO can jump forward — the
//         target label may appear after the GOTO in the source.
//       - collectSharedVars: finds SHARED declarations inside SUBs/FUNCTIONs so
//         those variables can be promoted to package-level.
//       - collectMainVariables (when GOTOs exist): hoists all variable
//         declarations to the top of main() to avoid Go's "goto jumps over
//         variable declaration" compile error.
//
//  2. Body emission
//     The generator iterates program.Statements and calls emitStatement() for
//     each one. SUB/FUNCTION declarations are routed to funcBuf instead of the
//     main buffer. This single-pass walk relies on the pre-pass data being
//     complete before it starts.
//
//  3. Assembly
//     The final output is assembled from separate pieces in the correct order:
//     "package main" → imports → suppress-unused helpers → package-level vars →
//     func main() { preamble + body } → SUB/FUNCTION declarations.
//
// This three-phase structure is a common pattern in transpilers: gather
// information first, emit second, assemble third.
func (g *CodeGenerator) Generate(program *ast.Program, table *semantic.SymbolTable) (string, error) {
	g.program = program
	g.table = table

	// Pre-pass: collect labels & DATA values.
	g.collectLabelsAndData(program.Statements)

	// Pre-pass: collect SHARED variables from SUB/FUNCTION bodies.
	// These will be emitted as package-level var declarations.
	g.collectSharedVars(program.Statements)

	// Pre-populate declared map for shared/package-level variables so they
	// are not re-declared inside main().
	for _, pv := range g.packageVars {
		g.declared[pv.name] = true
	}

	// Check if the program uses any GOTO/GOSUB statements.
	g.hasGoto = len(g.referencedLabels) > 0

	// If GOTOs exist, collect all main-level variables for hoisting to avoid
	// goto-over-declaration errors in Go.
	if g.hasGoto {
		// Only collect from main-level statements (skip SUB/FUNCTION).
		mainStmts := make([]ast.Statement, 0, len(program.Statements))
		for _, stmt := range program.Statements {
			switch stmt.(type) {
			case *ast.SubDeclaration, *ast.FunctionDeclaration:
				continue
			}
			mainStmts = append(mainStmts, stmt)
		}
		g.collectMainVariables(mainStmts)

		// Pre-populate declared map so emitLet/emitFor/etc. use assignment form.
		for _, hv := range g.hoistedVars {
			g.declared[hv.name] = true
		}
	}

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

	// Emit package-level variable declarations for SHARED variables.
	if len(g.packageVars) > 0 {
		out.WriteString("// Package-level variables (SHARED across SUBs).\n")
		for _, pv := range g.packageVars {
			out.WriteString(fmt.Sprintf("var %s %s\n", pv.name, pv.typ))
		}
		out.WriteString("\n")
	}

	// main function.
	out.WriteString("func main() {\n")

	// Declare rng if needed.
	if g.needRng {
		out.WriteString("\trng := rt.NewRNG()\n")
		out.WriteString("\t_ = rng\n")
	}

	// Declare errState if needed (for ERR / ERL builtins).
	if g.needErrState {
		out.WriteString("\terrState := rt.NewErrorState()\n")
		out.WriteString("\t_ = errState\n")
	}

	// Declare FileManager if needed (for OPEN/CLOSE file I/O).
	if g.needFileManager {
		out.WriteString("\tfm := rt.NewFileManager()\n")
		out.WriteString("\tdefer fm.FileCloseAll()\n")
	}

	// Declare DATA pool if needed.
	if hasData {
		out.WriteString("\tvar dataPool []interface{}\n")
		out.WriteString("\t_ = dataPool\n")
		g.emitDataPoolInit(&out)
		out.WriteString("\tdataIdx := 0\n")
		out.WriteString("\t_ = dataIdx\n")
	}

	// Emit hoisted variable declarations (goto-over-declaration fix).
	if len(g.hoistedVars) > 0 {
		out.WriteString("\t// Hoisted variable declarations (avoids goto-over-declaration errors).\n")
		for _, hv := range g.hoistedVars {
			out.WriteString(fmt.Sprintf("\tvar %s %s\n", hv.name, hv.typ))
		}
		out.WriteString("\n")
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
		case *ast.RestoreStatement:
			g.hasRead = true

		// Collect GOTO/GOSUB targets — these become goto labels in Go.
		case *ast.GotoStatement:
			g.referencedLabels[strings.ToUpper(s.Target)] = true
		case *ast.GosubStatement:
			g.referencedLabels[strings.ToUpper(s.Target)] = true
		case *ast.OnErrorGotoStatement:
			if s.Target != "0" && s.Target != "" {
				g.referencedLabels[strings.ToUpper(s.Target)] = true
			}
		case *ast.ResumeStatement:
			if s.Type != "" && s.Type != "NEXT" {
				g.referencedLabels[strings.ToUpper(s.Type)] = true
			}
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
		case *ast.DefFnDeclaration:
			g.collectLabelsAndData(s.Body)
		}
	}
}

// ---------------------------------------------------------------------------
// Pre-pass: collect variables for hoisting (goto-over-declaration fix)
// ---------------------------------------------------------------------------

// collectMainVariables walks the main-level statements (not SUB/FUNCTION) and
// records every variable that will need a `var` declaration. When the program
// contains GOTOs, these declarations are hoisted to the top of main() so that
// no goto can jump over a variable declaration — which is a compile error in Go.
func (g *CodeGenerator) collectMainVariables(stmts []ast.Statement) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.LetStatement:
			name := mangleName(s.Name.Name + s.Name.TypeSuffix)
			goT := g.goTypeForIdent(s.Name.Name + s.Name.TypeSuffix)
			g.addHoistedVar(name, goT)

		case *ast.ForStatement:
			counter := mangleName(s.Counter.Name + s.Counter.TypeSuffix)
			goT := g.goTypeForIdent(s.Counter.Name + s.Counter.TypeSuffix)
			g.addHoistedVar(counter, goT)
			g.collectMainVariables(s.Body)

		case *ast.DimStatement:
			for _, d := range s.Declarations {
				name := mangleName(d.Name + d.TypeSuffix)
				if len(d.Dimensions) == 0 {
					goT := g.goTypeForDecl(d)
					g.addHoistedVar(name, goT)
				} else if len(d.Dimensions) == 1 {
					goT := g.goTypeForDecl(d)
					g.addHoistedVar(name, "[]"+goT)
				} else {
					goT := g.goTypeForDecl(d)
					g.addHoistedVar(name, "[]"+goT)
				}
			}

		case *ast.ReadStatement:
			for _, v := range s.Variables {
				if ident, ok := v.(*ast.Identifier); ok {
					name := mangleName(ident.Name + ident.TypeSuffix)
					goT := g.goTypeForIdent(ident.Name + ident.TypeSuffix)
					g.addHoistedVar(name, goT)
				}
			}

		// Recurse into nested blocks (but not SUB/FUNCTION — those are top-level).
		case *ast.IfStatement:
			g.collectMainVariables(s.ThenBlock)
			for _, clause := range s.ElseIfClauses {
				g.collectMainVariables(clause.Body)
			}
			g.collectMainVariables(s.ElseBlock)
		case *ast.WhileStatement:
			g.collectMainVariables(s.Body)
		case *ast.DoLoopStatement:
			g.collectMainVariables(s.Body)
		case *ast.SelectCaseStatement:
			for _, c := range s.Cases {
				g.collectMainVariables(c.Body)
			}
		case *ast.DefFnDeclaration:
			g.collectMainVariables(s.Body)
		}
	}
}

// collectSharedVars walks all SUB/FUNCTION declarations and finds
// SHARED statements, marking those variables as package-level.
func (g *CodeGenerator) collectSharedVars(stmts []ast.Statement) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.SubDeclaration:
			g.collectSharedFromBody(s.Body)
		case *ast.FunctionDeclaration:
			g.collectSharedFromBody(s.Body)
		}
	}
}

// collectSharedFromBody scans a SUB/FUNCTION body for SHARED scope statements
// and registers each listed variable as a package-level variable.
func (g *CodeGenerator) collectSharedFromBody(body []ast.Statement) {
	for _, stmt := range body {
		scope, ok := stmt.(*ast.ScopeStatement)
		if !ok || scope.Modifier != "SHARED" {
			continue
		}
		for _, v := range scope.Variables {
			name := mangleName(v)
			goT := g.goTypeForIdent(v)
			g.sharedVars[name] = true
			if !g.packageVarSet[name] {
				g.packageVarSet[name] = true
				g.packageVars = append(g.packageVars, hoistedVar{name: name, typ: goT})
			}
		}
	}
}

// addPackageVar records a variable for package-level emission, avoiding duplicates.
func (g *CodeGenerator) addPackageVar(name, goType string) {
	if !g.packageVarSet[name] {
		g.packageVarSet[name] = true
		g.packageVars = append(g.packageVars, hoistedVar{name: name, typ: goType})
		g.sharedVars[name] = true
	}
}

// addHoistedVar records a variable for hoisting, avoiding duplicates.
// Variables that are already package-level (SHARED) are skipped.
func (g *CodeGenerator) addHoistedVar(name, goType string) {
	if g.sharedVars[name] {
		return // already declared at package level
	}
	if !g.hoistedSet[name] {
		g.hoistedSet[name] = true
		g.hoistedVars = append(g.hoistedVars, hoistedVar{name: name, typ: goType})
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
