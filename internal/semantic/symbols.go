package semantic

import (
	"fmt"
	"strings"
)

// Package semantic implements Phase 3 of the go-basic compiler pipeline:
// semantic analysis. This phase takes the AST produced by the parser and
// validates it for correctness beyond what grammar rules can check.
//
// TUTORIAL — What is semantic analysis?
//
// The parser (Phase 2) checks that the source text follows the grammar — that
// tokens appear in the right order. But grammar alone cannot catch many errors:
//
//   PRINT X * Y       ' grammar says this is fine
//   ' but what if X was never declared?  What type is it?
//
// Semantic analysis catches these program-meaning errors by building a model
// of the program's meaning (semantics) and checking it for consistency.
//
// In this compiler, Phase 3 does two main things:
//
//  1. Symbol table construction — tracks every named entity (variable, array,
//     function, label) in the program, recording its type, scope, and whether
//     it has been defined and/or used.
//
//  2. Reference resolution — ensures that every use of a name refers to a
//     declared entity, and that the number of arguments in function calls
//     matches the function's declared parameter list.
//
// The output of Phase 3 is a populated SymbolTable that Phase 4 (codegen) uses
// to look up type information when emitting Go code.

// ---------------------------------------------------------------------------
// SymbolType – classifies what a symbol represents
//
// TUTORIAL — Why distinguish symbol kinds?
//
// Not all named entities in a program are the same. A variable holds a value.
// An array holds a collection. A function is callable and returns a value. A
// label is a jump target. A constant cannot be reassigned.
//
// Code generators need to know the kind of each symbol to emit correct output.
// For example, in emitArrayAccess (codegen), the generator looks up whether a
// name is a SymArray or a SymFunction — if it is a function, it emits a
// function call instead of a slice index expression. Without this distinction,
// the generated code would be wrong.
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
//
// TUTORIAL — Representing types in a compiler
//
// The type system of the source language must be represented inside the
// compiler so that Phase 3 can record "this variable holds an integer" and
// Phase 4 can emit "var x int16". In a simple language like BASIC, an integer
// enum (iota) is sufficient. More complex languages use richer type
// representations (structs with generic parameters, function type signatures,
// pointer chains, etc.).
//
// BASIC's five fundamental types map directly to Go's numeric and string
// types. See codegen_types.go for the mapping table.
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
//
// TUTORIAL — What information does a symbol carry?
//
// A Symbol is the compiler's "record" for a named entity. Every time the
// compiler encounters a new name, it creates a Symbol and stores it in the
// symbol table. Every subsequent use of that name retrieves the Symbol to
// check that the usage is valid.
//
// The fields encode everything the compiler needs to know:
//
//   Name       — The original BASIC name (e.g., "COUNT%"). Used for error
//                messages so users see their own variable names.
//
//   Type       — The kind of entity (SymVariable, SymArray, SymFunction, …).
//                Prevents calling a variable as if it were a function.
//
//   DataType   — The BASIC type (Integer, Long, Single, Double, String).
//                Used by codegen to emit the correct Go type (int16, float32,…).
//
//   Scope      — The scope where the symbol was declared ("global", or the
//                name of the enclosing SUB/FUNCTION). Used to implement BASIC's
//                scoping rules where local variables shadow globals.
//
//   Defined    — True if the variable has been assigned a value. Allows the
//                compiler to warn about variables used before assignment.
//
//   Used       — True if the variable was ever read. Allows the compiler to
//                warn about declared-but-never-used variables.
//
//   Line/Column — Source position for error messages. Without this, a compiler
//                error like "undefined variable X" would be useless — the user
//                needs to know where in the source file the problem is.
//
//   Params     — For SUBs and FUNCTIONs, the types of formal parameters. Used
//                to check that call sites pass the right number of arguments.
//
//   ReturnType — For FUNCTIONs, the return type. Needed by codegen to emit
//                the correct Go function signature.
//
//   ArrayDims  — For arrays, the number of dimensions. A(10) is 1D; A(3,3)
//                is 2D. Needed by codegen to emit the correct [][]T type.
//
//   IsShared   — True for variables declared SHARED inside a SUB/FUNCTION.
//                SHARED variables live at package scope in the generated Go.
//
//   IsStatic   — True for STATIC variables (retain their value between calls).
//                Not fully supported in the transpiler — documented limitation.
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
//
// TUTORIAL — Lexical scoping and the scope chain
//
// Most programming languages have lexical scoping: the meaning of a name is
// determined by where it appears in the source code, not by the call stack at
// runtime. BASIC is no exception — variables declared inside a SUB are local
// to that SUB.
//
// A Scope is a named container of symbols. Scopes are linked in a parent chain
// (forming a tree rooted at the global scope). When the resolver looks up a
// name, it starts in the current scope and walks up the parent chain until it
// finds a match or reaches the root. This implements the "inner declarations
// shadow outer ones" rule that most languages share.
//
// Example for a BASIC program with a SUB:
//
//   Global scope
//   └── SUB MySub scope
//         contains: parameter variables, local variables
//
// The keys in the Symbols map are stored in upper-case because BASIC is
// case-insensitive: COUNT%, Count%, and count% all refer to the same variable.
// Upper-casing at storage time means lookups can also upper-case the key
// without any special case-folding logic at every use site.
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
//
// TUTORIAL — The symbol table: the compiler's global knowledge base
//
// The symbol table is the central data structure of semantic analysis. It is
// the compiler's "memory" of everything it has learned about the program so
// far. Almost every phase of compilation either reads from or writes to it:
//
//   Phase 3 (semantic analysis): WRITES — populates it by walking the AST
//   Phase 4 (codegen):           READS  — looks up types to emit correct Go
//
// SymbolTable holds three kinds of information:
//
//   Scopes     — The scope tree, giving access to all symbols in all scopes.
//                CurrentScope is a cursor that tracks which scope is "active"
//                during the resolver walk. EnterScope/ExitScope advance it.
//
//   DefTypes   — The DEFINT/DEFLNG/DEFSNG/DEFDBL/DEFSTR declarations. These
//                BASIC statements change the default type for variable names
//                starting with certain letters:
//                  DEFINT A-N     ' variables A through N are integers by default
//                  DEFDBL X-Z     ' X, Y, Z are doubles by default
//                This maps a first-letter byte to a DataType.
//
//   OptionBase — BASIC's OPTION BASE 0 or OPTION BASE 1 setting. Affects
//                the default lower bound of array indices. The symbol table
//                stores it so codegen can adjust array allocation sizes.
//
// TUTORIAL — Why is SymbolTable separate from the Resolver?
//
// The SymbolTable is separated from the Resolver (the walker that builds it)
// because it is consumed by a different phase. The Resolver lives in Phase 3
// and is used once. The SymbolTable is passed to Phase 4 and may be consulted
// many times as the generator walks every variable reference.
//
// Separating data (SymbolTable) from the algorithm that builds it (Resolver)
// is a clean design that makes both easier to test and reason about.
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
