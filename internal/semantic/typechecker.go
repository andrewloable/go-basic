package semantic

import (
	"fmt"
	"strings"

	"github.com/loabletech/go-basic/internal/ast"
)

// TypeChecker walks the AST and enforces Turbo BASIC type rules.
type TypeChecker struct {
	Table  *SymbolTable
	Errors []string
}

// NewTypeChecker creates a new TypeChecker using the given symbol table.
func NewTypeChecker(table *SymbolTable) *TypeChecker {
	return &TypeChecker{
		Table: table,
	}
}

// Check runs type checking over the entire program.
func (tc *TypeChecker) Check(prog *ast.Program) []string {
	for _, stmt := range prog.Statements {
		tc.checkStatement(stmt)
	}
	return tc.Errors
}

func (tc *TypeChecker) errorf(pos ast.Position, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	tc.Errors = append(tc.Errors, fmt.Sprintf("%d:%d: %s", pos.Line, pos.Column, msg))
}

// resolveExprType returns the DataType of an expression.
func (tc *TypeChecker) resolveExprType(expr ast.Expression) DataType {
	if expr == nil {
		return TypeUnknown
	}

	switch e := expr.(type) {
	case *ast.NumberLiteral:
		switch e.NumType {
		case ast.NumInt:
			return TypeInteger
		case ast.NumLong:
			return TypeLong
		case ast.NumSingle:
			return TypeSingle
		case ast.NumDouble:
			return TypeDouble
		default:
			return TypeSingle
		}

	case *ast.StringLiteral:
		return TypeString

	case *ast.Identifier:
		sym := tc.Table.Lookup(e.Name)
		if sym != nil {
			return sym.DataType
		}
		return tc.Table.ResolveType(e.Name)

	case *ast.BinaryExpr:
		return tc.resolveBinaryType(e)

	case *ast.UnaryExpr:
		return tc.resolveExprType(e.Operand)

	case *ast.FunctionCall:
		return tc.resolveFunctionReturnType(e)

	case *ast.ArrayAccess:
		sym := tc.Table.Lookup(e.Name)
		if sym != nil {
			return sym.DataType
		}
		return tc.Table.ResolveType(e.Name)

	case *ast.GroupExpr:
		return tc.resolveExprType(e.Inner)
	}

	return TypeUnknown
}

// resolveBinaryType determines the result type of a binary expression.
func (tc *TypeChecker) resolveBinaryType(expr *ast.BinaryExpr) DataType {
	op := strings.ToUpper(expr.Operator)

	// Relational operators always return integer (boolean)
	switch op {
	case "=", "<>", "<", ">", "<=", ">=":
		leftType := tc.resolveExprType(expr.Left)
		rightType := tc.resolveExprType(expr.Right)
		// Both must be same category (numeric or string)
		if isStringType(leftType) != isStringType(rightType) {
			tc.errorf(expr.BasePos, "type mismatch: cannot compare %s with %s",
				DataTypeName(leftType), DataTypeName(rightType))
		}
		return TypeInteger
	}

	// Logical operators require numeric operands, return integer
	switch op {
	case "AND", "OR", "XOR", "EQV", "IMP", "NOT":
		leftType := tc.resolveExprType(expr.Left)
		rightType := tc.resolveExprType(expr.Right)
		if isStringType(leftType) || isStringType(rightType) {
			tc.errorf(expr.BasePos, "type mismatch: logical operator %s requires numeric operands", op)
		}
		return TypeInteger
	}

	// String concatenation
	if op == "+" {
		leftType := tc.resolveExprType(expr.Left)
		rightType := tc.resolveExprType(expr.Right)
		if isStringType(leftType) && isStringType(rightType) {
			return TypeString
		}
		if isStringType(leftType) != isStringType(rightType) {
			tc.errorf(expr.BasePos, "type mismatch: cannot add %s and %s",
				DataTypeName(leftType), DataTypeName(rightType))
			return TypeUnknown
		}
	}

	// Arithmetic operators: widen to the wider type
	leftType := tc.resolveExprType(expr.Left)
	rightType := tc.resolveExprType(expr.Right)

	if isStringType(leftType) || isStringType(rightType) {
		tc.errorf(expr.BasePos, "type mismatch: arithmetic operator %s requires numeric operands", op)
		return TypeUnknown
	}

	// Integer division always returns integer
	if op == "\\" {
		return TypeInteger
	}

	// MOD returns integer
	if op == "MOD" {
		return TypeInteger
	}

	// Regular division promotes to at least single
	if op == "/" {
		wider := widenType(leftType, rightType)
		if wider == TypeInteger || wider == TypeLong {
			return TypeSingle
		}
		return wider
	}

	// Exponentiation: result is at least single
	if op == "^" {
		wider := widenType(leftType, rightType)
		if wider == TypeInteger || wider == TypeLong {
			return TypeSingle
		}
		return wider
	}

	return widenType(leftType, rightType)
}

// resolveFunctionReturnType determines the return type of a function call.
func (tc *TypeChecker) resolveFunctionReturnType(call *ast.FunctionCall) DataType {
	name := strings.ToUpper(call.Name)

	// String-returning built-in functions (ending with $)
	if strings.HasSuffix(name, "$") {
		return TypeString
	}

	// Specific built-in function return types
	switch name {
	// Integer-returning functions
	case "ASC", "LEN", "INSTR", "CSRLIN", "POS", "SCREEN", "PEEK",
		"INP", "EOF", "LOF", "LOC", "FREEFILE", "FRE",
		"INKEY", "LBOUND", "UBOUND", "ERDEV", "ERR", "ERL",
		"CINT", "INT", "FIX":
		return TypeInteger

	// Long-returning functions
	case "CLNG", "SADD", "VARPTR", "SETMEM":
		return TypeLong

	// Single-returning functions
	case "SIN", "COS", "TAN", "ATN", "SQR", "LOG", "EXP",
		"ABS", "SGN", "RND", "TIMER", "CSNG", "VAL",
		"POINT", "PMAP":
		return TypeSingle

	// Double-returning functions
	case "CDBL":
		return TypeDouble
	}

	// User-defined function: check symbol table
	sym := tc.Table.Lookup(name)
	if sym != nil && (sym.Type == SymFunction || sym.Type == SymDefFn) {
		return sym.ReturnType
	}

	return TypeSingle // default
}

// checkStatement dispatches type checking for a statement.
func (tc *TypeChecker) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		tc.checkLetStatement(s)
	case *ast.ArrayAssignment:
		tc.checkArrayAssignment(s)
	case *ast.PrintStatement:
		tc.checkPrintStatement(s)
	case *ast.IfStatement:
		tc.checkIfStatement(s)
	case *ast.ForStatement:
		tc.checkForStatement(s)
	case *ast.WhileStatement:
		tc.checkExpression(s.Condition)
		tc.checkBlock(s.Body)
	case *ast.DoLoopStatement:
		if s.Condition != nil {
			tc.checkExpression(s.Condition)
		}
		tc.checkBlock(s.Body)
	case *ast.SelectCaseStatement:
		tc.checkSelectCase(s)
	case *ast.SubDeclaration:
		tc.checkBlock(s.Body)
	case *ast.FunctionDeclaration:
		tc.checkBlock(s.Body)
	case *ast.DefFnDeclaration:
		if s.SingleLineExpr != nil {
			tc.checkExpression(s.SingleLineExpr)
		}
		tc.checkBlock(s.Body)
	case *ast.SwapStatement:
		tc.checkSwap(s)
	case *ast.IncrStatement:
		tc.checkIncrDecr(s.Variable, s.Amount, s.BasePos)
	case *ast.DecrStatement:
		tc.checkIncrDecr(s.Variable, s.Amount, s.BasePos)
	case *ast.ErrorStatement:
		tc.checkExpression(s.Code)
	case *ast.SoundStatement:
		tc.checkExpression(s.Frequency)
		tc.checkExpression(s.Duration)
	case *ast.DataStatement:
		// DATA values are not type-checked at compile time
	}
}

func (tc *TypeChecker) checkBlock(stmts []ast.Statement) {
	for _, stmt := range stmts {
		tc.checkStatement(stmt)
	}
}

func (tc *TypeChecker) checkExpression(expr ast.Expression) {
	if expr == nil {
		return
	}
	tc.resolveExprType(expr) // triggers any type error reporting
}

func (tc *TypeChecker) checkLetStatement(s *ast.LetStatement) {
	varType := tc.Table.ResolveType(s.Name.Name)
	sym := tc.Table.Lookup(s.Name.Name)
	if sym != nil {
		varType = sym.DataType
	}

	exprType := tc.resolveExprType(s.Value)

	if !isAssignmentCompatible(varType, exprType) {
		tc.errorf(s.BasePos, "type mismatch: cannot assign %s to %s variable '%s'",
			DataTypeName(exprType), DataTypeName(varType), s.Name.Name)
	}
}

func (tc *TypeChecker) checkArrayAssignment(s *ast.ArrayAssignment) {
	varType := tc.Table.ResolveType(s.Array.Name)
	sym := tc.Table.Lookup(s.Array.Name)
	if sym != nil {
		varType = sym.DataType
	}

	exprType := tc.resolveExprType(s.Value)

	if !isAssignmentCompatible(varType, exprType) {
		tc.errorf(s.BasePos, "type mismatch: cannot assign %s to %s array '%s'",
			DataTypeName(exprType), DataTypeName(varType), s.Array.Name)
	}

	// Check all index expressions are numeric
	for _, idx := range s.Array.Indices {
		idxType := tc.resolveExprType(idx)
		if isStringType(idxType) {
			tc.errorf(idx.Pos(), "type mismatch: array index must be numeric, got %s",
				DataTypeName(idxType))
		}
	}
}

func (tc *TypeChecker) checkPrintStatement(s *ast.PrintStatement) {
	for _, expr := range s.Expressions {
		tc.checkExpression(expr)
	}
	if s.Format != nil {
		fmtType := tc.resolveExprType(s.Format)
		if !isStringType(fmtType) {
			tc.errorf(s.Format.Pos(), "PRINT USING format must be a string, got %s",
				DataTypeName(fmtType))
		}
	}
}

func (tc *TypeChecker) checkIfStatement(s *ast.IfStatement) {
	condType := tc.resolveExprType(s.Condition)
	if isStringType(condType) {
		tc.errorf(s.BasePos, "IF condition must be numeric, got STRING")
	}

	tc.checkBlock(s.ThenBlock)
	for _, elif := range s.ElseIfClauses {
		elifCondType := tc.resolveExprType(elif.Condition)
		if isStringType(elifCondType) {
			tc.errorf(elif.BasePos, "ELSEIF condition must be numeric, got STRING")
		}
		tc.checkBlock(elif.Body)
	}
	tc.checkBlock(s.ElseBlock)
}

func (tc *TypeChecker) checkForStatement(s *ast.ForStatement) {
	counterType := tc.Table.ResolveType(s.Counter.Name)
	if isStringType(counterType) {
		tc.errorf(s.BasePos, "FOR counter variable must be numeric")
	}

	startType := tc.resolveExprType(s.Start)
	endType := tc.resolveExprType(s.End)
	if isStringType(startType) || isStringType(endType) {
		tc.errorf(s.BasePos, "FOR start/end values must be numeric")
	}

	if s.Step != nil {
		stepType := tc.resolveExprType(s.Step)
		if isStringType(stepType) {
			tc.errorf(s.BasePos, "FOR STEP value must be numeric")
		}
	}

	tc.checkBlock(s.Body)
}

func (tc *TypeChecker) checkSelectCase(s *ast.SelectCaseStatement) {
	testType := tc.resolveExprType(s.TestExpr)

	for _, c := range s.Cases {
		for _, cv := range c.Values {
			valType := tc.resolveExprType(cv.Value)
			if isStringType(testType) != isStringType(valType) {
				tc.errorf(cv.BasePos, "type mismatch in CASE: test expression is %s but case value is %s",
					DataTypeName(testType), DataTypeName(valType))
			}
			if cv.IsRange && cv.EndValue != nil {
				endType := tc.resolveExprType(cv.EndValue)
				if isStringType(testType) != isStringType(endType) {
					tc.errorf(cv.BasePos, "type mismatch in CASE TO: test expression is %s but range end is %s",
						DataTypeName(testType), DataTypeName(endType))
				}
			}
		}
		tc.checkBlock(c.Body)
	}
	tc.checkBlock(s.ElseBlock)
}

func (tc *TypeChecker) checkSwap(s *ast.SwapStatement) {
	type1 := tc.resolveExprType(s.Var1)
	type2 := tc.resolveExprType(s.Var2)
	if isStringType(type1) != isStringType(type2) {
		tc.errorf(s.BasePos, "type mismatch in SWAP: cannot swap %s with %s",
			DataTypeName(type1), DataTypeName(type2))
	}
}

func (tc *TypeChecker) checkIncrDecr(variable ast.Expression, amount ast.Expression, pos ast.Position) {
	varType := tc.resolveExprType(variable)
	if isStringType(varType) {
		tc.errorf(pos, "INCR/DECR variable must be numeric, got STRING")
	}
	if amount != nil {
		amtType := tc.resolveExprType(amount)
		if isStringType(amtType) {
			tc.errorf(pos, "INCR/DECR amount must be numeric, got STRING")
		}
	}
}

// --- Type helper functions ---

// isStringType returns true if the data type is string.
func isStringType(dt DataType) bool {
	return dt == TypeString
}

// isNumericType returns true if the data type is numeric.
func isNumericType(dt DataType) bool {
	return dt == TypeInteger || dt == TypeLong || dt == TypeSingle || dt == TypeDouble
}

// widenType returns the wider of two numeric types.
// Widening order: Integer < Long < Single < Double
func widenType(a, b DataType) DataType {
	if a == TypeUnknown {
		return b
	}
	if b == TypeUnknown {
		return a
	}
	order := map[DataType]int{
		TypeInteger: 0,
		TypeLong:    1,
		TypeSingle:  2,
		TypeDouble:  3,
	}
	if order[a] >= order[b] {
		return a
	}
	return b
}

// isAssignmentCompatible checks if a value of exprType can be assigned to varType.
// Numeric-to-numeric is always allowed (with implicit conversion).
// String-to-string is allowed.
// String-to-numeric or numeric-to-string is not allowed.
func isAssignmentCompatible(varType, exprType DataType) bool {
	if varType == TypeUnknown || exprType == TypeUnknown {
		return true // can't verify, assume ok
	}
	if isStringType(varType) && isStringType(exprType) {
		return true
	}
	if isNumericType(varType) && isNumericType(exprType) {
		return true
	}
	return false
}

// DataTypeName returns a human-readable name for a DataType.
func DataTypeName(dt DataType) string {
	switch dt {
	case TypeInteger:
		return "INTEGER"
	case TypeLong:
		return "LONG"
	case TypeSingle:
		return "SINGLE"
	case TypeDouble:
		return "DOUBLE"
	case TypeString:
		return "STRING"
	default:
		return "UNKNOWN"
	}
}
