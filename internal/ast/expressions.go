package ast

// ===========================================================================
// Expression nodes
//
// Tutorial note — why separate expression types?
//
// A compiler could represent all expressions as a single tagged union
// (a struct with a "kind" field and various value fields). Go favors
// interfaces instead: each expression kind is its own concrete struct, and
// they all satisfy the Expression interface.
//
// The benefit: later compiler phases (codegen, semantic analysis) use Go's
// type-switch to dispatch on the concrete type:
//
//   switch e := expr.(type) {
//   case *BinaryExpr:   emit binary operation ...
//   case *FunctionCall: emit function call ...
//   case *Identifier:   look up variable ...
//   }
//
// The compiler enforces exhaustiveness (missing cases become obvious bugs),
// and each case gets a correctly-typed value without any casting boilerplate.
// ===========================================================================

// NumberLiteral represents a numeric constant (e.g. 42, 3.14, 1&, 2.5#).
//
// The lexer already classifies the number's type (int/long/single/double) so
// the parser stores that classification in NumType. Downstream phases can then
// generate the correct Go type without reparsing the original text.
//
// OriginalText is preserved because floating-point round-trips are lossy:
// converting "3.14" → float64 → string may not reproduce "3.14" exactly.
// Keeping the raw source text lets the code generator emit the literal
// verbatim rather than relying on fmt.Sprint(value).
type NumberLiteral struct {
	BasePos      Position
	Value        float64 // numeric value, always stored as float64 for uniformity
	OriginalText string  // exact source text (e.g. "3.14", "42") for faithful emission
	NumType      int     // NumInt / NumLong / NumSingle / NumDouble (see ast_base.go)
}

func (n *NumberLiteral) expressionNode()      {}
func (n *NumberLiteral) TokenLiteral() string  { return n.OriginalText }
func (n *NumberLiteral) Pos() Position         { return n.BasePos }

// StringLiteral represents a quoted string constant.
type StringLiteral struct {
	BasePos Position
	Value   string
}

func (n *StringLiteral) expressionNode()      {}
func (n *StringLiteral) TokenLiteral() string  { return n.Value }
func (n *StringLiteral) Pos() Position         { return n.BasePos }

// Identifier represents a variable or named reference (e.g. x, count%, name$).
//
// Turbo BASIC encodes type information directly in the variable name via a
// single-character suffix:
//
//	count%  → integer (16-bit)
//	total&  → long (32-bit)
//	ratio!  → single-precision float
//	pi#     → double-precision float
//	msg$    → string
//
// The parser strips the suffix character from Name and stores it separately in
// TypeSuffix. This separation makes later phases simpler: they can compare
// clean names without stripping suffixes themselves, while still accessing the
// declared type through TypeSuffix.
//
// TypeSuffix is the empty string when no suffix was present, indicating that
// the variable's type comes from a DEFTYPE rule or defaults to float32.
type Identifier struct {
	BasePos    Position
	Name       string // bare variable name without suffix (e.g. "count" not "count%")
	TypeSuffix string // one of "%", "&", "!", "#", "$", or "" (no suffix)
}

func (n *Identifier) expressionNode()      {}
func (n *Identifier) TokenLiteral() string  { return n.Name }
func (n *Identifier) Pos() Position         { return n.BasePos }

// BinaryExpr represents an infix expression (e.g. a + b, x AND y, i >= 10).
//
// A binary expression has exactly two operands (Left and Right) and one
// operator in between. The tree structure encodes precedence: in the expression
// "2 + 3 * 4" the multiplication node is nested inside the addition node
// because * binds tighter than +:
//
//	BinaryExpr(+)
//	  Left:  NumberLiteral(2)
//	  Right: BinaryExpr(*)
//	           Left:  NumberLiteral(3)
//	           Right: NumberLiteral(4)
//
// The Pratt parser (in parse_expressions.go) builds exactly this shape by
// comparing operator precedence levels before deciding which node to nest inside
// which. No further precedence information is needed after the AST is built —
// the tree structure IS the precedence.
type BinaryExpr struct {
	BasePos  Position
	Left     Expression // left-hand operand
	Operator string     // operator symbol: "+", "-", "*", "/", "AND", "OR", etc.
	Right    Expression // right-hand operand
}

func (n *BinaryExpr) expressionNode()      {}
func (n *BinaryExpr) TokenLiteral() string  { return n.Operator }
func (n *BinaryExpr) Pos() Position         { return n.BasePos }

// UnaryExpr represents a prefix expression (e.g. -x, NOT flag).
type UnaryExpr struct {
	BasePos  Position
	Operator string
	Operand  Expression
}

func (n *UnaryExpr) expressionNode()      {}
func (n *UnaryExpr) TokenLiteral() string  { return n.Operator }
func (n *UnaryExpr) Pos() Position         { return n.BasePos }

// FunctionCall represents a function invocation (e.g. ABS(x), MID$(s,1,3)).
//
// In BASIC, built-in functions (ABS, SIN, LEFT$, …) are syntactically
// identical to user-defined functions — both look like Name(args). The parser
// resolves the distinction using the isBuiltinFunction() lookup in
// parse_expressions.go; built-ins become FunctionCall nodes immediately while
// user-defined names initially become ArrayAccess nodes (because the same
// syntax also describes array subscripting). Semantic analysis later
// reclassifies ArrayAccess → FunctionCall when it discovers the name is a
// declared function.
//
// Args is a slice rather than a fixed-size array because functions may take
// varying numbers of arguments (e.g. MID$ takes 2 or 3).
type FunctionCall struct {
	BasePos Position
	Name    string       // upper-case function name including any suffix (e.g. "MID$", "ABS")
	Args    []Expression // positional arguments in source order; may be empty
}

func (n *FunctionCall) expressionNode()      {}
func (n *FunctionCall) TokenLiteral() string  { return n.Name }
func (n *FunctionCall) Pos() Position         { return n.BasePos }

// ArrayAccess represents an array element reference (e.g. a%(1), grid(r,c)).
//
// BASIC uses identical syntax for array subscripting and function calls:
//
//	grid(r, c)   — could be array element OR user-defined function
//	ABS(x)       — built-in function (resolved at parse time)
//
// The parser produces ArrayAccess for the ambiguous case and leaves it to
// semantic analysis to reclassify the node if the name turns out to be a
// function. This lazy disambiguation is simpler than trying to maintain a
// full symbol table inside the parser.
//
// Indices holds one Expression per dimension. A one-dimensional array like
// a(5) produces Indices with one element; a two-dimensional grid(r, c)
// produces Indices with two elements.
type ArrayAccess struct {
	BasePos    Position
	Name       string       // array name without type suffix
	TypeSuffix string       // type suffix stripped from the name (e.g. "%" for a%)
	Indices    []Expression // subscript expressions, one per dimension
}

func (n *ArrayAccess) expressionNode()      {}
func (n *ArrayAccess) TokenLiteral() string  { return n.Name }
func (n *ArrayAccess) Pos() Position         { return n.BasePos }

// GroupExpr represents a parenthesized expression (e.g. (a + b)).
type GroupExpr struct {
	BasePos Position
	Inner   Expression
}

func (n *GroupExpr) expressionNode()      {}
func (n *GroupExpr) TokenLiteral() string  { return "(" }
func (n *GroupExpr) Pos() Position         { return n.BasePos }

// FnCallExpression represents FN name(args) in expression context.
type FnCallExpression struct {
	BasePos Position
	Name    string
	Args    []Expression
}

func (n *FnCallExpression) expressionNode()      {}
func (n *FnCallExpression) TokenLiteral() string  { return "FN" }
func (n *FnCallExpression) Pos() Position         { return n.BasePos }

// FieldAccessExpression represents struct/TYPE member access: expr.field
// In Turbo BASIC, the TYPE keyword defines record types whose fields are
// accessed with dot notation — e.g., Point.X, BCoor(i).XCoor.
type FieldAccessExpression struct {
	BasePos Position
	Object  Expression // the base expression (identifier or array access)
	Field   string     // the field name
}

func (n *FieldAccessExpression) expressionNode()      {}
func (n *FieldAccessExpression) TokenLiteral() string  { return "." }
func (n *FieldAccessExpression) Pos() Position         { return n.BasePos }
