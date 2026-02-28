package ast

// ===========================================================================
// Control flow statement nodes
//
// Tutorial note — modeling branching and looping in an AST
//
// Control flow constructs are the most structurally complex nodes because they
// contain other statements as children (the "body" or "block"). This is what
// makes the data structure a tree rather than a flat list.
//
// A key design question is: should each control-flow variant (IF, FOR, WHILE,
// DO…LOOP, SELECT CASE) be its own node type, or should they share a generic
// "loop" / "branch" node with a tag field? This compiler uses separate types.
//
// Separate types are better because:
//   - Each type carries only the fields it needs (FOR has Counter/Step, WHILE
//     does not). A union type would waste space on inapplicable fields.
//   - Code generation switch-cases are explicit and easy to read.
//   - Adding a new construct requires adding a new type, not enlarging a catch-all
//     struct — change is localised.
// ===========================================================================

// IfStatement represents IF / THEN / ELSEIF / ELSE / END IF.
//
// BASIC supports two syntactic forms:
//
//   1. Single-line: IF cond THEN stmt [ELSE stmt]
//      IsSingleLine is true; ThenBlock and ElseBlock each hold exactly one stmt.
//
//   2. Block:
//        IF cond THEN
//          body...
//        ELSEIF cond2 THEN
//          body2...
//        ELSE
//          body3...
//        END IF
//      IsSingleLine is false; ElseIfClauses holds the ELSEIF chain.
//
// Having both forms in one struct (rather than two separate node types) keeps
// the code generator simpler: it checks IsSingleLine once and then handles
// ThenBlock / ElseIfClauses / ElseBlock uniformly.
type IfStatement struct {
	BasePos       Position
	Condition     Expression   // the tested expression (must evaluate to bool-like value)
	ThenBlock     []Statement  // statements in the THEN branch
	ElseIfClauses []ElseIfClause // zero or more ELSEIF branches, in source order
	ElseBlock     []Statement  // statements in the ELSE branch (empty if no ELSE)
	IsSingleLine  bool         // true for single-line form; affects code-gen layout
}

func (n *IfStatement) statementNode()      {}
func (n *IfStatement) TokenLiteral() string { return "IF" }
func (n *IfStatement) Pos() Position        { return n.BasePos }

// ElseIfClause represents a single ELSEIF branch within an IfStatement.
type ElseIfClause struct {
	BasePos   Position
	Condition Expression
	Body      []Statement
}

// ForStatement represents FOR / TO / STEP / NEXT.
//
// The FOR loop is the most structured loop in BASIC:
//
//   FOR counter = start TO end [STEP step]
//     body
//   NEXT [counter]
//
// Each semantic piece maps to a named field:
//
//   Counter — the loop-control variable (always an Identifier)
//   Start   — the initial value assigned to Counter before the first iteration
//   End     — the upper bound; the loop terminates when Counter exceeds End
//   Step    — the per-iteration increment; defaults to 1 when nil
//   Body    — the statements inside the loop, parsed into a slice
//
// The AST records Step as nil (not as a NumberLiteral(1)) when STEP is absent
// from the source. Code generators check for nil and substitute the default
// at output time. This preserves the distinction between "user wrote STEP 1"
// and "user omitted STEP" — useful if a future optimiser wants to treat the
// two cases differently.
type ForStatement struct {
	BasePos Position
	Counter *Identifier // loop-control variable (e.g. "i" in FOR i = 1 TO 10)
	Start   Expression  // first value assigned to Counter
	End     Expression  // upper bound (inclusive, assuming positive Step)
	Step    Expression  // per-iteration increment; nil means STEP 1 (default)
	Body    []Statement // statements executed each iteration
}

func (n *ForStatement) statementNode()      {}
func (n *ForStatement) TokenLiteral() string { return "FOR" }
func (n *ForStatement) Pos() Position        { return n.BasePos }

// WhileStatement represents WHILE / WEND.
type WhileStatement struct {
	BasePos   Position
	Condition Expression
	Body      []Statement
}

func (n *WhileStatement) statementNode()      {}
func (n *WhileStatement) TokenLiteral() string { return "WHILE" }
func (n *WhileStatement) Pos() Position        { return n.BasePos }

// DoLoopStatement represents DO / LOOP with optional WHILE or UNTIL.
//
// DO…LOOP is more flexible than WHILE because the condition can appear at
// either the top (pre-test) or the bottom (post-test) of the loop, and it
// can be either a "continue while" (WHILE) or "stop when" (UNTIL) condition.
//
// The four combinations are encoded with two boolean flags:
//
//   TestAtTop=true,  IsUntil=false  → DO WHILE cond … LOOP
//   TestAtTop=true,  IsUntil=true   → DO UNTIL cond … LOOP
//   TestAtTop=false, IsUntil=false  → DO … LOOP WHILE cond
//   TestAtTop=false, IsUntil=true   → DO … LOOP UNTIL cond
//   TestAtTop=false, IsUntil=false,
//     Condition=nil                 → DO … LOOP  (infinite loop, EXIT DO breaks)
//
// Code generation maps this to Go's for-loop with an appropriate condition
// placement. Infinite DO…LOOP maps to a bare "for { … }".
type DoLoopStatement struct {
	BasePos   Position
	Condition Expression  // the test expression; nil for an infinite DO…LOOP
	Body      []Statement // statements inside the loop
	TestAtTop bool        // true if condition is at DO (pre-test); false if at LOOP (post-test)
	IsUntil   bool        // true if the loop runs UNTIL (stop when true); false for WHILE (run while true)
}

func (n *DoLoopStatement) statementNode()      {}
func (n *DoLoopStatement) TokenLiteral() string { return "DO" }
func (n *DoLoopStatement) Pos() Position        { return n.BasePos }

// SelectCaseStatement represents SELECT CASE.
//
// SELECT CASE is BASIC's multi-way branch, analogous to Go's switch statement.
// The compiler evaluates TestExpr once and compares the result against each
// CASE clause in order; the first matching clause's body is executed.
//
// Unlike Go's switch, BASIC's SELECT CASE supports range tests and relational
// comparisons inside CASE labels (see CaseValue below), making it more
// expressive than a simple equality switch.
//
// ElseBlock holds CASE ELSE (the default branch). In Go output this becomes
// the "default:" case of a switch statement.
type SelectCaseStatement struct {
	BasePos   Position
	TestExpr  Expression  // the expression whose value is tested against each CASE
	Cases     []CaseClause // the CASE branches, in source order
	ElseBlock []Statement  // body of CASE ELSE (the default); nil if absent
}

func (n *SelectCaseStatement) statementNode()      {}
func (n *SelectCaseStatement) TokenLiteral() string { return "SELECT CASE" }
func (n *SelectCaseStatement) Pos() Position        { return n.BasePos }

// CaseClause represents a single CASE branch.
type CaseClause struct {
	BasePos Position
	Values  []CaseValue
	Body    []Statement
}

// CaseValue represents a single value, range, or comparison in a CASE clause.
//
// Turbo BASIC CASE labels support three forms, modelled by this struct:
//
//   1. Simple equality:  CASE 5
//      IsRange=false, IsComparison=false, Value=NumberLiteral(5)
//
//   2. Inclusive range:  CASE 1 TO 10
//      IsRange=true, Value=lower bound, EndValue=upper bound
//
//   3. Relational test:  CASE IS > 100
//      IsComparison=true, Comparison=">", Value=the right-hand operand
//
// A single CASE line can list multiple CaseValues separated by commas:
//
//   CASE 1, 3, 5 TO 9, IS > 20
//
// The parser builds a slice of CaseValue (one per comma-separated item) and
// stores them in CaseClause.Values. Code generation evaluates them left-to-right
// and short-circuits on the first match.
type CaseValue struct {
	BasePos      Position
	Value        Expression // the literal, lower bound (range), or RHS (comparison)
	EndValue     Expression // upper bound for IS range form; nil otherwise
	Comparison   string     // relational operator for IS form (e.g. ">", "<="); empty otherwise
	IsRange      bool       // true for "low TO high" form
	IsComparison bool       // true for "IS op value" form
}

// GotoStatement represents GOTO.
type GotoStatement struct {
	BasePos Position
	Target  string
}

func (n *GotoStatement) statementNode()      {}
func (n *GotoStatement) TokenLiteral() string { return "GOTO" }
func (n *GotoStatement) Pos() Position        { return n.BasePos }

// GosubStatement represents GOSUB.
type GosubStatement struct {
	BasePos Position
	Target  string
}

func (n *GosubStatement) statementNode()      {}
func (n *GosubStatement) TokenLiteral() string { return "GOSUB" }
func (n *GosubStatement) Pos() Position        { return n.BasePos }

// ReturnStatement represents RETURN.
type ReturnStatement struct {
	BasePos Position
}

func (n *ReturnStatement) statementNode()      {}
func (n *ReturnStatement) TokenLiteral() string { return "RETURN" }
func (n *ReturnStatement) Pos() Position        { return n.BasePos }

// ExitStatement represents EXIT FOR, EXIT DO, EXIT WHILE, etc.
type ExitStatement struct {
	BasePos  Position
	ExitType string // "FOR", "DO", "WHILE", "SUB", "FUNCTION", "DEF"
}

func (n *ExitStatement) statementNode()      {}
func (n *ExitStatement) TokenLiteral() string { return "EXIT" }
func (n *ExitStatement) Pos() Position        { return n.BasePos }

// EndStatement represents END (terminate program).
type EndStatement struct {
	BasePos Position
}

func (n *EndStatement) statementNode()      {}
func (n *EndStatement) TokenLiteral() string { return "END" }
func (n *EndStatement) Pos() Position        { return n.BasePos }

// StopStatement represents STOP.
type StopStatement struct {
	BasePos Position
}

func (n *StopStatement) statementNode()      {}
func (n *StopStatement) TokenLiteral() string { return "STOP" }
func (n *StopStatement) Pos() Position        { return n.BasePos }

// SystemStatement represents SYSTEM.
type SystemStatement struct {
	BasePos Position
}

func (n *SystemStatement) statementNode()      {}
func (n *SystemStatement) TokenLiteral() string { return "SYSTEM" }
func (n *SystemStatement) Pos() Position        { return n.BasePos }
