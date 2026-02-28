package vm

import (
	"math"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Expression compilation — post-order traversal of the expression tree
//
// Every expression in a stack-based VM is compiled into a sequence of
// "push / operate" instructions that leaves exactly one value on the stack:
//
//   Literal      → OpPush <constant-pool-index>
//   Variable     → OpLoad <variable-index>
//   a + b        → [compile a] [compile b] OpAdd
//   f(x)         → [compile x] OpBuiltin <id>   (or OpCall for user functions)
//   (a + b) * c  → [compile a] [compile b] OpAdd [compile c] OpMul
//
// This is called "post-order traversal" because the operator instruction is
// emitted AFTER both operand sub-trees have been compiled.  The tree for
// (a + b) * c looks like:
//
//                  *
//                 / \
//                +   c
//               / \
//              a   b
//
// Post-order visits: a → b → (+) → c → (*).
// That produces exactly the bytecode sequence above, where each operator
// finds its operands already waiting on the stack.
//
// Because expressions are compiled recursively and each sub-expression leaves
// its result on the stack, the operands for any operator are always sitting
// on top of the stack in the correct order when the operator's opcode is
// reached.  This is why operand stacks and recursive descent compile so
// naturally together — no separate "shunting-yard" or precedence climbing is
// needed; the AST already encodes the correct evaluation order.
// ---------------------------------------------------------------------------

// compileExpression compiles one AST expression node, emitting zero or more
// instructions.  On completion, exactly one Value will have been added to the
// VM's operand stack (the expression's result).
func (c *Compiler) compileExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.NumberLiteral:
		c.compileNumberLiteral(e)
	case *ast.StringLiteral:
		idx := c.addConstant(StringVal(e.Value))
		c.emit(OpPush, idx, e.Pos().Line)
	case *ast.Identifier:
		name := e.Name + e.TypeSuffix
		// Check if this is a compile-time CONST before falling through to a variable.
		if v, ok := c.constMap[strings.ToUpper(name)]; ok {
			idx := c.addConstant(v)
			c.emit(OpPush, idx, e.Pos().Line)
			return
		}
		idx := c.getVarIndex(name)
		c.emit(OpLoad, idx, e.Pos().Line)
	case *ast.BinaryExpr:
		c.compileBinaryExpr(e)
	case *ast.UnaryExpr:
		c.compileUnaryExpr(e)
	case *ast.GroupExpr:
		c.compileExpression(e.Inner)
	case *ast.FunctionCall:
		c.compileFunctionCall(e)
	case *ast.ArrayAccess:
		c.compileArrayAccess(e)
	case *ast.FnCallExpression:
		c.compileFnCallExpr(e)
	case *ast.FieldAccessExpression:
		c.compileFieldAccessExpr(e)
	default:
		c.addError("unsupported expression type %T", expr)
	}
}

func (c *Compiler) compileNumberLiteral(n *ast.NumberLiteral) {
	var v Value
	if n.NumType == ast.NumInt || n.NumType == ast.NumLong {
		v = IntVal(int64(n.Value))
	} else {
		v = FloatVal(n.Value)
	}
	idx := c.addConstant(v)
	c.emit(OpPush, idx, n.Pos().Line)
}

// compileBinaryExpr compiles a binary (two-operand) expression.
//
// Operator-to-opcode mapping:
// After both operands are on the stack the operator string is mapped to a
// single opcode.  Every arithmetic, relational, and logical operator in BASIC
// has a dedicated opcode (OpAdd, OpEq, OpAnd, …) so the VM can implement each
// operation in a tight native Go switch with no string comparisons at runtime.
// All the string-to-opcode translation happens here, once, at compile time —
// this is one of the key performance advantages of compilation over
// interpretation.
//
// Left-before-right emission:
// The left operand is compiled first so it lands below the right operand on
// the stack.  When the VM pops two values, TOS is the right operand and NOS
// is the left, which is the correct order for non-commutative operators:
//   compile a → stack: [a]
//   compile b → stack: [a, b]
//   OpSub     → pops b (TOS), pops a (NOS), pushes a-b
func (c *Compiler) compileBinaryExpr(e *ast.BinaryExpr) {
	line := e.Pos().Line

	c.compileExpression(e.Left)
	c.compileExpression(e.Right)

	switch strings.ToUpper(e.Operator) {
	case "+":
		c.emit(OpAdd, 0, line)
	case "-":
		c.emit(OpSub, 0, line)
	case "*":
		c.emit(OpMul, 0, line)
	case "/":
		c.emit(OpDiv, 0, line)
	case "\\":
		c.emit(OpIDiv, 0, line)
	case "MOD":
		c.emit(OpMod, 0, line)
	case "^":
		c.emit(OpPow, 0, line)
	case "=":
		c.emit(OpEq, 0, line)
	case "<>", "><":
		c.emit(OpNe, 0, line)
	case "<":
		c.emit(OpLt, 0, line)
	case ">":
		c.emit(OpGt, 0, line)
	case "<=", "=<":
		c.emit(OpLe, 0, line)
	case ">=", "=>":
		c.emit(OpGe, 0, line)
	case "AND":
		c.emit(OpAnd, 0, line)
	case "OR":
		c.emit(OpOr, 0, line)
	case "XOR":
		c.emit(OpXor, 0, line)
	case "EQV":
		c.emit(OpEqv, 0, line)
	case "IMP":
		c.emit(OpImp, 0, line)
	default:
		c.addError("unsupported binary operator %q at line %d", e.Operator, line)
	}
}

func (c *Compiler) compileUnaryExpr(e *ast.UnaryExpr) {
	line := e.Pos().Line

	// Optimization: fold constant negation.
	if e.Operator == "-" {
		if num, ok := e.Operand.(*ast.NumberLiteral); ok {
			var v Value
			if num.NumType == ast.NumInt || num.NumType == ast.NumLong {
				v = IntVal(-int64(num.Value))
			} else {
				v = FloatVal(-num.Value)
			}
			idx := c.addConstant(v)
			c.emit(OpPush, idx, line)
			return
		}
	}

	c.compileExpression(e.Operand)

	switch strings.ToUpper(e.Operator) {
	case "-":
		c.emit(OpNeg, 0, line)
	case "NOT":
		c.emit(OpNot, 0, line)
	default:
		c.addError("unsupported unary operator %q at line %d", e.Operator, line)
	}
}

// ---------------------------------------------------------------------------
// Function calls (built-in and user-defined)
// ---------------------------------------------------------------------------

// builtinMap maps upper-cased function names to BuiltinID.
var builtinMap = map[string]BuiltinID{
	"ABS":     BuiltinAbs,
	"SGN":     BuiltinSgn,
	"INT":     BuiltinInt,
	"FIX":     BuiltinFix,
	"CEIL":    BuiltinCeil,
	"SQR":     BuiltinSqr,
	"EXP":     BuiltinExp,
	"EXP2":    BuiltinExp2,
	"EXP10":   BuiltinExp10,
	"LOG":     BuiltinLog,
	"LOG2":    BuiltinLog2,
	"LOG10":   BuiltinLog10,
	"SIN":     BuiltinSin,
	"COS":     BuiltinCos,
	"TAN":     BuiltinTan,
	"ATN":     BuiltinAtn,
	"CINT":    BuiltinCint,
	"CLNG":    BuiltinClng,
	"CSNG":    BuiltinCsng,
	"CDBL":    BuiltinCdbl,
	"RND":     BuiltinRnd,
	"LEFT$":   BuiltinLeft,
	"RIGHT$":  BuiltinRight,
	"MID$":    BuiltinMid,
	"LEN":     BuiltinLen,
	"ASC":     BuiltinAsc,
	"CHR$":    BuiltinChr,
	"STR$":    BuiltinStr,
	"VAL":     BuiltinVal,
	"INSTR":   BuiltinInstr,
	"UCASE$":  BuiltinUCase,
	"LCASE$":  BuiltinLCase,
	"LTRIM$":  BuiltinLTrim,
	"RTRIM$":  BuiltinRTrim,
	"TRIM$":   BuiltinTrim,
	"SPACE$":  BuiltinSpace,
	"STRING$": BuiltinString,
	"HEX$":    BuiltinHex,
	"OCT$":    BuiltinOct,
	"BIN$":    BuiltinBin,
	"MKI$":    BuiltinMki,
	"MKL$":    BuiltinMkl,
	"MKS$":    BuiltinMks,
	"MKD$":    BuiltinMkd,
	"CVI":     BuiltinCvi,
	"CVL":     BuiltinCvl,
	"CVS":     BuiltinCvs,
	"CVD":     BuiltinCvd,
	"TAB":     BuiltinTab,
	"SPC":     BuiltinSpc,
}

// compileFunctionCall compiles a function-call expression.
//
// Three categories of function calls are handled here:
//
// 1. Built-in functions (ABS, SQR, LEFT$, …):
//    Resolved at compile time via builtinMap (a static name→BuiltinID table).
//    Each argument expression is compiled in order so args land left-to-right
//    on the stack.  Then a single OpBuiltin <BuiltinID> instruction is emitted.
//    At runtime the VM pops all arguments, calls native Go code, and pushes
//    the result.  Using a numeric ID instead of a string name means dispatch is
//    an O(1) integer switch — no hash lookups during execution.
//
// 2. DEF FN functions (names starting with "FN"):
//    Resolved via defFnMap (populated when a DEF FN declaration is compiled).
//    Instead of emitting OpCall, the function body is inlined at the call site
//    by inlineDefFn().  This eliminates call-frame overhead entirely.
//
// 3. User-defined SUBs and FUNCTIONs:
//    Compiled to OpCall <addr>, where <addr> is the instruction index of the
//    function body.  If the function has not yet been declared (forward
//    reference), addr is left as 0 and the instruction index is appended to
//    subPatches so it can be backpatched after the full program is compiled.
func (c *Compiler) compileFunctionCall(fc *ast.FunctionCall) {
	line := fc.Pos().Line
	name := strings.ToUpper(fc.Name)

	// Check if this is a built-in function.
	if id, ok := builtinMap[name]; ok {
		// INSTR can take 2 or 3 args: INSTR(s$, find$) or INSTR(start%, s$, find$).
		// The VM always pops 3, so push a default start=1 when only 2 args given.
		if id == BuiltinInstr && len(fc.Args) == 2 {
			c.emit(OpPush, c.addConstant(IntVal(1)), line)
		}
		for _, arg := range fc.Args {
			c.compileExpression(arg)
		}
		c.emit(OpBuiltin, int32(id), line)
		return
	}

	// DEF FN functions (names starting with "FN") are inlined if defined.
	if defn, ok := c.defFnMap[name]; ok {
		c.inlineDefFn(defn, fc.Args, line)
		return
	}

	// User-defined FUNCTION or SUB call.
	for _, arg := range fc.Args {
		c.compileExpression(arg)
	}

	if addr, ok := c.subAddrs[name]; ok {
		c.emit(OpCall, int32(addr), line)
	} else {
		idx := c.emitJump(OpCall, line)
		c.subPatches = append(c.subPatches, subPatch{
			instrIndex: idx,
			name:       fc.Name,
		})
	}
}

func (c *Compiler) compileArrayAccess(aa *ast.ArrayAccess) {
	line := aa.Pos().Line
	name := aa.Name + aa.TypeSuffix
	upper := strings.ToUpper(name)

	// The parser emits *ast.ArrayAccess for any identifier-with-parentheses that
	// it does not recognise as a builtin or FN-prefixed function.  This means
	// user-defined SUBs and FUNCTIONs (e.g. Factorial(n)) are initially parsed
	// as array accesses.  Detect this case and compile as a function call instead.
	_, isKnownFunc := c.subAddrs[upper]
	if c.declaredFuncs[upper] || isKnownFunc {
		// Treat as a user-defined function call: push args then OpCall.
		for _, idx := range aa.Indices {
			c.compileExpression(idx)
		}
		if addr, ok := c.subAddrs[upper]; ok {
			c.emit(OpCall, int32(addr), line)
		} else {
			patchIdx := c.emitJump(OpCall, line)
			c.subPatches = append(c.subPatches, subPatch{
				instrIndex: patchIdx,
				name:       aa.Name,
			})
		}
		return
	}

	nameIdx := c.getVarIndex(name)
	for _, idx := range aa.Indices {
		c.compileExpression(idx)
	}
	c.emit(OpLoadArray, nameIdx, line)
}

// ---------------------------------------------------------------------------
// Utility
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// DEF FN call — inline expansion
// ---------------------------------------------------------------------------

// compileFnCallExpr inlines a DEF FN call (FN name(args) syntax).
func (c *Compiler) compileFnCallExpr(e *ast.FnCallExpression) {
	line := e.Pos().Line
	name := strings.ToUpper(e.Name)

	defn, ok := c.defFnMap[name]
	if !ok {
		// Unknown DEF FN — push zero as placeholder.
		idx := c.addConstant(FloatVal(0))
		c.emit(OpPush, idx, line)
		return
	}
	c.inlineDefFn(defn, e.Args, line)
}

// inlineDefFn performs inline expansion of a DEF FN function call.
//
// "Inline expansion" means that instead of emitting an OpCall instruction and
// setting up a call frame, the compiler emits the function's body bytecode
// directly at the call site — as if the programmer had copy-pasted the body.
// This is the simplest form of inlining and avoids all call overhead.
//
// The challenge with inlining is parameter passing: the function's parameter
// variables (e.g., X in DEF FN SQUARE(X) = X*X) might already have a value
// from the surrounding scope.  We must temporarily bind the argument values
// to the parameter names without permanently overwriting the caller's state.
//
// Inline expansion protocol:
//   Step 1 — compile all argument expressions (push left-to-right onto stack).
//   Step 2 — bind parameters right-to-left (last param is at TOS):
//             a. Save the parameter's current value into a hidden temp variable
//                named __fnsave_FNNAME_PARAMNAME (preserves caller's value).
//             b. Pop the argument from the stack into the parameter variable.
//   Step 3 — compile the function body in-place (single-line expression or
//             multi-line block).  Body uses the parameter variable normally.
//   Step 4 — save the body's result into a hidden result variable.
//   Step 5 — restore each parameter from its temp save slot (in forward order).
//   Step 6 — push the result variable back onto the stack (leave result at TOS).
//
// This keeps DEF FN calls pure (no lasting side effects on caller variables)
// while completely avoiding call frame allocation or jump overhead.
func (c *Compiler) inlineDefFn(defn *ast.DefFnDeclaration, args []ast.Expression, line int) {
	name := strings.ToUpper(defn.Name)

	// Compile all argument expressions (pushed L-to-R onto the stack).
	for _, arg := range args {
		c.compileExpression(arg)
	}

	// Bind parameters R-to-L (last param at top of stack).
	for i := len(defn.Params) - 1; i >= 0; i-- {
		paramName := defn.Params[i].Name
		paramIdx := c.getVarIndex(paramName)
		saveIdx := c.getVarIndex("__fnsave_" + name + "_" + paramName)

		// Save current param value to a temp slot, then store arg into param.
		c.emit(OpLoad, paramIdx, line)
		c.emit(OpStore, saveIdx, line)
		c.emit(OpStore, paramIdx, line) // pops argument from stack
	}

	// Compile the body inline.
	var resultIdx int32
	if defn.SingleLineExpr != nil {
		c.compileExpression(defn.SingleLineExpr)
		resultIdx = c.getVarIndex("__fnresult_" + name)
		c.emit(OpStore, resultIdx, line)
	} else {
		// Multi-line body: FnAssignStatement will store to the function's name var.
		for _, stmt := range defn.Body {
			c.compileStatement(stmt)
		}
		resultIdx = c.getVarIndex(defn.Name)
	}

	// Restore params in forward order.
	for _, param := range defn.Params {
		paramName := param.Name
		paramIdx := c.getVarIndex(paramName)
		saveIdx := c.getVarIndex("__fnsave_" + name + "_" + paramName)
		c.emit(OpLoad, saveIdx, line)
		c.emit(OpStore, paramIdx, line)
	}

	// Leave result on stack.
	c.emit(OpLoad, resultIdx, line)
}

// ---------------------------------------------------------------------------
// TYPE field access — flat variable mangling
// ---------------------------------------------------------------------------

// compileFieldAccessExpr compiles a TYPE record field read using flat variable
// mangling: "var.field" is stored as a separate VM global "var_field".
// This handles the common case where the object is a simple Identifier.
func (c *Compiler) compileFieldAccessExpr(e *ast.FieldAccessExpression) {
	line := e.Pos().Line
	id, ok := e.Object.(*ast.Identifier)
	if !ok {
		// Complex object (array element, etc.) — push zero as placeholder.
		idx := c.addConstant(FloatVal(0))
		c.emit(OpPush, idx, line)
		return
	}
	// Mangle "varname.field" into a single flat variable "varname_field".
	flatName := (id.Name + id.TypeSuffix) + "_" + e.Field
	flatIdx := c.getVarIndex(flatName)
	c.emit(OpLoad, flatIdx, line)
}

// isWholeNumber reports whether f has no fractional part.
func isWholeNumber(f float64) bool {
	return f == math.Trunc(f) && !math.IsInf(f, 0) && !math.IsNaN(f)
}

// evalConstExpr evaluates a constant expression at compile time and returns a
// Value. If the expression is not a simple literal or arithmetic on literals, it
// falls back to FloatVal(0). This is used by CONST statement compilation.
func (c *Compiler) evalConstExpr(expr ast.Expression) Value {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		switch e.NumType {
		case ast.NumInt, ast.NumLong:
			return IntVal(int64(e.Value))
		default:
			return FloatVal(e.Value)
		}
	case *ast.StringLiteral:
		return StringVal(e.Value)
	case *ast.GroupExpr:
		return c.evalConstExpr(e.Inner)
	case *ast.UnaryExpr:
		v := c.evalConstExpr(e.Operand)
		if e.Operator == "-" {
			switch v.Type {
			case ValInt:
				return IntVal(-v.asInt())
			default:
				return FloatVal(-v.asFloat())
			}
		}
		return v
	case *ast.BinaryExpr:
		l := c.evalConstExpr(e.Left)
		r := c.evalConstExpr(e.Right)
		lf, rf := l.asFloat(), r.asFloat()
		switch e.Operator {
		case "+":
			return FloatVal(lf + rf)
		case "-":
			return FloatVal(lf - rf)
		case "*":
			return FloatVal(lf * rf)
		case "/":
			if rf != 0 {
				return FloatVal(lf / rf)
			}
		}
		return FloatVal(lf)
	}
	return FloatVal(0)
}
