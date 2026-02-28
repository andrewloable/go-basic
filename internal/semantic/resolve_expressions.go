package semantic

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Expression resolver
//
// TUTORIAL — Expression resolution in semantic analysis
//
// Expression resolution serves two purposes:
//
//  1. Mark used symbols — when the resolver encounters an identifier like X,
//     it finds X's Symbol and sets Symbol.Used = true. This allows a later
//     "unused variable" warning pass to identify variables that were declared
//     but never read.
//
//  2. Validate call sites — when the resolver encounters a function call like
//     MySub(A, B), it looks up MySub's Symbol to find its Params list and
//     checks that the argument count matches.
//
// The resolver does NOT evaluate expressions or compute types at this point.
// Type inference (what Go type does this expression produce?) is handled in
// codegen by goTypeForExpr(), which is closer to where the type information
// is actually consumed.
//
// TUTORIAL — Implicit variable declaration on use
//
// BASIC allows using a variable without ever declaring it. The resolver handles
// this in resolveIdentifier(): if the name is not in the symbol table, it is
// defined implicitly as a variable with the type inferred from the name (via
// ResolveType() which looks at the sigil suffix). This matches BASIC's runtime
// semantics where the first reference to an undeclared name creates a variable
// with the default value (0 or "").
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
		// Recursively resolve both sides of the binary operator.
		// We do not perform type checking here — that is deferred to codegen.
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
		// literals – nothing to resolve; a literal like 42 or "hello" is
		// self-contained and does not reference any named symbol.
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
		// TUTORIAL: BASIC has many built-in functions (ABS, SIN, LEFT$, etc.)
		// that are not declared in the user program. The resolver does not
		// have a built-in registry, so it simply allows any unknown function
		// call to pass unchecked. The code generator (Phase 4) has the authoritative
		// list of built-ins and will report an error at generation time if a
		// function is truly unknown.
		return
	}
	sym.Used = true
	// Check argument count for user-defined subs/functions.
	// TUTORIAL: Argument count checking is a basic form of type checking.
	// We can only check count here, not argument types, because BASIC's
	// implicit type coercions make type-checking individual arguments complex.
	// Checking count is still valuable — passing 3 args to a 2-param SUB is
	// almost certainly a bug.
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
	// TUTORIAL — Error accumulation vs. early exit
	//
	// Many early compilers stopped at the first error. This is frustrating for
	// users who have to fix one error, recompile, discover the next, and repeat.
	// Modern compilers accumulate errors and report them all at once.
	//
	// addError appends to r.Errors instead of panicking or returning early.
	// The resolver continues walking the AST even after finding errors. Some
	// errors make subsequent analysis less meaningful (e.g., if a variable is
	// undefined, any operation on it will also seem wrong), so the resolver
	// must be careful not to cascade trivial follow-on errors — but in practice,
	// continuing analysis almost always produces more useful output than stopping.
	msg := fmt.Sprintf("%d:%d: %s", line, col, fmt.Sprintf(format, args...))
	r.Errors = append(r.Errors, msg)
}

// resolveElementType determines the DataType from an AS type keyword, a type
// suffix, or the default resolution rules.
//
// TUTORIAL — Multiple ways to declare a type in BASIC
//
// BASIC offers three ways to specify the type of a variable or array element:
//
//  1. Explicit AS keyword in DIM: DIM A(10) AS INTEGER
//     The elementType parameter captures this ("INTEGER").
//
//  2. Type suffix in the name: DIM A%(10)
//     The suffix parameter captures this ("%").
//
//  3. Default from DEFINT/DEFLNG/DEFxxx or the language default (SINGLE):
//     Handled by ResolveType() using the first letter of the name.
//
// This function tests them in priority order: explicit AS first, then suffix,
// then default. This priority matches the BASIC specification.
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
