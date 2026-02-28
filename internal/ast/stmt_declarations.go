package ast

// ===========================================================================
// Declaration and assignment statement nodes
//
// Tutorial note — declarations vs. assignments
//
// In BASIC, the boundary between "declaring" and "using" a variable is blurry.
// Variables spring into existence the first time they are assigned, so there is
// no mandatory declaration. However, DIM (dimension/declare), CONST, SUB, and
// FUNCTION provide explicit declarations that the compiler must track in its
// symbol table.
//
// This file contains both true declarations (SUB, FUNCTION, DIM, CONST, TYPE)
// and assignment-like statements (LET, array assignment, INCR/DECR). They are
// grouped together because they all affect the symbol table during semantic
// analysis.
// ===========================================================================

// LetStatement represents a LET (or implicit) assignment.
type LetStatement struct {
	BasePos Position
	Name    *Identifier
	Value   Expression
}

func (n *LetStatement) statementNode()      {}
func (n *LetStatement) TokenLiteral() string { return "LET" }
func (n *LetStatement) Pos() Position        { return n.BasePos }

// ArrayAssignment represents assignment to an array element.
type ArrayAssignment struct {
	BasePos Position
	Array   *ArrayAccess
	Value   Expression
}

func (n *ArrayAssignment) statementNode()      {}
func (n *ArrayAssignment) TokenLiteral() string { return "LET" }
func (n *ArrayAssignment) Pos() Position        { return n.BasePos }

// DimStatement represents DIM.
type DimStatement struct {
	BasePos      Position
	Declarations []DimDecl
}

func (n *DimStatement) statementNode()      {}
func (n *DimStatement) TokenLiteral() string { return "DIM" }
func (n *DimStatement) Pos() Position        { return n.BasePos }

// DimDecl describes a single variable/array declaration inside DIM.
type DimDecl struct {
	BasePos      Position
	Name         string
	TypeSuffix   string
	Dimensions   []DimRange
	ElementType  string
	StringLength Expression
}

// DimRange describes the lower..upper bounds of one array dimension.
type DimRange struct {
	BasePos Position
	Lower   Expression
	Upper   Expression
}

// RedimStatement represents REDIM.
type RedimStatement struct {
	BasePos      Position
	Declarations []DimDecl
}

func (n *RedimStatement) statementNode()      {}
func (n *RedimStatement) TokenLiteral() string { return "REDIM" }
func (n *RedimStatement) Pos() Position        { return n.BasePos }

// EraseStatement represents ERASE.
type EraseStatement struct {
	BasePos Position
	Names   []string
}

func (n *EraseStatement) statementNode()      {}
func (n *EraseStatement) TokenLiteral() string { return "ERASE" }
func (n *EraseStatement) Pos() Position        { return n.BasePos }

// OptionBaseStatement represents OPTION BASE 0 or OPTION BASE 1.
type OptionBaseStatement struct {
	BasePos Position
	Value   int // 0 or 1
}

func (n *OptionBaseStatement) statementNode()      {}
func (n *OptionBaseStatement) TokenLiteral() string { return "OPTION BASE" }
func (n *OptionBaseStatement) Pos() Position        { return n.BasePos }

// SubDeclaration represents SUB / END SUB.
//
// A SUB is a named subroutine that performs an action but does NOT return a
// value. This is the BASIC equivalent of a Go function that returns nothing.
// In generated Go, a SUB becomes a plain "func Name(params)" with no return
// type.
//
// Params describes the formal parameter list. BASIC passes arguments by
// reference by default (the callee can modify the caller's variable); BYVAL
// opts out of this, making the parameter a copy. Each Parameter's IsByVal
// flag records this choice so the code generator knows whether to use pointer
// semantics or value semantics.
//
// IsForward is set when this node was produced by DECLARE SUB (a forward
// declaration). Forward declarations exist only to tell the compiler about the
// parameter types before the full definition appears; they have no body and
// generate no output code.
type SubDeclaration struct {
	BasePos   Position
	Name      string       // SUB name (raw, not mangled)
	Params    []Parameter  // formal parameters in source order
	Body      []Statement  // the statements inside the SUB body; nil for DECLARE
	IsForward bool         // true when produced by DECLARE SUB (no body)
}

func (n *SubDeclaration) statementNode()      {}
func (n *SubDeclaration) TokenLiteral() string { return "SUB" }
func (n *SubDeclaration) Pos() Position        { return n.BasePos }

// FunctionDeclaration represents FUNCTION / END FUNCTION.
//
// The critical difference from SubDeclaration: a FUNCTION returns a value.
// In BASIC, the return value is set by assigning to a variable with the same
// name as the function inside the body:
//
//   FUNCTION Square(x AS INTEGER) AS INTEGER
//     Square = x * x        ← assignment to function name = return value
//   END FUNCTION
//
// ReturnType is the explicit "AS type" annotation. When omitted, the type
// is inferred from the function name's suffix (e.g. Square% → INTEGER).
// An empty ReturnType string means no explicit type was declared.
//
// In generated Go, this becomes a function with a named return variable
// matching the function name, so that "FunctionName = expr" assignments
// inside the body translate naturally to "functionName = expr".
type FunctionDeclaration struct {
	BasePos    Position
	Name       string      // FUNCTION name (raw, not mangled)
	Params     []Parameter // formal parameters in source order
	ReturnType string      // explicit return type from "AS type" clause; "" if omitted
	Body       []Statement // the statements inside the FUNCTION body; nil for DECLARE
	IsForward  bool        // true when produced by DECLARE FUNCTION (no body)
}

func (n *FunctionDeclaration) statementNode()      {}
func (n *FunctionDeclaration) TokenLiteral() string { return "FUNCTION" }
func (n *FunctionDeclaration) Pos() Position        { return n.BasePos }

// DefFnDeclaration represents DEF FN (single-line or multi-line).
//
// DEF FN is the oldest BASIC function mechanism, predating SUB and FUNCTION.
// It comes in two forms:
//
//   Single-line (inline function):
//     DEF FNSquare(x) = x * x
//     SingleLineExpr holds the expression; Body is nil.
//     In Go, this becomes a one-line lambda: var FNSquare = func(x float32) float32 { return x * x }
//
//   Multi-line (block function, Turbo BASIC extension):
//     DEF FNCompute(a, b)
//       ...statements...
//     END DEF
//     Body holds the statements; SingleLineExpr is nil.
//     In Go, this becomes a full function declaration.
//
// By convention, DEF FN names always begin with "FN" (e.g. FNSquare, FNMax).
// The parser stores the full name including the "FN" prefix so that call sites
// (FNSquare(5)) can be matched without any transformation.
type DefFnDeclaration struct {
	BasePos        Position
	Name           string      // full function name including "FN" prefix (e.g. "FNSquare")
	Params         []Parameter // formal parameters; may be empty
	Body           []Statement // multi-line body; nil for single-line form
	SingleLineExpr Expression  // body expression for single-line form; nil for multi-line
}

func (n *DefFnDeclaration) statementNode()      {}
func (n *DefFnDeclaration) TokenLiteral() string { return "DEF FN" }
func (n *DefFnDeclaration) Pos() Position        { return n.BasePos }

// Parameter describes a formal parameter of a SUB, FUNCTION, or DEF FN.
//
// BASIC passes arguments by reference by default: the callee receives a
// pointer to the caller's variable and can modify it. BYVAL changes this
// to pass-by-value (a copy), matching Go's default calling convention.
//
// When the code generator emits a SUB or FUNCTION, it uses IsByVal to decide
// whether to declare the parameter as a plain Go value (BYVAL) or as a
// pointer (*Type) that is automatically dereferenced inside the body.
//
// Type is the explicit AS-type annotation (e.g. "INTEGER", "STRING"). When
// empty, the type is inferred from the Name suffix character or from a
// DEFTYPE rule.
type Parameter struct {
	BasePos Position
	Name    string // parameter name (may include type suffix, e.g. "count%")
	Type    string // explicit AS-type, e.g. "INTEGER", "STRING"; "" when absent
	IsByVal bool   // true if BYVAL was specified (pass by value, not by reference)
}

// DefTypeStatement represents DEFINT, DEFLNG, DEFSNG, DEFDBL, DEFSTR.
type DefTypeStatement struct {
	BasePos      Position
	LetterRanges []LetterRange
	Type         string // "DEFINT", "DEFLNG", "DEFSNG", "DEFDBL", "DEFSTR"
}

func (n *DefTypeStatement) statementNode()      {}
func (n *DefTypeStatement) TokenLiteral() string { return n.Type }
func (n *DefTypeStatement) Pos() Position        { return n.BasePos }

// LetterRange represents a range of letters (e.g. A-Z) in a DEFTYPE statement.
type LetterRange struct {
	Start byte
	End   byte
}

// ScopeStatement represents SHARED, LOCAL, STATIC, or COMMON.
type ScopeStatement struct {
	BasePos   Position
	Modifier  string // "SHARED", "LOCAL", "STATIC", "COMMON"
	Variables []string
}

func (n *ScopeStatement) statementNode()      {}
func (n *ScopeStatement) TokenLiteral() string { return n.Modifier }
func (n *ScopeStatement) Pos() Position        { return n.BasePos }

// ConstStatement represents CONST name = expr (compile-time constant).
type ConstStatement struct {
	BasePos Position
	Name    string
	Value   Expression
}

func (n *ConstStatement) statementNode()      {}
func (n *ConstStatement) TokenLiteral() string { return "CONST" }
func (n *ConstStatement) Pos() Position        { return n.BasePos }

// ClearStatement represents the CLEAR statement (resets all variables).
type ClearStatement struct {
	BasePos Position
}

func (n *ClearStatement) statementNode()      {}
func (n *ClearStatement) TokenLiteral() string { return "CLEAR" }
func (n *ClearStatement) Pos() Position        { return n.BasePos }

// TypeBlockStatement represents a TYPE name ... END TYPE user-defined type.
type TypeBlockStatement struct {
	BasePos Position
	Name    string
	Fields  []TypeField
}

// TypeField represents a single field in a TYPE block.
type TypeField struct {
	Name     string
	TypeName string
}

func (n *TypeBlockStatement) statementNode()      {}
func (n *TypeBlockStatement) TokenLiteral() string { return "TYPE" }
func (n *TypeBlockStatement) Pos() Position        { return n.BasePos }

// FnAssignStatement represents FN name = expr inside a DEF FN block (return value assignment).
type FnAssignStatement struct {
	BasePos Position
	Name    string
	Value   Expression
}

func (n *FnAssignStatement) statementNode()      {}
func (n *FnAssignStatement) TokenLiteral() string { return "FN" }
func (n *FnAssignStatement) Pos() Position        { return n.BasePos }

// FieldAssignStatement represents struct/TYPE member assignment: expr.field = value
type FieldAssignStatement struct {
	BasePos Position
	Object  Expression // the base expression
	Field   string     // the field name
	Value   Expression // the value to assign
}

func (n *FieldAssignStatement) statementNode()      {}
func (n *FieldAssignStatement) TokenLiteral() string { return "." }
func (n *FieldAssignStatement) Pos() Position        { return n.BasePos }
