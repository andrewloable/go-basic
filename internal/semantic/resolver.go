package semantic

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Resolver
//
// TUTORIAL — What is a Resolver and what does it do?
//
// The Resolver is the engine of Phase 3 (semantic analysis). It walks the
// AST produced by Phase 2 (parsing) and performs two jobs:
//
//  1. Symbol definition — the first time a name appears in the source, the
//     Resolver creates a Symbol for it and adds it to the current scope's
//     symbol table. This is called "defining" or "declaring" the symbol.
//
//  2. Symbol resolution — subsequent uses of the same name look up the
//     existing Symbol in the table to verify that it exists and is used
//     correctly (right number of arguments, etc.).
//
// The Resolver does not transform the AST or generate any output code. Its
// only outputs are:
//   - A populated SymbolTable (the primary output, consumed by codegen)
//   - A slice of error strings (semantic errors found in the program)
//
// TUTORIAL — Why is semantic analysis a separate phase?
//
// The parser (Phase 2) focuses on syntactic structure: it only checks that
// tokens appear in the right order according to the grammar. It does not and
// cannot check that names are valid, because the grammar allows any identifier
// anywhere a name is expected.
//
// Separating semantic analysis into its own phase has several advantages:
//  1. Cleaner responsibility: the parser handles syntax; the resolver handles
//     meaning. Each is simpler when focused on one concern.
//  2. Better error recovery: the resolver can report multiple semantic errors
//     in one pass (by appending to r.Errors) without stopping at the first.
//  3. Reusability: the same AST can be passed to multiple analysis passes
//     (type inference, optimization, codegen) without re-parsing.
// ---------------------------------------------------------------------------

// Resolver walks an AST and populates a SymbolTable, collecting errors.
type Resolver struct {
	Program *ast.Program
	Table   *SymbolTable
	Errors  []string

	// labelRefs tracks (label -> position) for GOTO/GOSUB targets so we can
	// verify them after the full walk.
	labelRefs []labelRef
}

type labelRef struct {
	name   string
	line   int
	column int
}

// NewResolver creates a Resolver for the given program.
func NewResolver(prog *ast.Program) *Resolver {
	return &Resolver{
		Program: prog,
		Table:   NewSymbolTable(),
	}
}

// Resolve performs a two-pass walk of the AST:
//
//  1. Register forward declarations (SUB, FUNCTION, LABEL, DEF FN, line numbers).
//  2. Walk every statement, defining variables on first use and resolving references.
//
// It returns the populated SymbolTable and any accumulated errors.
//
// TUTORIAL — Why two passes?
//
// Pass 1 (forward declarations) solves the "forward reference" problem: in
// BASIC, a program can call a SUB before the SUB is defined in the source file:
//
//   CALL MySub()         ' call appears before definition
//   ...
//   SUB MySub()          ' definition is here
//   END SUB
//
// If we tried to resolve "MySub" in a single forward pass, we would reach the
// CALL before we have registered "MySub" as a symbol, and we would incorrectly
// report it as undefined.
//
// The two-pass solution:
//   Pass 1: scan the entire program and register all SUB, FUNCTION, LABEL, and
//           DEF FN declarations in the global scope. After this pass, every
//           callable name is known.
//   Pass 2: walk every statement in full. When we encounter the CALL, "MySub"
//           is already in the symbol table from Pass 1, so it resolves correctly.
//
// Two-pass analysis is a standard technique in compilers. Languages that
// require forward declarations (like early C) sidestep the problem by making
// the programmer declare names before use. Languages without that requirement
// (like modern Go) use multi-pass analysis internally.
//
// The post-pass label verification (checking GOTO targets exist) cannot be
// done during Pass 2 because labels may also appear after the GOTO that
// references them. It must be done after Pass 2 is complete — only then are
// all labels guaranteed to be in the symbol table.
func (r *Resolver) Resolve() (*SymbolTable, []string) {
	// ---- Pass 1: collect forward declarations ----
	for _, stmt := range r.Program.Statements {
		r.registerForwardDecl(stmt)
	}

	// ---- Pass 2: full walk ----
	for _, stmt := range r.Program.Statements {
		r.resolveStatement(stmt)
	}

	// ---- Post: verify label references ----
	for _, ref := range r.labelRefs {
		if r.Table.GlobalScope.Symbols[strings.ToUpper(ref.name)] == nil {
			r.addError(ref.line, ref.column, "undefined label %q", ref.name)
		}
	}

	return r.Table, r.Errors
}

// ---------------------------------------------------------------------------
// Pass 1 helpers
// ---------------------------------------------------------------------------

func (r *Resolver) registerForwardDecl(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.SubDeclaration:
		r.defineGlobal(s.Name, &Symbol{
			Name:     s.Name,
			Type:     SymSub,
			DataType: TypeUnknown,
			Scope:    "global",
			Defined:  true,
			Line:     s.Pos().Line,
			Column:   s.Pos().Column,
			Params:   paramTypes(s.Params, r.Table),
		})
	case *ast.FunctionDeclaration:
		rt := resolveReturnType(s.ReturnType, s.Name, r.Table)
		r.defineGlobal(s.Name, &Symbol{
			Name:       s.Name,
			Type:       SymFunction,
			DataType:   rt,
			Scope:      "global",
			Defined:    true,
			Line:       s.Pos().Line,
			Column:     s.Pos().Column,
			Params:     paramTypes(s.Params, r.Table),
			ReturnType: rt,
		})
	case *ast.DefFnDeclaration:
		r.defineGlobal(s.Name, &Symbol{
			Name:    s.Name,
			Type:    SymDefFn,
			Scope:   "global",
			Defined: true,
			Line:    s.Pos().Line,
			Column:  s.Pos().Column,
			Params:  paramTypes(s.Params, r.Table),
		})
	case *ast.LabelStatement:
		r.defineGlobal(s.Name, &Symbol{
			Name:    s.Name,
			Type:    SymLabel,
			Scope:   "global",
			Defined: true,
			Line:    s.Pos().Line,
			Column:  s.Pos().Column,
		})
	case *ast.LineNumberStatement:
		label := fmt.Sprintf("%d", s.Number)
		r.defineGlobal(label, &Symbol{
			Name:    label,
			Type:    SymLabel,
			Scope:   "global",
			Defined: true,
			Line:    s.Pos().Line,
			Column:  s.Pos().Column,
		})
	}
}

// defineGlobal defines a symbol in the global scope, ignoring duplicate errors
// (which will be reported later if they are genuine conflicts).
func (r *Resolver) defineGlobal(name string, sym *Symbol) {
	key := strings.ToUpper(name)
	if existing, ok := r.Table.GlobalScope.Symbols[key]; ok {
		// Allow forward-declaration followed by actual definition.
		if existing.Type == sym.Type {
			return
		}
		r.addError(sym.Line, sym.Column, "duplicate definition of %q (previously at %d:%d)", name, existing.Line, existing.Column)
		return
	}
	r.Table.GlobalScope.Symbols[key] = sym
}

// ---------------------------------------------------------------------------
// Pass 2: statement dispatcher
// ---------------------------------------------------------------------------

func (r *Resolver) resolveStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.LetStatement:
		r.resolveLet(s)
	case *ast.DimStatement:
		r.resolveDim(s)
	case *ast.RedimStatement:
		r.resolveRedim(s)
	case *ast.ForStatement:
		r.resolveFor(s)
	case *ast.WhileStatement:
		r.resolveWhile(s)
	case *ast.DoLoopStatement:
		r.resolveDoLoop(s)
	case *ast.IfStatement:
		r.resolveIf(s)
	case *ast.SelectCaseStatement:
		r.resolveSelectCase(s)
	case *ast.PrintStatement:
		r.resolvePrint(s)
	case *ast.SubDeclaration:
		r.resolveSubDecl(s)
	case *ast.FunctionDeclaration:
		r.resolveFunctionDecl(s)
	case *ast.DefFnDeclaration:
		r.resolveDefFn(s)
	case *ast.GotoStatement:
		r.resolveGoto(s)
	case *ast.GosubStatement:
		r.resolveGosub(s)
	case *ast.OnErrorGotoStatement:
		r.resolveOnErrorGoto(s)
	case *ast.OnEventGosubStatement:
		r.resolveOnEventGosub(s)
	case *ast.DefTypeStatement:
		r.resolveDefType(s)
	case *ast.OptionBaseStatement:
		r.resolveOptionBase(s)
	case *ast.ScopeStatement:
		r.resolveScopeStmt(s)
	case *ast.ArrayAssignment:
		r.resolveArrayAssignment(s)
	case *ast.ReadStatement:
		r.resolveRead(s)
	case *ast.SwapStatement:
		r.resolveSwap(s)
	case *ast.IncrStatement:
		r.resolveIncr(s)
	case *ast.DecrStatement:
		r.resolveDecr(s)
	case *ast.OpenStatement:
		r.resolveOpen(s)
	case *ast.CloseStatement:
		r.resolveClose(s)
	case *ast.FileInputStatement:
		r.resolveFileInput(s)
	case *ast.FilePrintStatement:
		r.resolveFilePrint(s)
	case *ast.FileWriteStatement:
		r.resolveFileWrite(s)
	case *ast.GetStatement:
		r.resolveGet(s)
	case *ast.PutStatement:
		r.resolvePut(s)
	case *ast.SeekStatement:
		r.resolveSeek(s)
	case *ast.ErrorStatement:
		r.resolveExpressionIfNotNil(s.Code)

	// Statements that carry no meaningful symbol information:
	case *ast.LabelStatement, *ast.LineNumberStatement, *ast.RemStatement,
		*ast.EndStatement, *ast.StopStatement, *ast.SystemStatement,
		*ast.ReturnStatement, *ast.ExitStatement, *ast.DataStatement,
		*ast.RestoreStatement, *ast.BeepStatement, *ast.EraseStatement,
		*ast.ResumeStatement:
		// nothing to resolve

	// Graphics / sound – just resolve contained expressions.
	case *ast.ScreenStatement:
		r.resolveExpressionIfNotNil(s.Mode)
		r.resolveExpressionIfNotNil(s.ColorSwitch)
		r.resolveExpressionIfNotNil(s.ActivePage)
		r.resolveExpressionIfNotNil(s.VisualPage)
	case *ast.ColorStatement:
		r.resolveExpressionIfNotNil(s.Foreground)
		r.resolveExpressionIfNotNil(s.Background)
		r.resolveExpressionIfNotNil(s.Border)
	case *ast.PsetStatement:
		r.resolveExpressionIfNotNil(s.X)
		r.resolveExpressionIfNotNil(s.Y)
		r.resolveExpressionIfNotNil(s.Color)
	case *ast.LineStmt:
		r.resolveExpressionIfNotNil(s.X1)
		r.resolveExpressionIfNotNil(s.Y1)
		r.resolveExpressionIfNotNil(s.X2)
		r.resolveExpressionIfNotNil(s.Y2)
		r.resolveExpressionIfNotNil(s.Color)
	case *ast.CircleStmt:
		r.resolveExpressionIfNotNil(s.X)
		r.resolveExpressionIfNotNil(s.Y)
		r.resolveExpressionIfNotNil(s.Radius)
		r.resolveExpressionIfNotNil(s.Color)
		r.resolveExpressionIfNotNil(s.Start)
		r.resolveExpressionIfNotNil(s.End)
		r.resolveExpressionIfNotNil(s.Aspect)
	case *ast.PaintStmt:
		r.resolveExpressionIfNotNil(s.X)
		r.resolveExpressionIfNotNil(s.Y)
		r.resolveExpressionIfNotNil(s.FillColor)
		r.resolveExpressionIfNotNil(s.BorderColor)
	case *ast.DrawStmt:
		r.resolveExpressionIfNotNil(s.CommandString)
	case *ast.SoundStatement:
		r.resolveExpressionIfNotNil(s.Frequency)
		r.resolveExpressionIfNotNil(s.Duration)
	case *ast.PlayStatement:
		r.resolveExpressionIfNotNil(s.CommandString)
	case *ast.LocateStatement:
		r.resolveExpressionIfNotNil(s.Row)
		r.resolveExpressionIfNotNil(s.Col)
	case *ast.ClsStatement:
		r.resolveExpressionIfNotNil(s.Mode)
	case *ast.ViewStatement:
		r.resolveExpressionIfNotNil(s.X1)
		r.resolveExpressionIfNotNil(s.Y1)
		r.resolveExpressionIfNotNil(s.X2)
		r.resolveExpressionIfNotNil(s.Y2)
		r.resolveExpressionIfNotNil(s.FillColor)
		r.resolveExpressionIfNotNil(s.BorderColor)
		r.resolveExpressionIfNotNil(s.Top)
		r.resolveExpressionIfNotNil(s.Bottom)
	case *ast.FieldStatement:
		r.resolveExpressionIfNotNil(s.FileNum)
	case *ast.LsetStatement:
		r.resolveExpressionIfNotNil(s.Value)
	case *ast.RsetStatement:
		r.resolveExpressionIfNotNil(s.Value)
	case *ast.KillStatement:
		r.resolveExpressionIfNotNil(s.Filename)
	case *ast.NameStatement:
		r.resolveExpressionIfNotNil(s.OldName)
		r.resolveExpressionIfNotNil(s.NewName)
	case *ast.ChdirStatement:
		r.resolveExpressionIfNotNil(s.Path)
	case *ast.MkdirStatement:
		r.resolveExpressionIfNotNil(s.Path)
	case *ast.RmdirStatement:
		r.resolveExpressionIfNotNil(s.Path)
	}
}
