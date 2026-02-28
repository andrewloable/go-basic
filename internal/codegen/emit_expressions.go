package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// ---------------------------------------------------------------------------
// Expression emitter – visitor pattern for expression nodes
//
// emitExpr() is the expression-side visitor.  It mirrors emitStatement() but
// returns a string instead of writing to the buffer directly.  This is the
// key architectural difference between statements and expressions in a
// tree-walking code generator:
//
//	Statement emitters  → write text to g.buf as a side effect (l-value
//	                       semantics: "do this").
//	Expression emitters → return a string that represents the value and can
//	                       be composed into a larger expression (r-value
//	                       semantics: "compute this").
//
// For example, emitPrint() calls g.emitExpr(expr) to get the Go string for
// each sub-expression, then assembles them into a fmt.Println() call.
//
// Recursion is the natural mechanism for nested expressions.  A BinaryExpr
// node calls emitExpr on both its Left and Right children, then wraps the
// results in parentheses with the mapped operator in the middle.  This
// bottom-up composition means deeply nested expressions like:
//
//	A * (B + C) ^ 2
//
// are correctly parenthesised in the output without any explicit precedence
// tracking — parentheses are inserted at every level.
//
// TUTORIAL — Statements vs. expressions in a tree-walking generator
//
// Understanding the statement/expression split is fundamental to code generation.
//
// Statements produce side effects (print to stdout, modify a variable, jump to
// a label). They are executed for what they DO. In the generator, statement
// emitters write directly to g.buf because their purpose is to contribute a
// complete line (or block) of Go code.
//
// Expressions produce values. They are evaluated for what they RETURN. In the
// generator, expression emitters return a Go source string that represents the
// value, allowing the caller to embed it anywhere a Go expression is needed:
//   - As an argument to fmt.Println()
//   - As the right-hand side of an assignment
//   - As a subexpression in a larger arithmetic expression
//
// This two-mode design is universal in tree-walking generators. The boundary
// between the two modes is: "will I write a complete statement, or will I be
// composed into a larger expression?" If the former, write to buf. If the
// latter, return a string.
//
// TUTORIAL — Recursive expression emission and operator precedence
//
// Consider A * (B + C) ^ 2 in BASIC. The parser builds this AST:
//
//   BinaryExpr{*,
//     Left:  Identifier{A},
//     Right: BinaryExpr{^,
//              Left:  GroupExpr{
//                       Inner: BinaryExpr{+, Left: Identifier{B}, Right: Identifier{C}}
//                     },
//              Right: NumberLiteral{2}
//            }
//   }
//
// emitExpr walks this bottom-up:
//   1. Identifier{A}        → "A"
//   2. Identifier{B}        → "B"
//   3. Identifier{C}        → "C"
//   4. BinaryExpr{+, B, C}  → "(B + C)"        (parentheses around every binary)
//   5. GroupExpr{...}       → "((B + C))"      (extra parens from GroupExpr)
//   6. NumberLiteral{2}     → "2"
//   7. BinaryExpr{^, ...}   → "math.Pow(((B + C)), 2)"  (^ maps to math.Pow)
//   8. BinaryExpr{*, A, ...}→ "(A * math.Pow(((B + C)), 2))"
//
// The extra parentheses are harmless and ensure correctness — the Go compiler
// ignores redundant parens. This is safer than trying to track operator
// precedence manually across multiple levels.
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
	case "ERR":
		g.needErrState = true
		return "float64(errState.Err())"
	case "ERL":
		g.needErrState = true
		return "float64(errState.Erl())"
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
		left, right = g.promoteNumericPair(e.Left, e.Right, left, right)
		return fmt.Sprintf("(%s %s %s)", left, e.Operator, right)
	case "/":
		left, right = g.promoteNumericPair(e.Left, e.Right, left, right)
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
		left, right = g.promoteNumericPair(e.Left, e.Right, left, right)
		return fmt.Sprintf("(%s == %s)", left, right)
	case "<>", "><":
		left, right = g.promoteNumericPair(e.Left, e.Right, left, right)
		return fmt.Sprintf("(%s != %s)", left, right)
	case "<":
		left, right = g.promoteNumericPair(e.Left, e.Right, left, right)
		return fmt.Sprintf("(%s < %s)", left, right)
	case ">":
		left, right = g.promoteNumericPair(e.Left, e.Right, left, right)
		return fmt.Sprintf("(%s > %s)", left, right)
	case "<=", "=<":
		left, right = g.promoteNumericPair(e.Left, e.Right, left, right)
		return fmt.Sprintf("(%s <= %s)", left, right)
	case ">=", "=>":
		left, right = g.promoteNumericPair(e.Left, e.Right, left, right)
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

// promoteNumericPair inspects the Go types of left and right AST expressions
// and, when they differ and both are numeric, wraps the narrower one in an
// explicit cast to the wider type.  This prevents Go "mismatched types" errors
// such as "invalid operation: float32 + float64".
//
// String operands (concatenation with +) are left untouched.
// Untyped numeric literals (NumberLiteral without a suffix) are also left
// untouched because Go untyped constants convert automatically at the use site.
func (g *CodeGenerator) promoteNumericPair(
	leftExpr, rightExpr ast.Expression,
	leftStr, rightStr string,
) (string, string) {
	lt := g.goTypeForExpr(leftExpr)
	rt := g.goTypeForExpr(rightExpr)

	// If either side is string, skip — string concatenation needs no cast.
	if lt == "string" || rt == "string" {
		return leftStr, rightStr
	}

	// If types are the same, nothing to do.
	if lt == rt {
		return leftStr, rightStr
	}

	// Untyped NumberLiterals (NumType == NumSingle by default when no suffix)
	// emit as plain untyped Go constants; Go coerces them automatically.
	// We still need to handle the case where one side IS typed (identifier)
	// and the other IS an untyped literal — the literal will coerce fine on
	// its own, but when both are typed and different we must cast.
	_, leftIsLiteral := leftExpr.(*ast.NumberLiteral)
	_, rightIsLiteral := rightExpr.(*ast.NumberLiteral)

	lw := numericWidth(lt)
	rw := numericWidth(rt)

	if lw < 0 || rw < 0 {
		// Non-standard types — leave unchanged.
		return leftStr, rightStr
	}

	wider := lt
	if rw > lw {
		wider = rt
	}

	if lw < rw {
		// left is narrower — cast it unless it's an untyped literal
		if !leftIsLiteral {
			leftStr = fmt.Sprintf("%s(%s)", wider, leftStr)
		}
	} else {
		// right is narrower — cast it unless it's an untyped literal
		if !rightIsLiteral {
			rightStr = fmt.Sprintf("%s(%s)", wider, rightStr)
		}
	}

	return leftStr, rightStr
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
//	func() float64 { v_, _ := rt.Sqr(x); return v_ }()
//
// This is an inline closure that discards the error — a pragmatic choice that
// keeps expression context simple at the cost of ignoring runtime errors.
//
// The "default" case handles user-defined SUB/FUNCTION calls: the name is
// mangled and called directly, since user functions end up as top-level Go
// functions in the same package.
//
// TUTORIAL — Built-in functions and the runtime library
//
// Every language has a standard library of built-in functions. In BASIC, these
// are part of the language itself — SIN, LEFT$, CHR$, etc. are keywords, not
// library calls. In the generated Go, there is no such built-in support, so
// we need a bridge.
//
// The bridge is the internal/runtime package. It is a purpose-built Go library
// that implements all BASIC built-ins as regular Go functions. This approach is
// called a "runtime library" pattern and is used by virtually all compilers and
// transpilers:
//   - C has libc (printf, malloc, strcpy, ...)
//   - Java has java.lang (Math, String, ...)
//   - This transpiler has internal/runtime (rt.Sin, rt.Left, rt.Chr, ...)
//
// The emitFunctionCall switch is the mapping table between the BASIC built-in
// name and its runtime equivalent. For functions where the mapping is trivial
// (same semantics, different name), the case is one line:
//
//   case "SIN": return fmt.Sprintf("rt.Sin(%s)", castF64(0))
//
// For functions where the BASIC and Go semantics differ (e.g., BASIC's SQR
// can be called in an expression context but Go's equivalent returns an error
// value), an inline closure bridges the gap:
//
//   case "SQR":
//     return "func() float64 { v_, _ := rt.Sqr(x); return v_ }()"
//
// This immediately-invoked function literal (IIFE) is a Go idiom for handling
// multi-return functions in expression context. It creates a tiny anonymous
// function, calls it, and returns its result — all in one expression.
//
// TUTORIAL — The castF64 helper and Go's strict typing
//
// BASIC treats all numeric arguments to built-in functions as float64 (or at
// least automatically converts them). Go does not: passing a float32 where
// float64 is expected is a compile error.
//
// The castF64 helper wraps an argument in float64(...) when its inferred type
// is not already float64. This is a common pattern in transpilers from weakly-
// typed to strongly-typed languages: insert explicit coercions wherever the
// source language would have performed an implicit conversion.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitFunctionCall(fc *ast.FunctionCall) string {
	name := strings.ToUpper(fc.Name)
	args := make([]string, 0, len(fc.Args))
	for _, a := range fc.Args {
		args = append(args, g.emitExpr(a))
	}

	// castF64 wraps the emitted string for fc.Args[i] in float64(...) when the
	// inferred Go type of that argument is not already float64.  This ensures
	// that passing a float32 or int16 variable to a runtime function that
	// accepts float64 does not cause a Go compilation error.
	castF64 := func(i int) string {
		s := g.argN(args, i)
		if i < len(fc.Args) && g.goTypeForExpr(fc.Args[i]) == "float64" {
			return s
		}
		return "float64(" + s + ")"
	}

	switch name {
	// Math functions.
	case "ABS":
		return fmt.Sprintf("rt.Abs(%s)", castF64(0))
	case "SGN":
		return fmt.Sprintf("rt.Sgn(%s)", castF64(0))
	case "INT":
		return fmt.Sprintf("rt.IntFloor(%s)", castF64(0))
	case "FIX":
		return fmt.Sprintf("rt.Fix(%s)", castF64(0))
	case "CEIL":
		return fmt.Sprintf("rt.Ceil(%s)", castF64(0))
	case "SQR":
		return fmt.Sprintf("func() float64 { v_, _ := rt.Sqr(%s); return v_ }()", castF64(0))
	case "EXP":
		return fmt.Sprintf("rt.Exp(%s)", castF64(0))
	case "EXP2":
		return fmt.Sprintf("rt.Exp2(%s)", castF64(0))
	case "EXP10":
		return fmt.Sprintf("rt.Exp10(%s)", castF64(0))
	case "LOG":
		return fmt.Sprintf("func() float64 { v_, _ := rt.Log(%s); return v_ }()", castF64(0))
	case "LOG2":
		return fmt.Sprintf("func() float64 { v_, _ := rt.Log2(%s); return v_ }()", castF64(0))
	case "LOG10":
		return fmt.Sprintf("func() float64 { v_, _ := rt.Log10(%s); return v_ }()", castF64(0))
	case "SIN":
		return fmt.Sprintf("rt.Sin(%s)", castF64(0))
	case "COS":
		return fmt.Sprintf("rt.Cos(%s)", castF64(0))
	case "TAN":
		return fmt.Sprintf("rt.Tan(%s)", castF64(0))
	case "ATN":
		return fmt.Sprintf("rt.Atn(%s)", castF64(0))

	// Conversion functions.
	case "CINT":
		return fmt.Sprintf("func() int16 { v_, _ := rt.Cint(%s); return v_ }()", castF64(0))
	case "CLNG":
		return fmt.Sprintf("func() int32 { v_, _ := rt.Clng(%s); return v_ }()", castF64(0))
	case "CSNG":
		return fmt.Sprintf("rt.Csng(%s)", castF64(0))
	case "CDBL":
		return fmt.Sprintf("rt.Cdbl(%s)", castF64(0))

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
		return fmt.Sprintf("rt.Str(%s)", castF64(0))
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
		return fmt.Sprintf("rt.Mkd(%s)", castF64(0))
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
			return fmt.Sprintf("rng.Rnd(%s)", castF64(0))
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

	// Memory / hardware stubs.
	case "FRE":
		if len(args) > 0 {
			return fmt.Sprintf("rt.Fre(%s)", castF64(0))
		}
		return "rt.Fre(0)"
	case "PEEK":
		if len(args) > 0 {
			return fmt.Sprintf("rt.Peek(%s)", castF64(0))
		}
		return "rt.Peek(0)"

	// File EOF function.
	case "EOF":
		// EOF(filenum) returns -1 (true) or 0 (false) in BASIC.
		if len(args) > 0 {
			return fmt.Sprintf("func() float64 { eofResult_, eofErr_ := fm.Eof(int(%s)); if eofErr_ == nil && eofResult_ { return -1 }; return 0 }()", args[0])
		}
		return "float64(0) /* EOF: no file number */"

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
//	DIM A(10)  →  a := make([]float64, 10+1)   // indices 0..10
//
// This wastes one slot at index 0 but keeps the indexing arithmetic trivial:
// A(i) in BASIC simply becomes a[int(i)] in Go — no adjustment needed.
// The cost is one unused slot per array, which is almost always acceptable.
//
// Note that emitArrayAccess does NOT subtract 1 from the index.  The "+1"
// in emitDim is the only adaptation required.  This design choice must be
// understood when reading both functions together.
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitArrayAccess(aa *ast.ArrayAccess) string {
	// Check if this name resolves to a user-defined FUNCTION or SUB in the
	// symbol table.  The parser cannot always distinguish a function call from
	// an array access when there are no parenthesised DECLARE or DIM hints,
	// so it may produce an ArrayAccess node for what is really a call site.
	// Here we do a symbol-table lookup and, when the name is a function/sub,
	// emit a proper Go function call instead of a slice index expression.
	if g.table != nil {
		if sym := g.table.Lookup(aa.Name); sym != nil {
			if sym.Type == semantic.SymFunction || sym.Type == semantic.SymSub || sym.Type == semantic.SymDefFn {
				mangledName := mangleName(aa.Name)
				args := make([]string, 0, len(aa.Indices))
				for _, idx := range aa.Indices {
					args = append(args, g.emitExpr(idx))
				}
				return fmt.Sprintf("%s(%s)", mangledName, strings.Join(args, ", "))
			}
		}
	}

	name := mangleName(aa.Name + aa.TypeSuffix)
	if len(aa.Indices) == 0 {
		// No indices — array passed by reference (e.g., as a SUB parameter)
		return name
	}
	if len(aa.Indices) == 1 {
		idx := g.emitExpr(aa.Indices[0])
		return fmt.Sprintf("%s[int(%s)]", name, idx)
	}
	// Multi-dimensional: emit chained bracket access, e.g. A[int(i)][int(j)]
	var buf strings.Builder
	buf.WriteString(name)
	for _, idx := range aa.Indices {
		buf.WriteString(fmt.Sprintf("[int(%s)]", g.emitExpr(idx)))
	}
	return buf.String()
}

// emitFnCallExpression emits a DEF FN function call.
func (g *CodeGenerator) emitFnCallExpression(e *ast.FnCallExpression) string {
	var argStrs []string
	for _, a := range e.Args {
		argStrs = append(argStrs, g.emitExpr(a))
	}
	return fmt.Sprintf("fn_%s(%s)", mangleName(e.Name), strings.Join(argStrs, ", "))
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
