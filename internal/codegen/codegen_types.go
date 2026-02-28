package codegen

import (
	"fmt"
	"math"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

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
//
// TUTORIAL — Why name mangling is necessary
//
// Name mangling is a technique used by virtually all compilers that must map
// source-language identifiers to target-language identifiers when the two
// identifier namespaces are incompatible.
//
// This function handles two incompatibilities:
//
// 1. BASIC type-sigil characters are not valid in Go identifiers.
//    BASIC encodes type information in the variable name itself using a
//    trailing punctuation character (called a "sigil" or "type suffix"):
//      COUNT%   → integer variable
//      NAME$    → string variable
//      TOTAL&   → long integer variable
//    Go identifiers can only contain letters, digits, and underscores. The
//    mangler replaces each sigil with a readable suffix string:
//      COUNT%   → COUNT_pct   (% → _pct for "percent", the actual character)
//      NAME$    → NAME_str    ($ → _str for "string")
//      TOTAL&   → TOTAL_lng   (& → _lng for "long")
//    The suffix choices are arbitrary but consistent — what matters is that
//    every use of the same BASIC name maps to the same Go name.
//
// 2. Go reserved words must not be used as identifiers.
//    BASIC programs freely use names like RETURN, FOR, TYPE, STRING because
//    they are not reserved in BASIC. These names are reserved in Go. The
//    mangler prepends "b_" (for "BASIC") to avoid the collision:
//      RETURN   → b_return
//      TYPE     → b_type
//      STRING   → b_string
//    The "b_" prefix was chosen to be short, readable, and unlikely to clash
//    with real BASIC variable names (which rarely start with "b_").
//
// TUTORIAL — Consistency requirement
//
// The most important property of a name mangler is consistency: every
// occurrence of the same source name must produce the same mangled name.
// If a variable is declared as "COUNT_pct" but later referenced as "count_pct"
// (or without the "_pct" suffix), the generated Go code will have an "undefined
// variable" error. This is why every code path that touches a variable name —
// declarations in emitLet/emitDim/emitFor, reads in emitIdentifier, and writes
// in emitArrayAssignment — must call mangleName() rather than using the raw
// name directly.
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
//
// TUTORIAL — Mapping source types to target types
//
// Every compiled or transpiled language needs a type mapping: a table that
// says "a value of type X in the source language becomes a value of type Y
// in the target language." This is one of the most consequential design
// decisions in a transpiler because it affects the semantics of every
// arithmetic operation and function call in the generated code.
//
// BASIC's type system (Turbo BASIC dialect):
//
//   BASIC type  Suffix  Range                      Go equivalent
//   ─────────── ──────  ─────────────────────────  ─────────────
//   INTEGER       %     -32768 to 32767            int16
//   LONG          &     -2,147,483,648 to ...      int32
//   SINGLE        !     ~7 significant digits      float32
//   DOUBLE        #     ~15 significant digits     float64
//   STRING        $     variable-length text       string
//   (default)     none  same as SINGLE             float32
//
// The mapping is straightforward here because BASIC's numeric types correspond
// to standard IEEE 754 and two's-complement integer sizes that Go also uses.
// Not all transpilation targets are this lucky — if the target language uses
// different sizes or different floating-point representations, the transpiler
// must insert range checks or normalization calls to preserve semantics.
//
// Note that BASIC's default type (no suffix) maps to float32 (SINGLE), not
// float64. This is Turbo BASIC's historically-determined default: in the
// 1980s, 32-bit floats were the performance-friendly choice. Modern Go code
// would typically use float64 as the default, but we must match the original
// language's semantics for correctness.
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
//
// TUTORIAL — Two-tier type resolution
//
// Type resolution follows a priority chain:
//  1. If a SymbolTable is available (normal compilation), consult it.
//     The table was populated during Phase 3 (semantic analysis) and
//     has the authoritative type for every named symbol, including types
//     declared via DEFINT/DEFDBL letter-range statements.
//  2. If no table is available (e.g., unit-testing codegen in isolation),
//     fall back to goTypeFromSuffix(), which examines the last character of
//     the name to infer the type from the sigil alone.
//
// The fallback exists because codegen and semantic analysis are separate
// packages: codegen can be tested without running the full semantic pass.
// This defensive design is good practice — each phase should be testable
// independently.
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

// zeroValueForType returns the Go zero-value literal for a given Go type.
//
// TUTORIAL — Zero-value initialization
//
// Every typed language has a concept of a "default" or "zero" value for each
// type — the value a variable holds before it is explicitly assigned.
//
// BASIC initialises all numeric variables to 0 and all string variables to ""
// at program start. Go has the same zero-value rules: numeric types default to
// 0, strings to "". This alignment makes the mapping trivial here.
//
// zeroValueForType is used in two contexts:
//  1. When generating a "var x T" declaration and an explicit initialiser is
//     needed (e.g., for return statements in DEF FN functions).
//  2. When emitting CLEAR statement logic to reset all variables.
//
// In a language with a richer type system (structs, slices, pointers), the
// zero-value function would need to handle more cases. For this transpiler,
// the BASIC type set is small enough that a two-branch function suffices.
func zeroValueForType(goT string) string {
	if goT == "string" {
		return `""`
	}
	return "0"
}

// isStringType returns true if the BASIC name denotes a string variable.
func isStringType(name string) bool {
	return len(name) > 0 && name[len(name)-1] == '$'
}

// isStringExpr returns true if the expression resolves to a string Go type.
func (g *CodeGenerator) isStringExpr(expr ast.Expression) bool {
	return g.goTypeForExpr(expr) == "string"
}

// ---------------------------------------------------------------------------
// goTypeForExpr — infer Go type of an arbitrary AST expression
// ---------------------------------------------------------------------------

// numericWidth returns a numeric width rank for Go numeric type names.
// Higher rank = wider type. Returns -1 for non-numeric types.
func numericWidth(goTypeName string) int {
	switch goTypeName {
	case "int16":
		return 0
	case "int32":
		return 1
	case "float32":
		return 2
	case "float64":
		return 3
	default:
		return -1 // non-numeric (e.g. string)
	}
}

// widenType returns the wider of two Go numeric type names.
// If both are the same, returns that type. If either is non-numeric, returns "float64".
func widenType(a, b string) string {
	wa := numericWidth(a)
	wb := numericWidth(b)
	if wa < 0 || wb < 0 {
		return "float64"
	}
	if wa >= wb {
		return a
	}
	return b
}

// goTypeForExpr infers the Go type string for an arbitrary AST expression.
// It is used by emitBinaryExpr to detect type mismatches and emit the
// necessary widening casts so Go accepts the generated code.
func (g *CodeGenerator) goTypeForExpr(expr ast.Expression) string {
	if expr == nil {
		return "float64"
	}
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		// Use the NumType annotation set by the parser/lexer.
		switch e.NumType {
		case ast.NumInt:
			return "int16"
		case ast.NumLong:
			return "int32"
		case ast.NumSingle:
			return "float32"
		case ast.NumDouble:
			return "float64"
		default:
			// Untyped literal: use float64 if fractional, otherwise
			// it's an untyped constant and causes no mismatch on its own.
			// Return "float64" as the safe default so widening works.
			return "float64"
		}
	case *ast.StringLiteral:
		return "string"
	case *ast.Identifier:
		return g.goTypeForIdent(e.Name + e.TypeSuffix)
	case *ast.ArrayAccess:
		return g.goTypeForIdent(e.Name + e.TypeSuffix)
	case *ast.GroupExpr:
		return g.goTypeForExpr(e.Inner)
	case *ast.UnaryExpr:
		return g.goTypeForExpr(e.Operand)
	case *ast.BinaryExpr:
		lt := g.goTypeForExpr(e.Left)
		rt := g.goTypeForExpr(e.Right)
		return widenType(lt, rt)
	case *ast.FunctionCall:
		return goTypeForBuiltin(strings.ToUpper(e.Name))
	case *ast.FnCallExpression:
		// DEF FN functions: infer from suffix of function name.
		return g.goTypeForIdent(e.Name)
	case *ast.FieldAccessExpression:
		// Struct field: conservative default.
		return "float64"
	default:
		return "float64"
	}
}

// goTypeForBuiltin returns the Go return type of a known BASIC built-in function.
func goTypeForBuiltin(name string) string {
	switch name {
	// Integer-returning functions.
	case "LEN", "INSTR", "ASC", "PEEK":
		return "int"
	// int16-returning conversion functions.
	case "CINT", "CVI":
		return "int16"
	// int32-returning conversion functions.
	case "CLNG", "CVL":
		return "int32"
	// float32-returning conversion functions.
	case "CSNG", "CVS":
		return "float32"
	// String-returning functions.
	case "LEFT$", "RIGHT$", "MID$", "CHR$", "STR$", "HEX$", "OCT$", "BIN$",
		"UCASE$", "LCASE$", "LTRIM$", "RTRIM$", "TRIM$", "SPACE$", "STRING$",
		"MKI$", "MKL$", "MKS$", "MKD$", "DATE$", "TIME$", "INKEY$",
		"COMMAND$", "ENVIRON$", "TAB", "SPC":
		return "string"
	// float64-returning functions (the vast majority).
	default:
		return "float64"
	}
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
