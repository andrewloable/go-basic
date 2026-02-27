package semantic

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// SymbolType – classifies what a symbol represents
// ---------------------------------------------------------------------------

// SymbolType indicates the kind of entity a symbol represents.
type SymbolType int

const (
	SymVariable SymbolType = iota
	SymArray
	SymFunction
	SymSub
	SymLabel
	SymConst
	SymDefFn
)

// String returns the human-readable name for a SymbolType.
func (st SymbolType) String() string {
	switch st {
	case SymVariable:
		return "Variable"
	case SymArray:
		return "Array"
	case SymFunction:
		return "Function"
	case SymSub:
		return "Sub"
	case SymLabel:
		return "Label"
	case SymConst:
		return "Const"
	case SymDefFn:
		return "DefFn"
	default:
		return "Unknown"
	}
}

// ---------------------------------------------------------------------------
// DataType – the BASIC data types
// ---------------------------------------------------------------------------

// DataType represents a Turbo BASIC data type.
type DataType int

const (
	TypeInteger DataType = iota // %
	TypeLong                    // &
	TypeSingle                  // !
	TypeDouble                  // #
	TypeString                  // $
	TypeUnknown
)

// String returns the human-readable name (with suffix) for a DataType.
func (dt DataType) String() string {
	switch dt {
	case TypeInteger:
		return "Integer(%)"
	case TypeLong:
		return "Long(&)"
	case TypeSingle:
		return "Single(!)"
	case TypeDouble:
		return "Double(#)"
	case TypeString:
		return "String($)"
	default:
		return "Unknown"
	}
}

// ---------------------------------------------------------------------------
// Symbol
// ---------------------------------------------------------------------------

// Symbol describes a named entity (variable, array, sub, function, etc.).
type Symbol struct {
	Name       string
	Type       SymbolType
	DataType   DataType
	Scope      string // "global", or the SUB/FUNCTION name
	Defined    bool   // has been assigned / defined
	Used       bool   // has been referenced
	Line       int    // line where first declared
	Column     int    // column where first declared
	Params     []DataType
	ReturnType DataType
	ArrayDims  int  // number of dimensions for arrays
	IsShared   bool // SHARED declaration
	IsStatic   bool // STATIC declaration
}

// ---------------------------------------------------------------------------
// Scope
// ---------------------------------------------------------------------------

// Scope represents a lexical scope (global, or inside a SUB/FUNCTION).
type Scope struct {
	Name    string
	Parent  *Scope
	Symbols map[string]*Symbol // keys are upper-cased for case-insensitivity
}

// newScope creates a Scope with an initialised symbol map.
func newScope(name string, parent *Scope) *Scope {
	return &Scope{
		Name:    name,
		Parent:  parent,
		Symbols: make(map[string]*Symbol),
	}
}

// ---------------------------------------------------------------------------
// SymbolTable
// ---------------------------------------------------------------------------

// SymbolTable holds scopes, DEFtype mappings, and OPTION BASE for a program.
type SymbolTable struct {
	GlobalScope  *Scope
	CurrentScope *Scope
	Scopes       map[string]*Scope
	DefTypes     map[byte]DataType // letter -> default type (set by DEFINT etc.)
	OptionBase   int               // 0 or 1
}

// NewSymbolTable creates a ready-to-use SymbolTable with a global scope.
func NewSymbolTable() *SymbolTable {
	global := newScope("global", nil)
	st := &SymbolTable{
		GlobalScope:  global,
		CurrentScope: global,
		Scopes:       make(map[string]*Scope),
		DefTypes:     make(map[byte]DataType),
		OptionBase:   0,
	}
	st.Scopes["global"] = global
	return st
}

// EnterScope creates a new child scope with the given name, sets it as the
// current scope, and registers it in the Scopes map.
func (st *SymbolTable) EnterScope(name string) {
	key := strings.ToUpper(name)
	s := newScope(key, st.CurrentScope)
	st.Scopes[key] = s
	st.CurrentScope = s
}

// ExitScope returns to the parent scope.  If we are already at the global
// scope, this is a no-op.
func (st *SymbolTable) ExitScope() {
	if st.CurrentScope.Parent != nil {
		st.CurrentScope = st.CurrentScope.Parent
	}
}

// Define adds a symbol to the current scope.  It returns an error if a symbol
// with the same name (case-insensitive) already exists in the current scope.
func (st *SymbolTable) Define(name string, sym *Symbol) error {
	key := strings.ToUpper(name)
	if _, exists := st.CurrentScope.Symbols[key]; exists {
		return fmt.Errorf("symbol %q already defined in scope %s", name, st.CurrentScope.Name)
	}
	st.CurrentScope.Symbols[key] = sym
	return nil
}

// Lookup searches for a symbol starting from the current scope, then walking
// up to parent scopes (ultimately reaching global).  Returns nil when the
// symbol is not found in any enclosing scope.
func (st *SymbolTable) Lookup(name string) *Symbol {
	key := strings.ToUpper(name)
	for s := st.CurrentScope; s != nil; s = s.Parent {
		if sym, ok := s.Symbols[key]; ok {
			return sym
		}
	}
	return nil
}

// LookupLocal searches only the current scope.
func (st *SymbolTable) LookupLocal(name string) *Symbol {
	key := strings.ToUpper(name)
	if sym, ok := st.CurrentScope.Symbols[key]; ok {
		return sym
	}
	return nil
}

// ResolveType determines the DataType of a variable name by examining:
//  1. An explicit type suffix (%, &, !, #, $).
//  2. The DEFtype letter-range mapping for the first letter.
//  3. A default of TypeSingle (Turbo BASIC default).
func (st *SymbolTable) ResolveType(name string) DataType {
	// Strip any trailing suffix character to get the base name, then use the
	// suffix to determine the type.
	if len(name) > 0 {
		last := name[len(name)-1]
		switch last {
		case '%':
			return TypeInteger
		case '&':
			return TypeLong
		case '!':
			return TypeSingle
		case '#':
			return TypeDouble
		case '$':
			return TypeString
		}
	}

	// Check DEFtype mapping for the first letter.
	if len(name) > 0 {
		first := name[0]
		// Normalise to upper case.
		if first >= 'a' && first <= 'z' {
			first = first - 32
		}
		if dt, ok := st.DefTypes[first]; ok {
			return dt
		}
	}

	// Turbo BASIC default: SINGLE.
	return TypeSingle
}

// SetDefType sets the default DataType for all letters in [startLetter, endLetter].
// Letters are normalised to upper-case.
func (st *SymbolTable) SetDefType(startLetter, endLetter byte, dt DataType) {
	start := normaliseUpper(startLetter)
	end := normaliseUpper(endLetter)
	if start > end {
		start, end = end, start
	}
	for ch := start; ch <= end; ch++ {
		st.DefTypes[ch] = dt
	}
}

func normaliseUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

// ---------------------------------------------------------------------------
// Resolver
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

// ---------------------------------------------------------------------------
// Statement-specific resolvers
// ---------------------------------------------------------------------------

func (r *Resolver) resolveLet(s *ast.LetStatement) {
	// Resolve the right-hand side first.
	r.resolveExpression(s.Value)

	// Implicitly declare the variable if it is not yet known.
	name := s.Name.Name
	suffix := s.Name.TypeSuffix
	fullName := name + suffix
	dt := r.Table.ResolveType(fullName)

	if sym := r.Table.Lookup(fullName); sym != nil {
		sym.Defined = true
		sym.Used = true
	} else {
		_ = r.Table.Define(fullName, &Symbol{
			Name:     fullName,
			Type:     SymVariable,
			DataType: dt,
			Scope:    r.Table.CurrentScope.Name,
			Defined:  true,
			Line:     s.Pos().Line,
			Column:   s.Pos().Column,
		})
	}
}

func (r *Resolver) resolveDim(s *ast.DimStatement) {
	for _, d := range s.Declarations {
		name := d.Name + d.TypeSuffix
		dims := len(d.Dimensions)
		dt := r.resolveElementType(d.ElementType, d.TypeSuffix, d.Name)

		symType := SymVariable
		if dims > 0 {
			symType = SymArray
		}

		if existing := r.Table.LookupLocal(name); existing != nil {
			r.addError(d.BasePos.Line, d.BasePos.Column,
				"duplicate DIM of %q in scope %s", name, r.Table.CurrentScope.Name)
			continue
		}

		_ = r.Table.Define(name, &Symbol{
			Name:      name,
			Type:      symType,
			DataType:  dt,
			Scope:     r.Table.CurrentScope.Name,
			Defined:   true,
			Line:      d.BasePos.Line,
			Column:    d.BasePos.Column,
			ArrayDims: dims,
		})

		// Resolve dimension expressions.
		for _, dr := range d.Dimensions {
			r.resolveExpressionIfNotNil(dr.Lower)
			r.resolveExpressionIfNotNil(dr.Upper)
		}
		r.resolveExpressionIfNotNil(d.StringLength)
	}
}

func (r *Resolver) resolveRedim(s *ast.RedimStatement) {
	for _, d := range s.Declarations {
		name := d.Name + d.TypeSuffix
		dt := r.resolveElementType(d.ElementType, d.TypeSuffix, d.Name)
		dims := len(d.Dimensions)

		// REDIM replaces existing dimension info.
		if existing := r.Table.Lookup(name); existing != nil {
			existing.ArrayDims = dims
			existing.DataType = dt
		} else {
			_ = r.Table.Define(name, &Symbol{
				Name:      name,
				Type:      SymArray,
				DataType:  dt,
				Scope:     r.Table.CurrentScope.Name,
				Defined:   true,
				Line:      d.BasePos.Line,
				Column:    d.BasePos.Column,
				ArrayDims: dims,
			})
		}
		for _, dr := range d.Dimensions {
			r.resolveExpressionIfNotNil(dr.Lower)
			r.resolveExpressionIfNotNil(dr.Upper)
		}
	}
}

func (r *Resolver) resolveFor(s *ast.ForStatement) {
	// The counter variable is implicitly declared.
	name := s.Counter.Name + s.Counter.TypeSuffix
	dt := r.Table.ResolveType(name)
	if sym := r.Table.Lookup(name); sym != nil {
		sym.Defined = true
		sym.Used = true
	} else {
		_ = r.Table.Define(name, &Symbol{
			Name:     name,
			Type:     SymVariable,
			DataType: dt,
			Scope:    r.Table.CurrentScope.Name,
			Defined:  true,
			Line:     s.Counter.Pos().Line,
			Column:   s.Counter.Pos().Column,
		})
	}
	r.resolveExpression(s.Start)
	r.resolveExpression(s.End)
	r.resolveExpressionIfNotNil(s.Step)
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolveWhile(s *ast.WhileStatement) {
	r.resolveExpression(s.Condition)
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolveDoLoop(s *ast.DoLoopStatement) {
	r.resolveExpressionIfNotNil(s.Condition)
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolveIf(s *ast.IfStatement) {
	r.resolveExpression(s.Condition)
	for _, stmt := range s.ThenBlock {
		r.resolveStatement(stmt)
	}
	for _, clause := range s.ElseIfClauses {
		r.resolveExpression(clause.Condition)
		for _, stmt := range clause.Body {
			r.resolveStatement(stmt)
		}
	}
	for _, stmt := range s.ElseBlock {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolveSelectCase(s *ast.SelectCaseStatement) {
	r.resolveExpression(s.TestExpr)
	for _, c := range s.Cases {
		for _, cv := range c.Values {
			r.resolveExpressionIfNotNil(cv.Value)
			r.resolveExpressionIfNotNil(cv.EndValue)
		}
		for _, stmt := range c.Body {
			r.resolveStatement(stmt)
		}
	}
	for _, stmt := range s.ElseBlock {
		r.resolveStatement(stmt)
	}
}

func (r *Resolver) resolvePrint(s *ast.PrintStatement) {
	for _, expr := range s.Expressions {
		r.resolveExpression(expr)
	}
	r.resolveExpressionIfNotNil(s.Format)
}

func (r *Resolver) resolveSubDecl(s *ast.SubDeclaration) {
	if s.IsForward {
		return // forward declaration already registered in pass 1
	}
	r.Table.EnterScope(s.Name)
	// Register parameters as local variables.
	for _, p := range s.Params {
		dt := paramDataType(p.Type, p.Name, r.Table)
		_ = r.Table.Define(p.Name, &Symbol{
			Name:     p.Name,
			Type:     SymVariable,
			DataType: dt,
			Scope:    strings.ToUpper(s.Name),
			Defined:  true,
			Line:     p.BasePos.Line,
			Column:   p.BasePos.Column,
		})
	}
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
	r.Table.ExitScope()
}

func (r *Resolver) resolveFunctionDecl(s *ast.FunctionDeclaration) {
	if s.IsForward {
		return
	}
	r.Table.EnterScope(s.Name)
	for _, p := range s.Params {
		dt := paramDataType(p.Type, p.Name, r.Table)
		_ = r.Table.Define(p.Name, &Symbol{
			Name:     p.Name,
			Type:     SymVariable,
			DataType: dt,
			Scope:    strings.ToUpper(s.Name),
			Defined:  true,
			Line:     p.BasePos.Line,
			Column:   p.BasePos.Column,
		})
	}
	// The function name itself is a local variable that holds the return value.
	rt := resolveReturnType(s.ReturnType, s.Name, r.Table)
	_ = r.Table.Define(s.Name, &Symbol{
		Name:       s.Name,
		Type:       SymVariable,
		DataType:   rt,
		Scope:      strings.ToUpper(s.Name),
		Defined:    true,
		Line:       s.Pos().Line,
		Column:     s.Pos().Column,
		ReturnType: rt,
	})
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
	r.Table.ExitScope()
}

func (r *Resolver) resolveDefFn(s *ast.DefFnDeclaration) {
	r.Table.EnterScope(s.Name)
	for _, p := range s.Params {
		dt := paramDataType(p.Type, p.Name, r.Table)
		_ = r.Table.Define(p.Name, &Symbol{
			Name:     p.Name,
			Type:     SymVariable,
			DataType: dt,
			Scope:    strings.ToUpper(s.Name),
			Defined:  true,
			Line:     p.BasePos.Line,
			Column:   p.BasePos.Column,
		})
	}
	r.resolveExpressionIfNotNil(s.SingleLineExpr)
	for _, stmt := range s.Body {
		r.resolveStatement(stmt)
	}
	r.Table.ExitScope()
}

func (r *Resolver) resolveGoto(s *ast.GotoStatement) {
	r.labelRefs = append(r.labelRefs, labelRef{
		name:   s.Target,
		line:   s.Pos().Line,
		column: s.Pos().Column,
	})
}

func (r *Resolver) resolveGosub(s *ast.GosubStatement) {
	r.labelRefs = append(r.labelRefs, labelRef{
		name:   s.Target,
		line:   s.Pos().Line,
		column: s.Pos().Column,
	})
}

func (r *Resolver) resolveOnErrorGoto(s *ast.OnErrorGotoStatement) {
	if s.Target != "0" {
		r.labelRefs = append(r.labelRefs, labelRef{
			name:   s.Target,
			line:   s.Pos().Line,
			column: s.Pos().Column,
		})
	}
}

func (r *Resolver) resolveOnEventGosub(s *ast.OnEventGosubStatement) {
	r.resolveExpressionIfNotNil(s.EventParam)
	if s.Target != "" {
		r.labelRefs = append(r.labelRefs, labelRef{
			name:   s.Target,
			line:   s.Pos().Line,
			column: s.Pos().Column,
		})
	}
}

func (r *Resolver) resolveDefType(s *ast.DefTypeStatement) {
	var dt DataType
	switch strings.ToUpper(s.Type) {
	case "DEFINT":
		dt = TypeInteger
	case "DEFLNG":
		dt = TypeLong
	case "DEFSNG":
		dt = TypeSingle
	case "DEFDBL":
		dt = TypeDouble
	case "DEFSTR":
		dt = TypeString
	default:
		return
	}
	for _, lr := range s.LetterRanges {
		r.Table.SetDefType(lr.Start, lr.End, dt)
	}
}

func (r *Resolver) resolveOptionBase(s *ast.OptionBaseStatement) {
	if r.Table.OptionBase != 0 && r.Table.OptionBase != s.Value {
		r.addError(s.Pos().Line, s.Pos().Column,
			"OPTION BASE conflict: previously set to %d, now %d", r.Table.OptionBase, s.Value)
	}
	r.Table.OptionBase = s.Value
}

func (r *Resolver) resolveScopeStmt(s *ast.ScopeStatement) {
	mod := strings.ToUpper(s.Modifier)
	for _, varName := range s.Variables {
		key := strings.ToUpper(varName)
		switch mod {
		case "SHARED":
			// SHARED inside a SUB/FUNCTION makes a global variable visible.
			if gsym := r.Table.GlobalScope.Symbols[key]; gsym != nil {
				gsym.IsShared = true
				// Also add an alias in the current scope.
				r.Table.CurrentScope.Symbols[key] = gsym
			} else {
				// Implicit global declaration.
				dt := r.Table.ResolveType(varName)
				sym := &Symbol{
					Name:     varName,
					Type:     SymVariable,
					DataType: dt,
					Scope:    "global",
					Defined:  true,
					IsShared: true,
					Line:     s.Pos().Line,
					Column:   s.Pos().Column,
				}
				r.Table.GlobalScope.Symbols[key] = sym
				r.Table.CurrentScope.Symbols[key] = sym
			}
		case "STATIC":
			if sym := r.Table.LookupLocal(varName); sym != nil {
				sym.IsStatic = true
			} else {
				dt := r.Table.ResolveType(varName)
				_ = r.Table.Define(varName, &Symbol{
					Name:     varName,
					Type:     SymVariable,
					DataType: dt,
					Scope:    r.Table.CurrentScope.Name,
					Defined:  true,
					IsStatic: true,
					Line:     s.Pos().Line,
					Column:   s.Pos().Column,
				})
			}
		default:
			// LOCAL, COMMON – just ensure a symbol exists.
			if r.Table.LookupLocal(varName) == nil {
				dt := r.Table.ResolveType(varName)
				_ = r.Table.Define(varName, &Symbol{
					Name:     varName,
					Type:     SymVariable,
					DataType: dt,
					Scope:    r.Table.CurrentScope.Name,
					Defined:  true,
					Line:     s.Pos().Line,
					Column:   s.Pos().Column,
				})
			}
		}
	}
}

func (r *Resolver) resolveArrayAssignment(s *ast.ArrayAssignment) {
	r.resolveExpression(s.Value)
	// Resolve index expressions.
	for _, idx := range s.Array.Indices {
		r.resolveExpression(idx)
	}
	// Mark the array as used/defined.
	name := s.Array.Name + s.Array.TypeSuffix
	if sym := r.Table.Lookup(name); sym != nil {
		sym.Defined = true
		sym.Used = true
	}
}

func (r *Resolver) resolveRead(s *ast.ReadStatement) {
	for _, v := range s.Variables {
		r.resolveExpression(v)
	}
}

func (r *Resolver) resolveSwap(s *ast.SwapStatement) {
	r.resolveExpression(s.Var1)
	r.resolveExpression(s.Var2)
}

func (r *Resolver) resolveIncr(s *ast.IncrStatement) {
	r.resolveExpression(s.Variable)
	r.resolveExpressionIfNotNil(s.Amount)
}

func (r *Resolver) resolveDecr(s *ast.DecrStatement) {
	r.resolveExpression(s.Variable)
	r.resolveExpressionIfNotNil(s.Amount)
}

func (r *Resolver) resolveOpen(s *ast.OpenStatement) {
	r.resolveExpressionIfNotNil(s.Filename)
	r.resolveExpressionIfNotNil(s.FileNum)
	r.resolveExpressionIfNotNil(s.RecLen)
}

func (r *Resolver) resolveClose(s *ast.CloseStatement) {
	for _, f := range s.FileNums {
		r.resolveExpression(f)
	}
}

func (r *Resolver) resolveFileInput(s *ast.FileInputStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	for _, v := range s.Variables {
		r.resolveExpression(v)
	}
}

func (r *Resolver) resolveFilePrint(s *ast.FilePrintStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	for _, e := range s.Expressions {
		r.resolveExpression(e)
	}
	r.resolveExpressionIfNotNil(s.Format)
}

func (r *Resolver) resolveFileWrite(s *ast.FileWriteStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	for _, e := range s.Expressions {
		r.resolveExpression(e)
	}
}

func (r *Resolver) resolveGet(s *ast.GetStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	r.resolveExpressionIfNotNil(s.RecordOrPos)
	r.resolveExpressionIfNotNil(s.Variable)
}

func (r *Resolver) resolvePut(s *ast.PutStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	r.resolveExpressionIfNotNil(s.RecordOrPos)
	r.resolveExpressionIfNotNil(s.Variable)
}

func (r *Resolver) resolveSeek(s *ast.SeekStatement) {
	r.resolveExpressionIfNotNil(s.FileNum)
	r.resolveExpressionIfNotNil(s.Position)
}

// ---------------------------------------------------------------------------
// Expression resolver
// ---------------------------------------------------------------------------

func (r *Resolver) resolveExpressionIfNotNil(expr ast.Expression) {
	if expr != nil {
		r.resolveExpression(expr)
	}
}

func (r *Resolver) resolveExpression(expr ast.Expression) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		r.resolveIdentifier(e)
	case *ast.BinaryExpr:
		r.resolveExpression(e.Left)
		r.resolveExpression(e.Right)
	case *ast.UnaryExpr:
		r.resolveExpression(e.Operand)
	case *ast.GroupExpr:
		r.resolveExpression(e.Inner)
	case *ast.FunctionCall:
		r.resolveFunctionCall(e)
	case *ast.ArrayAccess:
		r.resolveArrayAccess(e)
	case *ast.NumberLiteral, *ast.StringLiteral:
		// literals – nothing to resolve
	}
}

func (r *Resolver) resolveIdentifier(id *ast.Identifier) {
	name := id.Name + id.TypeSuffix
	sym := r.Table.Lookup(name)
	if sym != nil {
		sym.Used = true
		return
	}
	// Implicit variable declaration on first reference.
	dt := r.Table.ResolveType(name)
	_ = r.Table.Define(name, &Symbol{
		Name:     name,
		Type:     SymVariable,
		DataType: dt,
		Scope:    r.Table.CurrentScope.Name,
		Defined:  false, // referenced but never explicitly assigned yet
		Used:     true,
		Line:     id.Pos().Line,
		Column:   id.Pos().Column,
	})
}

func (r *Resolver) resolveFunctionCall(fc *ast.FunctionCall) {
	// Resolve arguments.
	for _, arg := range fc.Args {
		r.resolveExpression(arg)
	}
	// Check if the function was declared.
	sym := r.Table.Lookup(fc.Name)
	if sym == nil {
		// Could be a built-in – we allow it, but don't register it.
		return
	}
	sym.Used = true
	// Check argument count for user-defined subs/functions.
	if (sym.Type == SymFunction || sym.Type == SymSub || sym.Type == SymDefFn) && sym.Params != nil {
		if len(fc.Args) != len(sym.Params) {
			r.addError(fc.Pos().Line, fc.Pos().Column,
				"wrong number of arguments for %s %q: expected %d, got %d",
				sym.Type, fc.Name, len(sym.Params), len(fc.Args))
		}
	}
}

func (r *Resolver) resolveArrayAccess(aa *ast.ArrayAccess) {
	for _, idx := range aa.Indices {
		r.resolveExpression(idx)
	}
	name := aa.Name + aa.TypeSuffix
	sym := r.Table.Lookup(name)
	if sym != nil {
		sym.Used = true
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *Resolver) addError(line, col int, format string, args ...interface{}) {
	msg := fmt.Sprintf("%d:%d: %s", line, col, fmt.Sprintf(format, args...))
	r.Errors = append(r.Errors, msg)
}

// resolveElementType determines the DataType from an AS type keyword, a type
// suffix, or the default resolution rules.
func (r *Resolver) resolveElementType(elementType, suffix, baseName string) DataType {
	if elementType != "" {
		switch strings.ToUpper(elementType) {
		case "INTEGER":
			return TypeInteger
		case "LONG":
			return TypeLong
		case "SINGLE":
			return TypeSingle
		case "DOUBLE":
			return TypeDouble
		case "STRING":
			return TypeString
		}
	}
	return r.Table.ResolveType(baseName + suffix)
}

// paramTypes extracts the DataType slice from a list of ast.Parameters.
func paramTypes(params []ast.Parameter, table *SymbolTable) []DataType {
	out := make([]DataType, len(params))
	for i, p := range params {
		out[i] = paramDataType(p.Type, p.Name, table)
	}
	return out
}

// paramDataType resolves a single parameter's DataType from its declared type
// string or the table defaults.
func paramDataType(typeStr, name string, table *SymbolTable) DataType {
	switch strings.ToUpper(typeStr) {
	case "INTEGER", "%":
		return TypeInteger
	case "LONG", "&":
		return TypeLong
	case "SINGLE", "!":
		return TypeSingle
	case "DOUBLE", "#":
		return TypeDouble
	case "STRING", "$":
		return TypeString
	}
	return table.ResolveType(name)
}

// resolveReturnType determines a function's return type from the declared
// return-type string or from the function name's suffix/DEFtype defaults.
func resolveReturnType(retType, name string, table *SymbolTable) DataType {
	switch strings.ToUpper(retType) {
	case "INTEGER", "%":
		return TypeInteger
	case "LONG", "&":
		return TypeLong
	case "SINGLE", "!":
		return TypeSingle
	case "DOUBLE", "#":
		return TypeDouble
	case "STRING", "$":
		return TypeString
	}
	return table.ResolveType(name)
}
